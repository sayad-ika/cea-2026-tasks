package main

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sayad-ika/craftsbite/internal/dateutil"
	"github.com/sayad-ika/craftsbite/internal/discord"
	"github.com/sayad-ika/craftsbite/internal/repository"
	"github.com/sayad-ika/craftsbite/internal/services"
)

func testDateParser(t *testing.T) *dateutil.DateParser {
	t.Helper()
	loc, err := time.LoadLocation("Asia/Dhaka")
	if err != nil {
		t.Fatalf("NewDateParser() returned unexpected error: %v", err)
	}
	return dateutil.NewDateParser(loc)
}

func TestRunScheduledHeadcount_NoMeals(t *testing.T) {
	deps := scheduledDeps{
		availableMeals: func(ctx context.Context, date string) ([]string, error) {
			return nil, nil
		},
		sendDiscord: func(message discord.Message) error {
			t.Error("sendDiscord should not be called when no meals")
			return nil
		},
		sendGChat: func(ctx context.Context, body []byte) error {
			t.Error("sendGChat should not be called when no meals")
			return nil
		},
		headcount: func(ctx context.Context, date string) (*services.HeadcountResult, error) {
			t.Error("headcount should not be called when no meals")
			return nil, nil
		},
		listAudience: func(ctx context.Context, roles ...string) ([]repository.User, error) {
			return nil, nil
		},
	}

	err := runScheduledHeadcount(context.Background(), testDateParser(t), deps, ScheduledHeadcountEvent{})
	if err != nil {
		t.Fatalf("expected nil for no meals, got %v", err)
	}
}

func TestRunScheduledHeadcount_HeadcountError(t *testing.T) {
	deps := scheduledDeps{
		availableMeals: func(ctx context.Context, date string) ([]string, error) {
			return []string{"lunch"}, nil
		},
		headcount: func(ctx context.Context, date string) (*services.HeadcountResult, error) {
			return nil, errors.New("dynamo down")
		},
		sendDiscord: func(message discord.Message) error {
			t.Error("sendDiscord should not be called on headcount error")
			return nil
		},
		sendGChat: func(ctx context.Context, body []byte) error {
			t.Error("sendGChat should not be called on headcount error")
			return nil
		},
		listAudience: func(ctx context.Context, roles ...string) ([]repository.User, error) {
			return nil, nil
		},
	}

	err := runScheduledHeadcount(context.Background(), testDateParser(t), deps, ScheduledHeadcountEvent{})
	if err == nil {
		t.Fatal("expected error when headcount fails")
	}
}

func TestRunScheduledHeadcount_BothSucceed(t *testing.T) {
	var discordCalled atomic.Bool
	var gchatCalled atomic.Bool

	deps := scheduledDeps{
		availableMeals: func(ctx context.Context, date string) ([]string, error) {
			return []string{"lunch"}, nil
		},
		headcount: func(ctx context.Context, date string) (*services.HeadcountResult, error) {
			return &services.HeadcountResult{Date: date, TotalUsers: 5}, nil
		},
		sendDiscord: func(message discord.Message) error {
			discordCalled.Store(true)
			if len(message.Embeds) == 0 {
				t.Error("expected discord embed payload")
			}
			return nil
		},
		sendGChat: func(ctx context.Context, body []byte) error {
			gchatCalled.Store(true)
			return nil
		},
		listAudience: func(ctx context.Context, roles ...string) ([]repository.User, error) {
			return nil, nil
		},
	}

	err := runScheduledHeadcount(context.Background(), testDateParser(t), deps, ScheduledHeadcountEvent{})
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if !discordCalled.Load() {
		t.Error("expected discord to be called")
	}
	if !gchatCalled.Load() {
		t.Error("expected gchat to be called")
	}
}

func TestRunScheduledHeadcount_UsesCompactSummaryPayload(t *testing.T) {
	deps := scheduledDeps{
		availableMeals: func(ctx context.Context, date string) ([]string, error) {
			return []string{"lunch"}, nil
		},
		headcount: func(ctx context.Context, date string) (*services.HeadcountResult, error) {
			return &services.HeadcountResult{
				Date:           date,
				DayStatus:      "normal",
				TotalUsers:     8,
				LocationCounts: services.LocationCount{Office: 5, WFH: 3},
				MealCounts: map[string]services.MealCount{
					"lunch": {MealType: "lunch", OptedIn: 5, OptedOut: 3},
				},
				Teams: []services.TeamHeadcount{{TeamName: "Engineering", MemberCount: 4}},
			}, nil
		},
		sendDiscord: func(message discord.Message) error {
			if len(message.Embeds) != 1 {
				t.Fatalf("expected 1 compact embed, got %d", len(message.Embeds))
			}
			for _, field := range message.Embeds[0].Fields {
				if field.Name == "Engineering" || strings.Contains(field.Value, "Engineering") {
					t.Fatal("did not expect team-specific details in scheduled discord payload")
				}
			}
			return nil
		},
		sendGChat: func(ctx context.Context, body []byte) error {
			if strings.Contains(string(body), "Engineering") {
				t.Fatal("did not expect team-specific details in scheduled gchat payload")
			}
			return nil
		},
		listAudience: func(ctx context.Context, roles ...string) ([]repository.User, error) {
			return nil, nil
		},
	}

	err := runScheduledHeadcount(context.Background(), testDateParser(t), deps, ScheduledHeadcountEvent{Date: "2026-05-01"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunScheduledHeadcount_DiscordFails_GChatSucceeds(t *testing.T) {
	var gchatCalled atomic.Bool

	deps := scheduledDeps{
		availableMeals: func(ctx context.Context, date string) ([]string, error) {
			return []string{"lunch"}, nil
		},
		headcount: func(ctx context.Context, date string) (*services.HeadcountResult, error) {
			return &services.HeadcountResult{Date: date, TotalUsers: 5}, nil
		},
		sendDiscord: func(message discord.Message) error {
			return errors.New("discord 403")
		},
		sendGChat: func(ctx context.Context, body []byte) error {
			gchatCalled.Store(true)
			return nil
		},
		listAudience: func(ctx context.Context, roles ...string) ([]repository.User, error) {
			return nil, nil
		},
	}

	err := runScheduledHeadcount(context.Background(), testDateParser(t), deps, ScheduledHeadcountEvent{})
	if err == nil {
		t.Fatal("expected error when discord fails")
	}
	if !gchatCalled.Load() {
		t.Error("expected gchat to still be called even when discord fails")
	}
}

func TestRunScheduledHeadcount_GChatFails_DiscordSucceeds(t *testing.T) {
	var discordCalled atomic.Bool

	deps := scheduledDeps{
		availableMeals: func(ctx context.Context, date string) ([]string, error) {
			return []string{"lunch"}, nil
		},
		headcount: func(ctx context.Context, date string) (*services.HeadcountResult, error) {
			return &services.HeadcountResult{Date: date, TotalUsers: 5}, nil
		},
		sendDiscord: func(message discord.Message) error {
			discordCalled.Store(true)
			return nil
		},
		sendGChat: func(ctx context.Context, body []byte) error {
			return errors.New("gchat 403")
		},
		listAudience: func(ctx context.Context, roles ...string) ([]repository.User, error) {
			return nil, nil
		},
	}

	err := runScheduledHeadcount(context.Background(), testDateParser(t), deps, ScheduledHeadcountEvent{})
	if err == nil {
		t.Fatal("expected error when gchat fails")
	}
	if !discordCalled.Load() {
		t.Error("expected discord to still be called even when gchat fails")
	}
}

func TestRunScheduledHeadcount_DeliveriesRunInParallel(t *testing.T) {
	discordStarted := make(chan struct{})
	gchatStarted := make(chan struct{})
	releaseDiscord := make(chan struct{})

	deps := scheduledDeps{
		availableMeals: func(ctx context.Context, date string) ([]string, error) {
			return []string{"lunch"}, nil
		},
		headcount: func(ctx context.Context, date string) (*services.HeadcountResult, error) {
			return &services.HeadcountResult{Date: date, TotalUsers: 5}, nil
		},
		sendDiscord: func(message discord.Message) error {
			close(discordStarted)
			<-releaseDiscord
			return nil
		},
		sendGChat: func(ctx context.Context, body []byte) error {
			close(gchatStarted)
			return nil
		},
		listAudience: func(ctx context.Context, roles ...string) ([]repository.User, error) {
			return nil, nil
		},
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- runScheduledHeadcount(context.Background(), testDateParser(t), deps, ScheduledHeadcountEvent{Date: "2026-05-01"})
	}()

	select {
	case <-discordStarted:
	case <-time.After(time.Second):
		t.Fatal("discord delivery did not start")
	}

	select {
	case <-gchatStarted:
	case <-time.After(time.Second):
		t.Fatal("gchat delivery did not start while discord delivery was blocked")
	}

	close(releaseDiscord)

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("runScheduledHeadcount did not return")
	}
}

func TestRunScheduledHeadcount_BothFail(t *testing.T) {
	deps := scheduledDeps{
		availableMeals: func(ctx context.Context, date string) ([]string, error) {
			return []string{"lunch"}, nil
		},
		headcount: func(ctx context.Context, date string) (*services.HeadcountResult, error) {
			return &services.HeadcountResult{Date: date, TotalUsers: 5}, nil
		},
		sendDiscord: func(message discord.Message) error {
			return errors.New("discord error")
		},
		sendGChat: func(ctx context.Context, body []byte) error {
			return errors.New("gchat error")
		},
		listAudience: func(ctx context.Context, roles ...string) ([]repository.User, error) {
			return nil, nil
		},
	}

	err := runScheduledHeadcount(context.Background(), testDateParser(t), deps, ScheduledHeadcountEvent{})
	if err == nil {
		t.Fatal("expected error when both deliveries fail")
	}
}

func TestRunScheduledHeadcount_DateOverride(t *testing.T) {
	var capturedDate string

	deps := scheduledDeps{
		availableMeals: func(ctx context.Context, date string) ([]string, error) {
			capturedDate = date
			return []string{"lunch"}, nil
		},
		headcount: func(ctx context.Context, date string) (*services.HeadcountResult, error) {
			return &services.HeadcountResult{Date: date, TotalUsers: 3}, nil
		},
		sendDiscord: func(message discord.Message) error { return nil },
		sendGChat:   func(ctx context.Context, body []byte) error { return nil },
		listAudience: func(ctx context.Context, roles ...string) ([]repository.User, error) {
			return nil, nil
		},
	}

	err := runScheduledHeadcount(context.Background(), testDateParser(t), deps, ScheduledHeadcountEvent{Date: "2026-05-01"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedDate != "2026-05-01" {
		t.Errorf("expected date override '2026-05-01', got %q", capturedDate)
	}
}
