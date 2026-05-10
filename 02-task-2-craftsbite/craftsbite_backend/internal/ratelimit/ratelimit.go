package ratelimit

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type dynamoClient interface {
	GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
}

type Limiter struct {
	client     dynamoClient
	table      string
	maxTokens  int
	refillSecs int
	timezone   *time.Location
}

func NewLimiter(client dynamoClient, table string, maxTokens, refillSecs int, tz *time.Location) *Limiter {
	return &Limiter{
		client:     client,
		table:      table,
		maxTokens:  maxTokens,
		refillSecs: refillSecs,
		timezone:   tz,
	}
}

func (l *Limiter) Allow(ctx context.Context, userID, commandName string) (bool, error) {
	pk := "RATELIMIT"
	sk := "USER#" + userID + "#" + commandName
	return l.allow(ctx, pk, sk)
}

func (l *Limiter) allow(ctx context.Context, pk, sk string) (bool, error) {
	out, err := l.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(l.table),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: pk},
			"SK": &types.AttributeValueMemberS{Value: sk},
		},
	})
	if err != nil {
		return false, fmt.Errorf("ratelimit get: %w", err)
	}

	if out.Item == nil {
		return l.tryFirstWrite(ctx, pk, sk)
	}

	return l.tryConsume(ctx, pk, sk, out.Item)
}

func (l *Limiter) tryFirstWrite(ctx context.Context, pk, sk string) (bool, error) {
	now := time.Now().In(l.timezone)
	_, err := l.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(l.table),
		Item: map[string]types.AttributeValue{
			"PK":          &types.AttributeValueMemberS{Value: pk},
			"SK":          &types.AttributeValueMemberS{Value: sk},
			"tokens":      &types.AttributeValueMemberN{Value: strconv.Itoa(l.maxTokens - 1)},
			"lastUpdated": &types.AttributeValueMemberS{Value: now.Format(time.RFC3339Nano)},
		},
		ConditionExpression: aws.String("attribute_not_exists(PK)"),
	})
	if err != nil {
		if isConditionalCheckFailed(err) {
			return l.allow(ctx, pk, sk)
		}
		return false, fmt.Errorf("ratelimit first put: %w", err)
	}
	return true, nil
}

func (l *Limiter) tryConsume(ctx context.Context, pk, sk string, item map[string]types.AttributeValue) (bool, error) {
	readTokens, _ := strconv.Atoi(item["tokens"].(*types.AttributeValueMemberN).Value)
	readLastUpdated := item["lastUpdated"].(*types.AttributeValueMemberS).Value

	last, err := time.Parse(time.RFC3339Nano, readLastUpdated)
	if err != nil {
		last = time.Now().In(l.timezone)
	}

	now := time.Now().In(l.timezone)
	elapsed := now.Sub(last)
	refill := int(elapsed.Seconds()) / l.refillSecs

	tokens := min(readTokens + refill, l.maxTokens)

	if tokens < 1 {
		return false, nil
	}

	tokens--
	consumedRefillDuration := time.Duration(refill*l.refillSecs) * time.Second
	newLastUpdated := last.Add(consumedRefillDuration).Format(time.RFC3339Nano)

	_, err = l.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(l.table),
		Item: map[string]types.AttributeValue{
			"PK":          &types.AttributeValueMemberS{Value: pk},
			"SK":          &types.AttributeValueMemberS{Value: sk},
			"tokens":      &types.AttributeValueMemberN{Value: strconv.Itoa(tokens)},
			"lastUpdated": &types.AttributeValueMemberS{Value: newLastUpdated},
		},
		ConditionExpression: aws.String("#tokens = :rt AND #lastUpdated = :rl"),
		ExpressionAttributeNames: map[string]string{
			"#tokens":      "tokens",
			"#lastUpdated": "lastUpdated",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":rt": &types.AttributeValueMemberN{Value: strconv.Itoa(readTokens)},
			":rl": &types.AttributeValueMemberS{Value: readLastUpdated},
		},
	})
	if err != nil {
		if isConditionalCheckFailed(err) {
			slog.Debug("ratelimit optimistic lock conflict, retrying", "sk", sk)
			return l.allow(ctx, pk, sk)
		}
		return false, fmt.Errorf("ratelimit put: %w", err)
	}

	return true, nil
}

func isConditionalCheckFailed(err error) bool {
	var ccf *types.ConditionalCheckFailedException
	return errors.As(err, &ccf)
}
