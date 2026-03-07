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
