package main

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/aws/aws-lambda-go/lambda"
	appconfig "github.com/sayad-ika/craftsbite/internal/config"
	"github.com/sayad-ika/craftsbite/internal/dateutil"
	"github.com/sayad-ika/craftsbite/internal/discord"
	"github.com/sayad-ika/craftsbite/internal/dynamo"
	"github.com/sayad-ika/craftsbite/internal/gchat"
	"github.com/sayad-ika/craftsbite/internal/headcountreport"
	"github.com/sayad-ika/craftsbite/internal/repository"
	"github.com/sayad-ika/craftsbite/internal/services"
)

type ScheduledHeadcountEvent struct {
	Date string `json:"date,omitempty"`
}

type scheduledDeps struct {
	headcount      func(ctx context.Context, date string) (*services.HeadcountResult, error)
	sendDiscord    func(content string) error
	sendGChat      func(ctx context.Context, body []byte) error
	listAudience   func(ctx context.Context, roles ...string) ([]repository.User, error)
	availableMeals func(ctx context.Context, date string) ([]string, error)
}

func handler(ctx context.Context, event ScheduledHeadcountEvent) error {
	cfg := appconfig.MustLoad()

	if cfg.DiscordHeadcountChannelID == "" {
		return fmt.Errorf("notifier: DISCORD_HEADCOUNT_CHANNEL_ID is required")
	}
	if cfg.GChatHeadcountSpace == "" {
		return fmt.Errorf("notifier: GCHAT_HEADCOUNT_SPACE is required")
	}

	client := dynamo.GetClient(cfg)

	deps := scheduledDeps{
		headcount: func(ctx context.Context, date string) (*services.HeadcountResult, error) {
			return services.GetHeadcount(ctx, client, cfg.DynamoDBTable, date)
		},
		sendDiscord: func(content string) error {
			return discord.CreateChannelMessage(cfg.DiscordBotToken, cfg.DiscordHeadcountChannelID, content)
		},
		sendGChat: func(ctx context.Context, body []byte) error {
			return gchat.CreateSpaceMessage(ctx, cfg.GChatServiceAccountJSON, cfg.GChatHeadcountSpace, body)
		},
		listAudience: func(ctx context.Context, roles ...string) ([]repository.User, error) {
			return repository.ListActiveUsersByRoles(ctx, client, cfg.DynamoDBTable, roles...)
		},
		availableMeals: func(ctx context.Context, date string) ([]string, error) {
			return repository.GetAvailableMeals(ctx, client, cfg.DynamoDBTable, date)
		},
	}
	return runScheduledHeadcount(ctx, deps, event)
}

func runScheduledHeadcount(ctx context.Context, deps scheduledDeps, event ScheduledHeadcountEvent) error {
	date := event.Date
	if date == "" {
		var err error
		date, err = dateutil.ParseDateWithDefaults("")
		if err != nil {
			return fmt.Errorf("resolve date: %w", err)
		}
	}

	meals, err := deps.availableMeals(ctx, date)
	if err != nil {
		return fmt.Errorf("check available meals: %w", err)
	}
	if len(meals) == 0 {
		slog.Info("no meals configured, skipping notification", "date", date)
		return nil
	}

	audience, err := deps.listAudience(ctx, "admin", "logistics")
	if err != nil {
		slog.Warn("could not list audience for audit", "error", err)
	} else {
		slog.Info("scheduled headcount audience", "date", date, "eligible_users", len(audience))
	}

	result, err := deps.headcount(ctx, date)
	if err != nil {
		return fmt.Errorf("get headcount: %w", err)
	}

	discordBody := headcountreport.BuildDiscordMessage(result)
	gchatBody, err := headcountreport.BuildGChatCard(result)
	if err != nil {
		return fmt.Errorf("build gchat card: %w", err)
	}

	var discordErr, gchatErr error
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		if err := deps.sendDiscord(discordBody); err != nil {
			slog.Error("discord delivery failed", "error", err)
			discordErr = err
		} else {
			slog.Info("discord delivery succeeded", "date", date)
		}
	}()

	go func() {
		defer wg.Done()
		if err := deps.sendGChat(ctx, gchatBody); err != nil {
			slog.Error("gchat delivery failed", "error", err)
			gchatErr = err
		} else {
			slog.Info("gchat delivery succeeded", "date", date)
		}
	}()

	wg.Wait()

	if discordErr != nil || gchatErr != nil {
		return fmt.Errorf("delivery incomplete: discord=%v; gchat=%v", discordErr, gchatErr)
	}
	return nil
}

func main() {
	lambda.Start(handler)
}
