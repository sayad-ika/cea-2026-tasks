package main

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// --- Item builders ---
func teamItem(id, name, leadID string, active bool) map[string]types.AttributeValue {
	return map[string]types.AttributeValue{
		"PK":         &types.AttributeValueMemberS{Value: "TEAM#" + id},
		"SK":         &types.AttributeValueMemberS{Value: "#METADATA"},
		"entityType": &types.AttributeValueMemberS{Value: "TEAM"},
		"id":         &types.AttributeValueMemberS{Value: id},
		"name":       &types.AttributeValueMemberS{Value: name},
		"teamLeadID": &types.AttributeValueMemberS{Value: leadID},
		"active":     &types.AttributeValueMemberBOOL{Value: active},
		// GSI1 — team lead lookup
		"GSI1PK": &types.AttributeValueMemberS{Value: "TEAMLEAD#" + leadID},
		"GSI1SK": &types.AttributeValueMemberS{Value: "TEAM#" + id},
	}
}

func teamListingItem(id string, active bool) map[string]types.AttributeValue {
	activeStr := boolPrefix(active)
	return map[string]types.AttributeValue{
		"PK":         &types.AttributeValueMemberS{Value: "TEAM#" + id},
		"SK":         &types.AttributeValueMemberS{Value: "#LISTING"},
		"entityType": &types.AttributeValueMemberS{Value: "TEAM_LISTING"},
		"id":         &types.AttributeValueMemberS{Value: id},
		// GSI1 — entity listing
		"GSI1PK": &types.AttributeValueMemberS{Value: "ENTITY#TEAM"},
		"GSI1SK": &types.AttributeValueMemberS{Value: activeStr + "#TEAM#" + id},
	}
}

func userProfileItem(id, email, name, role string, active bool) map[string]types.AttributeValue {
	return map[string]types.AttributeValue{
		"PK":         &types.AttributeValueMemberS{Value: "USER#" + id},
		"SK":         &types.AttributeValueMemberS{Value: "#PROFILE"},
		"entityType": &types.AttributeValueMemberS{Value: "USER"},
		"id":         &types.AttributeValueMemberS{Value: id},
		"email":      &types.AttributeValueMemberS{Value: email},
		"name":       &types.AttributeValueMemberS{Value: name},
		"role":       &types.AttributeValueMemberS{Value: role},
		"active":     &types.AttributeValueMemberBOOL{Value: active},
		// GSI1 — email lookup
		"GSI1PK": &types.AttributeValueMemberS{Value: "EMAIL#" + email},
		"GSI1SK": &types.AttributeValueMemberS{Value: "USER#" + id},
	}
}

func userListingItem(id, role string, active bool) map[string]types.AttributeValue {
	activeStr := boolPrefix(active)
	return map[string]types.AttributeValue{
		"PK":         &types.AttributeValueMemberS{Value: "USER#" + id},
		"SK":         &types.AttributeValueMemberS{Value: "#LISTING"},
		"entityType": &types.AttributeValueMemberS{Value: "USER_LISTING"},
		"id":         &types.AttributeValueMemberS{Value: id},
		"role":       &types.AttributeValueMemberS{Value: role},
		// GSI1 — entity listing
		"GSI1PK": &types.AttributeValueMemberS{Value: "ENTITY#USER"},
		"GSI1SK": &types.AttributeValueMemberS{Value: activeStr + "#USER#" + id},
	}
}

func teamMemberItem(teamID, userID string) map[string]types.AttributeValue {
	return map[string]types.AttributeValue{
		"PK":         &types.AttributeValueMemberS{Value: "TEAM#" + teamID},
		"SK":         &types.AttributeValueMemberS{Value: "MEMBER#" + userID},
		"entityType": &types.AttributeValueMemberS{Value: "TEAM_MEMBER"},
		"teamID":     &types.AttributeValueMemberS{Value: teamID},
		"userID":     &types.AttributeValueMemberS{Value: userID},
		// GSI1 — user → team reverse lookup
		"GSI1PK": &types.AttributeValueMemberS{Value: "USER_TEAMS#" + userID},
		"GSI1SK": &types.AttributeValueMemberS{Value: "TEAM#" + teamID},
	}
}

func discordLookupItem(discordID, userID, role string) map[string]types.AttributeValue {
	return map[string]types.AttributeValue{
		"PK":         &types.AttributeValueMemberS{Value: "DISCORD#" + discordID},
		"SK":         &types.AttributeValueMemberS{Value: "#LOOKUP"},
		"entityType": &types.AttributeValueMemberS{Value: "DISCORD_LOOKUP"},
		"discordId":  &types.AttributeValueMemberS{Value: discordID},
		"userID":     &types.AttributeValueMemberS{Value: userID},
		"role":       &types.AttributeValueMemberS{Value: role},
	}
}

// --- Helpers ---

func putItem(ctx context.Context, client *dynamodb.Client, table string, item map[string]types.AttributeValue) error {
	_, err := client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(table),
		Item:      item,
	})
	return err
}

func boolPrefix(active bool) string {
	if active {
		return "true"
	}
	return "false"
}