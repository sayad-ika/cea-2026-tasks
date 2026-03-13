package config

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

type Config struct {
	DiscordPublicKey string
	AWSRegion        string
	DynamoDBEndpoint string
	DynamoDBTable    string
}

const paramPrefix = "/craftsbite/"

func Load() (*Config, error) {
	ctx := context.Background()

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion("ap-southeast-1"))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	ssmClient := ssm.NewFromConfig(awsCfg)

	params, err := fetchParams(ctx, ssmClient, []string{
		paramPrefix + "DISCORD_PUBLIC_KEY",
		paramPrefix + "AWS_REGION",
		paramPrefix + "DYNAMODB_ENDPOINT",
		paramPrefix + "DYNAMODB_TABLE",
	})
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		DiscordPublicKey: params[paramPrefix+"DISCORD_PUBLIC_KEY"],
		AWSRegion:        params[paramPrefix+"AWS_REGION"],
		DynamoDBEndpoint: params[paramPrefix+"DYNAMODB_ENDPOINT"],
		DynamoDBTable:    params[paramPrefix+"DYNAMODB_TABLE"],
	}

	if cfg.AWSRegion == "" {
		cfg.AWSRegion = "ap-southeast-1"
	}
	if cfg.DynamoDBTable == "" {
		cfg.DynamoDBTable = "craftsbite"
	}

	var missing []string
	if cfg.DiscordPublicKey == "" {
		missing = append(missing, "DISCORD_PUBLIC_KEY")
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required parameters: %s", strings.Join(missing, ", "))
	}

	return cfg, nil
}

func fetchParams(ctx context.Context, client *ssm.Client, names []string) (map[string]string, error) {
	resp, err := client.GetParameters(ctx, &ssm.GetParametersInput{
		Names:          names,
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch parameters: %w", err)
	}

	result := make(map[string]string)
	for _, p := range resp.Parameters {
		result[*p.Name] = *p.Value
	}
	return result, nil
}

func MustLoad() *Config {
	cfg, err := Load()
	if err != nil {
		panic("config: " + err.Error())
	}
	return cfg
}
