package main

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/sayad-ika/craftsbite/internal/dateutil"
	"github.com/sayad-ika/craftsbite/internal/discord"
	"github.com/sayad-ika/craftsbite/internal/payload"
	"github.com/sayad-ika/craftsbite/internal/repository"
	"github.com/sayad-ika/craftsbite/internal/services"
)

type initTestStore struct {
	day            *repository.DaySchedule
	days           map[string]*repository.DaySchedule
	availableMeals []string
	mealsByDate    map[string][]string
	participations map[string]repository.MealParticipation
	locations      map[string]repository.WorkLocation
}

func newInitTestStore() *initTestStore {
	return &initTestStore{
		days:           map[string]*repository.DaySchedule{},
		mealsByDate:    map[string][]string{},
		participations: map[string]repository.MealParticipation{},
		locations:      map[string]repository.WorkLocation{},
	}
}

func (s *initTestStore) GetDay(ctx context.Context, date string) (*repository.DaySchedule, error) {
	if day, ok := s.days[date]; ok {
		return day, nil
	}
	return s.day, nil
}

func (s *initTestStore) GetAvailableMeals(ctx context.Context, date string) ([]string, error) {
	if meals, ok := s.mealsByDate[date]; ok {
		return append([]string(nil), meals...), nil
	}
	return append([]string(nil), s.availableMeals...), nil
}

func (s *initTestStore) GetParticipationsByUserDate(ctx context.Context, userID, date string) ([]repository.MealParticipation, error) {
	var result []repository.MealParticipation
	for _, meal := range s.availableMeals {
		if p, ok := s.participations[initMealKey(userID, date, meal)]; ok {
			result = append(result, p)
		}
	}
	return result, nil
}

func (s *initTestStore) GetParticipationsByDate(ctx context.Context, date string) ([]repository.MealParticipation, error) {
	return nil, nil
}

func (s *initTestStore) UpsertParticipation(ctx context.Context, p repository.MealParticipation, prevUpdatedAt time.Time) error {
	s.participations[initMealKey(p.UserID, p.Date, p.MealType)] = p
	return nil
}

func (s *initTestStore) GetWorkLocation(ctx context.Context, userID, date string) (*repository.WorkLocation, error) {
	wl, ok := s.locations[initLocationKey(userID, date)]
	if !ok {
		return nil, nil
	}
	copy := wl
	return &copy, nil
}

func (s *initTestStore) GetWorkLocationsByDate(ctx context.Context, date string) ([]repository.WorkLocation, error) {
	return nil, nil
}

func (s *initTestStore) UpsertWorkLocation(ctx context.Context, wl repository.WorkLocation, prevUpdatedAt time.Time) error {
	s.locations[initLocationKey(wl.UserID, wl.Date)] = wl
	return nil
}

func initMealKey(userID, date, meal string) string {
	return userID + ":" + date + ":" + meal
}

func initLocationKey(userID, date string) string {
	return userID + ":" + date
}

func mustInitDateParser(t *testing.T) *dateutil.DateParser {
	t.Helper()
	return dateutil.NewDateParser(time.UTC)
}

func mustInitCutoff(t *testing.T) *services.CutoffChecker {
	t.Helper()
	cutoff, err := services.NewCutoffChecker(services.CutoffConfig{Location: time.UTC, CutoffTime: "23:59"})
	if err != nil {
		t.Fatalf("NewCutoffChecker() error = %v", err)
	}
	return cutoff
}

func initPanelMessageForTest(date string) discord.Message {
	statuses := []services.ResolvedStatus{
		{MealType: "lunch", Status: "opted_in"},
		{MealType: "snacks", Status: "opted_out"},
	}
	availableDates := []initAvailableDate{{Date: date, Meals: []string{"lunch", "snacks"}}}
	draft := initDraft{Location: "office", Meals: []string{"lunch"}, Dates: []string{date}, Anchor: date, Saved: true}
	return buildDiscordInitMessage(date, "office", statuses, []string{"lunch", "snacks"}, availableDates, draft, "", discord.NoticeToneInfo)
}

func TestDiscordComponentPayload_InitLocation(t *testing.T) {
	commandName, raw, err := discordComponentPayload(interactionBody{
		Data: interactionData{
			CustomID: initCustomID(initActionLocationButton+"wfh", "2026-05-15", initDraft{Location: "office", Meals: []string{"lunch"}}),
		},
		Message: initPanelMessageForTest("2026-05-15"),
	})
	if err != nil {
		t.Fatalf("discordComponentPayload() error = %v", err)
	}
	if commandName != "init" {
		t.Fatalf("commandName = %q, want init", commandName)
	}
	var opts payload.InitOptions
	if err := json.Unmarshal(raw, &opts); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if opts.Action != initActionLocation || opts.Date != "2026-05-15" || opts.Location != "wfh" {
		t.Fatalf("unexpected init options: %+v", opts)
	}
	if len(opts.Meals) != 1 || opts.Meals[0] != "lunch" {
		t.Fatalf("expected existing meal draft to be preserved, got %+v", opts.Meals)
	}
}

func TestDiscordComponentPayload_InitApplyPreservesEmptyMeals(t *testing.T) {
	commandName, raw, err := discordComponentPayload(interactionBody{
		Data: interactionData{
			CustomID: initCustomID(initActionApply, "2026-05-15", initDraft{Location: "office", Meals: nil}),
		},
	})
	if err != nil {
		t.Fatalf("discordComponentPayload() error = %v", err)
	}
	if commandName != "init" {
		t.Fatalf("commandName = %q, want init", commandName)
	}
	var opts payload.InitOptions
	if err := json.Unmarshal(raw, &opts); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if opts.Meals == nil || len(opts.Meals) != 0 {
		t.Fatalf("expected explicit empty meals slice, got %#v", opts.Meals)
	}
}

func TestDiscordComponentPayload_InitMealsEmptySelection(t *testing.T) {
	commandName, raw, err := discordComponentPayload(interactionBody{
		Data: interactionData{
			CustomID: initCustomID(initActionMealButton+"lunch", "2026-05-15", initDraft{Location: "office", Meals: []string{"lunch"}}),
		},
	})
	if err != nil {
		t.Fatalf("discordComponentPayload() error = %v", err)
	}
	if commandName != "init" {
		t.Fatalf("commandName = %q, want init", commandName)
	}
	var opts payload.InitOptions
	if err := json.Unmarshal(raw, &opts); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if opts.Meals == nil {
		t.Fatal("expected empty meal selection to be preserved as an explicit empty slice")
	}
	if len(opts.Meals) != 0 {
		t.Fatalf("expected no selected meals, got %#v", opts.Meals)
	}
}

func TestDiscordComponentPayload_InitDateToggle(t *testing.T) {
	commandName, raw, err := discordComponentPayload(interactionBody{
		Data: interactionData{
			CustomID: initCustomID(initActionDateButton+"2026-05-16", "2026-05-15", initDraft{Location: "office", Meals: []string{"lunch"}, Dates: []string{"2026-05-15"}}),
		},
	})
	if err != nil {
		t.Fatalf("discordComponentPayload() error = %v", err)
	}
	if commandName != "init" {
		t.Fatalf("commandName = %q, want init", commandName)
	}
	var opts payload.InitOptions
	if err := json.Unmarshal(raw, &opts); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if opts.Action != initActionDate {
		t.Fatalf("action = %q, want date", opts.Action)
	}
	if strings.Join(opts.Dates, ",") != "2026-05-15,2026-05-16" {
		t.Fatalf("dates = %#v", opts.Dates)
	}
}

func TestInitCustomID_StaysWithinDiscordLimit(t *testing.T) {
	draft := initDraft{
		Location: "office",
		Meals:    []string{"lunch", "snacks", "optional_dinner"},
		Dates:    []string{"2026-05-15", "2026-05-16", "2026-05-17", "2026-05-18", "2026-05-19", "2026-05-20", "2026-05-21"},
	}
	id := initCustomID(initActionDateButton+"2026-05-21", "2026-05-15", draft)
	if len(id) > 100 {
		t.Fatalf("custom id length = %d, want <= 100: %q", len(id), id)
	}
}

func TestHandleInitInteraction_OpenPanel(t *testing.T) {
	store := newInitTestStore()
	store.availableMeals = []string{"lunch", "snacks"}
	store.locations[initLocationKey("u1", "2026-05-15")] = repository.WorkLocation{UserID: "u1", Date: "2026-05-15", Location: "wfh"}
	store.participations[initMealKey("u1", "2026-05-15", "lunch")] = repository.MealParticipation{UserID: "u1", Date: "2026-05-15", MealType: "lunch", IsParticipating: true}
	store.participations[initMealKey("u1", "2026-05-15", "snacks")] = repository.MealParticipation{UserID: "u1", Date: "2026-05-15", MealType: "snacks", IsParticipating: false}

	ctx, recorder := withReplyRecorder(context.Background(), HandlerRequest{Platform: PlatformDiscord, Raw: events.APIGatewayV2HTTPRequest{}})
	err := handleInitInteraction(ctx, nil, store, mustInitDateParser(t), nil, payload.CommandEvent{
		UserID:  "u1",
		Source:  "discord",
		Options: json.RawMessage(`{"action":"open","date":"2026-05-15"}`),
	})
	if err != nil {
		t.Fatalf("handleInitInteraction() error = %v", err)
	}

	var body RouterResponse
	if err := json.Unmarshal([]byte(recorder.finalResponse().Body), &body); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if body.Type != 4 {
		t.Fatalf("response type = %d, want 4", body.Type)
	}
	if body.Data == nil || len(body.Data.Embeds) != 1 {
		t.Fatal("expected one embed in init panel")
	}
	if body.Data.Embeds[0].Title != initPanelTitle {
		t.Fatalf("embed title = %q, want %q", body.Data.Embeds[0].Title, initPanelTitle)
	}
	if body.Data.Embeds[0].Description != "Choose dates, location, and meals. Then save.\nFri, May 15" {
		t.Fatalf("embed description = %q", body.Data.Embeds[0].Description)
	}
	if body.Data.Flags != 64 {
		t.Fatalf("flags = %d, want 64", body.Data.Flags)
	}
	if len(body.Data.Components) != 5 {
		t.Fatalf("expected 5 component rows, got %d", len(body.Data.Components))
	}
	dateButtons := append([]discord.Component{}, body.Data.Components[0].Components...)
	dateButtons = append(dateButtons, body.Data.Components[1].Components...)
	if len(dateButtons) != initConfiguredDateLimit {
		t.Fatalf("expected %d date buttons, got %d", initConfiguredDateLimit, len(dateButtons))
	}
	if dateButtons[0].Style != discord.ButtonStylePrimary {
		t.Fatal("expected first configured date to be selected")
	}
	locationButtons := body.Data.Components[2].Components
	if len(locationButtons) != 2 {
		t.Fatalf("expected 2 location buttons, got %d", len(locationButtons))
	}
	if locationButtons[0].Style != discord.ButtonStyleSecondary || locationButtons[1].Style != discord.ButtonStylePrimary {
		t.Fatalf("location button styles = %d/%d", locationButtons[0].Style, locationButtons[1].Style)
	}
	openDraft := initDraft{Location: "wfh", Meals: []string{"lunch"}, Dates: []string{"2026-05-15"}, Anchor: "2026-05-15", Saved: true}
	if locationButtons[1].CustomID != initCustomID(initActionLocationButton+"wfh", "2026-05-15", openDraft) {
		t.Fatalf("location button custom id = %q", locationButtons[1].CustomID)
	}
	mealButtons := body.Data.Components[3].Components
	if len(mealButtons) != 2 {
		t.Fatalf("expected 2 meal buttons, got %d", len(mealButtons))
	}
	if mealButtons[0].CustomID != initCustomID(initActionMealButton+"lunch", "2026-05-15", openDraft) {
		t.Fatalf("meal button custom id = %q", mealButtons[0].CustomID)
	}
	applyButton := body.Data.Components[4].Components[0]
	if applyButton.Label != "Save" {
		t.Fatalf("apply button label = %q", applyButton.Label)
	}
	if applyButton.CustomID != initCustomID(initActionApply, "2026-05-15", openDraft) {
		t.Fatalf("apply button custom id = %q", applyButton.CustomID)
	}
	refreshButton := body.Data.Components[4].Components[1]
	if refreshButton.Label != "Reset" {
		t.Fatalf("refresh button label = %q", refreshButton.Label)
	}
	if refreshButton.CustomID != initCustomID(initActionRefresh, "2026-05-15", openDraft) {
		t.Fatalf("refresh button custom id = %q", refreshButton.CustomID)
	}
	if mealButtons[0].Style != discord.ButtonStylePrimary {
		t.Fatal("expected lunch button to be selected")
	}
	if mealButtons[1].Style != discord.ButtonStyleSecondary {
		t.Fatal("expected snacks button to be unselected")
	}
}

func TestHandleInitInteraction_OpenPanelScansConfiguredMealDays(t *testing.T) {
	store := newInitTestStore()
	store.mealsByDate["2026-05-15"] = []string{"lunch"}
	store.mealsByDate["2026-05-16"] = nil
	store.days["2026-05-17"] = &repository.DaySchedule{Date: "2026-05-17", DayStatus: "office_closed"}
	store.mealsByDate["2026-05-17"] = []string{"lunch"}
	store.mealsByDate["2026-05-18"] = []string{"lunch"}
	store.mealsByDate["2026-05-19"] = []string{"lunch"}
	store.mealsByDate["2026-05-20"] = []string{"lunch"}
	store.mealsByDate["2026-05-21"] = []string{"lunch"}
	store.mealsByDate["2026-05-22"] = []string{"lunch"}
	store.mealsByDate["2026-05-23"] = []string{"lunch"}
	store.mealsByDate["2026-05-24"] = []string{"lunch"}

	ctx, recorder := withReplyRecorder(context.Background(), HandlerRequest{Platform: PlatformDiscord, Raw: events.APIGatewayV2HTTPRequest{}})
	err := handleInitInteraction(ctx, nil, store, mustInitDateParser(t), nil, payload.CommandEvent{
		UserID:  "u1",
		Source:  "discord",
		Options: json.RawMessage(`{"action":"open","date":"2026-05-15"}`),
	})
	if err != nil {
		t.Fatalf("handleInitInteraction() error = %v", err)
	}

	var body RouterResponse
	if err := json.Unmarshal([]byte(recorder.finalResponse().Body), &body); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	dateButtons := append([]discord.Component{}, body.Data.Components[0].Components...)
	dateButtons = append(dateButtons, body.Data.Components[1].Components...)
	gotLabels := make([]string, 0, len(dateButtons))
	for _, button := range dateButtons {
		gotLabels = append(gotLabels, button.Label)
	}
	wantLabels := []string{"Fri 15", "Mon 18", "Tue 19", "Wed 20", "Thu 21", "Fri 22", "Sat 23"}
	if strings.Join(gotLabels, ",") != strings.Join(wantLabels, ",") {
		t.Fatalf("date labels = %#v, want %#v", gotLabels, wantLabels)
	}
}

func TestHandleInitInteraction_LocationDraftDoesNotPersistUntilApply(t *testing.T) {
	store := newInitTestStore()
	store.availableMeals = []string{"lunch"}
	dateParser := mustInitDateParser(t)
	date, err := dateParser.ParseDateWithDefaults("")
	if err != nil {
		t.Fatalf("ParseDateWithDefaults() error = %v", err)
	}
	store.locations[initLocationKey("u1", date)] = repository.WorkLocation{UserID: "u1", Date: date, Location: "office"}

	ctx, recorder := withReplyRecorder(context.Background(), HandlerRequest{Platform: PlatformDiscord})
	err = handleInitInteraction(ctx, nil, store, dateParser, mustInitCutoff(t), payload.CommandEvent{
		UserID:  "u1",
		Source:  "discord",
		Options: json.RawMessage(`{"action":"location","date":"` + date + `","location":"wfh"}`),
	})
	if err != nil {
		t.Fatalf("handleInitInteraction() error = %v", err)
	}

	updated, err := store.GetWorkLocation(context.Background(), "u1", date)
	if err != nil {
		t.Fatalf("GetWorkLocation() error = %v", err)
	}
	if updated == nil || updated.Location != "office" {
		t.Fatalf("location persisted unexpectedly = %#v, want office", updated)
	}

	var body RouterResponse
	if err := json.Unmarshal([]byte(recorder.finalResponse().Body), &body); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if body.Type != 7 {
		t.Fatalf("response type = %d, want 7", body.Type)
	}
	locationButtons := body.Data.Components[2].Components
	if locationButtons[1].Style != discord.ButtonStylePrimary {
		t.Fatal("expected draft location to switch to WFH")
	}
	if body.Data.Embeds[0].Color != discord.BrandColor {
		t.Fatalf("embed color = %d, want brand color for draft update", body.Data.Embeds[0].Color)
	}
}

func TestHandleInitInteraction_MealDraftDoesNotPersistUntilApply(t *testing.T) {
	store := newInitTestStore()
	store.availableMeals = []string{"lunch", "snacks"}
	dateParser := mustInitDateParser(t)
	date, err := dateParser.ParseDateWithDefaults("")
	if err != nil {
		t.Fatalf("ParseDateWithDefaults() error = %v", err)
	}
	store.participations[initMealKey("u1", date, "lunch")] = repository.MealParticipation{UserID: "u1", Date: date, MealType: "lunch", IsParticipating: true}
	store.participations[initMealKey("u1", date, "snacks")] = repository.MealParticipation{UserID: "u1", Date: date, MealType: "snacks", IsParticipating: false}

	ctx, recorder := withReplyRecorder(context.Background(), HandlerRequest{Platform: PlatformDiscord})
	err = handleInitInteraction(ctx, nil, store, dateParser, mustInitCutoff(t), payload.CommandEvent{
		UserID:  "u1",
		Source:  "discord",
		Options: json.RawMessage(`{"action":"meals","date":"` + date + `","meals":["snacks"]}`),
	})
	if err != nil {
		t.Fatalf("handleInitInteraction() error = %v", err)
	}

	lunch := store.participations[initMealKey("u1", date, "lunch")]
	if !lunch.IsParticipating {
		t.Fatal("expected lunch participation to remain unchanged before apply")
	}
	snacks := store.participations[initMealKey("u1", date, "snacks")]
	if snacks.IsParticipating {
		t.Fatal("expected snacks participation to remain unchanged before apply")
	}

	var body RouterResponse
	if err := json.Unmarshal([]byte(recorder.finalResponse().Body), &body); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if body.Type != 7 {
		t.Fatalf("response type = %d, want 7", body.Type)
	}
	mealButtons := body.Data.Components[3].Components
	if mealButtons[0].Style != discord.ButtonStyleSecondary {
		t.Fatal("expected lunch draft to be deselected")
	}
	if mealButtons[1].Style != discord.ButtonStylePrimary {
		t.Fatal("expected snacks draft to be selected")
	}
}

func TestHandleInitInteraction_ApplyPersistsDraft(t *testing.T) {
	store := newInitTestStore()
	store.availableMeals = []string{"lunch", "snacks"}
	dateParser := mustInitDateParser(t)
	date, err := dateParser.ParseDateWithDefaults("")
	if err != nil {
		t.Fatalf("ParseDateWithDefaults() error = %v", err)
	}
	store.locations[initLocationKey("u1", date)] = repository.WorkLocation{UserID: "u1", Date: date, Location: "office"}
	store.participations[initMealKey("u1", date, "lunch")] = repository.MealParticipation{UserID: "u1", Date: date, MealType: "lunch", IsParticipating: true}
	store.participations[initMealKey("u1", date, "snacks")] = repository.MealParticipation{UserID: "u1", Date: date, MealType: "snacks", IsParticipating: false}

	ctx, recorder := withReplyRecorder(context.Background(), HandlerRequest{Platform: PlatformDiscord})
	err = handleInitInteraction(ctx, nil, store, dateParser, mustInitCutoff(t), payload.CommandEvent{
		UserID:  "u1",
		Source:  "discord",
		Options: json.RawMessage(`{"action":"apply","date":"` + date + `","location":"wfh","meals":["snacks"]}`),
	})
	if err != nil {
		t.Fatalf("handleInitInteraction() error = %v", err)
	}

	updatedLocation, _ := store.GetWorkLocation(context.Background(), "u1", date)
	if updatedLocation == nil || updatedLocation.Location != "wfh" {
		t.Fatalf("location = %#v, want wfh", updatedLocation)
	}
	if store.participations[initMealKey("u1", date, "lunch")].IsParticipating {
		t.Fatal("expected lunch to be opted out after apply")
	}
	if !store.participations[initMealKey("u1", date, "snacks")].IsParticipating {
		t.Fatal("expected snacks to be opted in after apply")
	}

	var body RouterResponse
	if err := json.Unmarshal([]byte(recorder.finalResponse().Body), &body); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if body.Type != 7 {
		t.Fatalf("response type = %d, want 7", body.Type)
	}
	if body.Data == nil || body.Data.Embeds[0].Color != discord.SuccessColor {
		t.Fatal("expected success response after apply")
	}
}

func TestHandleInitInteraction_ApplyPersistsMultipleDates(t *testing.T) {
	store := newInitTestStore()
	store.availableMeals = []string{"lunch", "snacks"}
	dateParser := mustInitDateParser(t)

	ctx, recorder := withReplyRecorder(context.Background(), HandlerRequest{Platform: PlatformDiscord})
	err := handleInitInteraction(ctx, nil, store, dateParser, mustInitCutoff(t), payload.CommandEvent{
		UserID: "u1",
		Source: "discord",
		Options: json.RawMessage(`{
			"action":"apply",
			"date":"2026-05-15",
			"dates":["2026-05-15","2026-05-16"],
			"location":"wfh",
			"meals":["lunch"]
		}`),
	})
	if err != nil {
		t.Fatalf("handleInitInteraction() error = %v", err)
	}

	for _, date := range []string{"2026-05-15", "2026-05-16"} {
		updatedLocation, _ := store.GetWorkLocation(context.Background(), "u1", date)
		if updatedLocation == nil || updatedLocation.Location != "wfh" {
			t.Fatalf("location for %s = %#v, want wfh", date, updatedLocation)
		}
		if !store.participations[initMealKey("u1", date, "lunch")].IsParticipating {
			t.Fatalf("expected lunch to be opted in for %s", date)
		}
		if store.participations[initMealKey("u1", date, "snacks")].IsParticipating {
			t.Fatalf("expected snacks to be opted out for %s", date)
		}
	}

	var body RouterResponse
	if err := json.Unmarshal([]byte(recorder.finalResponse().Body), &body); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if body.Data == nil || body.Data.Embeds[0].Color != discord.SuccessColor {
		t.Fatal("expected success response after multi-date apply")
	}
}

func TestHandleInitInteraction_ApplySupportsEmptyMealSelection(t *testing.T) {
	store := newInitTestStore()
	store.availableMeals = []string{"lunch", "snacks"}
	dateParser := mustInitDateParser(t)
	date, err := dateParser.ParseDateWithDefaults("")
	if err != nil {
		t.Fatalf("ParseDateWithDefaults() error = %v", err)
	}
	store.participations[initMealKey("u1", date, "lunch")] = repository.MealParticipation{UserID: "u1", Date: date, MealType: "lunch", IsParticipating: true}
	store.participations[initMealKey("u1", date, "snacks")] = repository.MealParticipation{UserID: "u1", Date: date, MealType: "snacks", IsParticipating: true}

	ctx, recorder := withReplyRecorder(context.Background(), HandlerRequest{Platform: PlatformDiscord})
	err = handleInitInteraction(ctx, nil, store, dateParser, mustInitCutoff(t), payload.CommandEvent{
		UserID:  "u1",
		Source:  "discord",
		Options: json.RawMessage(`{"action":"apply","date":"` + date + `","location":"office","meals":[]}`),
	})
	if err != nil {
		t.Fatalf("handleInitInteraction() error = %v", err)
	}

	if store.participations[initMealKey("u1", date, "lunch")].IsParticipating {
		t.Fatal("expected lunch to be opted out after empty apply")
	}
	if store.participations[initMealKey("u1", date, "snacks")].IsParticipating {
		t.Fatal("expected snacks to be opted out after empty apply")
	}

	var body RouterResponse
	if err := json.Unmarshal([]byte(recorder.finalResponse().Body), &body); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	mealButtons := body.Data.Components[3].Components
	if mealButtons[0].Style != discord.ButtonStyleSecondary || mealButtons[1].Style != discord.ButtonStyleSecondary {
		t.Fatal("expected all meals to be deselected after apply")
	}
}

func TestHandleInitInteraction_RefreshDiscardsDraft(t *testing.T) {
	store := newInitTestStore()
	store.availableMeals = []string{"lunch", "snacks"}
	dateParser := mustInitDateParser(t)
	date, err := dateParser.ParseDateWithDefaults("")
	if err != nil {
		t.Fatalf("ParseDateWithDefaults() error = %v", err)
	}
	store.locations[initLocationKey("u1", date)] = repository.WorkLocation{UserID: "u1", Date: date, Location: "office"}
	store.participations[initMealKey("u1", date, "lunch")] = repository.MealParticipation{UserID: "u1", Date: date, MealType: "lunch", IsParticipating: true}
	store.participations[initMealKey("u1", date, "snacks")] = repository.MealParticipation{UserID: "u1", Date: date, MealType: "snacks", IsParticipating: false}

	ctx, recorder := withReplyRecorder(context.Background(), HandlerRequest{Platform: PlatformDiscord})
	err = handleInitInteraction(ctx, nil, store, dateParser, mustInitCutoff(t), payload.CommandEvent{
		UserID:  "u1",
		Source:  "discord",
		Options: json.RawMessage(`{"action":"refresh","date":"` + date + `","location":"wfh","meals":["snacks"]}`),
	})
	if err != nil {
		t.Fatalf("handleInitInteraction() error = %v", err)
	}

	var body RouterResponse
	if err := json.Unmarshal([]byte(recorder.finalResponse().Body), &body); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	locationButtons := body.Data.Components[2].Components
	if locationButtons[0].Style != discord.ButtonStylePrimary {
		t.Fatal("expected saved office location after refresh")
	}
	mealButtons := body.Data.Components[3].Components
	if mealButtons[0].Style != discord.ButtonStylePrimary || mealButtons[1].Style != discord.ButtonStyleSecondary {
		t.Fatal("expected saved meal selection after refresh")
	}
}
