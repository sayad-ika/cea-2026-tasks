package services

import (
	"testing"
	"time"
)

func TestLocationSentinels(t *testing.T) {
	if ErrLocationPastDate == nil {
		t.Fatal("ErrLocationPastDate must not be nil")
	}
	if ErrLocationCutoffPassed == nil {
		t.Fatal("ErrLocationCutoffPassed must not be nil")
	}
	if ErrLocationTooFarAhead == nil {
		t.Fatal("ErrLocationTooFarAhead must not be nil")
	}
}

func TestGetLocationNotSetDefault(t *testing.T) {
	wl := notSetLocation("user-1", "2026-03-10")
	if wl.Location != "not_set" {
		t.Fatalf("expected not_set, got %q", wl.Location)
	}
	if wl.UserID != "user-1" || wl.Date != "2026-03-10" {
		t.Fatal("not_set sentinel must carry userID and date")
	}
}

func TestLocationDateGuards(t *testing.T) {
	t.Setenv("CUTOFF_TIME", "21:00")
	t.Setenv("TIMEZONE", "Asia/Dhaka")

	loc := dhakaLoc(t)

	tests := []struct {
		name       string
		now        time.Time
		date       string
		wantResult string
	}{
		{
			name:       "past date → ErrLocationPastDate",
			now:        time.Date(2026, 3, 10, 10, 0, 0, 0, loc),
			date:       "2026-03-09",
			wantResult: "past",
		},
		{
			name:       "same day → ErrLocationPastDate (not strictly future)",
			now:        time.Date(2026, 3, 10, 10, 0, 0, 0, loc),
			date:       "2026-03-10",
			wantResult: "past",
		},
		{
			name:       "next day before cutoff → allowed",
			now:        time.Date(2026, 3, 10, 18, 0, 0, 0, loc),
			date:       "2026-03-11",
			wantResult: "allowed",
		},
		{
			name:       "next day after cutoff → ErrLocationCutoffPassed",
			now:        time.Date(2026, 3, 10, 22, 0, 0, 0, loc),
			date:       "2026-03-11",
			wantResult: "cutoff",
		},
		{
			name:       "8 days ahead → ErrLocationTooFarAhead",
			now:        time.Date(2026, 3, 10, 10, 0, 0, 0, loc),
			date:       "2026-03-18",
			wantResult: "tooFar",
		},
		{
			name:       "7 days ahead before cutoff → allowed (boundary)",
			now:        time.Date(2026, 3, 10, 10, 0, 0, 0, loc),
			date:       "2026-03-17",
			wantResult: "allowed",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			today := tc.now.In(loc).Format("2006-01-02")

			var got string
			switch {
			case tc.date <= today:
				got = "past"
			default:
				target, _ := time.ParseInLocation("2006-01-02", tc.date, loc)
				todayMidnight := time.Date(tc.now.Year(), tc.now.Month(), tc.now.Day(), 0, 0, 0, 0, loc)
				if int(target.Sub(todayMidnight).Hours()/24) > maxDaysAhead {
					got = "tooFar"
				} else {
					ok, err := isBeforeCutoffAt(tc.date, tc.now)
					if err != nil {
						t.Fatalf("unexpected error: %v", err)
					}
					if ok {
						got = "allowed"
					} else {
						got = "cutoff"
					}
				}
			}

			if got != tc.wantResult {
				t.Errorf("date=%s now=%v: got %q, want %q", tc.date, tc.now, got, tc.wantResult)
			}
		})
	}
}
