package dynamo_test

import (
	"testing"

	"github.com/sayad-ika/craftsbite/internal/config"
	"github.com/sayad-ika/craftsbite/internal/dynamo"
)

func TestGetClient_WithLocalEndpoint(t *testing.T) {
	cfg := &config.Config{
		AWSRegion:        "ap-southeast-1",
		DynamoDBEndpoint: "http://localhost:8000",
		DynamoDBTable:    "craftsbite",
		DiscordPublicKey: "aabbccdd",
	}

	c := dynamo.GetClient(cfg)
	if c == nil {
		t.Fatal("GetClient() returned nil; expected a valid *dynamodb.Client")
	}
}

func TestGetClient_WithoutLocalEndpoint(t *testing.T) {
	cfg := &config.Config{
		AWSRegion:        "ap-southeast-1",
		DynamoDBEndpoint: "",
		DynamoDBTable:    "craftsbite",
		DiscordPublicKey: "aabbccdd",
	}

	c := dynamo.GetClient(cfg)
	if c == nil {
		t.Fatal("GetClient() returned nil without DynamoDBEndpoint set")
	}
}
