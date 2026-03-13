package config

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

type Config struct {
	DiscordApplicationID string
	DiscordBotToken      string
	DiscordPublicKey     string

	AWSRegion string

	DynamoDBEndpoint string

	DynamoDBTable string

	LambdaSelfFunctionName       string
	LambdaManagementFunctionName string
	LambdaOpsFunctionName        string
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
		paramPrefix + "DISCORD_BOT_TOKEN",
		paramPrefix + "DISCORD_PUBLIC_KEY",
	})
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		DiscordApplicationID:         os.Getenv("DISCORD_APPLICATION_ID"),
		AWSRegion:                    os.Getenv("AWS_REGION"),
		DynamoDBEndpoint:             os.Getenv("DYNAMODB_ENDPOINT"),
		DynamoDBTable:                os.Getenv("DYNAMODB_TABLE"),
		LambdaSelfFunctionName:       os.Getenv("LAMBDA_SELF_FUNCTION_NAME"),
		LambdaManagementFunctionName: os.Getenv("LAMBDA_MANAGEMENT_FUNCTION_NAME"),
		LambdaOpsFunctionName:        os.Getenv("LAMBDA_OPS_FUNCTION_NAME"),
		DiscordBotToken:  params[paramPrefix+"DISCORD_BOT_TOKEN"],
		DiscordPublicKey: params[paramPrefix+"DISCORD_PUBLIC_KEY"],
	}

	if cfg.AWSRegion == "" {
		cfg.AWSRegion = "ap-southeast-1"
	}
	if cfg.DynamoDBTable == "" {
		cfg.DynamoDBTable = "craftsbite"
	}

	var missing []string
	if cfg.DiscordBotToken == "" {
		missing = append(missing, "DISCORD_BOT_TOKEN")
	}
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
