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

// UpsertDaySchedule creates or updates a day schedule
func UpsertDaySchedule(ctx context.Context, client *dynamodb.Client, table string, schedule DaySchedule) error {
	now := time.Now().UTC().Format(rfc3339)

	// Validate day status
	if !IsValidDayStatus(schedule.DayStatus) {
		return fmt.Errorf("invalid day status: %s", schedule.DayStatus)
	}

	// Validate meal types
	for _, meal := range schedule.AvailableMeals {
		if !IsValidMealType(meal) {
			return fmt.Errorf("invalid meal type: %s", meal)
		}
	}

	item := dayScheduleItem{
		PK:             "DAY#" + schedule.Date,
		SK:             "METADATA",
		Date:           schedule.Date,
		DayStatus:      schedule.DayStatus,
		AvailableMeals: schedule.AvailableMeals,
		Reason:         schedule.Reason,
		CreatedBy:      schedule.CreatedBy,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return fmt.Errorf("repository: UpsertDaySchedule marshal: %w", err)
	}

	_, err = client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(table),
		Item:      av,
	})
	if err != nil {
		return fmt.Errorf("repository: UpsertDaySchedule put: %w", err)
	}

	// Also update the MEALS record for quick lookup
	return upsertDayMeals(ctx, client, table, schedule.Date, schedule.AvailableMeals)
}

// upsertDayMeals updates the meals lookup record
func upsertDayMeals(ctx context.Context, client *dynamodb.Client, table, date string, meals []string) error {
	item := dayMealsItem{
		PK:    "DAY#" + date,
		SK:    "MEALS",
		Meals: meals,
	}

	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return fmt.Errorf("repository: upsertDayMeals marshal: %w", err)
	}

	_, err = client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(table),
		Item:      av,
	})
	return err
}

// DeleteDaySchedule removes a day schedule (resets to default)
func DeleteDaySchedule(ctx context.Context, client *dynamodb.Client, table, date string) error {
	// Delete METADATA record
	_, err := client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(table),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "DAY#" + date},
			"SK": &types.AttributeValueMemberS{Value: "METADATA"},
		},
	})
	if err != nil {
		return fmt.Errorf("repository: DeleteDaySchedule metadata: %w", err)
	}

	// Delete MEALS record
	_, err = client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(table),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "DAY#" + date},
			"SK": &types.AttributeValueMemberS{Value: "MEALS"},
		},
	})
	if err != nil {
		return fmt.Errorf("repository: DeleteDaySchedule meals: %w", err)
	}

	return nil
}
