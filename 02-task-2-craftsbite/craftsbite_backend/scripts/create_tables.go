//go:build ignore

// create_tables.go creates the craftsbite DynamoDB table.
// Safe to re-run — it is a no-op if the table already exists.
//
// Usage:
//
//	go run ./scripts/create_tables.go
package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env if present — fine to ignore missing file in CI.
	_ = godotenv.Load()

	tableName := os.Getenv("DYNAMODB_TABLE")
	if tableName == "" {
		tableName = "craftsbite"
	}

	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "ap-southeast-1"
	}

	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(region),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load AWS config: %v\n", err)
		os.Exit(1)
	}

	client := dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
		if endpoint := os.Getenv("DYNAMODB_ENDPOINT"); endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
		}
	})

	ctx := context.Background()

	_, err = client.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName: aws.String(tableName),
		BillingMode: types.BillingModePayPerRequest,

		// Primary key schema.
		AttributeDefinitions: []types.AttributeDefinition{
			{AttributeName: aws.String("PK"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("SK"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("GSI1PK"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("GSI1SK"), AttributeType: types.ScalarAttributeTypeS},
		},
		KeySchema: []types.KeySchemaElement{
			{AttributeName: aws.String("PK"), KeyType: types.KeyTypeHash},
			{AttributeName: aws.String("SK"), KeyType: types.KeyTypeRange},
		},

		// GSI1 — overloaded index serving date, team, entity listing, and audit patterns.
		GlobalSecondaryIndexes: []types.GlobalSecondaryIndex{
			{
				IndexName: aws.String("GSI1"),
				KeySchema: []types.KeySchemaElement{
					{AttributeName: aws.String("GSI1PK"), KeyType: types.KeyTypeHash},
					{AttributeName: aws.String("GSI1SK"), KeyType: types.KeyTypeRange},
				},
				Projection: &types.Projection{
					ProjectionType: types.ProjectionTypeAll,
				},
			},
		},
	})

	if err != nil {
		// ResourceInUseException means the table already exists — treat as success.
		var inUse *types.ResourceInUseException
		if errors.As(err, &inUse) {
			fmt.Printf("Table %q already exists — no-op.\n", tableName)
			return
		}
		fmt.Fprintf(os.Stderr, "CreateTable error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Table %q created successfully.\n", tableName)
}
