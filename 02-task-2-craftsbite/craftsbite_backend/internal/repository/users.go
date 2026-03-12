package repository

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type discordLookupItem struct {
	UserID string `dynamodbav:"user_id"`
	Role   string `dynamodbav:"role"`
}

func GetUserByDiscordID(ctx context.Context, client *dynamodb.Client, tableName, discordID string) (userID, role string, err error) {
	out, err := client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "DISCORD#" + discordID},
			"SK": &types.AttributeValueMemberS{Value: "LOOKUP"},
		},
	})
	if err != nil {
		return "", "", fmt.Errorf("repository: GetUserByDiscordID: %w", err)
	}
	if out.Item == nil {
		return "", "", nil
	}

	var item discordLookupItem
	if err := attributevalue.UnmarshalMap(out.Item, &item); err != nil {
		return "", "", fmt.Errorf("repository: GetUserByDiscordID unmarshal: %w", err)
	}

	return item.UserID, item.Role, nil
}

func ListActiveUsers(ctx context.Context, client *dynamodb.Client, tableName string) ([]User, error) {
	out, err := client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(tableName),
		IndexName:              aws.String("GSI1"),
		KeyConditionExpression: aws.String("GSI1PK = :pk AND begins_with(GSI1SK, :prefix)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk":     &types.AttributeValueMemberS{Value: "ENTITY#USER"},
			":prefix": &types.AttributeValueMemberS{Value: "true#"},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("repository: ListActiveUsers: %w", err)
	}

	results := make([]User, 0, len(out.Items))
	for _, rawItem := range out.Items {
		var item userProfileItem
		if err := attributevalue.UnmarshalMap(rawItem, &item); err != nil {
			return nil, fmt.Errorf("repository: ListActiveUsers unmarshal: %w", err)
		}
		results = append(results, User{
			ID:     item.ID,
			Email:  item.Email,
			Name:   item.Name,
			Role:   item.Role,
			TeamID: item.TeamID,
			Active: item.Active,
		})
	}
	return results, nil
}