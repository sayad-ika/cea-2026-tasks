package main

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/sayad-ika/craftsbite/internal/cmdutil"
	appconfig "github.com/sayad-ika/craftsbite/internal/config"
	"github.com/sayad-ika/craftsbite/internal/dateutil"
	"github.com/sayad-ika/craftsbite/internal/gchat"
	"github.com/sayad-ika/craftsbite/internal/payload"
	"github.com/sayad-ika/craftsbite/internal/services"
)

func handleGChatInitInteraction(ctx context.Context, cfg *appconfig.Config, store initStore, dateParser *dateutil.DateParser, cutoff *services.CutoffChecker, event payload.CommandEvent) error {
	var opts payload.InitOptions
	if len(event.Options) > 0 {
		if err := event.ParseOptions(&opts); err != nil {
			return sendWarningReply(ctx, cfg, event, "Invalid `/init` interaction payload.")
		}
	}
	if opts.Action == "" {
		opts.Action = initActionOpen
	}
	slog.Info("gchat init handling started", "action", opts.Action, "date_present", opts.Date != "", "dates_count", len(opts.Dates), "meals_count", len(opts.Meals), "location_present", opts.Location != "")

	if opts.Action == initActionCancel {
		slog.Info("gchat init canceled")
		return sendReply(ctx, cfg, event, "Setup canceled.")
	}
	if opts.Action == initActionApply {
		return saveGChatInitCard(ctx, cfg, store, dateParser, cutoff, event, opts)
	}

	anchorDate, err := dateParser.ParseDateWithDefaults(opts.Date)
	if err != nil {
		slog.Warn("gchat init date parse failed", "error", err, "date_present", opts.Date != "")
		return sendWarningReply(ctx, cfg, event, fmt.Sprintf("Invalid init date: %v", err))
	}
	slog.Info("gchat init open date resolved", "anchor_date", anchorDate)

	availableDates, err := loadInitAvailableDates(ctx, store, anchorDate)
	if err != nil {
		slog.Error("gchat init available dates load failed", "error", err, "anchor_date", anchorDate)
		return sendWarningReply(ctx, cfg, event, "Unable to load upcoming meal days. Please try again shortly.")
	}
	if len(availableDates) == 0 {
		slog.Warn("gchat init no available dates", "anchor_date", anchorDate)
		return sendWarningReply(ctx, cfg, event, "No configured meal days were found in the next 15 days.")
	}

	selectedDates := normalizeInitDates(opts.Dates, anchorDate, availableDates)
	state, err := loadInitState(ctx, store, event.UserID, selectedDates[0])
	if err != nil {
		slog.Error("gchat init state load failed", "error", err, "anchor_date", anchorDate, "selected_dates_count", len(selectedDates))
		return sendWarningReply(ctx, cfg, event, "Unable to load your current setup. Please try again shortly.")
	}
	state.AvailableDates = availableDates
	state.AvailableMeals = commonInitMeals(availableDates, selectedDates)
	slog.Info("gchat init open state ready", "anchor_date", anchorDate, "available_dates_count", len(availableDates), "selected_dates_count", len(selectedDates), "available_meals_count", len(state.AvailableMeals), "current_meals_count", len(selectedMealsFromStatuses(state.Statuses)), "location", state.Location)

	draft := initDraftFromState(state)
	draft.Dates = selectedDates
	draft.Anchor = anchorDate
	draft.Saved = true

	return sendGChatInitCard(ctx, cfg, event, gchatInitCardInput(anchorDate, state, draft))
}

func saveGChatInitCard(ctx context.Context, cfg *appconfig.Config, store initStore, dateParser *dateutil.DateParser, cutoff *services.CutoffChecker, event payload.CommandEvent, opts payload.InitOptions) error {
	slog.Info("gchat init save started", "date_present", opts.Date != "", "dates_count", len(opts.Dates), "meals_count", len(opts.Meals), "location", opts.Location)
	if len(opts.Dates) == 0 {
		slog.Warn("gchat init save rejected: no dates")
		return sendWarningReply(ctx, cfg, event, "Choose at least one date before saving.")
	}
	if opts.Location != "office" && opts.Location != "wfh" {
		slog.Warn("gchat init save rejected: invalid location", "location", opts.Location)
		return sendWarningReply(ctx, cfg, event, "Choose either Office or WFH before saving.")
	}

	anchorDate, err := dateParser.ParseDateWithDefaults(opts.Date)
	if err != nil {
		slog.Warn("gchat init save date parse failed", "error", err, "date_present", opts.Date != "")
		return sendWarningReply(ctx, cfg, event, fmt.Sprintf("Invalid init date: %v", err))
	}

	availableDates, err := loadInitAvailableDates(ctx, store, anchorDate)
	if err != nil {
		slog.Error("gchat init save available dates load failed", "error", err, "anchor_date", anchorDate)
		return sendWarningReply(ctx, cfg, event, "Unable to load upcoming meal days. Please try again shortly.")
	}
	targetDates := filterRequestedInitDates(opts.Dates, availableDates)
	if len(targetDates) == 0 {
		slog.Warn("gchat init save rejected: no available requested dates", "requested_dates_count", len(opts.Dates), "available_dates_count", len(availableDates), "anchor_date", anchorDate)
		return sendWarningReply(ctx, cfg, event, "Choose at least one available date before saving.")
	}

	selectedMeals := filterInitMeals(opts.Meals, commonInitMeals(availableDates, targetDates))
	slog.Info("gchat init save normalized", "anchor_date", anchorDate, "available_dates_count", len(availableDates), "target_dates_count", len(targetDates), "selected_meals_count", len(selectedMeals), "location", opts.Location)
	failures := 0
	for _, date := range targetDates {
		if err := applyInitMealSelection(ctx, store, cutoff, event.UserID, date, selectedMeals); err != nil {
			slog.Warn("gchat init save meal update failed", "error", err, "date", date)
			failures++
			continue
		}
		if _, err := services.SetLocation(ctx, store, event.UserID, date, opts.Location, cutoff); err != nil {
			slog.Warn("gchat init save location update failed", "error", err, "date", date, "location", opts.Location)
			failures++
		}
	}

	if failures == len(targetDates) {
		slog.Warn("gchat init save all targets failed", "target_dates_count", len(targetDates), "failures", failures)
		return sendWarningReply(ctx, cfg, event, "No selected dates could be saved. Please review cutoff and availability.")
	}

	saved := len(targetDates) - failures
	note := fmt.Sprintf("Saved setup for %d date(s).", saved)
	if failures > 0 {
		note = fmt.Sprintf("Saved setup for %d date(s). %d date(s) could not be updated.", saved, failures)
	}
	if len(selectedMeals) == 0 && failures == 0 {
		note = fmt.Sprintf("Saved location for %d date(s). No meals selected.", saved)
	}
	slog.Info("gchat init save completed", "saved_dates_count", saved, "failures", failures, "selected_meals_count", len(selectedMeals))
	return sendReply(ctx, cfg, event, note)
}

func gchatInitCardInput(anchorDate string, state initState, draft initDraft) gchat.InitCardInput {
	return gchat.InitCardInput{
		Title:      "CraftsBite Setup",
		Subtitle:   "Minimal setup for upcoming meal days",
		Intro:      "<b>Set it once.</b><br>Choose dates, your work location, and the meals you want included.",
		AnchorDate: anchorDate,
		Dates:      gchatInitDateItems(state.AvailableDates, draft.Dates),
		Locations:  gchatInitLocationItems(draft.Location),
		Meals:      gchatInitMealItems(state.AvailableMeals, draft.Meals),
		SummaryRows: []gchat.TeamRow{
			{Label: "Current location", Value: displayLocationLabel(draft.Location)},
			{Label: "Selected meals", Value: gchatInitMealSummary(draft.Meals)},
		},
	}
}

func gchatInitDateItems(availableDates []initAvailableDate, selectedDates []string) []gchat.SelectionItem {
	selected := stringSet(selectedDates)
	items := make([]gchat.SelectionItem, 0, len(availableDates))
	for _, available := range availableDates {
		items = append(items, gchat.SelectionItem{
			Text:     displayInitDate(available.Date),
			Value:    available.Date,
			Selected: setContains(selected, available.Date),
		})
	}
	return items
}

func gchatInitLocationItems(selectedLocation string) []gchat.SelectionItem {
	if selectedLocation != "wfh" {
		selectedLocation = "office"
	}
	return []gchat.SelectionItem{
		{Text: "Office", Value: "office", Selected: selectedLocation == "office"},
		{Text: "WFH", Value: "wfh", Selected: selectedLocation == "wfh"},
	}
}

func gchatInitMealItems(availableMeals []string, selectedMeals []string) []gchat.SelectionItem {
	selected := stringSet(selectedMeals)
	items := make([]gchat.SelectionItem, 0, len(availableMeals))
	for _, meal := range availableMeals {
		items = append(items, gchat.SelectionItem{
			Text:     cmdutil.DisplayMealName(meal),
			Value:    meal,
			Selected: setContains(selected, meal),
		})
	}
	return items
}

func gchatInitMealSummary(meals []string) string {
	if len(meals) == 0 {
		return "No meals selected"
	}
	return displayMealList(meals)
}

func filterRequestedInitDates(requested []string, availableDates []initAvailableDate) []string {
	available := make(map[string]struct{}, len(availableDates))
	for _, date := range availableDates {
		available[date.Date] = struct{}{}
	}

	result := make([]string, 0, len(requested))
	seen := make(map[string]struct{}, len(requested))
	for _, date := range requested {
		if _, ok := available[date]; !ok {
			continue
		}
		if _, ok := seen[date]; ok {
			continue
		}
		seen[date] = struct{}{}
		result = append(result, date)
	}
	return result
}

func stringSet(values []string) map[string]struct{} {
	set := make(map[string]struct{}, len(values))
	for _, value := range values {
		set[value] = struct{}{}
	}
	return set
}

func setContains(set map[string]struct{}, value string) bool {
	_, ok := set[value]
	return ok
}

func sendGChatInitCard(ctx context.Context, cfg *appconfig.Config, event payload.CommandEvent, input gchat.InitCardInput) error {
	if input.ActionFunction == "" {
		input.ActionFunction = gchatActionFunctionFromContext(ctx)
	}
	body, err := gchat.InitCardResponse(input)
	if err != nil {
		slog.Error("gchat init card build failed", "error", err)
		return err
	}
	slog.Info("gchat init card built", "dates_count", len(input.Dates), "locations_count", len(input.Locations), "meals_count", len(input.Meals), "summary_rows_count", len(input.SummaryRows), "body_len", len(body), "action_function_present", input.ActionFunction != "")
	return sendGChatCard(ctx, cfg, event, body)
}

func gchatActionFunctionFromContext(ctx context.Context) string {
	recorder := recorderFromContext(ctx)
	if recorder == nil {
		return ""
	}
	return gchatActionFunctionURL(recorder.req.Raw)
}

func gchatActionFunctionURL(req events.APIGatewayV2HTTPRequest) string {
	host := firstHeaderValue(req.Headers, "x-forwarded-host")
	if host == "" {
		host = firstHeaderValue(req.Headers, "host")
	}
	if host == "" {
		return ""
	}
	scheme := firstHeaderValue(req.Headers, "x-forwarded-proto")
	if scheme == "" {
		scheme = "https"
	}
	path := req.RawPath
	if path == "" {
		path = req.RequestContext.HTTP.Path
	}
	if path == "" {
		path = "/gchat"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return strings.TrimRight(scheme, ":/") + "://" + host + path
}

func firstHeaderValue(headers map[string]string, name string) string {
	value := getHeader(headers, name)
	if value == "" {
		return ""
	}
	first, _, _ := strings.Cut(value, ",")
	return strings.TrimSpace(first)
}
