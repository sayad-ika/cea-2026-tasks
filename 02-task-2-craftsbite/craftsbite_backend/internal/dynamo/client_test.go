package dynamo_test

import (
	"testing"

	"github.com/sayad-ika/craftsbite/internal/config"
	"github.com/sayad-ika/craftsbite/internal/dynamo"
)

func TestNewClient_WithLocalEndpoint(t *testing.T) {
	cfg := &config.Config{
		AWSRegion:        "ap-southeast-1",
		DynamoDBEndpoint: "http://localhost:8000",
		DynamoDBTable:    "craftsbite",
		DiscordPublicKey: "aabbccdd",
	}

	c, err := dynamo.NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient() returned unexpected error: %v", err)
	}
	if c == nil {
		t.Fatal("NewClient() returned nil; expected a valid *dynamodb.Client")
	}
}

func TestNewClient_WithoutLocalEndpoint(t *testing.T) {
	cfg := &config.Config{
		AWSRegion:        "ap-southeast-1",
		DynamoDBEndpoint: "",
		DynamoDBTable:    "craftsbite",
		DiscordPublicKey: "aabbccdd",
	}

	c, err := dynamo.NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient() returned unexpected error: %v", err)
	}
	if c == nil {
		t.Fatal("NewClient() returned nil without DynamoDBEndpoint set")
	}
}
