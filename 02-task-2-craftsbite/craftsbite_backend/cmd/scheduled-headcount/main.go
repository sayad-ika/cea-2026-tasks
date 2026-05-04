package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"sync"
	"time"

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
	sendDiscord    func(message discord.Message) error
	sendGChat      func(ctx context.Context, body []byte) error
	listAudience   func(ctx context.Context, roles ...string) ([]repository.User, error)
	availableMeals func(ctx context.Context, date string) ([]string, error)
}

func handler(ctx context.Context, dateParser *dateutil.DateParser, cfg *appconfig.Config, deps scheduledDeps, event ScheduledHeadcountEvent) error {
	if cfg.DiscordHeadcountChannelID == "" {
		return fmt.Errorf("notifier: DISCORD_HEADCOUNT_CHANNEL_ID is required")
	}
	if cfg.GChatHeadcountSpace == "" {
		return fmt.Errorf("notifier: GCHAT_HEADCOUNT_SPACE is required")
	}
	return runScheduledHeadcount(ctx, dateParser, deps, event)
}

func runScheduledHeadcount(ctx context.Context, dateParser *dateutil.DateParser, deps scheduledDeps, event ScheduledHeadcountEvent) error {
	date := event.Date
	if date == "" {
		var err error
		date, err = dateParser.ParseDateWithDefaults("")
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

	discordBody := headcountreport.BuildScheduledDiscordMessage(result)
	gchatBody, err := headcountreport.BuildScheduledGChatCard(result)
	if err != nil {
		return fmt.Errorf("build gchat card: %w", err)
	}

	var discordErr error
	var gchatErr error
	var mu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		if err := deps.sendDiscord(discordBody); err != nil {
			slog.Error("discord delivery failed", "error", err)
			mu.Lock()
			discordErr = err
			mu.Unlock()
		} else {
			slog.Info("discord delivery succeeded", "date", date)
		}
	}()

	go func() {
		defer wg.Done()
		if err := deps.sendGChat(ctx, gchatBody); err != nil {
			slog.Error("gchat delivery failed", "error", err)
			mu.Lock()
			gchatErr = err
			mu.Unlock()
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
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

	cfg := appconfig.MustLoad()
	client, err := dynamo.NewClient(cfg)
	if err != nil {
		log.Fatalf("scheduled-headcount: %v", err)
	}
	dateParser, err := dateutil.NewDateParser(cfg.Timezone)
	if err != nil {
		log.Fatalf("scheduled-headcount: %v", err)
	}
	store := repository.NewStore(client, cfg.DynamoDBTable)

	deps := scheduledDeps{
		headcount: func(ctx context.Context, date string) (*services.HeadcountResult, error) {
			return services.GetHeadcount(ctx, store, date)
		},
		sendDiscord: func(message discord.Message) error {
			return discord.CreateChannelMessageObject(cfg.DiscordBotToken, cfg.DiscordHeadcountChannelID, message)
		},
		sendGChat: func(ctx context.Context, body []byte) error {
			return gchat.CreateSpaceMessage(ctx, cfg.GChatServiceAccountJSON, cfg.GChatHeadcountSpace, body)
		},
		listAudience: func(ctx context.Context, roles ...string) ([]repository.User, error) {
			return store.ListActiveUsersByRoles(ctx, roles...)
		},
		availableMeals: func(ctx context.Context, date string) ([]string, error) {
			return store.GetAvailableMeals(ctx, date)
		},
	}

	const handlerTimeout = 28 * time.Second
	lambda.Start(func(ctx context.Context, event ScheduledHeadcountEvent) error {
		ctx, cancel := context.WithTimeout(ctx, handlerTimeout)
		defer cancel()
		return handler(ctx, dateParser, cfg, deps, event)
	})
}
