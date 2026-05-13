package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

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
	initActionApply    = "apply"

	initCustomIDPrefix = "init"
	initCustomIDSep    = "|"
	initPanelTitle     = "CraftsBite Quick Setup"
	initPanelSubtitle  = "Tomorrow's setup. Changes are not saved until you click Apply."
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
}

type initDraft struct {
	Location string
	Meals    []string
}

func handleInitCommand(ctx context.Context, deps handlerDeps, event payload.CommandEvent) error {
	return handleInitInteraction(ctx, deps.cfg, deps.store, deps.dateParser, deps.cutoff, event)
}

func handleInitInteraction(ctx context.Context, cfg *appconfig.Config, store initStore, dateParser *dateutil.DateParser, cutoff *services.CutoffChecker, event payload.CommandEvent) error {
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

	date, err := dateParser.ParseDateWithDefaults(opts.Date)
	if err != nil {
		return sendInitPanelError(ctx, fmt.Sprintf("Invalid init date: %v", err), opts.Action != initActionOpen)
	}

	state, err := loadInitState(ctx, store, event.UserID, date)
	if err != nil {
		return sendInitPanelError(ctx, "Unable to load your current setup. Please try again shortly.", opts.Action != initActionOpen)
	}

	savedDraft := initDraftFromState(state)

	switch opts.Action {
	case initActionOpen:
		return renderInitPanel(ctx, date, state, savedDraft, "", discord.NoticeToneInfo, false)
	case initActionRefresh:
		return renderInitPanel(ctx, date, state, savedDraft, "Draft discarded. Showing your saved setup.", discord.NoticeToneSuccess, true)
	case initActionLocation:
		draft := normalizeInitDraft(opts, state)
		if draft.Location != "office" && draft.Location != "wfh" {
			return renderInitPanel(ctx, date, state, savedDraft, "Please choose either Office or WFH.", discord.NoticeToneWarning, true)
		}
		return renderInitPanel(ctx, date, state, draft, fmt.Sprintf("Draft location set to %s. Click Apply to save.", displayLocationLabel(draft.Location)), discord.NoticeToneInfo, true)
	case initActionMeals:
		draft := normalizeInitDraft(opts, state)
		return renderInitPanel(ctx, date, state, draft, "Draft meal selection updated. Click Apply to save.", discord.NoticeToneInfo, true)
	case initActionApply:
		draft := normalizeInitDraft(opts, state)
		return applyInitDraftAndRender(ctx, store, cutoff, event.UserID, date, state, draft)
	default:
		return renderInitPanel(ctx, date, state, savedDraft, "This `/init` action is not supported.", discord.NoticeToneWarning, true)
	}
}

func applyInitDraftAndRender(ctx context.Context, store initStore, cutoff *services.CutoffChecker, userID, date string, state initState, draft initDraft) error {
	mealsChanged := !sameMealSelection(draft.Meals, initDraftFromState(state).Meals)
	locationChanged := draft.Location != state.Location

	if !mealsChanged && !locationChanged {
		return renderInitPanel(ctx, date, state, draft, "No changes to apply.", discord.NoticeToneInfo, true)
	}

	if mealsChanged {
		if err := applyInitMealSelection(ctx, store, cutoff, userID, date, draft.Meals); err != nil {
			return renderInitPanel(ctx, date, state, draft, mealErrorReply(err, date), discord.NoticeToneWarning, true)
		}
	}

	if locationChanged {
		if _, err := services.SetLocation(ctx, store, userID, date, draft.Location, cutoff); err != nil {
			if mealsChanged {
				refreshed, loadErr := loadInitState(ctx, store, userID, date)
				if loadErr != nil {
					return sendInitPanelError(ctx, "Meals were saved, but the panel could not be refreshed. Please run `/init` again.", true)
				}
				return renderInitPanel(ctx, date, refreshed, initDraftFromState(refreshed), "Meals were saved, but location could not be updated: "+locationErrorReply(err, date), discord.NoticeToneWarning, true)
			}
			return renderInitPanel(ctx, date, state, draft, locationErrorReply(err, date), discord.NoticeToneWarning, true)
		}
	}

	refreshed, err := loadInitState(ctx, store, userID, date)
	if err != nil {
		return sendInitPanelError(ctx, "Changes were saved, but the panel could not be refreshed. Please run `/init` again.", true)
	}

	note := "Draft applied successfully."
	if len(draft.Meals) == 0 && mealsChanged {
		note = "Draft applied successfully. All meals are now opted out."
	}
	return renderInitPanel(ctx, date, refreshed, initDraftFromState(refreshed), note, discord.NoticeToneSuccess, true)
}

func renderInitPanel(ctx context.Context, date string, state initState, draft initDraft, note string, tone discord.NoticeTone, update bool) error {
	message := buildDiscordInitMessage(date, draft.Location, draftStatuses(state, draft), state.AvailableMeals, note, tone)
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

func buildDiscordInitMessage(date, location string, statuses []services.ResolvedStatus, availableMeals []string, note string, tone discord.NoticeTone) discord.Message {
	fields := []discord.EmbedField{
		{Name: "Date", Value: date, Inline: true},
		{Name: "Work location", Value: displayLocationLabel(location), Inline: true},
		{Name: "Meals", Value: initMealSummary(statuses)},
	}
	if note != "" {
		fields = append(fields, discord.EmbedField{Name: "Update", Value: note})
	}

	embed := discord.BrandEmbed(initPanelTitle, initPanelSubtitle, fields)
	if note != "" {
		embed = discord.ToneEmbed(initPanelTitle, initPanelSubtitle, fields, tone)
	}

	message := discord.EmbedMessage(embed)
	message.Components = initMessageComponents(date, location, statuses, availableMeals)
	return message
}

func initMessageComponents(date, location string, statuses []services.ResolvedStatus, availableMeals []string) []discord.Component {
	draft := initDraft{Location: location, Meals: selectedMealsFromStatuses(statuses)}
	components := []discord.Component{{
		Type:       discord.ComponentTypeActionRow,
		Components: []discord.Component{buildInitLocationSelect(date, draft)},
	}}

	mealSelect := buildInitMealSelect(date, statuses, availableMeals, draft)
	if mealSelect.Type != 0 {
		components = append(components, discord.Component{
			Type:       discord.ComponentTypeActionRow,
			Components: []discord.Component{mealSelect},
		})
	}

	components = append(components, discord.Component{
		Type: discord.ComponentTypeActionRow,
		Components: []discord.Component{
			{
				Type:     discord.ComponentTypeButton,
				Style:    discord.ButtonStyleSuccess,
				Label:    "Apply",
				CustomID: initCustomID(initActionApply, date, draft),
			},
			{
				Type:     discord.ComponentTypeButton,
				Style:    discord.ButtonStyleSecondary,
				Label:    "Refresh",
				CustomID: initCustomID(initActionRefresh, date, draft),
			},
		},
	})

	return components
}

func buildInitLocationSelect(date string, draft initDraft) discord.Component {
	return discord.Component{
		Type:        discord.ComponentTypeStringSelect,
		CustomID:    initCustomID(initActionLocation, date, draft),
		Placeholder: "Choose your location",
		Options: []discord.SelectOption{
			{Label: "Office", Value: "office", Default: draft.Location != "wfh"},
			{Label: "WFH", Value: "wfh", Default: draft.Location == "wfh"},
		},
		MinValues: intPtr(1),
		MaxValues: intPtr(1),
	}
}

func buildInitMealSelect(date string, statuses []services.ResolvedStatus, availableMeals []string, draft initDraft) discord.Component {
	if len(availableMeals) == 0 {
		return discord.Component{}
	}

	statusByMeal := make(map[string]services.ResolvedStatus, len(statuses))
	allUnavailable := len(statuses) > 0
	for _, status := range statuses {
		statusByMeal[status.MealType] = status
		if status.Status != "unavailable" {
			allUnavailable = false
		}
	}

	options := make([]discord.SelectOption, 0, len(availableMeals))
	selected := make(map[string]struct{}, len(draft.Meals))
	for _, meal := range draft.Meals {
		selected[meal] = struct{}{}
	}
	for _, meal := range availableMeals {
		status := statusByMeal[meal]
		description := "Select meals to include before applying"
		if status.Status == "unavailable" {
			description = "This meal is currently unavailable"
		}
		options = append(options, discord.SelectOption{
			Label:       cmdutil.DisplayMealName(meal),
			Value:       meal,
			Description: description,
			Default:     initMealDefault(status, meal, selected),
		})
	}

	return discord.Component{
		Type:        discord.ComponentTypeStringSelect,
		CustomID:    initCustomID(initActionMeals, date, draft),
		Placeholder: "Choose the meals you want included",
		Options:     options,
		MinValues:   intPtr(0),
		MaxValues:   intPtr(len(options)),
		Disabled:    allUnavailable,
	}
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

func initMealDefault(status services.ResolvedStatus, meal string, selected map[string]struct{}) bool {
	if status.Status == "unavailable" {
		return false
	}
	_, ok := selected[meal]
	return ok
}

func initMealSummary(statuses []services.ResolvedStatus) string {
	if len(statuses) == 0 {
		return "No meals configured"
	}

	lines := make([]string, 0, len(statuses))
	for _, status := range statuses {
		lines = append(lines, fmt.Sprintf("%s: %s", cmdutil.DisplayMealName(status.MealType), mealStatusValue(status.Status)))
	}
	return strings.Join(lines, "\n")
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

func normalizeInitDraft(opts payload.InitOptions, state initState) initDraft {
	draft := initDraftFromState(state)
	if opts.Location == "office" || opts.Location == "wfh" {
		draft.Location = opts.Location
	}
	if opts.Meals != nil {
		draft.Meals = filterInitMeals(opts.Meals, state.AvailableMeals)
	}
	return draft
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
		meals = strings.Join(draft.Meals, ",")
	}
	return strings.Join([]string{initCustomIDPrefix, action, date, draft.Location, meals}, initCustomIDSep)
}

func parseInitCustomID(id string) (action, date string, draft initDraft, err error) {
	parts := strings.Split(id, initCustomIDSep)
	if len(parts) != 5 || parts[0] != initCustomIDPrefix || parts[1] == "" || parts[2] == "" {
		return "", "", initDraft{}, fmt.Errorf("unrecognized init control")
	}
	location := parts[3]
	if location != "office" && location != "wfh" {
		return "", "", initDraft{}, fmt.Errorf("unrecognized init location draft")
	}
	draft = initDraft{Location: location}
	if parts[4] != "" && parts[4] != "-" {
		draft.Meals = strings.Split(parts[4], ",")
	}
	return parts[1], parts[2], draft, nil
}

func discordComponentPayload(interaction interactionBody) (string, json.RawMessage, error) {
	action, date, draft, err := parseInitCustomID(interaction.Data.CustomID)
	if err != nil {
		return "", nil, fmt.Errorf("This interactive control is no longer recognized. Please run `/init` again.")
	}

	opts := payload.InitOptions{
		Date:     date,
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
		return "", nil, fmt.Errorf("This `/init` action is not supported.")
	}

	raw, _ := json.Marshal(opts)
	return "init", raw, nil
}

func intPtr(v int) *int {
	return &v
}
