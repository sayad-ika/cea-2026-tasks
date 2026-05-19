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
	"github.com/sayad-ika/craftsbite/internal/discord"
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

	if opts.Action == initActionCancel {
		return sendGChatInitCanceledCard(ctx, cfg, dateParser, event, opts)
	}
	if opts.Action == initActionApply {
		return saveGChatInitCard(ctx, cfg, store, dateParser, cutoff, event, opts)
	}

	anchorDate, err := dateParser.ParseDateWithDefaults(opts.Date)
	if err != nil {
		slog.Warn("gchat init date parse failed", "error", err, "date_present", opts.Date != "")
		return sendWarningReply(ctx, cfg, event, fmt.Sprintf("Invalid init date: %v", err))
	}

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

	draft := initDraftFromState(state)
	draft.Dates = selectedDates
	draft.Anchor = anchorDate
	draft.Saved = true

	return sendGChatInitCard(ctx, cfg, event, gchatInitCardInput(anchorDate, state, draft, ""))
}

func saveGChatInitCard(ctx context.Context, cfg *appconfig.Config, store initStore, dateParser *dateutil.DateParser, cutoff *services.CutoffChecker, event payload.CommandEvent, opts payload.InitOptions) error {
	if len(opts.Dates) == 0 {
		slog.Warn("gchat init save rejected: no dates")
		return sendGChatInitValidationCard(ctx, cfg, store, dateParser, event, opts, "<b>Review needed</b><br>Choose at least one meal day before saving.")
	}
	if opts.Location != "office" && opts.Location != "wfh" {
		slog.Warn("gchat init save rejected: invalid location", "location", opts.Location)
		return sendGChatInitValidationCard(ctx, cfg, store, dateParser, event, opts, "<b>Review needed</b><br>Choose either Office or WFH before saving.")
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
		return sendGChatInitValidationCard(ctx, cfg, store, dateParser, event, opts, "<b>Review needed</b><br>Choose at least one available meal day before saving.")
	}

	selectedMeals := filterInitMeals(opts.Meals, commonInitMeals(availableDates, targetDates))
	failures := 0
	savedDates := make([]string, 0, len(targetDates))
	for _, date := range targetDates {
		if err := applyInitMealSelection(ctx, store, cutoff, event.UserID, date, selectedMeals); err != nil {
			slog.Warn("gchat init save meal update failed", "error", err, "date", date)
			failures++
			continue
		}
		if _, err := services.SetLocation(ctx, store, event.UserID, date, opts.Location, cutoff); err != nil {
			slog.Warn("gchat init save location update failed", "error", err, "date", date, "location", opts.Location)
			failures++
			continue
		}
		savedDates = append(savedDates, date)
	}

	if failures == len(targetDates) {
		slog.Warn("gchat init save all targets failed", "target_dates_count", len(targetDates), "failures", failures)
		return sendGChatInitResultCard(ctx, cfg, event, anchorDate, gchatInitSaveFailureSummary(), gchat.InitSaveConfirmationInput{
			Title:          "Setup not saved",
			Subtitle:       "No selected dates were updated",
			EditButtonText: "Edit setup",
			Tone:           discord.NoticeToneWarning,
		})
	}

	note := gchatInitSaveSummary(savedDates, len(targetDates), selectedMeals, opts.Location, failures)
	return sendGChatInitSavedCard(ctx, cfg, event, anchorDate, note)
}

func sendGChatInitValidationCard(ctx context.Context, cfg *appconfig.Config, store initStore, dateParser *dateutil.DateParser, event payload.CommandEvent, opts payload.InitOptions, note string) error {
	anchorDate, err := dateParser.ParseDateWithDefaults(opts.Date)
	if err != nil {
		slog.Warn("gchat init validation date parse failed", "error", err, "date_present", opts.Date != "")
		return sendWarningReply(ctx, cfg, event, fmt.Sprintf("Invalid init date: %v", err))
	}

	availableDates, err := loadInitAvailableDates(ctx, store, anchorDate)
	if err != nil {
		slog.Error("gchat init validation available dates load failed", "error", err, "anchor_date", anchorDate)
		return sendWarningReply(ctx, cfg, event, "Unable to load upcoming meal days. Please try again shortly.")
	}
	if len(availableDates) == 0 {
		slog.Warn("gchat init validation no available dates", "anchor_date", anchorDate)
		return sendWarningReply(ctx, cfg, event, "No configured meal days were found in the next 15 days.")
	}

	selectedDates := filterRequestedInitDates(opts.Dates, availableDates)
	stateDate := availableDates[0].Date
	if len(selectedDates) > 0 {
		stateDate = selectedDates[0]
	}
	state, err := loadInitState(ctx, store, event.UserID, stateDate)
	if err != nil {
		slog.Error("gchat init validation state load failed", "error", err, "anchor_date", anchorDate)
		return sendWarningReply(ctx, cfg, event, "Unable to load your current setup. Please try again shortly.")
	}
	state.AvailableDates = availableDates
	mealDates := selectedDates
	if len(mealDates) == 0 {
		mealDates = []string{stateDate}
	}
	state.AvailableMeals = commonInitMeals(availableDates, mealDates)

	draft := initDraft{
		Location: opts.Location,
		Meals:    filterInitMeals(opts.Meals, state.AvailableMeals),
		Dates:    selectedDates,
		Anchor:   anchorDate,
	}
	return sendGChatInitCardUpdate(ctx, cfg, event, gchatInitCardInput(anchorDate, state, draft, note))
}

func gchatInitSaveFailureSummary() string {
	return "<b>Setup not saved</b><br>No selected dates could be updated.<br>Please review cutoff and availability, then try again."
}

func gchatInitSaveSummary(savedDates []string, targetCount int, selectedMeals []string, location string, failures int) string {
	saved := len(savedDates)
	result := fmt.Sprintf("Updated %d %s.", saved, pluralize(saved, "meal day", "meal days"))
	if failures > 0 {
		result = fmt.Sprintf("Updated %d of %d selected %s.", saved, targetCount, pluralize(targetCount, "meal day", "meal days"))
	}

	dateLabel := "Date"
	if len(savedDates) != 1 {
		dateLabel = "Dates"
	}
	meals := "No meals included"
	if len(selectedMeals) > 0 {
		meals = displayMealList(selectedMeals)
	}

	lines := []string{
		"<b>Setup saved</b>",
		result,
		fmt.Sprintf("<b>%s:</b> %s", dateLabel, gchatInitInlineDates(savedDates)),
		fmt.Sprintf("<b>Work location:</b> %s", displayLocationLabel(location)),
		fmt.Sprintf("<b>Meals included:</b> %s", meals),
	}
	if failures > 0 {
		lines = append(lines, fmt.Sprintf("%d selected date(s) could not be updated.", failures))
	}
	return strings.Join(lines, "<br>")
}

func pluralize(count int, singular, plural string) string {
	if count == 1 {
		return singular
	}
	return plural
}

func gchatInitInlineDates(dates []string) string {
	if len(dates) == 0 {
		return "None"
	}
	labels := make([]string, 0, len(dates))
	for _, date := range dates {
		labels = append(labels, displayInitDate(date))
	}
	return strings.Join(labels, "; ")
}

func gchatInitCardInput(anchorDate string, state initState, draft initDraft, note string) gchat.InitCardInput {
	return gchat.InitCardInput{
		Title:      "CraftsBite Setup",
		Subtitle:   "Minimal setup for upcoming meal days",
		Intro:      "<b>Set it once.</b><br>Choose dates, your work location, and the meals you want included.",
		Note:       note,
		AnchorDate: anchorDate,
		Dates:      gchatInitDateItems(state.AvailableDates, draft.Dates),
		Locations:  gchatInitLocationItems(draft.Location),
		Meals:      gchatInitMealItems(state.AvailableMeals, draft.Meals),
		SummaryRows: []gchat.TeamRow{
			{Label: "Selected dates", Value: initDateSummary(draft.Dates)},
			{Label: "Will save location", Value: displayLocationLabel(draft.Location)},
			{Label: "Meals included", Value: gchatInitMealSummary(draft.Meals)},
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
		return "No meals included"
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
	return sendGChatCard(ctx, cfg, event, body)
}

func sendGChatInitCardUpdate(ctx context.Context, cfg *appconfig.Config, event payload.CommandEvent, input gchat.InitCardInput) error {
	if input.ActionFunction == "" {
		input.ActionFunction = gchatActionFunctionFromContext(ctx)
	}
	body, err := gchat.InitCardResponse(input)
	if err != nil {
		slog.Error("gchat init card build failed", "error", err)
		return err
	}
	return sendGChatCardUpdate(ctx, cfg, event, body)

}

func sendGChatInitSavedCard(ctx context.Context, cfg *appconfig.Config, event payload.CommandEvent, anchorDate, summary string) error {
	return sendGChatInitResultCard(ctx, cfg, event, anchorDate, summary, gchat.InitSaveConfirmationInput{})
}

func sendGChatInitCanceledCard(ctx context.Context, cfg *appconfig.Config, dateParser *dateutil.DateParser, event payload.CommandEvent, opts payload.InitOptions) error {
	anchorDate, err := dateParser.ParseDateWithDefaults(opts.Date)
	if err != nil {
		slog.Warn("gchat init cancel date parse failed", "error", err, "date_present", opts.Date != "")
		return sendWarningReply(ctx, cfg, event, fmt.Sprintf("Invalid init date: %v", err))
	}
	return sendGChatInitResultCard(ctx, cfg, event, anchorDate, "<b>Setup canceled</b><br>No changes were saved.", gchat.InitSaveConfirmationInput{
		Title:          "Setup canceled",
		Subtitle:       "No changes were saved",
		EditButtonText: "Open setup",
		Tone:           discord.NoticeToneInfo,
	})
}

func sendGChatInitResultCard(ctx context.Context, cfg *appconfig.Config, event payload.CommandEvent, anchorDate, summary string, input gchat.InitSaveConfirmationInput) error {
	input.Summary = summary
	input.AnchorDate = anchorDate
	input.ActionFunction = gchatActionFunctionFromContext(ctx)
	body, err := gchat.InitSaveConfirmationCardResponse(input)
	if err != nil {
		slog.Error("gchat init saved card build failed", "error", err)
		return err
	}
	return sendGChatCardUpdate(ctx, cfg, event, body)
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
