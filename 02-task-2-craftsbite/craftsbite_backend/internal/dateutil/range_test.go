package dateutil

import (
	"testing"
	"time"
)

func TestParseDateRange_SingleDate(t *testing.T) {
	parser := mustDateParser(t)
	dates, err := parser.ParseDateRange("2026-03-20")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(dates) != 1 {
		t.Errorf("expected 1 date, got %d", len(dates))
	}
	if dates[0] != "2026-03-20" {
		t.Errorf("date = %q, want %q", dates[0], "2026-03-20")
	}
}

func TestParseDateRange_EmptyString(t *testing.T) {
	parser := mustDateParser(t)
	loc, _ := time.LoadLocation("Asia/Dhaka")
	now := time.Date(2026, 3, 19, 10, 0, 0, 0, loc)
	dates, err := parser.parseDateRangeAt("", now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(dates) != 1 {
		t.Errorf("expected 1 date, got %d", len(dates))
	}
	tomorrow := now.In(loc).AddDate(0, 0, 1).Format("2006-01-02")
	if dates[0] != tomorrow {
		t.Errorf("date = %q, want %q (tomorrow)", dates[0], tomorrow)
	}
}

func TestParseDateRange_Range(t *testing.T) {
	parser := mustDateParser(t)
	dates, err := parser.ParseDateRange("2026-03-20..2026-03-22")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := []string{"2026-03-20", "2026-03-21", "2026-03-22"}
	if len(dates) != len(expected) {
		t.Errorf("expected %d dates, got %d", len(expected), len(dates))
	}
	for i, want := range expected {
		if dates[i] != want {
			t.Errorf("dates[%d] = %q, want %q", i, dates[i], want)
		}
	}
}

func TestParseDateRange_RangeWithShortcuts(t *testing.T) {
	parser := mustDateParser(t)
	loc, _ := time.LoadLocation("Asia/Dhaka")
	dates, err := parser.parseDateRangeAt("today..+2", time.Date(2026, 3, 19, 10, 0, 0, 0, loc))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(dates) != 3 {
		t.Errorf("expected 3 dates, got %d", len(dates))
	}
	expected := []string{"2026-03-19", "2026-03-20", "2026-03-21"}
	for i, want := range expected {
		if dates[i] != want {
			t.Errorf("dates[%d] = %q, want %q", i, dates[i], want)
		}
	}
}

func TestParseDateRange_Week(t *testing.T) {
	parser := mustDateParser(t)
	loc, _ := time.LoadLocation("Asia/Dhaka")
	dates, err := parser.parseDateRangeAt("week", time.Date(2026, 3, 20, 10, 0, 0, 0, loc))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(dates) != 5 {
		t.Errorf("expected 5 business days, got %d", len(dates))
	}
	expected := []string{"2026-03-23", "2026-03-24", "2026-03-25", "2026-03-26", "2026-03-27"}
	for i, want := range expected {
		if dates[i] != want {
			t.Errorf("dates[%d] = %q, want %q", i, dates[i], want)
		}
	}
}

func TestParseDateRange_TooLarge(t *testing.T) {
	parser := mustDateParser(t)
	_, err := parser.ParseDateRange("2026-03-01..2026-03-20")
	if err == nil {
		t.Fatal("expected error for range > 14 days, got nil")
	}
}

func TestParseDateRange_InvalidRange(t *testing.T) {
	parser := mustDateParser(t)
	_, err := parser.ParseDateRange("2026-03-22..2026-03-20")
	if err == nil {
		t.Fatal("expected error for end before start, got nil")
	}
}

func TestParseDateRange_InvalidFormat(t *testing.T) {
	parser := mustDateParser(t)
	_, err := parser.ParseDateRange("2026-03-20..2026-03-22..2026-03-24")
	if err == nil {
		t.Fatal("expected error for invalid range format, got nil")
	}
}
