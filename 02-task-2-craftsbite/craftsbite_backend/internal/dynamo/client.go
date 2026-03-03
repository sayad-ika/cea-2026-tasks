package dynamo

import (
	"context"
	"log"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

var (
	mu     sync.Mutex
	client *dynamodb.Client
)

// GetClient returns a shared DynamoDB client.
// If DYNAMODB_ENDPOINT is set, it overrides the endpoint for local development.
// Otherwise the real AWS endpoint is resolved from the environment.
func GetClient() *dynamodb.Client {
	mu.Lock()
	defer mu.Unlock()

	if client != nil {
		return client
	}

	cfg, err := awsconfig.LoadDefaultConfig(context.Background())
	if err != nil {
	    log.Printf("failed to load AWS config: %v", err)
	    // In production, we want to fail hard if AWS config is not available.
		panic("failed to load AWS config: " + err.Error())
	}

	opts := []func(*dynamodb.Options){}

	if endpoint := os.Getenv("DYNAMODB_ENDPOINT"); endpoint != "" {
		opts = append(opts, func(o *dynamodb.Options) {
			o.BaseEndpoint = aws.String(endpoint)
		})
	}

	client = dynamodb.NewFromConfig(cfg, opts...)
	return client
}

// resetClient clears the cached client. For use in tests only.
func resetClient() {
	mu.Lock()
	defer mu.Unlock()
	client = nil
}
