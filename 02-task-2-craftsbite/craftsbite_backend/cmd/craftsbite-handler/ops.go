package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	appconfig "github.com/sayad-ika/craftsbite/internal/config"
	"github.com/sayad-ika/craftsbite/internal/dateutil"
	"github.com/sayad-ika/craftsbite/internal/headcountreport"
	"github.com/sayad-ika/craftsbite/internal/payload"
	"github.com/sayad-ika/craftsbite/internal/repository"
	"github.com/sayad-ika/craftsbite/internal/services"
)

func handleOpsCommand(ctx context.Context, deps handlerDeps, event payload.CommandEvent) error {
	switch event.CommandName {
	case "headcount":
		return handleHeadcountCommand(ctx, deps.store, deps.cfg, deps.dateParser, event)
	case "schedule-day":
		return handleScheduleDayCommand(ctx, deps.store, deps.cfg, deps.dateParser, event)
	case "admin":
		return sendReply(ctx, deps.cfg, event, "This feature is coming soon.")
	default:
		return sendReply(ctx, deps.cfg, event, fmt.Sprintf("Unknown command: /%s", event.CommandName))
	}
}

func handleHeadcountCommand(ctx context.Context, store *repository.Store, cfg *appconfig.Config, dateParser *dateutil.DateParser, event payload.CommandEvent) error {
	if event.Role != "admin" && event.Role != "logistics" {
		return sendReply(ctx, cfg, event, "You do not have permission to use `/headcount`.")
	}

	var opts payload.HeadcountOptions
	if err := event.ParseOptions(&opts); err != nil {
		return sendReply(ctx, cfg, event, "Invalid command options.")
	}
	date, err := dateParser.ParseDateWithDefaults(opts.Date)
	if err != nil {
		return sendReply(ctx, cfg, event, fmt.Sprintf("Invalid date: %v\nUse: tomorrow (default), +N, or YYYY-MM-DD", err))
	}

	result, err := services.GetHeadcount(ctx, store, date)
	if err != nil {
		return sendReply(ctx, cfg, event, "Something went wrong fetching headcount. Please try again later.")
	}

	if event.Source == "gchat" {
		card, _ := headcountreport.BuildGChatCard(result)
		return sendGChatCard(ctx, cfg, event, card)
	}

	return sendReply(ctx, cfg, event, headcountreport.BuildDiscordMessage(result))
}

func handleScheduleDayCommand(ctx context.Context, store *repository.Store, cfg *appconfig.Config, dateParser *dateutil.DateParser, event payload.CommandEvent) error {
	if event.Role != "admin" {
		return sendReply(ctx, cfg, event, "You do not have permission to use `/schedule-day`.")
	}

	var opts payload.ScheduleDayOptions
	if err := event.ParseOptions(&opts); err != nil {
		return sendReply(ctx, cfg, event, "Invalid command options.")
	}
	dates, err := dateParser.ParseDateRange(opts.Date)
	if err != nil {
		return sendReply(ctx, cfg, event, fmt.Sprintf("Invalid date: %v\nUse: YYYY-MM-DD, YYYY-MM-DD..YYYY-MM-DD, or week", err))
	}

	if opts.Status == "" {
		return sendReply(ctx, cfg, event, fmt.Sprintf("Day status is required. Valid values: %v", repository.ValidDayStatuses()))
	}

	var meals []string
	if opts.Meals != "" {
		meals = strings.Split(strings.ReplaceAll(opts.Meals, " ", ""), ",")
	}

	reason := opts.Reason

	if len(dates) > 1 {
		return handleBulkScheduleDayCommand(ctx, store, cfg, event, dates, opts.Status, meals, reason)
	}

	input := services.SetDayScheduleInput{
		Date:           dates[0],
		DayStatus:      opts.Status,
		AvailableMeals: meals,
		Reason:         reason,
		SetBy:          event.UserID,
	}

	schedule, err := services.SetDaySchedule(ctx, store, input)
	if err != nil {
		return sendReply(ctx, cfg, event, formatScheduleDayError(err))
	}

	return sendReply(ctx, cfg, event, formatScheduleDaySuccess(schedule))
}

func handleBulkScheduleDayCommand(ctx context.Context, store *repository.Store, cfg *appconfig.Config, event payload.CommandEvent, dates []string, statusStr string, meals []string, reason string) error {
	input := services.SetDayScheduleInput{
		DayStatus:      statusStr,
		AvailableMeals: meals,
		Reason:         reason,
		SetBy:          event.UserID,
	}

	result, err := services.BulkSetDaySchedule(ctx, store, dates, input)
	if err != nil {
		return sendReply(ctx, cfg, event, formatBulkScheduleDayError(err))
	}

	return sendReply(ctx, cfg, event, formatBulkScheduleDaySuccess(result, statusStr))
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
		fmt.Fprintf(&sb, "**Reason**: %s", schedule.Reason)
	}

	return sb.String()
}
