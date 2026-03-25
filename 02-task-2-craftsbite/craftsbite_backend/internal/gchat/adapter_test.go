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
	opts, err := parseMealArgs("in lunch 20-03-2026")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Adapter now passes raw date string - Lambda will validate
	if opts["date"] != "20-03-2026" {
		t.Errorf("date = %q, want %q", opts["date"], "20-03-2026")
	}
}

func TestParseMealArgs_Empty(t *testing.T) {
	_, err := parseMealArgs("")
	if err == nil {
		t.Fatal("expected error for empty input, got nil")
	}
}

func TestParseHeadcountArgs_MissingDate(t *testing.T) {
	opts, err := parseHeadcountArgs("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Adapter passes empty string - Lambda will default to tomorrow
	if opts["date"] != "" {
		t.Errorf("date = %q, want empty string", opts["date"])
	}
}

func TestParseHeadcountArgs_InvalidDateFormat(t *testing.T) {
	opts, err := parseHeadcountArgs("20-03-2026")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Adapter now passes raw date string - Lambda will validate
	if opts["date"] != "20-03-2026" {
		t.Errorf("date = %q, want %q", opts["date"], "20-03-2026")
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
		5: "status",
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
				Message:            &Message{ArgumentText: "invalid-date-format"},
			},
		},
	}
	ce, err := ToCommandEvent(evt, "user1", "admin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Adapter passes empty string - Lambda will handle default
	if ce.Options["date"] != "" {
		t.Errorf("date = %q, want empty string", ce.Options["date"])
	}
}
