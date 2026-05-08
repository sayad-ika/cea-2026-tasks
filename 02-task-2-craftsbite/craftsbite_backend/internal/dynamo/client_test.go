package dynamo_test

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/sayad-ika/craftsbite/internal/dynamo"
)

func TestNewClient_WithLocalEndpoint(t *testing.T) {
	c := dynamo.NewClient(aws.Config{}, "http://localhost:8000")
	if c == nil {
		t.Fatal("NewClient() returned nil; expected a valid *dynamodb.Client")
	}
}

func TestNewClient_WithoutLocalEndpoint(t *testing.T) {
	c := dynamo.NewClient(aws.Config{}, "")
	if c == nil {
		t.Fatal("NewClient() returned nil without DynamoDBEndpoint set")
	}
}
