package dynamo

import (
	"testing"
)

func TestGetClient_WithLocalEndpoint_InitialisesWithoutError(t *testing.T) {
	resetClient()
	t.Cleanup(resetClient)

	t.Setenv("DYNAMODB_ENDPOINT", "http://localhost:8000")
	t.Setenv("AWS_REGION", "us-east-1")
	t.Setenv("AWS_ACCESS_KEY_ID", "local")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "local")

	c := GetClient()
	if c == nil {
		t.Fatal("expected non-nil DynamoDB client")
	}
}
