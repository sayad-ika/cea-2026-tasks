package dynamo

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

var (
	once   sync.Once
	client *dynamodb.Client
)

func GetClient() *dynamodb.Client {
	once.Do(func() {
		cfg, err := config.LoadDefaultConfig(context.Background(),
			config.WithRegion(os.Getenv("AWS_REGION")),
		)
		if err != nil {
			panic(fmt.Sprintf("failed to load AWS config: %v", err))
		}

		client = dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
			if endpoint := os.Getenv("DYNAMODB_ENDPOINT"); endpoint != "" {
				o.BaseEndpoint = aws.String(endpoint)
			}
		})
	})

	return client
}
