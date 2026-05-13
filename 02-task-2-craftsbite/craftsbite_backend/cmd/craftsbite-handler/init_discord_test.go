package main

import (
	"context"
	"encoding/json"
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
	availableMeals []string
	participations map[string]repository.MealParticipation
	locations      map[string]repository.WorkLocation
}

func newInitTestStore() *initTestStore {
	return &initTestStore{
		participations: map[string]repository.MealParticipation{},
		locations:      map[string]repository.WorkLocation{},
	}
}

func (s *initTestStore) GetDay(ctx context.Context, date string) (*repository.DaySchedule, error) {
	return s.day, nil
}

func (s *initTestStore) GetAvailableMeals(ctx context.Context, date string) ([]string, error) {
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
	return buildDiscordInitMessage(date, "office", statuses, []string{"lunch", "snacks"}, "", discord.NoticeToneInfo)
}

func TestDiscordComponentPayload_InitLocation(t *testing.T) {
	commandName, raw, err := discordComponentPayload(interactionBody{
		Data: interactionData{
			CustomID: initCustomID(initActionLocation, "2026-05-15", initDraft{Location: "office", Meals: []string{"lunch"}}),
			Values:   []string{"wfh"},
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
			CustomID: initCustomID(initActionMeals, "2026-05-15", initDraft{Location: "office", Meals: []string{"lunch", "snacks"}}),
			Values:   nil,
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
	if body.Data.Flags != 64 {
		t.Fatalf("flags = %d, want 64", body.Data.Flags)
	}
	if len(body.Data.Components) != 3 {
		t.Fatalf("expected 3 component rows, got %d", len(body.Data.Components))
	}
	locationSelect := body.Data.Components[0].Components[0]
	if locationSelect.CustomID != initCustomID(initActionLocation, "2026-05-15", initDraft{Location: "wfh", Meals: []string{"lunch"}}) {
		t.Fatalf("location select custom id = %q", locationSelect.CustomID)
	}
	mealSelect := body.Data.Components[1].Components[0]
	if mealSelect.CustomID != initCustomID(initActionMeals, "2026-05-15", initDraft{Location: "wfh", Meals: []string{"lunch"}}) {
		t.Fatalf("meal select custom id = %q", mealSelect.CustomID)
	}
	if mealSelect.MinValues == nil || *mealSelect.MinValues != 0 {
		t.Fatalf("meal select min values = %#v, want 0", mealSelect.MinValues)
	}
	applyButton := body.Data.Components[2].Components[0]
	if applyButton.CustomID != initCustomID(initActionApply, "2026-05-15", initDraft{Location: "wfh", Meals: []string{"lunch"}}) {
		t.Fatalf("apply button custom id = %q", applyButton.CustomID)
	}
	refreshButton := body.Data.Components[2].Components[1]
	if refreshButton.CustomID != initCustomID(initActionRefresh, "2026-05-15", initDraft{Location: "wfh", Meals: []string{"lunch"}}) {
		t.Fatalf("refresh button custom id = %q", refreshButton.CustomID)
	}
	if !mealSelect.Options[0].Default {
		t.Fatal("expected lunch to be selected by default")
	}
	if mealSelect.Options[1].Default {
		t.Fatal("expected snacks to be unselected by default")
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
	locationSelect := body.Data.Components[0].Components[0]
	if !locationSelect.Options[1].Default {
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
	mealSelect := body.Data.Components[1].Components[0]
	if mealSelect.Options[0].Default {
		t.Fatal("expected lunch draft to be deselected")
	}
	if !mealSelect.Options[1].Default {
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
	mealSelect := body.Data.Components[1].Components[0]
	if mealSelect.Options[0].Default || mealSelect.Options[1].Default {
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
	locationSelect := body.Data.Components[0].Components[0]
	if !locationSelect.Options[0].Default {
		t.Fatal("expected saved office location after refresh")
	}
	mealSelect := body.Data.Components[1].Components[0]
	if !mealSelect.Options[0].Default || mealSelect.Options[1].Default {
		t.Fatal("expected saved meal selection after refresh")
	}
}
