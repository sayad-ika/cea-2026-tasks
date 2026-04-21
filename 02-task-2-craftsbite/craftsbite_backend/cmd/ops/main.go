package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/sayad-ika/craftsbite/internal/cmdutil"
	appconfig "github.com/sayad-ika/craftsbite/internal/config"
	"github.com/sayad-ika/craftsbite/internal/dateutil"
	"github.com/sayad-ika/craftsbite/internal/dynamo"
	"github.com/sayad-ika/craftsbite/internal/headcountreport"
	"github.com/sayad-ika/craftsbite/internal/payload"
	"github.com/sayad-ika/craftsbite/internal/repository"
	"github.com/sayad-ika/craftsbite/internal/services"
)

func handler(ctx context.Context, client *dynamodb.Client, cfg *appconfig.Config, dateParser *dateutil.DateParser, event payload.CommandEvent) error {
	switch event.CommandName {
	case "headcount":
		return handleHeadcountCommand(ctx, client, cfg, dateParser, event)
	case "schedule-day":
		return handleScheduleDayCommand(ctx, client, cfg, dateParser, event)
	case "admin":
		return cmdutil.SendReply(ctx, cfg, event, "This feature is coming soon.")
	default:
		return cmdutil.SendReply(ctx, cfg, event, fmt.Sprintf("Unknown command: /%s", event.CommandName))
	}
}

func handleHeadcountCommand(ctx context.Context, client *dynamodb.Client, cfg *appconfig.Config, dateParser *dateutil.DateParser, event payload.CommandEvent) error {
	if event.Role != "admin" && event.Role != "logistics" {
		return cmdutil.SendReply(ctx, cfg, event, "You do not have permission to use `/headcount`.")
	}

	dateStr, _ := cmdutil.OptString(event.Options, "date")
	date, err := dateParser.ParseDateWithDefaults(dateStr)
	if err != nil {
		return cmdutil.SendReply(ctx, cfg, event, fmt.Sprintf("Invalid date: %v\nUse: tomorrow (default), +N, or YYYY-MM-DD", err))
	}

	result, err := services.GetHeadcount(ctx, client, cfg.DynamoDBTable, date)
	if err != nil {
		return cmdutil.SendReply(ctx, cfg, event, "Something went wrong fetching headcount. Please try again later.")
	}

	if event.Source == "gchat" {
		card, _ := headcountreport.BuildGChatCard(result)
		return cmdutil.SendGChatCard(ctx, cfg, event, card)
	}

	return cmdutil.SendReply(ctx, cfg, event, headcountreport.BuildDiscordMessage(result))
}

func handleScheduleDayCommand(ctx context.Context, client *dynamodb.Client, cfg *appconfig.Config, dateParser *dateutil.DateParser, event payload.CommandEvent) error {
	if event.Role != "admin" {
		return cmdutil.SendReply(ctx, cfg, event, "You do not have permission to use `/schedule-day`.")
	}

	dateStr, _ := cmdutil.OptString(event.Options, "date")
	dates, err := dateParser.ParseDateRange(dateStr)
	if err != nil {
		return cmdutil.SendReply(ctx, cfg, event, fmt.Sprintf("Invalid date: %v\nUse: YYYY-MM-DD, YYYY-MM-DD..YYYY-MM-DD, or week", err))
	}

	statusStr, ok := cmdutil.OptString(event.Options, "status")
	if !ok || statusStr == "" {
		return cmdutil.SendReply(ctx, cfg, event, fmt.Sprintf("Day status is required. Valid values: %v", repository.ValidDayStatuses()))
	}

	mealsStr, _ := cmdutil.OptString(event.Options, "meals")
	var meals []string
	if mealsStr != "" {
		meals = strings.Split(strings.ReplaceAll(mealsStr, " ", ""), ",")
	}

	reason, _ := cmdutil.OptString(event.Options, "reason")

	if len(dates) > 1 {
		return handleBulkScheduleDayCommand(ctx, client, cfg, event, dates, statusStr, meals, reason)
	}

	input := services.SetDayScheduleInput{
		Date:           dates[0],
		DayStatus:      statusStr,
		AvailableMeals: meals,
		Reason:         reason,
		SetBy:          event.UserID,
	}

	schedule, err := services.SetDaySchedule(ctx, client, cfg.DynamoDBTable, input)
	if err != nil {
		return cmdutil.SendReply(ctx, cfg, event, formatScheduleDayError(err))
	}

	return cmdutil.SendReply(ctx, cfg, event, formatScheduleDaySuccess(schedule))
}

func handleBulkScheduleDayCommand(ctx context.Context, client *dynamodb.Client, cfg *appconfig.Config, event payload.CommandEvent, dates []string, statusStr string, meals []string, reason string) error {
	input := services.SetDayScheduleInput{
		DayStatus:      statusStr,
		AvailableMeals: meals,
		Reason:         reason,
		SetBy:          event.UserID,
	}

	result, err := services.BulkSetDaySchedule(ctx, client, cfg.DynamoDBTable, dates, input)
	if err != nil {
		return cmdutil.SendReply(ctx, cfg, event, formatBulkScheduleDayError(err))
	}

	return cmdutil.SendReply(ctx, cfg, event, formatBulkScheduleDaySuccess(result, statusStr))
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
	cfg := appconfig.MustLoad()
	client, err := dynamo.NewClient(cfg)
	if err != nil {
		log.Fatalf("ops: %v", err)
	}
	dateParser, err := dateutil.NewDateParser(cfg.Timezone)
	if err != nil {
		log.Fatalf("ops: %v", err)
	}

	lambda.Start(func(ctx context.Context, event payload.CommandEvent) error {
		return handler(ctx, client, cfg, dateParser, event)
	})
}
