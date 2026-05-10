package ratelimit

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type mockDynamoClient struct {
	getItemFn func(ctx context.Context, params *dynamodb.GetItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	putItemFn func(ctx context.Context, params *dynamodb.PutItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
}

func (m *mockDynamoClient) GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	return m.getItemFn(ctx, params, optFns...)
}

func (m *mockDynamoClient) PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	return m.putItemFn(ctx, params, optFns...)
}

func TestAllow_NewUser_Allows(t *testing.T) {
	m := &mockDynamoClient{
		getItemFn: func(_ context.Context, _ *dynamodb.GetItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
			return &dynamodb.GetItemOutput{Item: nil}, nil
		},
		putItemFn: func(_ context.Context, p *dynamodb.PutItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
			if cond := p.ConditionExpression; cond == nil || *cond != "attribute_not_exists(PK)" {
				return nil, fmt.Errorf("expected condition attribute_not_exists(PK), got %v", cond)
			}
			tokens := p.Item["tokens"].(*types.AttributeValueMemberN).Value
			if tokens != "4" {
				return nil, fmt.Errorf("expected tokens=4 (max 5 - 1), got %s", tokens)
			}
			return &dynamodb.PutItemOutput{}, nil
		},
	}

	lim := NewLimiter(m, "test", 5, 60, time.UTC)

	ok, err := lim.Allow(context.Background(), "u1", "meal")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected Allow to return true for new user")
	}
}

func TestAllow_ZeroTokens_Rejects(t *testing.T) {
	now := time.Now().UTC().Format(time.RFC3339Nano)

	callCount := 0
	m := &mockDynamoClient{
		getItemFn: func(_ context.Context, _ *dynamodb.GetItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
			callCount++
			return &dynamodb.GetItemOutput{
				Item: map[string]types.AttributeValue{
					"tokens":      &types.AttributeValueMemberN{Value: "0"},
					"lastUpdated": &types.AttributeValueMemberS{Value: now},
				},
			}, nil
		},
		putItemFn: func(_ context.Context, _ *dynamodb.PutItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
			t.Error("PutItem should not be called when tokens < 1")
			return nil, nil
		},
	}

	lim := NewLimiter(m, "test", 5, 60, time.UTC)

	ok, err := lim.Allow(context.Background(), "u1", "meal")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatal("expected Allow to return false when tokens=0 and recently updated")
	}
	if callCount != 1 {
		t.Fatalf("expected 1 GetItem call, got %d", callCount)
	}
}

func TestAllow_GetItemError_ReturnsError(t *testing.T) {
	expectedErr := errors.New("dynamo unavailable")

	m := &mockDynamoClient{
		getItemFn: func(_ context.Context, _ *dynamodb.GetItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
			return nil, expectedErr
		},
		putItemFn: nil,
	}

	lim := NewLimiter(m, "test", 5, 60, time.UTC)

	ok, err := lim.Allow(context.Background(), "u1", "meal")
	if err == nil {
		t.Fatal("expected error from GetItem failure")
	}
	if ok {
		t.Fatal("expected Allow to return false on error")
	}
}

func TestAllow_Refill_Allows(t *testing.T) {
	past := time.Now().UTC().Add(-120 * time.Second).Format(time.RFC3339Nano)

	callCount := 0
	m := &mockDynamoClient{
		getItemFn: func(_ context.Context, _ *dynamodb.GetItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
			callCount++
			return &dynamodb.GetItemOutput{
				Item: map[string]types.AttributeValue{
					"tokens":      &types.AttributeValueMemberN{Value: "0"},
					"lastUpdated": &types.AttributeValueMemberS{Value: past},
				},
			}, nil
		},
		putItemFn: func(_ context.Context, p *dynamodb.PutItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
			tokens := p.Item["tokens"].(*types.AttributeValueMemberN).Value
			v, _ := strconv.Atoi(tokens)
			if v != 1 {
				t.Errorf("after 120s elapsed at 60s refill: 2 tokens added, 1 consumed → expected 1, got %d", v)
			}
			return &dynamodb.PutItemOutput{}, nil
		},
	}

	lim := NewLimiter(m, "test", 5, 60, time.UTC)

	ok, err := lim.Allow(context.Background(), "u1", "meal")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected Allow to return true after refill")
	}
	if callCount != 1 {
		t.Fatalf("expected 1 GetItem call, got %d", callCount)
	}
}

func TestAllow_OptimisticLock_Retries(t *testing.T) {
	past := time.Now().UTC().Add(-10 * time.Second).Format(time.RFC3339Nano)

	getCount := 0
	m := &mockDynamoClient{
		getItemFn: func(_ context.Context, _ *dynamodb.GetItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
			getCount++
			return &dynamodb.GetItemOutput{
				Item: map[string]types.AttributeValue{
					"tokens":      &types.AttributeValueMemberN{Value: "1"},
					"lastUpdated": &types.AttributeValueMemberS{Value: past},
				},
			}, nil
		},
		putItemFn: func(_ context.Context, _ *dynamodb.PutItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
			if getCount == 1 {
				return nil, &types.ConditionalCheckFailedException{}
			}
			return &dynamodb.PutItemOutput{}, nil
		},
	}

	lim := NewLimiter(m, "test", 5, 60, time.UTC)

	ok, err := lim.Allow(context.Background(), "u1", "meal")
	if err != nil {
		t.Fatalf("unexpected error after retry: %v", err)
	}
	if !ok {
		t.Fatal("expected Allow to succeed after retry")
	}
	if getCount != 2 {
		t.Fatalf("expected 2 GetItem calls (1 initial + 1 retry), got %d", getCount)
	}
}

func TestAllow_FractionalRefillCarryover(t *testing.T) {
	// Scenario: user exhausted tokens 90s ago with refillSecs=60.
	// refill = 90/60 = 1. After consume, lastUpdated should advance by 60s (not 90s),
	// preserving the 30s remainder for the next calculation.
	refillSecs := 60
	baseTime := time.Now().UTC().Add(-90 * time.Second)
	past := baseTime.Format(time.RFC3339Nano)

	var capturedLastUpdated string
	m := &mockDynamoClient{
		getItemFn: func(_ context.Context, _ *dynamodb.GetItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
			return &dynamodb.GetItemOutput{
				Item: map[string]types.AttributeValue{
					"tokens":      &types.AttributeValueMemberN{Value: "0"},
					"lastUpdated": &types.AttributeValueMemberS{Value: past},
				},
			}, nil
		},
		putItemFn: func(_ context.Context, p *dynamodb.PutItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
			capturedLastUpdated = p.Item["lastUpdated"].(*types.AttributeValueMemberS).Value
			return &dynamodb.PutItemOutput{}, nil
		},
	}

	lim := NewLimiter(m, "test", 5, refillSecs, time.UTC)

	ok, err := lim.Allow(context.Background(), "u1", "meal")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected Allow=true (1 refill from 90s elapsed)")
	}

	// The fix: lastUpdated = baseTime + 60s, NOT time.Now()
	expectedLastUpdated := baseTime.Add(time.Duration(1*refillSecs) * time.Second).Format(time.RFC3339Nano)
	if capturedLastUpdated != expectedLastUpdated {
		t.Fatalf("lastUpdated should be baseTime+60s to carry over remainder.\nwant: %s\n got: %s", expectedLastUpdated, capturedLastUpdated)
	}

	// Validate the carryover effect: the 30s remainder means that at ~30s from now,
	// the user earns another refill. With the old bug (lastUpdated=now), they'd need
	// to wait the full 60s again.
	writtenLastUpdated, _ := time.Parse(time.RFC3339Nano, capturedLastUpdated)
	now := time.Now().UTC()
	remainingToNextRefill := now.Sub(writtenLastUpdated)
	if remainingToNextRefill >= time.Duration(refillSecs)*time.Second {
		t.Fatalf("carryover broken: expected <60s to next refill, got %v", remainingToNextRefill)
	}
}
