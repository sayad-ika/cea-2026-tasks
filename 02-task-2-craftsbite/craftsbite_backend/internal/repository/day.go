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



func GetDay(ctx context.Context, client *dynamodb.Client, table, date string) (*DaySchedule, error) {
	out, err := client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(table),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "DAY#" + date},
			"SK": &types.AttributeValueMemberS{Value: "METADATA"},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("repository: GetDay: %w", err)
	}
	if out.Item == nil {
		return nil, nil
	}

	var item dayScheduleItem
	if err := attributevalue.UnmarshalMap(out.Item, &item); err != nil {
		return nil, fmt.Errorf("repository: GetDay unmarshal: %w", err)
	}

	createdAt, _ := time.Parse(rfc3339, item.CreatedAt)
	updatedAt, _ := time.Parse(rfc3339, item.UpdatedAt)

	return &DaySchedule{
		Date:           item.Date,
		DayStatus:      item.DayStatus,
		AvailableMeals: item.AvailableMeals,
		Reason:         item.Reason,
		CreatedBy:      item.CreatedBy,
		CreatedAt:      createdAt,
		UpdatedAt:      updatedAt,
	}, nil
}

func GetAvailableMeals(ctx context.Context, client *dynamodb.Client, table, date string) ([]string, error) {
	out, err := client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(table),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "DAY#" + date},
			"SK": &types.AttributeValueMemberS{Value: "MEALS"},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("repository: GetAvailableMeals: %w", err)
	}
	if out.Item == nil {
		return nil, nil
	}

	var item dayMealsItem
	if err := attributevalue.UnmarshalMap(out.Item, &item); err != nil {
		return nil, fmt.Errorf("repository: GetAvailableMeals unmarshal: %w", err)
	}
	return item.Meals, nil
}
