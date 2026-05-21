package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/sayad-ika/craftsbite/internal/cmdutil"
	appconfig "github.com/sayad-ika/craftsbite/internal/config"
	"github.com/sayad-ika/craftsbite/internal/dateutil"
	"github.com/sayad-ika/craftsbite/internal/discord"
	"github.com/sayad-ika/craftsbite/internal/payload"
	"github.com/sayad-ika/craftsbite/internal/services"
)

const (
	initActionOpen     = "open"
	initActionRefresh  = "refresh"
	initActionLocation = "location"
	initActionMeals    = "meals"
	initActionDate     = "date"
	initActionApply    = "apply"
	initActionCancel   = "cancel"

	initActionLocationButton = "location:"
	initActionMealButton     = "meal:"
	initActionDateButton     = "date:"

	initCustomIDPrefix = "i"
	initCustomIDSep    = "|"
	initPanelTitle     = "Setup"
	initPanelSubtitle  = "Choose dates, location, and meals. Then save."

	initConfiguredDateLimit = 7
	initConfiguredDateScan  = 15
)

type initStore interface {
	services.DayScheduleReader
	services.ParticipationWriter
	services.LocationWriter
}

type initState struct {
	Location       string
	Statuses       []services.ResolvedStatus
	AvailableMeals []string
	AvailableDates []initAvailableDate
}

type initDraft struct {
	Location string
	Meals    []string
	Dates    []string
	Anchor   string
	Saved    bool
}

type initAvailableDate struct {
	Date  string
	Meals []string
}

func handleInitCommand(ctx context.Context, deps handlerDeps, event payload.CommandEvent) error {
	return handleInitInteraction(ctx, deps.cfg, deps.store, deps.dateParser, deps.cutoff, event)
}

func handleInitInteraction(ctx context.Context, cfg *appconfig.Config, store initStore, dateParser *dateutil.DateParser, cutoff *services.CutoffChecker, event payload.CommandEvent) error {
	if event.Source == "gchat" {
		return handleGChatInitInteraction(ctx, cfg, store, dateParser, cutoff, event)
	}
	if event.Source != "discord" {
		return sendWarningReply(ctx, cfg, event, "`/init` is currently available only in Discord.")
	}

	var opts payload.InitOptions
	if len(event.Options) > 0 {
		if err := event.ParseOptions(&opts); err != nil {
			return sendInitPanelError(ctx, "Invalid `/init` interaction payload.", false)
		}
	}
	if opts.Action == "" {
		opts.Action = initActionOpen
	}

	anchorDate, err := dateParser.ParseDateWithDefaults(opts.Date)
	if err != nil {
		return sendInitPanelError(ctx, fmt.Sprintf("Invalid init date: %v", err), opts.Action != initActionOpen)
	}

	availableDates, err := loadInitAvailableDates(ctx, store, anchorDate)
	if err != nil {
		return sendInitPanelError(ctx, "Unable to load upcoming meal days. Please try again shortly.", opts.Action != initActionOpen)
	}
	if len(availableDates) == 0 {
		return sendInitPanelError(ctx, "No configured meal days were found in the next 15 days.", opts.Action != initActionOpen)
	}

	selectedDates := normalizeInitDates(opts.Dates, anchorDate, availableDates)
	date := selectedDates[0]
	state, err := loadInitState(ctx, store, event.UserID, date)
	if err != nil {
		return sendInitPanelError(ctx, "Unable to load your current setup. Please try again shortly.", opts.Action != initActionOpen)
	}
	state.AvailableDates = availableDates
	state.AvailableMeals = commonInitMeals(availableDates, selectedDates)

	savedDraft := initDraftFromState(state)
	savedDraft.Dates = selectedDates
	savedDraft.Anchor = anchorDate
	savedDraft.Saved = true

	switch opts.Action {
	case initActionOpen:
		return renderInitPanel(ctx, date, state, savedDraft, "", discord.NoticeToneInfo, false)
	case initActionRefresh:
		return renderInitPanel(ctx, date, state, savedDraft, "Reset to your saved setup.", discord.NoticeToneSuccess, true)
	case initActionDate:
		draft := normalizeInitDraft(opts, state, selectedDates, anchorDate)
		if len(draft.Dates) == 0 {
			draft.Dates = selectedDates
		}
		return renderInitPanel(ctx, draft.Dates[0], state, draft, "Dates updated. Save when ready.", discord.NoticeToneInfo, true)
	case initActionLocation:
		draft := normalizeInitDraft(opts, state, selectedDates, anchorDate)
		if draft.Location != "office" && draft.Location != "wfh" {
			return renderInitPanel(ctx, date, state, savedDraft, "Please choose either Office or WFH.", discord.NoticeToneWarning, true)
		}
		return renderInitPanel(ctx, date, state, draft, fmt.Sprintf("Location set to %s. Save when ready.", displayLocationLabel(draft.Location)), discord.NoticeToneInfo, true)
	case initActionMeals:
		draft := normalizeInitDraft(opts, state, selectedDates, anchorDate)
		return renderInitPanel(ctx, date, state, draft, "Meals updated. Save when ready.", discord.NoticeToneInfo, true)
	case initActionApply:
		draft := normalizeInitDraft(opts, state, selectedDates, anchorDate)
		return applyInitDraftAndRender(ctx, store, cutoff, event.UserID, date, state, draft)
	default:
		return renderInitPanel(ctx, date, state, savedDraft, "This `/init` action is not supported.", discord.NoticeToneWarning, true)
	}
}

func applyInitDraftAndRender(ctx context.Context, store initStore, cutoff *services.CutoffChecker, userID, date string, state initState, draft initDraft) error {
	if len(draft.Dates) == 0 {
		draft.Dates = []string{date}
	}

	if draft.Saved && !draftHasChanges(state, draft) {
		return renderInitPanel(ctx, date, state, draft, "No changes to apply.", discord.NoticeToneInfo, true)
	}

	failures := 0
	for _, targetDate := range draft.Dates {
		if err := applyInitMealSelection(ctx, store, cutoff, userID, targetDate, draft.Meals); err != nil {
			failures++
			continue
		}
		if _, err := services.SetLocation(ctx, store, userID, targetDate, draft.Location, cutoff); err != nil {
			failures++
		}
	}

	if failures == len(draft.Dates) {
		return renderInitPanel(ctx, date, state, draft, "No selected dates could be saved. Please review cutoff and availability.", discord.NoticeToneWarning, true)
	}

	refreshed, err := loadInitState(ctx, store, userID, draft.Dates[0])
	if err != nil {
		return sendInitPanelError(ctx, "Changes were saved, but the panel could not be refreshed. Please run `/init` again.", true)
	}
	refreshed.AvailableDates = state.AvailableDates
	refreshed.AvailableMeals = commonInitMeals(state.AvailableDates, draft.Dates)

	saved := len(draft.Dates) - failures
	note := fmt.Sprintf("Saved for %d date(s).", saved)
	if failures > 0 {
		note = fmt.Sprintf("Saved for %d date(s). %d date(s) could not be updated.", saved, failures)
	}
	if len(draft.Meals) == 0 && failures == 0 {
		note = "Saved. No meals selected."
	}
	savedDraft := initDraftFromState(refreshed)
	savedDraft.Dates = draft.Dates
	savedDraft.Anchor = draft.Anchor
	savedDraft.Saved = failures == 0
	return renderInitPanel(ctx, draft.Dates[0], refreshed, savedDraft, note, discord.NoticeToneSuccess, true)
}

func draftHasChanges(state initState, draft initDraft) bool {
	return draft.Location != state.Location || !sameMealSelection(draft.Meals, initDraftFromState(state).Meals)
}

func renderInitPanel(ctx context.Context, date string, state initState, draft initDraft, note string, tone discord.NoticeTone, update bool) error {
	message := buildDiscordInitMessage(date, draft.Location, draftStatuses(state, draft), state.AvailableMeals, state.AvailableDates, draft, note, tone)
	if update {
		return sendDiscordInteractionResponse(ctx, updateMessage(message))
	}
	return sendDiscordInteractionResponse(ctx, ephemeralMessage(message))
}

func sendInitPanelError(ctx context.Context, text string, update bool) error {
	message := discord.ToneMessage(discord.DefaultNoticeTitle(discord.NoticeToneError), text, discord.NoticeToneError)
	if update {
		return sendDiscordInteractionResponse(ctx, updateMessage(message))
	}
	return sendDiscordInteractionResponse(ctx, ephemeralMessage(message))
}

func loadInitState(ctx context.Context, store initStore, userID, date string) (initState, error) {
	location, err := services.GetLocation(ctx, store, userID, date)
	if err != nil {
		return initState{}, err
	}

	statuses, err := services.GetUserMealStatus(ctx, store, store, userID, date)
	if err != nil {
		return initState{}, err
	}

	availableMeals, err := store.GetAvailableMeals(ctx, date)
	if err != nil {
		return initState{}, err
	}
	if len(availableMeals) == 0 {
		availableMeals = mealTypesFromStatuses(statuses)
	}

	return initState{
		Location:       initLocationValue(location.Location),
		Statuses:       statuses,
		AvailableMeals: availableMeals,
	}, nil
}

func loadInitAvailableDates(ctx context.Context, store initStore, startDate string) ([]initAvailableDate, error) {
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return nil, err
	}

	dates := make([]initAvailableDate, 0, initConfiguredDateLimit)
	for offset := 0; offset < initConfiguredDateScan && len(dates) < initConfiguredDateLimit; offset++ {
		date := start.AddDate(0, 0, offset).Format("2006-01-02")
		schedule, err := store.GetDay(ctx, date)
		if err != nil {
			return nil, err
		}
		if schedule != nil && (schedule.DayStatus == "office_closed" || schedule.DayStatus == "govt_holiday") {
			continue
		}

		meals, err := store.GetAvailableMeals(ctx, date)
		if err != nil {
			return nil, err
		}
		if len(meals) == 0 {
			continue
		}
		dates = append(dates, initAvailableDate{Date: date, Meals: meals})
	}
	return dates, nil
}

func commonInitMeals(availableDates []initAvailableDate, selectedDates []string) []string {
	if len(selectedDates) == 0 {
		return nil
	}
	selected := make(map[string]struct{}, len(selectedDates))
	for _, date := range selectedDates {
		selected[date] = struct{}{}
	}

	var common map[string]struct{}
	var order []string
	for _, available := range availableDates {
		if _, ok := selected[available.Date]; !ok {
			continue
		}
		mealSet := make(map[string]struct{}, len(available.Meals))
		for _, meal := range available.Meals {
			mealSet[meal] = struct{}{}
		}
		if common == nil {
			common = mealSet
			order = append(order, available.Meals...)
			continue
		}
		for meal := range common {
			if _, ok := mealSet[meal]; !ok {
				delete(common, meal)
			}
		}
	}

	meals := make([]string, 0, len(common))
	for _, meal := range order {
		if _, ok := common[meal]; ok {
			meals = append(meals, meal)
		}
	}
	return meals
}

func applyInitMealSelection(ctx context.Context, store initStore, cutoff *services.CutoffChecker, userID, date string, selectedMeals []string) error {
	availableMeals, err := store.GetAvailableMeals(ctx, date)
	if err != nil {
		return err
	}
	if len(availableMeals) == 0 {
		statuses, statusErr := services.GetUserMealStatus(ctx, store, store, userID, date)
		if statusErr != nil {
			return statusErr
		}
		availableMeals = mealTypesFromStatuses(statuses)
	}
	if len(availableMeals) == 0 {
		return services.ErrNoMeals
	}

	allowed := make(map[string]struct{}, len(availableMeals))
	for _, meal := range availableMeals {
		allowed[meal] = struct{}{}
	}

	selected := make(map[string]struct{}, len(selectedMeals))
	for _, meal := range selectedMeals {
		if _, ok := allowed[meal]; !ok {
			return services.ErrMealUnavailable
		}
		selected[meal] = struct{}{}
	}

	for _, meal := range availableMeals {
		_, include := selected[meal]
		if _, err := services.UpdateParticipation(ctx, store, store, userID, date, meal, include, cutoff); err != nil {
			return err
		}
	}

	return nil
}

func buildDiscordInitMessage(date, location string, statuses []services.ResolvedStatus, availableMeals []string, availableDates []initAvailableDate, draft initDraft, note string, tone discord.NoticeTone) discord.Message {
	fields := []discord.EmbedField{
		{Name: "Dates", Value: initDateSummary(draft.Dates)},
		{Name: "Location", Value: displayLocationLabel(location), Inline: true},
		{Name: "Meals", Value: initMealSummary(statuses)},
	}

	description := initPanelDescription(date, note)
	embed := discord.BrandEmbed(initPanelTitle, description, fields)
	if note != "" {
		embed = discord.ToneEmbed(initPanelTitle, description, fields, tone)
	}

	message := discord.EmbedMessage(embed)
	message.Components = initMessageComponents(date, location, statuses, availableMeals, availableDates, draft)
	return message
}

func initMessageComponents(date, location string, statuses []services.ResolvedStatus, availableMeals []string, availableDates []initAvailableDate, draft initDraft) []discord.Component {
	draft.Location = location
	draft.Meals = selectedMealsFromStatuses(statuses)
	components := buildInitDateRows(date, availableDates, draft)
	components = append(components, discord.Component{
		Type:       discord.ComponentTypeActionRow,
		Components: buildInitLocationButtons(date, draft),
	})

	mealButtons := buildInitMealButtons(date, statuses, availableMeals, draft)
	if len(mealButtons) > 0 {
		components = append(components, discord.Component{
			Type:       discord.ComponentTypeActionRow,
			Components: mealButtons,
		})
	}

	components = append(components, discord.Component{
		Type: discord.ComponentTypeActionRow,
		Components: []discord.Component{
			{
				Type:     discord.ComponentTypeButton,
				Style:    discord.ButtonStyleSuccess,
				Label:    "Save",
				CustomID: initCustomID(initActionApply, date, draft),
			},
			{
				Type:     discord.ComponentTypeButton,
				Style:    discord.ButtonStyleSecondary,
				Label:    "Reset",
				CustomID: initCustomID(initActionRefresh, date, draft),
			},
		},
	})

	return components
}

func buildInitDateRows(date string, availableDates []initAvailableDate, draft initDraft) []discord.Component {
	if len(availableDates) == 0 {
		return nil
	}
	selected := make(map[string]struct{}, len(draft.Dates))
	for _, date := range draft.Dates {
		selected[date] = struct{}{}
	}

	buttons := make([]discord.Component, 0, len(availableDates))
	for _, available := range availableDates {
		style := discord.ButtonStyleSecondary
		if _, ok := selected[available.Date]; ok {
			style = discord.ButtonStylePrimary
		}
		buttons = append(buttons, discord.Component{
			Type:     discord.ComponentTypeButton,
			Style:    style,
			Label:    displayInitDay(available.Date),
			CustomID: initCustomID(initActionDateButton+available.Date, date, draft),
		})
	}

	rows := make([]discord.Component, 0, 2)
	for len(buttons) > 0 {
		rowSize := len(buttons)
		if rowSize > 5 {
			rowSize = 5
		}
		rows = append(rows, discord.Component{
			Type:       discord.ComponentTypeActionRow,
			Components: buttons[:rowSize],
		})
		buttons = buttons[rowSize:]
	}
	return rows
}

func buildInitLocationButtons(date string, draft initDraft) []discord.Component {
	locations := []struct {
		label string
		value string
	}{
		{label: "Office", value: "office"},
		{label: "WFH", value: "wfh"},
	}

	buttons := make([]discord.Component, 0, len(locations))
	for _, location := range locations {
		style := discord.ButtonStyleSecondary
		if draft.Location == location.value || draft.Location == "" && location.value == "office" {
			style = discord.ButtonStylePrimary
		}
		buttons = append(buttons, discord.Component{
			Type:     discord.ComponentTypeButton,
			Style:    style,
			Label:    location.label,
			CustomID: initCustomID(initActionLocationButton+location.value, date, draft),
		})
	}
	return buttons
}

func buildInitMealButtons(date string, statuses []services.ResolvedStatus, availableMeals []string, draft initDraft) []discord.Component {
	if len(availableMeals) == 0 {
		return nil
	}

	statusByMeal := make(map[string]services.ResolvedStatus, len(statuses))
	for _, status := range statuses {
		statusByMeal[status.MealType] = status
	}

	selected := make(map[string]struct{}, len(draft.Meals))
	for _, meal := range draft.Meals {
		selected[meal] = struct{}{}
	}

	buttons := make([]discord.Component, 0, len(availableMeals))
	for _, meal := range availableMeals {
		status := statusByMeal[meal]
		style := discord.ButtonStyleSecondary
		if _, ok := selected[meal]; ok && status.Status != "unavailable" {
			style = discord.ButtonStylePrimary
		}
		buttons = append(buttons, discord.Component{
			Type:     discord.ComponentTypeButton,
			Style:    style,
			Label:    cmdutil.DisplayMealName(meal),
			CustomID: initCustomID(initActionMealButton+meal, date, draft),
			Disabled: status.Status == "unavailable",
		})
	}
	return buttons
}

func draftStatuses(state initState, draft initDraft) []services.ResolvedStatus {
	selected := make(map[string]struct{}, len(draft.Meals))
	for _, meal := range draft.Meals {
		selected[meal] = struct{}{}
	}

	statusByMeal := make(map[string]services.ResolvedStatus, len(state.Statuses))
	for _, status := range state.Statuses {
		statusByMeal[status.MealType] = status
	}

	statuses := make([]services.ResolvedStatus, 0, len(state.AvailableMeals))
	for _, meal := range state.AvailableMeals {
		base := statusByMeal[meal]
		if base.Status == "unavailable" {
			statuses = append(statuses, base)
			continue
		}
		status := services.ResolvedStatus{MealType: meal, Status: "opted_out", Source: base.Source}
		if _, ok := selected[meal]; ok {
			status.Status = "opted_in"
		}
		statuses = append(statuses, status)
	}
	return statuses
}

func initMealSummary(statuses []services.ResolvedStatus) string {
	if len(statuses) == 0 {
		return "No meals configured"
	}

	included := make([]string, 0, len(statuses))
	notIncluded := make([]string, 0, len(statuses))
	unavailable := make([]string, 0, len(statuses))
	for _, status := range statuses {
		label := cmdutil.DisplayMealName(status.MealType)
		switch status.Status {
		case "opted_in":
			included = append(included, label)
		case "unavailable":
			unavailable = append(unavailable, label)
		default:
			notIncluded = append(notIncluded, label)
		}
	}

	lines := make([]string, 0, 3)
	if len(included) > 0 {
		lines = append(lines, "Included: "+strings.Join(included, ", "))
	} else {
		lines = append(lines, "No meals selected")
	}
	if len(notIncluded) > 0 {
		lines = append(lines, "Not included: "+strings.Join(notIncluded, ", "))
	}
	if len(unavailable) > 0 {
		lines = append(lines, "Unavailable: "+strings.Join(unavailable, ", "))
	}
	return strings.Join(lines, "\n")
}

func initPanelDescription(date, note string) string {
	description := initPanelSubtitle + "\n" + displayInitDate(date)
	if note != "" {
		description = note + "\n" + description
	}
	return description
}

func initDateSummary(dates []string) string {
	if len(dates) == 0 {
		return "No dates selected"
	}
	labels := make([]string, 0, len(dates))
	for _, date := range dates {
		labels = append(labels, displayInitDate(date))
	}
	return strings.Join(labels, "\n")
}

func displayInitDate(date string) string {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return date
	}
	return t.Format("Mon, Jan 2")
}

func displayInitDay(date string) string {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return date
	}
	return t.Format("Mon 2")
}

func initDraftFromState(state initState) initDraft {
	return initDraft{
		Location: state.Location,
		Meals:    selectedMealsFromStatuses(state.Statuses),
	}
}

func selectedMealsFromStatuses(statuses []services.ResolvedStatus) []string {
	meals := make([]string, 0, len(statuses))
	for _, status := range statuses {
		if status.Status == "opted_in" {
			meals = append(meals, status.MealType)
		}
	}
	return meals
}

func initLocationValue(location string) string {
	if location == "wfh" {
		return "wfh"
	}
	return "office"
}

func normalizeInitDraft(opts payload.InitOptions, state initState, selectedDates []string, anchorDate string) initDraft {
	draft := initDraftFromState(state)
	draft.Dates = selectedDates
	draft.Anchor = anchorDate
	if opts.Location == "office" || opts.Location == "wfh" {
		draft.Location = opts.Location
	}
	if len(opts.Dates) > 0 {
		draft.Dates = normalizeInitDates(opts.Dates, anchorDate, state.AvailableDates)
	}
	if opts.Meals != nil {
		draft.Meals = filterInitMeals(opts.Meals, commonInitMeals(state.AvailableDates, draft.Dates))
	}
	return draft
}

func normalizeInitDates(requested []string, anchorDate string, availableDates []initAvailableDate) []string {
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
	if len(result) > 0 {
		return result
	}
	if _, ok := available[anchorDate]; ok {
		return []string{anchorDate}
	}
	return []string{availableDates[0].Date}
}

func filterInitMeals(selected []string, availableMeals []string) []string {
	if len(availableMeals) == 0 {
		return nil
	}
	allowed := make(map[string]struct{}, len(availableMeals))
	for _, meal := range availableMeals {
		allowed[meal] = struct{}{}
	}

	result := make([]string, 0, len(selected))
	seen := make(map[string]struct{}, len(selected))
	for _, meal := range selected {
		if _, ok := allowed[meal]; !ok {
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

func sameMealSelection(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	setA := make(map[string]struct{}, len(a))
	for _, meal := range a {
		setA[meal] = struct{}{}
	}
	for _, meal := range b {
		if _, ok := setA[meal]; !ok {
			return false
		}
	}
	return true
}

func mealTypesFromStatuses(statuses []services.ResolvedStatus) []string {
	meals := make([]string, 0, len(statuses))
	for _, status := range statuses {
		meals = append(meals, status.MealType)
	}
	return meals
}

func initCustomID(action, date string, draft initDraft) string {
	meals := "-"
	if len(draft.Meals) > 0 {
		meals = encodeInitMeals(draft.Meals)
	}
	dates := "-"
	if len(draft.Dates) > 0 {
		dates = encodeInitDates(draft.Dates)
	}
	return strings.Join([]string{initCustomIDPrefix, encodeInitAction(action), compactInitDate(date), compactInitLocation(draft.Location), meals, dates}, initCustomIDSep)
}

func parseInitCustomID(id string) (action, date string, draft initDraft, err error) {
	parts := strings.Split(id, initCustomIDSep)
	if len(parts) != 6 || parts[0] != initCustomIDPrefix || parts[1] == "" || parts[2] == "" {
		return "", "", initDraft{}, fmt.Errorf("unrecognized init control")
	}
	location := expandInitLocation(parts[3])
	if location != "office" && location != "wfh" {
		return "", "", initDraft{}, fmt.Errorf("unrecognized init location draft")
	}
	draft = initDraft{Location: location}
	if parts[4] != "" && parts[4] != "-" {
		draft.Meals = decodeInitMeals(parts[4])
	}
	if parts[5] != "" && parts[5] != "-" {
		draft.Dates = decodeInitDates(parts[5])
	}
	draft.Anchor = expandInitDate(parts[2])
	return decodeInitAction(parts[1]), draft.Anchor, draft, nil
}

func encodeInitAction(action string) string {
	if strings.HasPrefix(action, initActionDateButton) {
		return initActionDateButton + compactInitDate(strings.TrimPrefix(action, initActionDateButton))
	}
	if strings.HasPrefix(action, initActionMealButton) {
		return initActionMealButton + compactInitMeal(strings.TrimPrefix(action, initActionMealButton))
	}
	return action
}

func decodeInitAction(action string) string {
	if strings.HasPrefix(action, initActionDateButton) {
		return initActionDateButton + expandInitDate(strings.TrimPrefix(action, initActionDateButton))
	}
	if strings.HasPrefix(action, initActionMealButton) {
		return initActionMealButton + expandInitMeal(strings.TrimPrefix(action, initActionMealButton))
	}
	return action
}

func compactInitDate(date string) string {
	return strings.ReplaceAll(date, "-", "")
}

func expandInitDate(date string) string {
	if len(date) == 8 {
		return date[:4] + "-" + date[4:6] + "-" + date[6:]
	}
	return date
}

func encodeInitDates(dates []string) string {
	parts := make([]string, 0, len(dates))
	for _, date := range dates {
		parts = append(parts, compactInitDate(date))
	}
	return strings.Join(parts, "")
}

func decodeInitDates(encoded string) []string {
	if encoded == "" {
		return nil
	}
	dates := make([]string, 0, len(encoded)/8)
	for len(encoded) >= 8 {
		dates = append(dates, expandInitDate(encoded[:8]))
		encoded = encoded[8:]
	}
	return dates
}

func compactInitLocation(location string) string {
	if location == "wfh" {
		return "w"
	}
	return "o"
}

func expandInitLocation(location string) string {
	if location == "w" || location == "wfh" {
		return "wfh"
	}
	return "office"
}

func encodeInitMeals(meals []string) string {
	parts := make([]string, 0, len(meals))
	for _, meal := range meals {
		parts = append(parts, compactInitMeal(meal))
	}
	return strings.Join(parts, "")
}

func decodeInitMeals(encoded string) []string {
	meals := make([]string, 0, len(encoded))
	for _, code := range encoded {
		meals = append(meals, expandInitMeal(string(code)))
	}
	return meals
}

func compactInitMeal(meal string) string {
	switch meal {
	case "lunch":
		return "l"
	case "snacks":
		return "s"
	case "iftar":
		return "i"
	case "event_dinner":
		return "e"
	case "optional_dinner":
		return "o"
	default:
		return meal
	}
}

func expandInitMeal(meal string) string {
	switch meal {
	case "l":
		return "lunch"
	case "s":
		return "snacks"
	case "i":
		return "iftar"
	case "e":
		return "event_dinner"
	case "o":
		return "optional_dinner"
	default:
		return meal
	}
}

func discordComponentPayload(interaction interactionBody) (string, json.RawMessage, error) {
	if strings.HasPrefix(interaction.Data.CustomID, adminInitDiscordCustomIDPrefix+adminInitDiscordCustomIDSep) {
		raw, err := discordAdminInitComponentPayload(interaction)
		return "admin-init", raw, err
	}

	action, date, draft, err := parseInitCustomID(interaction.Data.CustomID)
	if err != nil {
		return "", nil, fmt.Errorf("This interactive control is no longer recognized. Please run `/init` again.")
	}
	mealAction := strings.TrimPrefix(action, initActionMealButton)
	locationAction := strings.TrimPrefix(action, initActionLocationButton)
	dateAction := strings.TrimPrefix(action, initActionDateButton)

	opts := payload.InitOptions{
		Date:     date,
		Dates:    append(make([]string, 0, len(draft.Dates)), draft.Dates...),
		Action:   action,
		Location: draft.Location,
		Meals:    append(make([]string, 0, len(draft.Meals)), draft.Meals...),
	}
	if interaction.Data.Values == nil && action == initActionMeals {
		interaction.Data.Values = []string{}
	}

	switch action {
	case initActionLocation:
		if len(interaction.Data.Values) != 1 {
			return "", nil, fmt.Errorf("Please choose one location option.")
		}
		opts.Location = interaction.Data.Values[0]
	case initActionMeals:
		opts.Meals = append(make([]string, 0, len(interaction.Data.Values)), interaction.Data.Values...)
	case initActionRefresh, initActionApply:
	default:
		if strings.HasPrefix(action, initActionDateButton) {
			opts.Action = initActionDate
			opts.Dates = toggleInitDate(draft.Dates, dateAction)
		} else if strings.HasPrefix(action, initActionLocationButton) {
			opts.Action = initActionLocation
			opts.Location = locationAction
		} else if strings.HasPrefix(action, initActionMealButton) {
			opts.Action = initActionMeals
			opts.Meals = toggleInitMeal(draft.Meals, mealAction)
		} else {
			return "", nil, fmt.Errorf("This `/init` action is not supported.")
		}
	}

	raw, _ := json.Marshal(opts)
	return "init", raw, nil
}

func toggleInitMeal(meals []string, meal string) []string {
	if meal == "" {
		return append([]string(nil), meals...)
	}
	result := make([]string, 0, len(meals)+1)
	removed := false
	for _, current := range meals {
		if current == meal {
			removed = true
			continue
		}
		result = append(result, current)
	}
	if !removed {
		result = append(result, meal)
	}
	return result
}

func toggleInitDate(dates []string, date string) []string {
	if date == "" {
		return append([]string(nil), dates...)
	}
	result := make([]string, 0, len(dates)+1)
	removed := false
	for _, current := range dates {
		if current == date {
			removed = true
			continue
		}
		result = append(result, current)
	}
	if !removed {
		result = append(result, date)
	}
	if len(result) == 0 {
		return []string{date}
	}
	return result
}
