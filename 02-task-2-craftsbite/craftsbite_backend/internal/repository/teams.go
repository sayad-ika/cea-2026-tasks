package repository

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func GetTeamByID(ctx context.Context, client *dynamodb.Client, table, teamID string) (*Team, error) {
	out, err := client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(table),
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
		ID:          item.ID,
		Name:        item.Name,
		Description: item.Description,
		LeadID:      item.TeamLeadID,
	}, nil
}
