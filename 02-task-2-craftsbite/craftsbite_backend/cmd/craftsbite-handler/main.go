package main

import (
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	appconfig "github.com/sayad-ika/craftsbite/internal/config"
	"github.com/sayad-ika/craftsbite/internal/dateutil"
	"github.com/sayad-ika/craftsbite/internal/discord"
	"github.com/sayad-ika/craftsbite/internal/dynamo"
	"github.com/sayad-ika/craftsbite/internal/ratelimit"
	"github.com/sayad-ika/craftsbite/internal/repository"
	"github.com/sayad-ika/craftsbite/internal/services"
)

type handlerDeps struct {
	cfg        *appconfig.Config
	store      *repository.Store
	dateParser *dateutil.DateParser
	cutoff     *services.CutoffChecker
	limiter    *ratelimit.Limiter
}

func rawHandler(ctx context.Context, deps handlerDeps, raw json.RawMessage) (events.APIGatewayV2HTTPResponse, error) {
	var event events.APIGatewayV2HTTPRequest
	if err := json.Unmarshal(raw, &event); err != nil {
		slog.Error("api gateway event parse failed", "error", err)
		return events.APIGatewayV2HTTPResponse{}, err
	}
	return handler(ctx, deps, event)
}

func handler(ctx context.Context, deps handlerDeps, event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	req, gchatImmediate, discordImmediate, err := normalizeRequest(ctx, deps.cfg, deps.store, event)
	if gchatImmediate != nil || discordImmediate != nil || err != nil {
		if err != nil {
			slog.Warn("request normalization returned error", "error", err)
		}
		return immediateResponse(req.Platform, gchatImmediate, discordImmediate, err)
	}

	allowed, msg, err := applyRateLimit(ctx, deps.limiter, req.Command)
	if err != nil {
		slog.Error("rate limit check failed", "error", err, "userID", req.Command.UserID, "command", req.Command.CommandName)
		return platformNoticeResponse(req, msg, discord.NoticeToneError), nil
	}
	if !allowed {
		slog.Warn("rate limit exceeded", "userID", req.Command.UserID, "command", req.Command.CommandName)
		return platformNoticeResponse(req, msg, discord.NoticeToneWarning), nil
	}

	recordedCtx, recorder := withReplyRecorder(ctx, req)
	if err := route(recordedCtx, deps, req); err != nil {
		slog.Error("command route failed", "error", err, "platform", req.Platform, "command", req.Command.CommandName)
		return platformNoticeResponse(req, "An internal error occurred. Please try again.", discord.NoticeToneError), nil
	}

	return recorder.finalResponse(), nil
}

func immediateResponse(platform Platform, gchatResp *events.APIGatewayV2HTTPResponse, discordResp *RouterResponse, err error) (events.APIGatewayV2HTTPResponse, error) {
	if gchatResp != nil {
		return *gchatResp, nil
	}
	if discordResp != nil {
		return discordJSON(*discordResp), nil
	}
	if err != nil {
		status := 400
		if strings.HasPrefix(err.Error(), "401:") {
			status = 401
		}
		return events.APIGatewayV2HTTPResponse{StatusCode: status, Body: err.Error()}, nil
	}
	return events.APIGatewayV2HTTPResponse{StatusCode: 200}, nil
}

func platformNoticeResponse(req HandlerRequest, msg string, tone discord.NoticeTone) events.APIGatewayV2HTTPResponse {
	if req.Platform == PlatformDiscord {
		return discordJSON(ephemeralNotice(msg, tone))
	}
	return gchatNoticeText(msg, req.Command.GChatViewerName, tone)
}

func discordJSON(resp RouterResponse) events.APIGatewayV2HTTPResponse {
	body, _ := json.Marshal(resp)
	return events.APIGatewayV2HTTPResponse{
		StatusCode: 200,
		Body:       string(body),
		Headers:    map[string]string{"Content-Type": "application/json"},
	}
}

func prewarm(ctx context.Context, cfg *appconfig.Config, client *dynamodb.Client) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	_, _ = cfg.AwsConfig.Credentials.Retrieve(ctx)

	_, _ = client.DescribeEndpoints(ctx, &dynamodb.DescribeEndpointsInput{})
}

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

	cfg := appconfig.MustLoad()
	client := dynamo.NewClient(cfg.AwsConfig, cfg.DynamoDBEndpoint)
	prewarm(context.Background(), cfg, client)
	dateParser := dateutil.NewDateParser(cfg.Location)
	cutoff, err := services.NewCutoffChecker(services.CutoffConfig{
		Location:   cfg.Location,
		CutoffTime: cfg.CutoffTime,
	})
	if err != nil {
		log.Fatalf("craftsbite-handler: %v", err)
	}

	deps := handlerDeps{
		cfg:        cfg,
		store:      repository.NewStore(client, cfg.DynamoDBTable),
		dateParser: dateParser,
		cutoff:     cutoff,
		limiter:    ratelimit.NewLimiter(client, cfg.DynamoDBTable, cfg.RateLimitMaxTokens, cfg.RateLimitRefillSeconds, cfg.Location),
	}

	const handlerTimeout = 28 * time.Second
	lambda.Start(func(ctx context.Context, raw json.RawMessage) (events.APIGatewayV2HTTPResponse, error) {
		ctx, cancel := context.WithTimeout(ctx, handlerTimeout)
		defer cancel()
		return rawHandler(ctx, deps, raw)
	})
}
