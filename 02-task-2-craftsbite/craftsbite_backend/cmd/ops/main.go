package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	appconfig "github.com/sayad-ika/craftsbite/internal/config"
	"github.com/sayad-ika/craftsbite/internal/dateutil"
	"github.com/sayad-ika/craftsbite/internal/discord"
	"github.com/sayad-ika/craftsbite/internal/dynamo"
	"github.com/sayad-ika/craftsbite/internal/gchat"
	"github.com/sayad-ika/craftsbite/internal/headcountreport"
	"github.com/sayad-ika/craftsbite/internal/payload"
	"github.com/sayad-ika/craftsbite/internal/repository"
	"github.com/sayad-ika/craftsbite/internal/services"
)

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
	case "headcount":
		return handleHeadcountCommand(ctx, client, c, event)
	case "schedule-day":
		return handleScheduleDayCommand(ctx, client, c, event)
	case "admin":
		return sendOpsReply(ctx, c, event, "This feature is coming soon.")
	default:
		return sendOpsReply(ctx, c, event, fmt.Sprintf("Unknown command: /%s", event.CommandName))
	}
}

func sendOpsReply(ctx context.Context, c *appconfig.Config, event payload.CommandEvent, text string) error {
	if event.Source == "gchat" {
		card, _ := gchat.SimpleTextCard(text)
		return gchat.CreatePrivateMessage(ctx, c.GChatServiceAccountJSON, event.GChatSpaceName, event.GChatViewerName, card)
	}
	return discord.SendFollowup(event.ApplicationID, event.InteractionToken, text)
}

func handleHeadcountCommand(ctx context.Context, client *dynamodb.Client, c *appconfig.Config, event payload.CommandEvent) error {
	if event.Role != "admin" && event.Role != "logistics" {
		return sendOpsReply(ctx, c, event, "You do not have permission to use `/headcount`.")
	}

	dateStr, _ := optString(event.Options, "date")
	date, err := dateutil.ParseDateWithDefaults(dateStr)
	if err != nil {
		return sendOpsReply(ctx, c, event, fmt.Sprintf("Invalid date: %v\nUse: tomorrow (default), +N, or YYYY-MM-DD", err))
	}

	result, err := services.GetHeadcount(ctx, client, c.DynamoDBTable, date)
	if err != nil {
		return sendOpsReply(ctx, c, event, "Something went wrong fetching headcount. Please try again later.")
	}

	if event.Source == "gchat" {
		card, _ := headcountreport.BuildGChatCard(result)
		return gchat.CreatePrivateMessage(ctx, c.GChatServiceAccountJSON, event.GChatSpaceName, event.GChatViewerName, card)
	}

	return discord.SendFollowup(event.ApplicationID, event.InteractionToken, headcountreport.BuildDiscordMessage(result))
}

func optString(opts map[string]interface{}, key string) (string, bool) {
	v, ok := opts[key]
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

func handleScheduleDayCommand(ctx context.Context, client *dynamodb.Client, c *appconfig.Config, event payload.CommandEvent) error {
	if event.Role != "admin" {
		return sendOpsReply(ctx, c, event, "You do not have permission to use `/schedule-day`.")
	}

	dateStr, _ := optString(event.Options, "date")

	dates, err := dateutil.ParseDateRange(dateStr)
	if err != nil {
		return sendOpsReply(ctx, c, event, fmt.Sprintf("Invalid date: %v\nUse: YYYY-MM-DD, YYYY-MM-DD..YYYY-MM-DD, or week", err))
	}

	statusStr, ok := optString(event.Options, "status")
	if !ok || statusStr == "" {
		return sendOpsReply(ctx, c, event, fmt.Sprintf("Day status is required. Valid values: %v", repository.ValidDayStatuses()))
	}

	mealsStr, _ := optString(event.Options, "meals")
	var meals []string
	if mealsStr != "" {
		meals = strings.Split(strings.ReplaceAll(mealsStr, " ", ""), ",")
	}

	reason, _ := optString(event.Options, "reason")

	if len(dates) > 1 {
		return handleBulkScheduleDayCommand(ctx, client, c, event, dates, statusStr, meals, reason)
	}

	input := services.SetDayScheduleInput{
		Date:           dates[0],
		DayStatus:      statusStr,
		AvailableMeals: meals,
		Reason:         reason,
		SetBy:          event.UserID,
	}

	schedule, err := services.SetDaySchedule(ctx, client, c.DynamoDBTable, input)
	if err != nil {
		return sendOpsReply(ctx, c, event, formatScheduleDayError(err))
	}

	return sendOpsReply(ctx, c, event, formatScheduleDaySuccess(schedule))
}

func handleBulkScheduleDayCommand(ctx context.Context, client *dynamodb.Client, c *appconfig.Config, event payload.CommandEvent, dates []string, statusStr string, meals []string, reason string) error {
	input := services.SetDayScheduleInput{
		DayStatus:      statusStr,
		AvailableMeals: meals,
		Reason:         reason,
		SetBy:          event.UserID,
	}

	result, err := services.BulkSetDaySchedule(ctx, client, c.DynamoDBTable, dates, input)
	if err != nil {
		return sendOpsReply(ctx, c, event, formatBulkScheduleDayError(err))
	}

	return sendOpsReply(ctx, c, event, formatBulkScheduleDaySuccess(result, statusStr))
}

func formatBulkScheduleDaySuccess(result *services.BulkSetDayScheduleResult, statusStr string) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "✓ Day schedule set to **%s** for %d weekday(s):\n", headcountreport.DisplayDayStatus(statusStr), len(result.SuccessDates))
	for _, d := range result.SuccessDates {
		fmt.Fprintf(&sb, "  • %s\n", d)
	}
	sb.WriteString("_(Weekend dates in the range were automatically skipped.)_")
	return sb.String()
}

func formatBulkScheduleDayError(err error) string {
	switch {
	case errors.Is(err, services.ErrAllWeekend):
		return "All dates in the specified range fall on weekends. No schedule was set."
	case errors.Is(err, services.ErrInvalidDate):
		return "Invalid date format. Use YYYY-MM-DD (e.g., 2026-03-25)"
	default:
		return fmt.Sprintf("Bulk schedule failed (no changes saved): %s", err.Error())
	}
}

func formatScheduleDayError(err error) string {
	errMsg := err.Error()
	switch {
	case errors.Is(err, services.ErrInvalidDate):
		return "Invalid date format. Use YYYY-MM-DD (e.g., 2026-03-25)"
	case errors.Is(err, services.ErrInvalidDayStatus):
		return errMsg
	case errors.Is(err, services.ErrInvalidMealType):
		return errMsg
	case strings.Contains(errMsg, "cannot have meals"):
		return "Office closed and government holiday days cannot have meals."
	default:
		return fmt.Sprintf("Failed to set day schedule: %s", errMsg)
	}
}

func formatScheduleDaySuccess(schedule *repository.DaySchedule) string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "✓ Day schedule set for %s\n\n", schedule.Date)
	fmt.Fprintf(&sb, "**Status**: %s\n", headcountreport.DisplayDayStatus(schedule.DayStatus))

	if len(schedule.AvailableMeals) > 0 {
		fmt.Fprintf(&sb, "**Meals**: %s\n", headcountreport.FormatMealList(schedule.AvailableMeals))
	} else {
		fmt.Fprintf(&sb, "**Meals**: None\n")
	}

	if schedule.Reason != "" {
		fmt.Fprintf(&sb, "**Reason**: %s\n", schedule.Reason)
	}

	return sb.String()
}

func main() {
	lambda.Start(handler)
}
