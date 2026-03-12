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

func GetWorkLocation(ctx context.Context, client *dynamodb.Client, table, userID, date string) (*WorkLocation, error) {
	out, err := client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(table),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "USER#" + userID},
			"SK": &types.AttributeValueMemberS{Value: "WORKLOCATION#" + date},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("repository: GetWorkLocation: %w", err)
	}
	if out.Item == nil {
		return nil, nil
	}

	var item workLocationItem
	if err := attributevalue.UnmarshalMap(out.Item, &item); err != nil {
		return nil, fmt.Errorf("repository: GetWorkLocation unmarshal: %w", err)
	}

	return workLocationItemToRecord(item), nil
}

func UpsertWorkLocation(ctx context.Context, client *dynamodb.Client, table string, wl WorkLocation) error {
	now := time.Now().UTC().Format(rfc3339)

	createdAt := now
	if !wl.CreatedAt.IsZero() {
		createdAt = wl.CreatedAt.UTC().Format(rfc3339)
	}

	wire := workLocationItem{
		PK:        "USER#" + wl.UserID,
		SK:        "WORKLOCATION#" + wl.Date,
		GSI1PK:   wl.Date,
		GSI1SK:   wl.Location + "#USER#" + wl.UserID,
		UserID:    wl.UserID,
		Date:      wl.Date,
		Location:  wl.Location,
		SetBy:     wl.SetBy,
		Reason:    wl.Reason,
		CreatedAt: createdAt,
		UpdatedAt: now,
	}
	if wl.Location == "wfh" {
		wire.WFHMonth = wl.Date[:7] // "YYYY-MM"
	}

	item, err := attributevalue.MarshalMap(wire)
	if err != nil {
		return fmt.Errorf("repository: UpsertWorkLocation marshal: %w", err)
	}

	_, err = client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(table),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("repository: UpsertWorkLocation: %w", err)
	}
	return nil
}

func workLocationItemToRecord(item workLocationItem) *WorkLocation {
	createdAt, _ := time.Parse(rfc3339, item.CreatedAt)
	updatedAt, _ := time.Parse(rfc3339, item.UpdatedAt)
	return &WorkLocation{
		UserID:    item.UserID,
		Date:      item.Date,
		Location:  item.Location,
		SetBy:     item.SetBy,
		Reason:    item.Reason,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
}

func GetWorkLocationsByDate(ctx context.Context, client *dynamodb.Client, table, date string) ([]WorkLocation, error) {
	out, err := client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(table),
		IndexName:              aws.String("GSI1"),
		KeyConditionExpression: aws.String("GSI1PK = :date"),
		FilterExpression:       aws.String("attribute_exists(#loc)"),
		ExpressionAttributeNames: map[string]string{
			"#loc": "location",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":date": &types.AttributeValueMemberS{Value: date},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("repository: GetWorkLocationsByDate: %w", err)
	}

	results := make([]WorkLocation, 0, len(out.Items))
	for _, rawItem := range out.Items {
		var item workLocationItem
		if err := attributevalue.UnmarshalMap(rawItem, &item); err != nil {
			return nil, fmt.Errorf("repository: GetWorkLocationsByDate unmarshal: %w", err)
		}
		results = append(results, *workLocationItemToRecord(item))
	}
	return results, nil
}

