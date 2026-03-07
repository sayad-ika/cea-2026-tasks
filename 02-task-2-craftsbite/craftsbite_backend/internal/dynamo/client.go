package dynamo

import (
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	appconfig "github.com/sayad-ika/craftsbite/internal/config"
)

var (
	once   sync.Once
	client *dynamodb.Client
)

func GetClient(cfg *appconfig.Config) *dynamodb.Client {
	once.Do(func() {
		awscfg, err := awsconfig.LoadDefaultConfig(context.Background(),
			awsconfig.WithRegion(cfg.AWSRegion),
		)
		if err != nil {
			panic(fmt.Sprintf("dynamo: failed to load AWS config: %v", err))
		}

		client = dynamodb.NewFromConfig(awscfg, func(o *dynamodb.Options) {
			if cfg.DynamoDBEndpoint != "" {
				o.BaseEndpoint = aws.String(cfg.DynamoDBEndpoint)
			}
		})
	})

	return client
}
