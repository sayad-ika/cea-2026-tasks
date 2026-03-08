package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func GetParticipation(ctx context.Context, client *dynamodb.Client, table, userID, date, mealType string) (*MealParticipation, error) {
	out, err := client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(table),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "USER#" + userID},
			"SK": &types.AttributeValueMemberS{Value: "MEAL#" + date + "#" + mealType},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("repository: GetParticipation: %w", err)
	}
	if out.Item == nil {
		return nil, nil
	}

	var item mealItem
	if err := attributevalue.UnmarshalMap(out.Item, &item); err != nil {
		return nil, fmt.Errorf("repository: GetParticipation unmarshal: %w", err)
	}

	return mealItemToParticipation(item), nil
}

func GetParticipationsByUserDate(ctx context.Context, client *dynamodb.Client, table, userID, date string) ([]MealParticipation, error) {
	prefix := "MEAL#" + date + "#"
	out, err := client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(table),
		KeyConditionExpression: aws.String("PK = :pk AND begins_with(SK, :prefix)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk":     &types.AttributeValueMemberS{Value: "USER#" + userID},
			":prefix": &types.AttributeValueMemberS{Value: prefix},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("repository: GetParticipationsByUserDate: %w", err)
	}

	results := make([]MealParticipation, 0, len(out.Items))
	for _, rawItem := range out.Items {
		var item mealItem
		if err := attributevalue.UnmarshalMap(rawItem, &item); err != nil {
			return nil, fmt.Errorf("repository: GetParticipationsByUserDate unmarshal: %w", err)
		}
		results = append(results, *mealItemToParticipation(item))
	}
	return results, nil
}

func UpsertParticipation(ctx context.Context, client *dynamodb.Client, table string, p MealParticipation) error {
	now := time.Now().UTC().Format(rfc3339)

	createdAt := now
	if !p.CreatedAt.IsZero() {
		createdAt = p.CreatedAt.UTC().Format(rfc3339)
	}

	item, err := attributevalue.MarshalMap(mealItem{
		PK:              "USER#" + p.UserID,
		SK:              "MEAL#" + p.Date + "#" + p.MealType,
		GSI1PK:          p.Date,
		GSI1SK:          "MEAL#USER#" + p.UserID,
		UserID:          p.UserID,
		Date:            p.Date,
		MealType:        p.MealType,
		IsParticipating: p.IsParticipating,
		OverrideBy:      p.OverrideBy,
		OverrideReason:  p.OverrideReason,
		CreatedAt:       createdAt,
		UpdatedAt:       now,
	})
	if err != nil {
		return fmt.Errorf("repository: UpsertParticipation marshal: %w", err)
	}

	_, err = client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(table),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("repository: UpsertParticipation: %w", err)
	}
	return nil
}

func mealItemToParticipation(item mealItem) *MealParticipation {
	createdAt, _ := time.Parse(rfc3339, item.CreatedAt)
	updatedAt, _ := time.Parse(rfc3339, item.UpdatedAt)
	return &MealParticipation{
		UserID:          item.UserID,
		Date:            item.Date,
		MealType:        item.MealType,
		IsParticipating: item.IsParticipating,
		OverrideBy:      item.OverrideBy,
		OverrideReason:  item.OverrideReason,
		CreatedAt:       createdAt,
		UpdatedAt:       updatedAt,
	}
}
