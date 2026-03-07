package dynamo_test

import (
	"os"
	"testing"

	"github.com/sayad-ika/craftsbite/internal/dynamo"
)

func TestGetClient_WithLocalEndpoint(t *testing.T) {
	t.Setenv("DYNAMODB_ENDPOINT", "http://localhost:8000")
	t.Setenv("AWS_REGION", "ap-southeast-1")

	c := dynamo.GetClient()
	if c == nil {
		t.Fatal("GetClient() returned nil; expected a valid *dynamodb.Client")
	}
}

func TestGetClient_WithoutLocalEndpoint(t *testing.T) {
	os.Unsetenv("DYNAMODB_ENDPOINT")
	t.Setenv("AWS_REGION", "ap-southeast-1")

	c := dynamo.GetClient()
	if c == nil {
		t.Fatal("GetClient() returned nil without DYNAMODB_ENDPOINT set")
	}
}
