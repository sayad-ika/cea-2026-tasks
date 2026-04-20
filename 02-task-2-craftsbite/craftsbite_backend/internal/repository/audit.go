package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// AuditEntry represents an audit log entry
type AuditEntry struct {
	ActorUserID  string    // Who made the change
	Timestamp    time.Time // When the change was made
	EntityType   string    // What type of entity (e.g., "DAY_SCHEDULE")
	EntityKey    string    // The entity identifier (e.g., date for day schedules)
	Action       string    // "CREATE" or "UPDATE"
	OldValue     string    // JSON of old state (empty for CREATE)
	NewValue     string    // JSON of new state
	TargetEntity string    // For reverse lookup (e.g., "DAY#2026-03-26")
}

type auditItem struct {
	PK           string `dynamodbav:"PK"`
	SK           string `dynamodbav:"SK"`
	GSI1PK       string `dynamodbav:"GSI1PK"`
	GSI1SK       string `dynamodbav:"GSI1SK"`
	ActorUserID  string `dynamodbav:"actor_user_id"`
	Timestamp    string `dynamodbav:"timestamp"`
	EntityType   string `dynamodbav:"entity_type"`
	EntityKey    string `dynamodbav:"entity_key"`
	Action       string `dynamodbav:"action"`
	OldValue     string `dynamodbav:"old_value,omitempty"`
	NewValue     string `dynamodbav:"new_value"`
	TargetEntity string `dynamodbav:"target_entity"`
}

// WriteAuditEntry writes an audit log entry
func WriteAuditEntry(ctx context.Context, client *dynamodb.Client, table string, entry AuditEntry) error {
	timestamp := entry.Timestamp.UTC().Format(rfc3339)

	item := auditItem{
		PK:           "AUDIT#" + entry.ActorUserID,
		SK:           timestamp + "#" + entry.EntityType + "#" + entry.EntityKey,
		GSI1PK:       "AUDITEE#" + entry.TargetEntity,
		GSI1SK:       timestamp,
		ActorUserID:  entry.ActorUserID,
		Timestamp:    timestamp,
		EntityType:   entry.EntityType,
		EntityKey:    entry.EntityKey,
		Action:       entry.Action,
		OldValue:     entry.OldValue,
		NewValue:     entry.NewValue,
		TargetEntity: entry.TargetEntity,
	}

	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return fmt.Errorf("repository: WriteAuditEntry marshal: %w", err)
	}

	_, err = client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(table),
		Item:      av,
	})
	if err != nil {
		return fmt.Errorf("repository: WriteAuditEntry put: %w", err)
	}

	return nil
}


