package main

import (
	"context"
	"fmt"
	"html"
	"log/slog"
	"strings"

	"github.com/sayad-ika/craftsbite/internal/cmdutil"
	appconfig "github.com/sayad-ika/craftsbite/internal/config"
	"github.com/sayad-ika/craftsbite/internal/dateutil"
	"github.com/sayad-ika/craftsbite/internal/discord"
	"github.com/sayad-ika/craftsbite/internal/gchat"
	"github.com/sayad-ika/craftsbite/internal/headcountreport"
	"github.com/sayad-ika/craftsbite/internal/payload"
	"github.com/sayad-ika/craftsbite/internal/repository"
	"github.com/sayad-ika/craftsbite/internal/services"
)

const (
	adminInitActionOpen           = "open"
	adminInitActionScheduleApply  = "schedule_apply"
	adminInitActionScheduleCancel = "schedule_cancel"
	adminInitActionScheduleEdit   = "schedule_edit"
	adminInitActionRangeToggle    = "schedule_range_toggle"
)

type adminInitScheduleDraft struct {
	Date     string
	EndDate  string
	UseRange bool
	Status   string
	Meals    []string
	Reason   string
}

func handleAdminInitCommand(ctx context.Context, store services.DayScheduleWriter, cfg *appconfig.Config, dateParser *dateutil.DateParser, event payload.CommandEvent) error {
	if event.Role != "admin" {
		return sendWarningReply(ctx, cfg, event, "You do not have permission to use `/admin-init`.")
	}
	if event.Source == "discord" {
		return handleDiscordAdminInitInteraction(ctx, cfg, store, dateParser, event)
	}
	if event.Source != "gchat" {
		return sendWarningReply(ctx, cfg, event, "`/admin-init` is currently available in Google Chat and Discord.")
	}

	var opts payload.AdminInitOptions
	if err := event.ParseOptions(&opts); err != nil {
		return sendWarningReply(ctx, cfg, event, "Invalid command options.")
	}
	if opts.Action == "" {
		opts.Action = adminInitActionOpen
	}

	switch opts.Action {
	case adminInitActionOpen:
		return openGChatAdminInitScheduleCard(ctx, cfg, store, dateParser, event, opts, false)
	case adminInitActionScheduleEdit:
		return openGChatAdminInitScheduleCard(ctx, cfg, store, dateParser, event, opts, true)
	case adminInitActionRangeToggle:
		return toggleGChatAdminInitScheduleRange(ctx, cfg, event, opts)
	case adminInitActionScheduleApply:
		return saveGChatAdminInitSchedule(ctx, cfg, store, dateParser, event, opts)
	case adminInitActionScheduleCancel:
		return sendGChatAdminInitCanceledCard(ctx, cfg, dateParser, event, opts)
	default:
		return sendWarningReply(ctx, cfg, event, "Unknown admin setup action.")
	}
}

func openGChatAdminInitScheduleCard(ctx context.Context, cfg *appconfig.Config, store services.DayScheduleWriter, dateParser *dateutil.DateParser, event payload.CommandEvent, opts payload.AdminInitOptions, update bool) error {
	date, err := dateParser.ParseDateWithDefaults(opts.Date)
	if err != nil {
		return sendWarningReply(ctx, cfg, event, fmt.Sprintf("Invalid date: %v", err))
	}

	schedule, err := store.GetDay(ctx, date)
	if err != nil {
		slog.Error("admin init schedule load failed", "error", err, "date", date)
		return sendErrorReply(ctx, cfg, event, "Unable to load the current schedule. Please try again shortly.")
	}

	draft := adminInitDraftFromSchedule(date, schedule)
	draft.UseRange = true
	if update {
		draft.UseRange = opts.UseRange
	}
	draft.EndDate = strings.TrimSpace(opts.EndDate)
	input := gchatAdminInitScheduleCardInput(draft, "")
	if update {
		return sendGChatAdminInitScheduleCardUpdate(ctx, cfg, event, input)
	}
	return sendGChatAdminInitScheduleCard(ctx, cfg, event, input)
}

func toggleGChatAdminInitScheduleRange(ctx context.Context, cfg *appconfig.Config, event payload.CommandEvent, opts payload.AdminInitOptions) error {
	draft := adminInitDraftFromOptions(opts)
	return sendGChatAdminInitScheduleCardUpdate(ctx, cfg, event, gchatAdminInitScheduleCardInput(draft, ""))
}

func saveGChatAdminInitSchedule(ctx context.Context, cfg *appconfig.Config, store services.DayScheduleWriter, dateParser *dateutil.DateParser, event payload.CommandEvent, opts payload.AdminInitOptions) error {
	draft := adminInitDraftFromOptions(opts)
	dates, err := adminInitTargetDates(dateParser, draft)
	if err != nil {
		return sendGChatAdminInitScheduleValidationCard(ctx, cfg, event, draft, adminInitReviewNote(fmt.Sprintf("Invalid date: %v", err)))
	}
	draft.Date = dates[0]
	if draft.UseRange {
		draft.EndDate = dates[len(dates)-1]
	}
	if draft.Status == "" {
		return sendGChatAdminInitScheduleValidationCard(ctx, cfg, event, draft, adminInitReviewNote("Choose a day status before saving."))
	}
	if adminInitStatusBlocksMeals(draft.Status) {
		draft.Meals = nil
	}

	input := services.SetDayScheduleInput{
		Date:           draft.Date,
		DayStatus:      draft.Status,
		AvailableMeals: draft.Meals,
		Reason:         draft.Reason,
		SetBy:          event.UserID,
	}
	if len(dates) > 1 {
		result, err := services.BulkSetDaySchedule(ctx, store, dates, input)
		if err != nil {
			return sendGChatAdminInitScheduleValidationCard(ctx, cfg, event, draft, adminInitReviewNote(formatBulkScheduleDayError(err)))
		}
		return sendGChatAdminInitBulkSavedCard(ctx, cfg, event, draft, result)
	}

	schedule, err := services.SetDaySchedule(ctx, store, input)
	if err != nil {
		return sendGChatAdminInitScheduleValidationCard(ctx, cfg, event, draft, adminInitReviewNote(formatScheduleDayError(err)))
	}

	return sendGChatAdminInitSavedCard(ctx, cfg, event, schedule)
}

func adminInitTargetDates(dateParser *dateutil.DateParser, draft adminInitScheduleDraft) ([]string, error) {
	if !draft.UseRange {
		date, err := dateParser.ParseDateWithDefaults(draft.Date)
		if err != nil {
			return nil, err
		}
		return []string{date}, nil
	}
	if strings.TrimSpace(draft.EndDate) == "" {
		return nil, fmt.Errorf("end date is required when date range is enabled")
	}
	return dateParser.ParseDateRange(strings.TrimSpace(draft.Date) + ".." + strings.TrimSpace(draft.EndDate))
}

func sendGChatAdminInitScheduleValidationCard(ctx context.Context, cfg *appconfig.Config, event payload.CommandEvent, draft adminInitScheduleDraft, note string) error {
	return sendGChatAdminInitScheduleCardUpdate(ctx, cfg, event, gchatAdminInitScheduleCardInput(draft, note))
}

func sendGChatAdminInitScheduleCard(ctx context.Context, cfg *appconfig.Config, event payload.CommandEvent, input gchat.AdminInitScheduleCardInput) error {
	if input.ActionFunction == "" {
		input.ActionFunction = gchatActionFunctionFromContext(ctx)
	}
	body, err := gchat.AdminInitScheduleCardResponse(input)
	if err != nil {
		slog.Error("admin init schedule card build failed", "error", err)
		return err
	}
	return sendGChatCard(ctx, cfg, event, body)
}

func sendGChatAdminInitScheduleCardUpdate(ctx context.Context, cfg *appconfig.Config, event payload.CommandEvent, input gchat.AdminInitScheduleCardInput) error {
	if input.ActionFunction == "" {
		input.ActionFunction = gchatActionFunctionFromContext(ctx)
	}
	body, err := gchat.AdminInitScheduleCardResponse(input)
	if err != nil {
		slog.Error("admin init schedule card build failed", "error", err)
		return err
	}
	return sendGChatCardUpdate(ctx, cfg, event, body)
}

func sendGChatAdminInitSavedCard(ctx context.Context, cfg *appconfig.Config, event payload.CommandEvent, schedule *repository.DaySchedule) error {
	summary := fmt.Sprintf(
		"<b>Schedule updated</b><br>Date: %s<br>Status: %s<br>Meals: %s%s",
		schedule.Date,
		headcountreport.DisplayDayStatus(schedule.DayStatus),
		adminInitMealSummary(schedule.AvailableMeals),
		adminInitReasonLine(schedule.Reason),
	)
	return sendGChatAdminInitResultCard(ctx, cfg, event, schedule.Date, summary, gchat.AdminInitScheduleConfirmationInput{})
}

func sendGChatAdminInitBulkSavedCard(ctx context.Context, cfg *appconfig.Config, event payload.CommandEvent, draft adminInitScheduleDraft, result *services.BulkSetDayScheduleResult) error {
	summary := fmt.Sprintf(
		"<b>Schedule updated</b><br>Status: %s<br>Meals: %s<br>Weekdays updated: %d<br>Affected dates: %s%s",
		headcountreport.DisplayDayStatus(draft.Status),
		adminInitMealSummary(draft.Meals),
		len(result.SuccessDates),
		html.EscapeString(strings.Join(result.SuccessDates, ", ")),
		adminInitReasonLine(draft.Reason),
	)
	return sendGChatAdminInitResultCard(ctx, cfg, event, draft.Date, summary, gchat.AdminInitScheduleConfirmationInput{
		Subtitle: draft.Date + " to " + draft.EndDate,
		EndDate:  draft.EndDate,
		UseRange: true,
	})
}

func sendGChatAdminInitCanceledCard(ctx context.Context, cfg *appconfig.Config, dateParser *dateutil.DateParser, event payload.CommandEvent, opts payload.AdminInitOptions) error {
	date, err := dateParser.ParseDateWithDefaults(opts.Date)
	if err != nil {
		return sendWarningReply(ctx, cfg, event, fmt.Sprintf("Invalid date: %v", err))
	}
	return sendGChatAdminInitResultCard(ctx, cfg, event, date, "<b>Schedule setup canceled</b><br>No changes were saved.", gchat.AdminInitScheduleConfirmationInput{
		Title:          "Setup canceled",
		Subtitle:       "No schedule changes were saved",
		EditButtonText: "Open schedule setup",
		EndDate:        strings.TrimSpace(opts.EndDate),
		UseRange:       opts.UseRange,
		Tone:           discord.NoticeToneInfo,
	})
}

func sendGChatAdminInitResultCard(ctx context.Context, cfg *appconfig.Config, event payload.CommandEvent, date, summary string, input gchat.AdminInitScheduleConfirmationInput) error {
	input.Summary = summary
	input.Date = date
	input.ActionFunction = gchatActionFunctionFromContext(ctx)
	body, err := gchat.AdminInitScheduleConfirmationCardResponse(input)
	if err != nil {
		slog.Error("admin init result card build failed", "error", err)
		return err
	}
	return sendGChatCardUpdate(ctx, cfg, event, body)
}

func adminInitDraftFromSchedule(date string, schedule *repository.DaySchedule) adminInitScheduleDraft {
	draft := adminInitScheduleDraft{Date: date, Status: string(repository.DayStatusNormal)}
	if schedule == nil {
		return draft
	}
	draft.Status = schedule.DayStatus
	draft.Meals = append([]string(nil), schedule.AvailableMeals...)
	draft.Reason = schedule.Reason
	return draft
}

func adminInitDraftFromOptions(opts payload.AdminInitOptions) adminInitScheduleDraft {
	return adminInitScheduleDraft{
		Date:     strings.TrimSpace(opts.Date),
		EndDate:  strings.TrimSpace(opts.EndDate),
		UseRange: opts.UseRange,
		Status:   strings.ToLower(strings.TrimSpace(opts.Status)),
		Meals:    normalizeAdminInitMeals(opts.Meals),
		Reason:   strings.TrimSpace(opts.Reason),
	}
}

func gchatAdminInitScheduleCardInput(draft adminInitScheduleDraft, note string) gchat.AdminInitScheduleCardInput {
	return gchat.AdminInitScheduleCardInput{
		Title:       "Admin Setup",
		Subtitle:    "Schedule a meal day",
		Intro:       "Configure a weekday date range by default. Select `Mark a single date` to update only one day.",
		Note:        note,
		Date:        draft.Date,
		EndDate:     draft.EndDate,
		UseRange:    draft.UseRange,
		Reason:      draft.Reason,
		Statuses:    adminInitStatusItems(draft.Status),
		Meals:       adminInitMealItems(draft.Meals),
		SummaryRows: adminInitSummaryRows(draft),
	}
}

func adminInitSummaryRows(draft adminInitScheduleDraft) []gchat.TeamRow {
	rows := []gchat.TeamRow{}
	if draft.UseRange {
		rows = append(rows,
			gchat.TeamRow{Label: "Start date", Value: adminInitEscapedValueOrNone(draft.Date)},
			gchat.TeamRow{Label: "End date", Value: adminInitEscapedValueOrNone(draft.EndDate)},
		)
	} else {
		rows = append(rows, gchat.TeamRow{Label: "Date", Value: adminInitEscapedValueOrNone(draft.Date)})
	}
	return append(rows,
		gchat.TeamRow{Label: "Day status", Value: adminInitDisplayStatus(draft.Status)},
		gchat.TeamRow{Label: "Meals", Value: adminInitMealSummary(draft.Meals)},
		gchat.TeamRow{Label: "Reason", Value: adminInitEscapedValueOrNone(draft.Reason)},
	)
}

func adminInitStatusItems(selected string) []gchat.SelectionItem {
	items := make([]gchat.SelectionItem, 0, len(repository.ValidDayStatuses()))
	for _, status := range repository.ValidDayStatuses() {
		items = append(items, gchat.SelectionItem{Text: headcountreport.DisplayDayStatus(status), Value: status, Selected: status == selected})
	}
	return items
}

func adminInitMealItems(selectedMeals []string) []gchat.SelectionItem {
	selected := make(map[string]struct{}, len(selectedMeals))
	for _, meal := range selectedMeals {
		selected[meal] = struct{}{}
	}
	items := make([]gchat.SelectionItem, 0, len(repository.ValidMealTypes()))
	for _, meal := range repository.ValidMealTypes() {
		_, ok := selected[meal]
		items = append(items, gchat.SelectionItem{Text: cmdutil.DisplayMealName(meal), Value: meal, Selected: ok})
	}
	return items
}

func normalizeAdminInitMeals(meals []string) []string {
	seen := make(map[string]struct{}, len(meals))
	result := make([]string, 0, len(meals))
	for _, meal := range meals {
		meal = strings.ToLower(strings.TrimSpace(meal))
		if meal == "" {
			continue
		}
		if _, ok := seen[meal]; ok {
			continue
		}
		seen[meal] = struct{}{}
		result = append(result, meal)
	}
	return result
}

func adminInitStatusBlocksMeals(status string) bool {
	switch status {
	case string(repository.DayStatusOfficeClosed), string(repository.DayStatusGovtHoliday):
		return true
	default:
		return false
	}
}

func adminInitMealSummary(meals []string) string {
	if len(meals) == 0 {
		return "None"
	}
	labels := make([]string, 0, len(meals))
	for _, meal := range meals {
		labels = append(labels, html.EscapeString(cmdutil.DisplayMealName(meal)))
	}
	return strings.Join(labels, ", ")
}

func adminInitDisplayStatus(status string) string {
	if strings.TrimSpace(status) == "" {
		return "None"
	}
	if repository.IsValidDayStatus(status) {
		return headcountreport.DisplayDayStatus(status)
	}
	return html.EscapeString(status)
}

func adminInitReviewNote(message string) string {
	return "<b>Review needed</b><br>" + html.EscapeString(message)
}

func adminInitReasonLine(reason string) string {
	if strings.TrimSpace(reason) == "" {
		return ""
	}
	return "<br>Reason: " + html.EscapeString(reason)
}

func adminInitEscapedValueOrNone(value string) string {
	if strings.TrimSpace(value) == "" {
		return "None"
	}
	return html.EscapeString(value)
}
