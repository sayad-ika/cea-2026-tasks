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
	"github.com/sayad-ika/craftsbite/internal/dateutil"
	"github.com/sayad-ika/craftsbite/internal/discord"
	"github.com/sayad-ika/craftsbite/internal/dynamo"
	"github.com/sayad-ika/craftsbite/internal/gchat"
	"github.com/sayad-ika/craftsbite/internal/payload"
	"github.com/sayad-ika/craftsbite/internal/services"
)

var validMealOptions = map[string]bool{
	"lunch":           true,
	"snacks":          true,
	"iftar":           true,
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

func handler(ctx context.Context, event payload.CommandEvent) error {
	c := getConfig()
	client := dynamo.GetClient(c)

	switch event.CommandName {
	case "meal":
		replyContent := handleMeal(ctx, client, c.DynamoDBTable, event)
		if event.Source == "gchat" {
			card, _ := gchat.SimpleTextCard(replyContent)
			return gchat.CreatePrivateMessage(ctx, c.GChatServiceAccountJSON, event.GChatSpaceName, event.GChatViewerName, card)
		}
		return discord.SendFollowup(event.ApplicationID, event.InteractionToken, replyContent)

	case "location":
		return handleLocationCommand(ctx, client, c, event)

	case "status":
		return handleStatusCommand(ctx, client, c, event)

	default:
		replyContent := fmt.Sprintf("Unknown command: /%s", event.CommandName)
		if event.Source == "gchat" {
			card, _ := gchat.SimpleTextCard(replyContent)
			return gchat.CreatePrivateMessage(ctx, c.GChatServiceAccountJSON, event.GChatSpaceName, event.GChatViewerName, card)
		}
		return discord.SendFollowup(event.ApplicationID, event.InteractionToken, replyContent)
	}
}

func handleStatusCommand(ctx context.Context, client *dynamodb.Client, c *appconfig.Config, event payload.CommandEvent) error {
	dateStr, _ := optString(event.Options, "date")
	date, err := dateutil.ParseDateWithDefaults(dateStr)
	if err != nil {
		return sendSelfReply(ctx, c, event, fmt.Sprintf("Invalid date: %v\nUse: tomorrow (default), today, +N, or YYYY-MM-DD", err))
	}

	mealStatuses, err := services.GetUserMealStatus(ctx, client, c.DynamoDBTable, event.UserID, date)
	if err != nil {
		return sendSelfReply(ctx, c, event, "Unable to fetch meal status. Please try again later.")
	}

	location, err := services.GetLocation(ctx, client, c.DynamoDBTable, event.UserID, date)
	if err != nil {
		return sendSelfReply(ctx, c, event, "Unable to fetch location status. Please try again later.")
	}

	return sendSelfReply(ctx, c, event, formatStatusView(date, location.Location, mealStatuses))
}

func handleLocationCommand(ctx context.Context, client *dynamodb.Client, c *appconfig.Config, event payload.CommandEvent) error {
	dateStr, _ := optString(event.Options, "date")
	date, err := dateutil.ParseDateWithDefaults(dateStr)
	if err != nil {
		return sendSelfReply(ctx, c, event, fmt.Sprintf("Invalid date: %v\nUse: tomorrow (default), today, +N, or YYYY-MM-DD", err))
	}

	loc, ok := optString(event.Options, "location")
	if !ok || (loc != "office" && loc != "wfh") {
		return sendSelfReply(ctx, c, event, "Please specify location as `office` or `wfh`.")
	}

	wl, err := services.SetLocation(ctx, client, c.DynamoDBTable, event.UserID, date, loc)
	if err != nil {
		return sendSelfReply(ctx, c, event, locationErrorReply(err, date))
	}

	mealStatuses, _ := services.GetUserMealStatus(ctx, client, c.DynamoDBTable, event.UserID, date)

	if event.Source == "gchat" {
		mealText := formatMealStatusLine(mealStatuses)
		card, _ := gchat.LocationCard(wl.Location, date, mealText)
		return gchat.CreatePrivateMessage(ctx, c.GChatServiceAccountJSON, event.GChatSpaceName, event.GChatViewerName, card)
	}

	return discord.SendFollowup(event.ApplicationID, event.InteractionToken, formatLocationStatus(date, wl.Location, mealStatuses))
}

func sendSelfReply(ctx context.Context, c *appconfig.Config, event payload.CommandEvent, text string) error {
	if event.Source == "gchat" {
		card, _ := gchat.SimpleTextCard(text)
		return gchat.CreatePrivateMessage(ctx, c.GChatServiceAccountJSON, event.GChatSpaceName, event.GChatViewerName, card)
	}
	return discord.SendFollowup(event.ApplicationID, event.InteractionToken, text)
}

func formatMealStatusLine(statuses []services.ResolvedStatus) string {
	if len(statuses) == 0 {
		return "No meals configured"
	}
	var parts []string
	for _, s := range statuses {
		icon := "✗"
		if s.Status == "opted_in" {
			icon = "✓"
		} else if s.Status == "unavailable" {
			icon = "—"
		}
		parts = append(parts, fmt.Sprintf("%s %s", displayMealName(s.MealType), icon))
	}
	return strings.Join(parts, "  ")
}

func handleMeal(ctx context.Context, client *dynamodb.Client, table string, event payload.CommandEvent) string {
	dateStr, _ := optString(event.Options, "date")
	date, err := dateutil.ParseDateWithDefaults(dateStr)
	if err != nil {
		return fmt.Sprintf("Invalid date: %v\nUse: tomorrow (default), today, +N, or YYYY-MM-DD", err)
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

	// Track which meals were changed
	var changedMeals []string
	if mealType == "all" {
		// All available meals were changed
		for _, s := range statuses {
			if s.Status != "unavailable" {
				changedMeals = append(changedMeals, s.MealType)
			}
		}
	} else {
		changedMeals = []string{mealType}
	}

	return formatMealStatusWithChanges(date, statuses, changedMeals)
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

func locationErrorReply(err error, date string) string {
	switch err {
	case services.ErrLocationPastDate:
		return "Cannot set work location for a past date."
	case services.ErrLocationCutoffPassed:
		return fmt.Sprintf("Updates for %s are closed. Cutoff was %s at 9:00 PM.", date, prevDay(date))
	case services.ErrLocationTooFarAhead:
		return fmt.Sprintf("Cannot set work location for %s — that's more than 7 days away.", date)
	default:
		return "Something went wrong. Please try again later."
	}
}

func formatLocationStatus(date, location string, mealStatuses []services.ResolvedStatus) string {
	locIcon := "🏢"
	locLabel := "Office"
	if location == "wfh" {
		locIcon = "🏠"
		locLabel = "WFH"
	}

	var sb strings.Builder
	// Highlight the location change with arrow prefix
	fmt.Fprintf(&sb, "Updated! Status for %s:\n → %s %s", date, locIcon, locLabel)

	for _, s := range mealStatuses {
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

func formatMealStatusWithChanges(date string, statuses []services.ResolvedStatus, changedMeals []string) string {
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

		// Highlight changed meals with arrow prefix
		prefix := "  "
		if containsString(changedMeals, s.MealType) {
			prefix = " →"
		}

		fmt.Fprintf(&sb, "%s %s %s", prefix, displayMealName(s.MealType), icon)
	}
	return sb.String()
}

func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
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

func formatStatusView(date, location string, mealStatuses []services.ResolvedStatus) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Status for %s:\n", date)

	// Location status
	locIcon := "🏢"
	locLabel := "Office"
	if location == "wfh" {
		locIcon = "🏠"
		locLabel = "WFH"
	} else if location == "not_set" {
		locIcon = "❓"
		locLabel = "Not Set"
	}
	fmt.Fprintf(&sb, "  %s %s", locIcon, locLabel)

	// Meal status
	if len(mealStatuses) == 0 {
		fmt.Fprintf(&sb, "\n  No meals configured")
	} else {
		for _, s := range mealStatuses {
			icon := "✗"
			if s.Status == "opted_in" {
				icon = "✓"
			} else if s.Status == "unavailable" {
				icon = "—"
			}
			fmt.Fprintf(&sb, "  %s %s", displayMealName(s.MealType), icon)
		}
	}

	return sb.String()
}

func main() {
	lambda.Start(handler)
}
