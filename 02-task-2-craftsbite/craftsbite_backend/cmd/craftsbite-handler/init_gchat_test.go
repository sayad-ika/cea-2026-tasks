package main

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/sayad-ika/craftsbite/internal/gchat"
	"github.com/sayad-ika/craftsbite/internal/payload"
	"github.com/sayad-ika/craftsbite/internal/repository"
)

func TestHandleGChatInitInteraction_OpenCardMessage(t *testing.T) {
	store := newInitTestStore()
	store.availableMeals = []string{"lunch", "snacks"}
	dateParser := mustInitDateParser(t)
	date := nextInitTestDates(t, dateParser, 1)[0]
	store.locations[initLocationKey("u1", date)] = repository.WorkLocation{UserID: "u1", Date: date, Location: "wfh"}
	store.participations[initMealKey("u1", date, "lunch")] = repository.MealParticipation{UserID: "u1", Date: date, MealType: "lunch", IsParticipating: true}
	store.participations[initMealKey("u1", date, "snacks")] = repository.MealParticipation{UserID: "u1", Date: date, MealType: "snacks", IsParticipating: false}

	ctx, recorder := withReplyRecorder(context.Background(), HandlerRequest{Platform: PlatformGChat, Raw: events.APIGatewayV2HTTPRequest{
		RawPath: "/gchat",
		Headers: map[string]string{
			"host":              "example.com",
			"x-forwarded-proto": "https",
		},
	}})
	err := handleGChatInitInteraction(ctx, nil, store, dateParser, mustInitCutoff(t), payload.CommandEvent{
		UserID:  "u1",
		Source:  "gchat",
		Options: json.RawMessage(`{"action":"open","date":"` + date + `"}`),
	})
	if err != nil {
		t.Fatalf("handleGChatInitInteraction() error = %v", err)
	}

	cards := gchatCreatedCards(t, recorder.finalResponse().Body)
	card := &cards[0].Card
	if card.Header == nil || card.Header.Title != "CraftsBite Setup" {
		t.Fatalf("header = %#v, want CraftsBite Setup", card.Header)
	}
	if card.FixedFooter != nil {
		t.Fatal("expected non-modal card message without fixedFooter")
	}
	buttons := card.Sections[len(card.Sections)-1].Widgets[0].ButtonList
	if buttons == nil || len(buttons.Buttons) != 2 {
		t.Fatalf("buttons = %#v, want Save and Cancel button list", buttons)
	}
	if buttons.Buttons[0].OnClick == nil || buttons.Buttons[0].OnClick.Action.Function != "https://example.com/gchat" {
		t.Fatalf("save function = %#v, want endpoint URL", buttons.Buttons[0].OnClick)
	}
	if gchatActionParameter(buttons.Buttons[0].OnClick.Action.Parameters, "action") != gchat.InitCardFunctionSave {
		t.Fatalf("save parameters = %#v, want action parameter", buttons.Buttons[0].OnClick.Action.Parameters)
	}
	dateInput := gchatSelectionInput(card, "dates")
	if dateInput == nil || len(dateInput.Items) == 0 || !dateInput.Items[0].Selected {
		t.Fatalf("date input = %#v, want first date selected", dateInput)
	}
	locationInput := gchatSelectionInput(card, "location")
	if locationInput == nil || len(locationInput.Items) != 2 || !locationInput.Items[1].Selected {
		t.Fatalf("location input = %#v, want WFH selected", locationInput)
	}
	mealInput := gchatSelectionInput(card, "meals")
	if mealInput == nil || len(mealInput.Items) != 2 || !mealInput.Items[0].Selected || mealInput.Items[1].Selected {
		t.Fatalf("meal input = %#v, want lunch selected only", mealInput)
	}
}

func TestHandleGChatInitInteraction_SavePersistsMultipleDates(t *testing.T) {
	store := newInitTestStore()
	store.availableMeals = []string{"lunch", "snacks"}
	dateParser := mustInitDateParser(t)
	dates := nextInitTestDates(t, dateParser, 2)

	ctx, recorder := withReplyRecorder(context.Background(), HandlerRequest{Platform: PlatformGChat})
	err := handleGChatInitInteraction(ctx, nil, store, dateParser, mustInitCutoff(t), payload.CommandEvent{
		UserID: "u1",
		Source: "gchat",
		Options: json.RawMessage(`{
			"action":"apply",
			"date":"` + dates[0] + `",
			"dates":["` + dates[0] + `","` + dates[1] + `"],
			"location":"wfh",
			"meals":["snacks"]
		}`),
	})
	if err != nil {
		t.Fatalf("handleGChatInitInteraction() error = %v", err)
	}

	for _, date := range dates {
		updatedLocation, _ := store.GetWorkLocation(context.Background(), "u1", date)
		if updatedLocation == nil || updatedLocation.Location != "wfh" {
			t.Fatalf("location for %s = %#v, want wfh", date, updatedLocation)
		}
		if store.participations[initMealKey("u1", date, "lunch")].IsParticipating {
			t.Fatalf("expected lunch to be opted out for %s", date)
		}
		if !store.participations[initMealKey("u1", date, "snacks")].IsParticipating {
			t.Fatalf("expected snacks to be opted in for %s", date)
		}
	}

	cards := gchatCreatedCards(t, recorder.finalResponse().Body)
	summary := gchatCardText(&cards[0].Card)
	for _, want := range []string{
		"Setup saved",
		"Updated 2 meal days.",
		"Dates:",
		displayInitDate(dates[0]),
		displayInitDate(dates[1]),
		"Work location:</b> WFH",
		"Meals included:</b> Snacks",
	} {
		if !strings.Contains(summary, want) {
			t.Fatalf("summary = %q, want %q", summary, want)
		}
	}
}

func TestHandleGChatInitInteraction_SaveSummarizesNoMeals(t *testing.T) {
	store := newInitTestStore()
	store.availableMeals = []string{"lunch", "snacks"}
	dateParser := mustInitDateParser(t)
	date := nextInitTestDates(t, dateParser, 1)[0]

	ctx, recorder := withReplyRecorder(context.Background(), HandlerRequest{Platform: PlatformGChat})
	err := handleGChatInitInteraction(ctx, nil, store, dateParser, mustInitCutoff(t), payload.CommandEvent{
		UserID: "u1",
		Source: "gchat",
		Options: json.RawMessage(`{
			"action":"apply",
			"date":"` + date + `",
			"dates":["` + date + `"],
			"location":"office",
			"meals":[]
		}`),
	})
	if err != nil {
		t.Fatalf("handleGChatInitInteraction() error = %v", err)
	}

	cards := gchatCreatedCards(t, recorder.finalResponse().Body)
	summary := gchatCardText(&cards[0].Card)
	for _, want := range []string{
		"Setup saved",
		"Updated 1 meal day.",
		"Date:</b> " + displayInitDate(date),
		"Work location:</b> Office",
		"Meals included:</b> No meals included",
	} {
		if !strings.Contains(summary, want) {
			t.Fatalf("summary = %q, want %q", summary, want)
		}
	}
}

func TestHandleGChatInitInteraction_SaveRequiresDateSelection(t *testing.T) {
	store := newInitTestStore()
	store.availableMeals = []string{"lunch", "snacks"}
	dateParser := mustInitDateParser(t)
	date := nextInitTestDates(t, dateParser, 1)[0]

	ctx, recorder := withReplyRecorder(context.Background(), HandlerRequest{Platform: PlatformGChat})
	err := handleGChatInitInteraction(ctx, nil, store, dateParser, mustInitCutoff(t), payload.CommandEvent{
		UserID:  "u1",
		Source:  "gchat",
		Options: json.RawMessage(`{"action":"apply","date":"` + date + `","dates":[],"location":"office","meals":["lunch"]}`),
	})
	if err != nil {
		t.Fatalf("handleGChatInitInteraction() error = %v", err)
	}
	if len(store.locations) != 0 || len(store.participations) != 0 {
		t.Fatalf("expected no writes, got locations=%#v participations=%#v", store.locations, store.participations)
	}

	gchatCreatedCards(t, recorder.finalResponse().Body)
	if !strings.Contains(recorder.finalResponse().Body, "at least one date") {
		t.Fatalf("response body = %s, want validation message", recorder.finalResponse().Body)
	}
}

func nextInitTestDates(t *testing.T, dateParser interface{ ParseDateWithDefaults(string) (string, error) }, count int) []string {
	t.Helper()
	first, err := dateParser.ParseDateWithDefaults("")
	if err != nil {
		t.Fatalf("ParseDateWithDefaults() error = %v", err)
	}
	start, err := time.Parse("2006-01-02", first)
	if err != nil {
		t.Fatalf("time.Parse() error = %v", err)
	}
	dates := make([]string, 0, count)
	for i := 0; i < count; i++ {
		dates = append(dates, start.AddDate(0, 0, i).Format("2006-01-02"))
	}
	return dates
}

func gchatSelectionInput(card *gchat.CardV2, name string) *gchat.SelectionInput {
	if card == nil {
		return nil
	}
	for _, section := range card.Sections {
		for _, widget := range section.Widgets {
			if widget.SelectionInput != nil && widget.SelectionInput.Name == name {
				return widget.SelectionInput
			}
		}
	}
	return nil
}

func gchatCreatedCards(t *testing.T, body string) []gchat.CardV2Wrapper {
	t.Helper()
	var envelope struct {
		HostAppDataAction struct {
			ChatDataAction struct {
				CreateMessageAction struct {
					Message struct {
						CardsV2 []gchat.CardV2Wrapper `json:"cardsV2"`
					} `json:"message"`
				} `json:"createMessageAction"`
			} `json:"chatDataAction"`
		} `json:"hostAppDataAction"`
	}
	if err := json.Unmarshal([]byte(body), &envelope); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	cards := envelope.HostAppDataAction.ChatDataAction.CreateMessageAction.Message.CardsV2
	if len(cards) == 0 {
		t.Fatalf("response body = %s, want createMessageAction.message.cardsV2", body)
	}
	return cards
}

func gchatActionParameter(params []gchat.ActionParameter, key string) string {
	for _, param := range params {
		if param.Key == key {
			return param.Value
		}
	}
	return ""
}

func gchatCardText(card *gchat.CardV2) string {
	if card == nil {
		return ""
	}
	parts := []string{}
	for _, section := range card.Sections {
		for _, widget := range section.Widgets {
			if widget.TextParagraph != nil {
				parts = append(parts, widget.TextParagraph.Text)
			}
			if widget.DecoratedText != nil {
				parts = append(parts, widget.DecoratedText.Text)
			}
		}
	}
	return strings.Join(parts, "\n")
}
