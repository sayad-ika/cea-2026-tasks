//go:build ignore

package main

import (
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

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