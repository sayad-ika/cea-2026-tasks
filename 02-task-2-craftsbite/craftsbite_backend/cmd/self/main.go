package main

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	appconfig "github.com/sayad-ika/craftsbite/internal/config"
	"github.com/sayad-ika/craftsbite/internal/discord"
	"github.com/sayad-ika/craftsbite/internal/dynamo"
	"github.com/sayad-ika/craftsbite/internal/services"
)

type CommandEvent struct {
	UserID           string                 `json:"userID"`
	Role             string                 `json:"role"`
	DiscordID        string                 `json:"discordId"`
	CommandName      string                 `json:"commandName"`
	Options          map[string]interface{} `json:"options"`
	InteractionToken string                 `json:"interactionToken"`
	ApplicationID    string                 `json:"applicationId"`
}

var validMealOptions = map[string]bool{
	"lunch":           true,
	"snacks":          true,
	"event_dinner":    true,
	"optional_dinner": true,
	"all":             true,
}

var (
	cfgOnce sync.Once
	cfg     *appconfig.Config
)

func getConfig() *appconfig.Config {
	cfgOnce.Do(func() {
		cfg = appconfig.MustLoad()
	})
	return cfg
}

func handler(ctx context.Context, event CommandEvent) error {
	c := getConfig()
	client := dynamo.GetClient(c)

	var replyContent string

	switch event.CommandName {
	case "meal":
		replyContent = handleMeal(ctx, client, c.DynamoDBTable, event)
	case "location":
		replyContent = "This feature is coming soon."
	case "status":
		replyContent = "This feature is coming soon."
	default:
		replyContent = fmt.Sprintf("Unknown command: /%s", event.CommandName)
	}

	return discord.SendFollowup(event.ApplicationID, event.InteractionToken, replyContent)
}

func handleMeal(ctx context.Context, client *dynamodb.Client, table string, event CommandEvent) string {
	date, ok := optString(event.Options, "date")
	if !ok || date == "" {
		return "Please provide a date. Example: `/meal date:2026-03-10 status:out`"
	}

	statusStr, ok := optString(event.Options, "status")
	if !ok || (statusStr != "in" && statusStr != "out") {
		return "Please specify status as `in` or `out`."
	}
	isParticipating := statusStr == "in"

	mealType, _ := optString(event.Options, "meal")
	if mealType == "" {
		mealType = "all"
	}
	if !validMealOptions[mealType] {
		return fmt.Sprintf("`%s` is not a valid meal type. Choose from: lunch, snacks, event_dinner, optional_dinner, all.", mealType)
	}

	statuses, err := services.UpdateParticipation(ctx, client, table, event.UserID, date, mealType, isParticipating)
	if err != nil {
		return mealErrorReply(err, date)
	}

	return formatMealStatus(date, statuses)
}

func optString(opts map[string]interface{}, key string) (string, bool) {
	v, ok := opts[key]
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

func mealErrorReply(err error, date string) string {
	switch err {
	case services.ErrPastDate:
		return "Cannot update participation for a past date."
	case services.ErrCutoffPassed:
		return fmt.Sprintf("Updates for %s are closed. Cutoff was %s at 9:00 PM.", date, prevDay(date))
	case services.ErrTooFarAhead:
		return fmt.Sprintf("Cannot update participation for %s — that's more than 7 days away.", date)
	case services.ErrDayClosed:
		return fmt.Sprintf("Office is closed on %s — no meals are available.", date)
	case services.ErrNoMeals:
		return fmt.Sprintf("No meals are configured for %s.", date)
	case services.ErrMealUnavailable:
		return fmt.Sprintf("That meal is not available on %s.", date)
	default:
		return "Something went wrong. Please try again later."
	}
}

func prevDay(targetDate string) string {
	t, err := time.Parse("2006-01-02", targetDate)
	if err != nil {
		return targetDate + " (previous day)"
	}
	return t.AddDate(0, 0, -1).Format("2006-01-02")
}

func formatMealStatus(date string, statuses []services.ResolvedStatus) string {
	if len(statuses) == 0 {
		return fmt.Sprintf("No meals are available on %s.", date)
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Updated! Meal status for %s:", date)
	for _, s := range statuses {
		icon := "✗"
		if s.Status == "opted_in" {
			icon = "✓"
		} else if s.Status == "unavailable" {
			icon = "—"
		}
		fmt.Fprintf(&sb, "  %s %s", displayMealName(s.MealType), icon)
	}
	return sb.String()
}

func displayMealName(s string) string {
	words := strings.Split(strings.ReplaceAll(s, "_", " "), " ")
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, " ")
}

func main() {
	lambda.Start(handler)
}
