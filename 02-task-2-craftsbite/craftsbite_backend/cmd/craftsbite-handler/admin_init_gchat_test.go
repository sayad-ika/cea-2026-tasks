package main

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/sayad-ika/craftsbite/internal/gchat"
	"github.com/sayad-ika/craftsbite/internal/payload"
	"github.com/sayad-ika/craftsbite/internal/repository"
)

type adminInitTestStore struct {
	schedules map[string]*repository.DaySchedule
}

func newAdminInitTestStore() *adminInitTestStore {
	return &adminInitTestStore{schedules: map[string]*repository.DaySchedule{}}
}

func (s *adminInitTestStore) GetDay(ctx context.Context, date string) (*repository.DaySchedule, error) {
	schedule, ok := s.schedules[date]
	if !ok {
		return nil, nil
	}
	copy := *schedule
	copy.AvailableMeals = append([]string(nil), schedule.AvailableMeals...)
	return &copy, nil
}

func (s *adminInitTestStore) GetAvailableMeals(ctx context.Context, date string) ([]string, error) {
	schedule, ok := s.schedules[date]
	if !ok {
		return nil, nil
	}
	return append([]string(nil), schedule.AvailableMeals...), nil
}

func (s *adminInitTestStore) UpsertDaySchedule(ctx context.Context, schedule repository.DaySchedule) error {
	copy := schedule
	copy.AvailableMeals = append([]string(nil), schedule.AvailableMeals...)
	s.schedules[schedule.Date] = &copy
	return nil
}

func (s *adminInitTestStore) DeleteDaySchedule(ctx context.Context, date string) error {
	delete(s.schedules, date)
	return nil
}

func TestHandleAdminInitCommand_OpenCreatesScheduleCard(t *testing.T) {
	store := newAdminInitTestStore()
	dateParser := mustInitDateParser(t)
	date := nextInitTestDates(t, dateParser, 1)[0]
	store.schedules[date] = &repository.DaySchedule{Date: date, DayStatus: "celebration", AvailableMeals: []string{"lunch"}, Reason: "Company event"}

	ctx, recorder := withReplyRecorder(context.Background(), HandlerRequest{Platform: PlatformGChat})
	err := handleAdminInitCommand(ctx, store, nil, dateParser, payload.CommandEvent{
		UserID:  "admin1",
		Role:    "admin",
		Source:  "gchat",
		Options: json.RawMessage(`{"action":"open","date":"` + date + `"}`),
	})
	if err != nil {
		t.Fatalf("handleAdminInitCommand() error = %v", err)
	}

	cards := gchatCreatedCards(t, recorder.finalResponse().Body)
	card := &cards[0].Card
	if card.Header == nil || card.Header.Title != "Admin Setup" {
		t.Fatalf("header = %#v, want Admin Setup", card.Header)
	}
	dateInput := gchatTextInput(card, "date")
	if dateInput == nil || dateInput.Value != date || dateInput.Label != "Schedule date" || dateInput.Type != "SINGLE_LINE" {
		t.Fatalf("date input = %#v, want %s", dateInput, date)
	}
	if endInput := gchatTextInput(card, "end_date"); endInput != nil {
		t.Fatalf("end date input = %#v, want hidden by default", endInput)
	}
	rangeInput := gchatSelectionInput(card, "date_range")
	if rangeInput == nil || gchatSelectionSelected(rangeInput, "true") {
		t.Fatalf("date range input = %#v, want unchecked", rangeInput)
	}
	reasonInput := gchatTextInput(card, "reason")
	if reasonInput == nil || reasonInput.Value != "Company event" || reasonInput.Type != "MULTIPLE_LINE" {
		t.Fatalf("reason input = %#v, want multiline existing reason", reasonInput)
	}
	statusInput := gchatSelectionInput(card, "status")
	if statusInput == nil || !gchatSelectionSelected(statusInput, "celebration") {
		t.Fatalf("status input = %#v, want celebration selected", statusInput)
	}
	mealInput := gchatSelectionInput(card, "meals")
	if mealInput == nil || !gchatSelectionSelected(mealInput, "lunch") {
		t.Fatalf("meal input = %#v, want lunch selected", mealInput)
	}
	if !strings.Contains(gchatCardText(card), "Company event") {
		t.Fatalf("card text = %q, want existing reason", gchatCardText(card))
	}
}

func TestHandleAdminInitCommand_RangeToggleUpdatesCard(t *testing.T) {
	store := newAdminInitTestStore()
	dateParser := mustInitDateParser(t)
	start, end, _ := nextAdminInitWeekdayRange(t, dateParser, 2)

	ctx, recorder := withReplyRecorder(context.Background(), HandlerRequest{Platform: PlatformGChat})
	err := handleAdminInitCommand(ctx, store, nil, dateParser, payload.CommandEvent{
		UserID:  "admin1",
		Role:    "admin",
		Source:  "gchat",
		Options: json.RawMessage(`{"action":"schedule_range_toggle","date":"` + start + `","end_date":"` + end + `","use_range":true,"status":"normal","meals":["lunch"],"reason":"Range ops"}`),
	})
	if err != nil {
		t.Fatalf("handleAdminInitCommand() error = %v", err)
	}

	cards := gchatUpdatedCards(t, recorder.finalResponse().Body)
	card := &cards[0].Card
	dateInput := gchatTextInput(card, "date")
	if dateInput == nil || dateInput.Label != "Start date" || dateInput.Value != start {
		t.Fatalf("date input = %#v, want Start date %s", dateInput, start)
	}
	endInput := gchatTextInput(card, "end_date")
	if endInput == nil || endInput.Label != "End date" || endInput.Value != end {
		t.Fatalf("end date input = %#v, want End date %s", endInput, end)
	}
	if rangeInput := gchatSelectionInput(card, "date_range"); rangeInput == nil || !gchatSelectionSelected(rangeInput, "true") {
		t.Fatalf("date range input = %#v, want checked", rangeInput)
	}
	if statusInput := gchatSelectionInput(card, "status"); statusInput == nil || !gchatSelectionSelected(statusInput, "normal") {
		t.Fatalf("status input = %#v, want normal selected", statusInput)
	}
	if !strings.Contains(gchatCardText(card), "Range ops") {
		t.Fatalf("card text = %q, want preserved reason", gchatCardText(card))
	}
}

func TestHandleAdminInitCommand_SaveUpdatesScheduleAndCard(t *testing.T) {
	store := newAdminInitTestStore()
	dateParser := mustInitDateParser(t)
	date := nextInitTestDates(t, dateParser, 1)[0]

	ctx, recorder := withReplyRecorder(context.Background(), HandlerRequest{Platform: PlatformGChat})
	err := handleAdminInitCommand(ctx, store, nil, dateParser, payload.CommandEvent{
		UserID:  "admin1",
		Role:    "admin",
		Source:  "gchat",
		Options: json.RawMessage(`{"action":"schedule_apply","date":"` + date + `","status":"normal","meals":["lunch","snacks"],"reason":"<b>Ops</b>"}`),
	})
	if err != nil {
		t.Fatalf("handleAdminInitCommand() error = %v", err)
	}

	schedule := store.schedules[date]
	if schedule == nil {
		t.Fatal("schedule was not saved")
	}
	if schedule.DayStatus != "normal" || strings.Join(schedule.AvailableMeals, ",") != "lunch,snacks" || schedule.Reason != "<b>Ops</b>" {
		t.Fatalf("schedule = %+v", schedule)
	}
	cards := gchatUpdatedCards(t, recorder.finalResponse().Body)
	card := &cards[0].Card
	if card.Header == nil || card.Header.Title != "Schedule updated" {
		t.Fatalf("header = %#v, want Schedule updated", card.Header)
	}
	button := gchatFirstButton(card, "Edit schedule")
	if button == nil || button.OnClick == nil {
		t.Fatalf("confirmation card missing Edit schedule button: %#v", card)
	}
	if gchatActionParameter(button.OnClick.Action.Parameters, "action") != gchat.AdminInitFunctionScheduleEdit {
		t.Fatalf("edit parameters = %#v, want admin init edit action", button.OnClick.Action.Parameters)
	}
	if !strings.Contains(gchatCardText(card), "Reason: &lt;b&gt;Ops&lt;/b&gt;") {
		t.Fatalf("card text = %q, want escaped reason", gchatCardText(card))
	}
}

func TestHandleAdminInitCommand_SaveValidationUpdatesCard(t *testing.T) {
	store := newAdminInitTestStore()
	dateParser := mustInitDateParser(t)
	date := nextInitTestDates(t, dateParser, 1)[0]

	ctx, recorder := withReplyRecorder(context.Background(), HandlerRequest{Platform: PlatformGChat})
	err := handleAdminInitCommand(ctx, store, nil, dateParser, payload.CommandEvent{
		UserID:  "admin1",
		Role:    "admin",
		Source:  "gchat",
		Options: json.RawMessage(`{"action":"schedule_apply","date":"` + date + `","meals":["lunch"]}`),
	})
	if err != nil {
		t.Fatalf("handleAdminInitCommand() error = %v", err)
	}
	if store.schedules[date] != nil {
		t.Fatalf("schedule saved on validation failure: %+v", store.schedules[date])
	}

	cards := gchatUpdatedCards(t, recorder.finalResponse().Body)
	text := gchatCardText(&cards[0].Card)
	for _, want := range []string{"Review needed", "Choose a day status before saving."} {
		if !strings.Contains(text, want) {
			t.Fatalf("card text = %q, want %q", text, want)
		}
	}
}

func TestHandleAdminInitCommand_SaveRangeRequiresEndDate(t *testing.T) {
	store := newAdminInitTestStore()
	dateParser := mustInitDateParser(t)
	date := nextInitTestDates(t, dateParser, 1)[0]

	ctx, recorder := withReplyRecorder(context.Background(), HandlerRequest{Platform: PlatformGChat})
	err := handleAdminInitCommand(ctx, store, nil, dateParser, payload.CommandEvent{
		UserID:  "admin1",
		Role:    "admin",
		Source:  "gchat",
		Options: json.RawMessage(`{"action":"schedule_apply","date":"` + date + `","use_range":true,"status":"normal","meals":["lunch"]}`),
	})
	if err != nil {
		t.Fatalf("handleAdminInitCommand() error = %v", err)
	}
	if len(store.schedules) != 0 {
		t.Fatalf("schedules saved on validation failure: %+v", store.schedules)
	}

	cards := gchatUpdatedCards(t, recorder.finalResponse().Body)
	text := gchatCardText(&cards[0].Card)
	for _, want := range []string{"Review needed", "end date is required"} {
		if !strings.Contains(text, want) {
			t.Fatalf("card text = %q, want %q", text, want)
		}
	}
}

func TestHandleAdminInitCommand_SaveRangeUpdatesWeekdays(t *testing.T) {
	store := newAdminInitTestStore()
	dateParser := mustInitDateParser(t)
	start, end, weekdays := nextAdminInitWeekdayRange(t, dateParser, 2)

	ctx, recorder := withReplyRecorder(context.Background(), HandlerRequest{Platform: PlatformGChat})
	err := handleAdminInitCommand(ctx, store, nil, dateParser, payload.CommandEvent{
		UserID:  "admin1",
		Role:    "admin",
		Source:  "gchat",
		Options: json.RawMessage(`{"action":"schedule_apply","date":"` + start + `","end_date":"` + end + `","use_range":true,"status":"normal","meals":["lunch","snacks"],"reason":"Bulk ops"}`),
	})
	if err != nil {
		t.Fatalf("handleAdminInitCommand() error = %v", err)
	}
	for _, date := range weekdays {
		schedule := store.schedules[date]
		if schedule == nil {
			t.Fatalf("schedule for %s was not saved", date)
		}
		if schedule.DayStatus != "normal" || strings.Join(schedule.AvailableMeals, ",") != "lunch,snacks" || schedule.Reason != "Bulk ops" {
			t.Fatalf("schedule[%s] = %+v", date, schedule)
		}
	}
	if len(store.schedules) != len(weekdays) {
		t.Fatalf("saved schedules = %d, want %d weekdays", len(store.schedules), len(weekdays))
	}

	cards := gchatUpdatedCards(t, recorder.finalResponse().Body)
	text := gchatCardText(&cards[0].Card)
	for _, want := range []string{"Schedule updated", "Weekdays updated: 2", start, end} {
		if !strings.Contains(text, want) {
			t.Fatalf("card text = %q, want %q", text, want)
		}
	}
	button := gchatFirstButton(&cards[0].Card, "Edit schedule")
	if button == nil || gchatActionParameter(button.OnClick.Action.Parameters, "use_range") != "true" || gchatActionParameter(button.OnClick.Action.Parameters, "end_date") != end {
		t.Fatalf("edit parameters = %#v, want range state", button.OnClick.Action.Parameters)
	}
}

func TestHandleAdminInitCommand_OfficeClosedClearsMeals(t *testing.T) {
	store := newAdminInitTestStore()
	dateParser := mustInitDateParser(t)
	date := nextInitTestDates(t, dateParser, 1)[0]

	ctx, _ := withReplyRecorder(context.Background(), HandlerRequest{Platform: PlatformGChat})
	err := handleAdminInitCommand(ctx, store, nil, dateParser, payload.CommandEvent{
		UserID:  "admin1",
		Role:    "admin",
		Source:  "gchat",
		Options: json.RawMessage(`{"action":"schedule_apply","date":"` + date + `","status":"office_closed","meals":["lunch","snacks"],"reason":"Maintenance"}`),
	})
	if err != nil {
		t.Fatalf("handleAdminInitCommand() error = %v", err)
	}
	schedule := store.schedules[date]
	if schedule == nil {
		t.Fatal("schedule was not saved")
	}
	if schedule.DayStatus != "office_closed" || len(schedule.AvailableMeals) != 0 {
		t.Fatalf("schedule = %+v, want office_closed with no meals", schedule)
	}
}

func TestHandleAdminInitCommand_CancelUpdatesCard(t *testing.T) {
	store := newAdminInitTestStore()
	dateParser := mustInitDateParser(t)
	date := nextInitTestDates(t, dateParser, 1)[0]

	ctx, recorder := withReplyRecorder(context.Background(), HandlerRequest{Platform: PlatformGChat})
	err := handleAdminInitCommand(ctx, store, nil, dateParser, payload.CommandEvent{
		UserID:  "admin1",
		Role:    "admin",
		Source:  "gchat",
		Options: json.RawMessage(`{"action":"schedule_cancel","date":"` + date + `"}`),
	})
	if err != nil {
		t.Fatalf("handleAdminInitCommand() error = %v", err)
	}

	cards := gchatUpdatedCards(t, recorder.finalResponse().Body)
	card := &cards[0].Card
	if card.Header == nil || card.Header.Title != "Setup canceled" {
		t.Fatalf("header = %#v, want Setup canceled", card.Header)
	}
	if !strings.Contains(gchatCardText(card), "No changes were saved.") {
		t.Fatalf("card text = %q, want cancel summary", gchatCardText(card))
	}
}

func gchatTextInput(card *gchat.CardV2, name string) *gchat.TextInput {
	if card == nil {
		return nil
	}
	for _, section := range card.Sections {
		for _, widget := range section.Widgets {
			if widget.TextInput != nil && widget.TextInput.Name == name {
				return widget.TextInput
			}
		}
	}
	return nil
}

func gchatSelectionSelected(input *gchat.SelectionInput, value string) bool {
	if input == nil {
		return false
	}
	for _, item := range input.Items {
		if item.Value == value && item.Selected {
			return true
		}
	}
	return false
}

func nextAdminInitWeekdayRange(t *testing.T, dateParser interface{ ParseDateWithDefaults(string) (string, error) }, count int) (string, string, []string) {
	t.Helper()
	first, err := dateParser.ParseDateWithDefaults("")
	if err != nil {
		t.Fatalf("ParseDateWithDefaults() error = %v", err)
	}
	current, err := time.Parse("2006-01-02", first)
	if err != nil {
		t.Fatalf("time.Parse() error = %v", err)
	}
	weekdays := []string{}
	for len(weekdays) < count {
		if current.Weekday() != time.Saturday && current.Weekday() != time.Sunday {
			weekdays = append(weekdays, current.Format("2006-01-02"))
		}
		current = current.AddDate(0, 0, 1)
	}
	return weekdays[0], weekdays[len(weekdays)-1], weekdays
}
