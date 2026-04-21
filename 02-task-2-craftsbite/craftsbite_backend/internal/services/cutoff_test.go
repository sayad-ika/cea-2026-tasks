package services

import (
	"testing"
	"time"
)

func dhakaLoc(t *testing.T) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation("Asia/Dhaka")
	if err != nil {
		t.Skipf("timezone Asia/Dhaka not available: %v", err)
	}
	return loc
}

func TestScenario_RequestMar10_TargetMar11(t *testing.T) {
	checker := mustCutoffChecker(t)
	loc := dhakaLoc(t)

	const target = "2026-03-11"

	t.Run("before cutoff (18:00) — should be allowed", func(t *testing.T) {
		now := time.Date(2026, 3, 10, 18, 0, 0, 0, loc)
		got, err := checker.isBeforeCutoffAt(target, now)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !got {
			t.Errorf("expected true (update allowed): request=2026-03-10 18:00, target=%s", target)
		}
	})

	t.Run("after cutoff (22:00) — should be denied", func(t *testing.T) {
		now := time.Date(2026, 3, 10, 22, 0, 0, 0, loc)
		got, err := checker.isBeforeCutoffAt(target, now)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got {
			t.Errorf("expected false (update denied): request=2026-03-10 22:00, target=%s", target)
		}
	})
}

func TestScenario_RequestMar10_2300_TargetMar10(t *testing.T) {
	checker := mustCutoffChecker(t)
	loc := dhakaLoc(t)

	now := time.Date(2026, 3, 10, 23, 0, 0, 0, loc)
	const target = "2026-03-10"

	got, err := checker.isBeforeCutoffAt(target, now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got {
		t.Errorf("expected false: same-day update at 23:00 must be denied for target=%s", target)
	}
}

func TestMultiDayLookahead(t *testing.T) {
	checker := mustCutoffChecker(t)
	loc := dhakaLoc(t)

	now := time.Date(2026, 3, 10, 10, 0, 0, 0, loc)

	tests := []struct {
		name   string
		target string
		wantOk bool
	}{
		{"1 day ahead (Mar 11) — allowed", "2026-03-11", true},
		{"2 days ahead (Mar 12) — allowed", "2026-03-12", true},
		{"3 days ahead (Mar 13) — allowed", "2026-03-13", true},
		{"7 days ahead (Mar 17) — allowed (boundary)", "2026-03-17", true},
		{"8 days ahead (Mar 18) — denied by cap", "2026-03-18", false},
		{"14 days ahead (Mar 24) — denied by cap", "2026-03-24", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := checker.isBeforeCutoffAt(tc.target, now)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.wantOk {
				t.Errorf("CutoffChecker.isBeforeCutoffAt(%q, 2026-03-10 10:00) = %v; want %v",
					tc.target, got, tc.wantOk)
			}
		})
	}
}

func TestPostCutoff_StillAllowsFurtherDays(t *testing.T) {
	checker := mustCutoffChecker(t)
	loc := dhakaLoc(t)

	now := time.Date(2026, 3, 10, 23, 45, 0, 0, loc)

	tests := []struct {
		name   string
		target string
		wantOk bool
	}{
		{"Mar 11 — denied (cutoff was Mar 10 21:00)", "2026-03-11", false},
		{"Mar 12 — allowed (cutoff is Mar 11 21:00)", "2026-03-12", true},
		{"Mar 13 — allowed (cutoff is Mar 12 21:00)", "2026-03-13", true},
		{"Mar 17 — allowed (7-day boundary)", "2026-03-17", true},
		{"Mar 18 — denied (8 days ahead)", "2026-03-18", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := checker.isBeforeCutoffAt(tc.target, now)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.wantOk {
				t.Errorf("CutoffChecker.isBeforeCutoffAt(%q, 2026-03-10 23:45) = %v; want %v",
					tc.target, got, tc.wantOk)
			}
		})
	}
}

func TestIsBeforeCutoff_InvalidTimezone(t *testing.T) {
	_, err := NewCutoffChecker(CutoffConfig{Timezone: "NotAReal/Timezone"})
	if err == nil {
		t.Fatal("expected error for invalid timezone; got nil")
	}
}

func mustCutoffChecker(t *testing.T) *CutoffChecker {
	t.Helper()
	checker, err := NewCutoffChecker(CutoffConfig{CutoffTime: "21:00", Timezone: "Asia/Dhaka"})
	if err != nil {
		t.Fatalf("NewCutoffChecker() returned unexpected error: %v", err)
	}
	return checker
}
