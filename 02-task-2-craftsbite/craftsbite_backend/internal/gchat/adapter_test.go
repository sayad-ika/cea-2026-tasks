package gchat

import (
	"strings"
	"testing"
)

func TestParseMealArgs_Valid(t *testing.T) {
	opts, err := parseMealArgs("in lunch 2026-03-20")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opts["status"] != "in" {
		t.Errorf("status = %q, want %q", opts["status"], "in")
	}
	if opts["meal"] != "lunch" {
		t.Errorf("meal = %q, want %q", opts["meal"], "lunch")
	}
	if opts["date"] != "2026-03-20" {
		t.Errorf("date = %q, want %q", opts["date"], "2026-03-20")
	}
}

func TestParseMealArgs_StatusOnly(t *testing.T) {
	opts, err := parseMealArgs("out")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opts["status"] != "out" {
		t.Errorf("status = %q, want %q", opts["status"], "out")
	}
	if opts["meal"] != "all" {
		t.Errorf("meal = %q, want %q", opts["meal"], "all")
	}
}

func TestParseMealArgs_InvalidStatus(t *testing.T) {
	_, err := parseMealArgs("maybe lunch")
	if err == nil {
		t.Fatal("expected error for invalid status, got nil")
	}
	if !strings.Contains(err.Error(), "Usage: /meal") {
		t.Errorf("error = %q, want usage message", err.Error())
	}
}

func TestParseMealArgs_InvalidMealType(t *testing.T) {
	_, err := parseMealArgs("in brunch")
	if err == nil {
		t.Fatal("expected error for invalid meal type, got nil")
	}
	if !strings.Contains(err.Error(), "Invalid meal_type") {
		t.Errorf("error = %q, want invalid meal_type message", err.Error())
	}
}

func TestParseMealArgs_InvalidDateFormat(t *testing.T) {
	_, err := parseMealArgs("in lunch 20-03-2026")
	if err == nil {
		t.Fatal("expected error for invalid date format, got nil")
	}
	if !strings.Contains(err.Error(), "Invalid date format") {
		t.Errorf("error = %q, want invalid date format message", err.Error())
	}
}

func TestParseMealArgs_Empty(t *testing.T) {
	_, err := parseMealArgs("")
	if err == nil {
		t.Fatal("expected error for empty input, got nil")
	}
}

func TestParseHeadcountArgs_MissingDate(t *testing.T) {
	_, err := parseHeadcountArgs("")
	if err == nil {
		t.Fatal("expected error for missing date, got nil")
	}
	if err.Error() != "Usage: /headcount YYYY-MM-DD" {
		t.Errorf("error = %q, want %q", err.Error(), "Usage: /headcount YYYY-MM-DD")
	}
}

func TestParseHeadcountArgs_InvalidDateFormat(t *testing.T) {
	_, err := parseHeadcountArgs("20-03-2026")
	if err == nil {
		t.Fatal("expected error for invalid date, got nil")
	}
	if !strings.Contains(err.Error(), "Invalid date format") {
		t.Errorf("error = %q, want invalid date format message", err.Error())
	}
}

func TestParseHeadcountArgs_Valid(t *testing.T) {
	opts, err := parseHeadcountArgs("2026-03-20")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opts["date"] != "2026-03-20" {
		t.Errorf("date = %q, want %q", opts["date"], "2026-03-20")
	}
}

func TestCommandMapping(t *testing.T) {
	expected := map[int64]string{
		1: "meal",
		2: "location",
		3: "team-summary",
		4: "headcount",
	}
	for id, want := range expected {
		got, ok := gchatCommandNames[id]
		if !ok {
			t.Errorf("command ID %d not found in mapping", id)
			continue
		}
		if got != want {
			t.Errorf("command ID %d = %q, want %q", id, got, want)
		}
	}
}

func TestToCommandEvent_NilPayload(t *testing.T) {
	evt := Event{Chat: ChatEvent{}}
	_, err := ToCommandEvent(evt, "user1", "member")
	if err == nil {
		t.Fatal("expected error for nil payload, got nil")
	}
}

func TestToCommandEvent_UnknownCommandID(t *testing.T) {
	evt := Event{
		Chat: ChatEvent{
			AppCommandPayload: &AppCommandPayload{
				AppCommandMetadata: AppCommandMetadata{AppCommandID: 99},
			},
		},
	}
	_, err := ToCommandEvent(evt, "user1", "member")
	if err == nil {
		t.Fatal("expected error for unknown command ID, got nil")
	}
	if !strings.Contains(err.Error(), "unknown command ID") {
		t.Errorf("error = %q, want unknown command ID message", err.Error())
	}
}

func TestToCommandEvent_MealParseError(t *testing.T) {
	evt := Event{
		Chat: ChatEvent{
			AppCommandPayload: &AppCommandPayload{
				AppCommandMetadata: AppCommandMetadata{AppCommandID: 1},
				Message:            &Message{ArgumentText: "maybe"},
			},
		},
	}
	_, err := ToCommandEvent(evt, "user1", "member")
	if err == nil {
		t.Fatal("expected error for invalid meal status, got nil")
	}
}

func TestToCommandEvent_HeadcountParseError(t *testing.T) {
	evt := Event{
		Chat: ChatEvent{
			AppCommandPayload: &AppCommandPayload{
				AppCommandMetadata: AppCommandMetadata{AppCommandID: 4},
				Message:            &Message{ArgumentText: ""},
			},
		},
	}
	_, err := ToCommandEvent(evt, "user1", "admin")
	if err == nil {
		t.Fatal("expected error for missing headcount date, got nil")
	}
}
