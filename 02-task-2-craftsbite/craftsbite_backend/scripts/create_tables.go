//go:build ignore

package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/joho/godotenv"

	"craftsbite-backend/internal/dynamo"
)

func main() {
    _ = godotenv.Load()

    client := dynamo.GetClient()
    ctx := context.Background()

    tables := []dynamodb.CreateTableInput{
        usersTable(),
        mealsTable(),
        workTable(),
    }

    for _, input := range tables {
        _, err := client.CreateTable(ctx, &input)
        if err != nil {
            // Ignore "already exists" errors so the script is safe to re-run
            var resErr *types.ResourceInUseException
            if errors.As(err, &resErr) {
                fmt.Printf("table %s already exists, skipping\n", *input.TableName)
                continue
            }
            log.Fatalf("failed to create table %s: %v", *input.TableName, err)
        }
        fmt.Printf("created table: %s\n", *input.TableName)
    }
}

func usersTable() dynamodb.CreateTableInput {
    return dynamodb.CreateTableInput{
        TableName:   aws.String(os.Getenv("USERS_TABLE")),
        BillingMode: types.BillingModePayPerRequest,
        KeySchema: []types.KeySchemaElement{
            {AttributeName: aws.String("PK"), KeyType: types.KeyTypeHash},
            {AttributeName: aws.String("SK"), KeyType: types.KeyTypeRange},
        },
        AttributeDefinitions: []types.AttributeDefinition{
            {AttributeName: aws.String("PK"),     AttributeType: types.ScalarAttributeTypeS},
            {AttributeName: aws.String("SK"),     AttributeType: types.ScalarAttributeTypeS},
            {AttributeName: aws.String("GSI1PK"), AttributeType: types.ScalarAttributeTypeS},
            {AttributeName: aws.String("GSI1SK"), AttributeType: types.ScalarAttributeTypeS},
        },
        GlobalSecondaryIndexes: []types.GlobalSecondaryIndex{
            {
                IndexName: aws.String("GSI1"),
                KeySchema: []types.KeySchemaElement{
                    {AttributeName: aws.String("GSI1PK"), KeyType: types.KeyTypeHash},
                    {AttributeName: aws.String("GSI1SK"), KeyType: types.KeyTypeRange},
                },
                Projection: &types.Projection{ProjectionType: types.ProjectionTypeAll},
            },
        },
    }
}

func mealsTable() dynamodb.CreateTableInput {
    return dynamodb.CreateTableInput{
        TableName:   aws.String(os.Getenv("MEALS_TABLE")),
        BillingMode: types.BillingModePayPerRequest,
        KeySchema: []types.KeySchemaElement{
            {AttributeName: aws.String("PK"), KeyType: types.KeyTypeHash},
            {AttributeName: aws.String("SK"), KeyType: types.KeyTypeRange},
        },
        AttributeDefinitions: []types.AttributeDefinition{
            {AttributeName: aws.String("PK"),     AttributeType: types.ScalarAttributeTypeS},
            {AttributeName: aws.String("SK"),     AttributeType: types.ScalarAttributeTypeS},
            {AttributeName: aws.String("GSI1PK"), AttributeType: types.ScalarAttributeTypeS},
            {AttributeName: aws.String("GSI1SK"), AttributeType: types.ScalarAttributeTypeS},
        },
        GlobalSecondaryIndexes: []types.GlobalSecondaryIndex{
            {
                IndexName: aws.String("GSI1"),
                KeySchema: []types.KeySchemaElement{
                    {AttributeName: aws.String("GSI1PK"), KeyType: types.KeyTypeHash},
                    {AttributeName: aws.String("GSI1SK"), KeyType: types.KeyTypeRange},
                },
                Projection: &types.Projection{ProjectionType: types.ProjectionTypeAll},
            },
        },
    }
}

func workTable() dynamodb.CreateTableInput {
    return dynamodb.CreateTableInput{
        TableName:   aws.String(os.Getenv("WORK_TABLE")),
        BillingMode: types.BillingModePayPerRequest,
        KeySchema: []types.KeySchemaElement{
            {AttributeName: aws.String("PK"), KeyType: types.KeyTypeHash},
            {AttributeName: aws.String("SK"), KeyType: types.KeyTypeRange},
        },
        AttributeDefinitions: []types.AttributeDefinition{
            {AttributeName: aws.String("PK"),     AttributeType: types.ScalarAttributeTypeS},
            {AttributeName: aws.String("SK"),     AttributeType: types.ScalarAttributeTypeS},
            {AttributeName: aws.String("GSI1PK"), AttributeType: types.ScalarAttributeTypeS},
            {AttributeName: aws.String("GSI1SK"), AttributeType: types.ScalarAttributeTypeS},
        },
        GlobalSecondaryIndexes: []types.GlobalSecondaryIndex{
            {
                IndexName: aws.String("GSI1"),
                KeySchema: []types.KeySchemaElement{
                    {AttributeName: aws.String("GSI1PK"), KeyType: types.KeyTypeHash},
                    {AttributeName: aws.String("GSI1SK"), KeyType: types.KeyTypeRange},
                },
                Projection: &types.Projection{ProjectionType: types.ProjectionTypeAll},
            },
        },
    }
}