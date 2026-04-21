package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	lambdaclient "github.com/aws/aws-sdk-go-v2/service/lambda"
	appconfig "github.com/sayad-ika/craftsbite/internal/config"
	"github.com/sayad-ika/craftsbite/internal/dynamo"
	"github.com/sayad-ika/craftsbite/internal/gchat"
)

func newLambdaClient(c *appconfig.Config) (*lambdaclient.Client, error) {
	awscfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion(c.AWSRegion),
	)
	if err != nil {
		return nil, fmt.Errorf("gchat-router: failed to load AWS config: %w", err)
	}
	return lambdaclient.NewFromConfig(awscfg), nil
}

func handler(ctx context.Context, cfg *appconfig.Config, client *dynamodb.Client, lambdaClient *lambdaclient.Client, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	if err := verifyGChatToken(ctx, req.Headers["authorization"]); err != nil {
		log.Printf("gchat-router: JWT verification failed: %v", err)
		return events.APIGatewayV2HTTPResponse{StatusCode: 401}, nil
	}

	body := req.Body
	if req.IsBase64Encoded {
		decoded, err := base64.StdEncoding.DecodeString(body)
		if err != nil {
			log.Printf("gchat-router: base64 decode failed: %v", err)
			return events.APIGatewayV2HTTPResponse{StatusCode: 400}, nil
		}
		body = string(decoded)
	}

	var evt gchat.Event
	if err := json.Unmarshal([]byte(body), &evt); err != nil {
		log.Printf("gchat-router: failed to unmarshal request body: %v", err)
		return events.APIGatewayV2HTTPResponse{StatusCode: 400}, nil
	}

	if evt.Chat.AppCommandPayload != nil {
		resp, err := handleMessage(ctx, cfg, client, lambdaClient, evt)
		if err != nil {
			log.Printf("gchat-router: handleMessage: %v", err)
			return gchatText("An error occurred. Please try again.", evt.Chat.User.Name), nil
		}
		return resp, nil
	}

	return ok("")
}

func ok(body string) (events.APIGatewayV2HTTPResponse, error) {
	return events.APIGatewayV2HTTPResponse{StatusCode: 200, Body: body}, nil
}

func main() {
	cfg := appconfig.MustLoad()
	client, err := dynamo.NewClient(cfg)
	if err != nil {
		log.Fatalf("gchat-router: %v", err)
	}
	lambdaClient, err := newLambdaClient(cfg)
	if err != nil {
		log.Fatalf("gchat-router: %v", err)
	}

	lambda.Start(func(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
		return handler(ctx, cfg, client, lambdaClient, req)
	})
}
