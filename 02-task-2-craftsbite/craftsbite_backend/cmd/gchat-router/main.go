package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	lambdaclient "github.com/aws/aws-sdk-go-v2/service/lambda"
	appconfig "github.com/sayad-ika/craftsbite/internal/config"
	"github.com/sayad-ika/craftsbite/internal/dynamo"
	"github.com/sayad-ika/craftsbite/internal/gchat"
	"github.com/sayad-ika/craftsbite/internal/repository"
)

const handlerTimeout = 28 * time.Second

func newLambdaClient(c *appconfig.Config) (*lambdaclient.Client, error) {
	awscfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion(c.AWSRegion),
	)
	if err != nil {
		return nil, fmt.Errorf("gchat-router: failed to load AWS config: %w", err)
	}
	return lambdaclient.NewFromConfig(awscfg), nil
}

func handler(ctx context.Context, cfg *appconfig.Config, store *repository.Store, lambdaClient *lambdaclient.Client, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	if err := verifyGChatToken(ctx, req.Headers["authorization"]); err != nil {
		slog.Error("JWT verification failed", "error", err)
		return events.APIGatewayV2HTTPResponse{StatusCode: 401}, nil
	}

	body := req.Body
	if req.IsBase64Encoded {
		decoded, err := base64.StdEncoding.DecodeString(body)
		if err != nil {
			slog.Error("base64 decode failed", "error", err)
			return events.APIGatewayV2HTTPResponse{StatusCode: 400}, nil
		}
		body = string(decoded)
	}

	var evt gchat.Event
	if err := json.Unmarshal([]byte(body), &evt); err != nil {
		slog.Error("failed to unmarshal request body", "error", err)
		return events.APIGatewayV2HTTPResponse{StatusCode: 400}, nil
	}

	if evt.Chat.AppCommandPayload != nil {
		resp, err := handleMessage(ctx, cfg, store, lambdaClient, evt)
		if err != nil {
			slog.Error("handleMessage failed", "error", err)
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
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

	cfg := appconfig.MustLoad()
	client, err := dynamo.NewClient(cfg)
	if err != nil {
		log.Fatalf("gchat-router: %v", err)
	}
	lambdaClient, err := newLambdaClient(cfg)
	if err != nil {
		log.Fatalf("gchat-router: %v", err)
	}
	store := repository.NewStore(client, cfg.DynamoDBTable)

	lambda.Start(func(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
		ctx, cancel := context.WithTimeout(ctx, handlerTimeout)
		defer cancel()
		return handler(ctx, cfg, store, lambdaClient, req)
	})
}
