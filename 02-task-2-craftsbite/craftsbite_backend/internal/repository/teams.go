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

func GetTeamByID(ctx context.Context, client *dynamodb.Client, tableName, teamID string) (*Team, error) {
	out, err := client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "TEAM#" + teamID},
			"SK": &types.AttributeValueMemberS{Value: "METADATA"},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("repository: GetTeamByID: %w", err)
	}
	if out.Item == nil {
		return nil, nil
	}

	var item teamMetadataItem
	if err := attributevalue.UnmarshalMap(out.Item, &item); err != nil {
		return nil, fmt.Errorf("repository: GetTeamByID unmarshal: %w", err)
	}
	return &Team{
		ID:         item.ID,
		Name:       item.Name,
		TeamLeadID: item.TeamLeadID,
		Active:     item.Active,
	}, nil
}

func GetTeamMembers(ctx context.Context, client *dynamodb.Client, tableName, teamID string) ([]TeamMember, error) {
	out, err := client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(tableName),
		KeyConditionExpression: aws.String("PK = :pk AND begins_with(SK, :prefix)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk":     &types.AttributeValueMemberS{Value: "TEAM#" + teamID},
			":prefix": &types.AttributeValueMemberS{Value: "MEMBER#"},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("repository: GetTeamMembers: %w", err)
	}

	results := make([]TeamMember, 0, len(out.Items))
	for _, rawItem := range out.Items {
		var item teamMemberItem
		if err := attributevalue.UnmarshalMap(rawItem, &item); err != nil {
			return nil, fmt.Errorf("repository: GetTeamMembers unmarshal: %w", err)
		}
		joinedAt, _ := time.Parse(rfc3339, item.JoinedAt)
		results = append(results, TeamMember{
			TeamID:   item.TeamID,
			UserID:   item.UserID,
			JoinedAt: joinedAt,
		})
	}
	return results, nil
}

func FindTeamsByLeadID(ctx context.Context, client *dynamodb.Client, tableName, leadUserID string) ([]Team, error) {
	out, err := client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(tableName),
		IndexName:              aws.String("GSI1"),
		KeyConditionExpression: aws.String("GSI1PK = :pk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: "TEAMLEAD#" + leadUserID},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("repository: FindTeamsByLeadID: %w", err)
	}

	results := make([]Team, 0, len(out.Items))
	for _, rawItem := range out.Items {
		var item teamMetadataItem
		if err := attributevalue.UnmarshalMap(rawItem, &item); err != nil {
			return nil, fmt.Errorf("repository: FindTeamsByLeadID unmarshal: %w", err)
		}
		results = append(results, Team{
			ID:         item.ID,
			Name:       item.Name,
			TeamLeadID: item.TeamLeadID,
			Active:     item.Active,
		})
	}
	return results, nil
}
