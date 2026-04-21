package dateutil

import (
	"testing"
	"time"
)

func TestParseDateWithDefaults(t *testing.T) {
	parser := mustDateParser(t)
	loc, _ := time.LoadLocation("Asia/Dhaka")
	now := time.Date(2026, 3, 19, 10, 0, 0, 0, loc)
	today := now.Format("2006-01-02")
	tomorrow := now.AddDate(0, 0, 1).Format("2006-01-02")

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "empty string defaults to tomorrow",
			input: "",
			want:  tomorrow,
		},
		{
			name:  "today keyword",
			input: "today",
			want:  today,
		},
		{
			name:  "tomorrow keyword",
			input: "tomorrow",
			want:  tomorrow,
		},
		{
			name:  "relative +1",
			input: "+1",
			want:  tomorrow,
		},
		{
			name:  "relative +3",
			input: "+3",
			want:  now.AddDate(0, 0, 3).Format("2006-01-02"),
		},
		{
			name:  "relative +7",
			input: "+7",
			want:  now.AddDate(0, 0, 7).Format("2006-01-02"),
		},
		{
			name:  "explicit date YYYY-MM-DD",
			input: "2026-03-25",
			want:  "2026-03-25",
		},
		{
			name:    "invalid date format",
			input:   "2026/03/25",
			wantErr: true,
		},
		{
			name:    "invalid relative format",
			input:   "+abc",
			wantErr: true,
		},
		{
			name:    "negative relative days",
			input:   "+-1",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parser.parseDateWithDefaultsAt(tt.input, now)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseDateWithDefaultsAt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("parseDateWithDefaultsAt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTodayInTimezone(t *testing.T) {
	parser := mustDateParser(t)
	loc, _ := time.LoadLocation("Asia/Dhaka")
	result := parser.todayInTimezoneAt(time.Date(2026, 3, 19, 10, 0, 0, 0, loc))
	if len(result) != 10 {
		t.Errorf("todayInTimezoneAt() returned invalid format: %s", result)
	}

	// Verify it's a valid date
	_, err := time.Parse("2006-01-02", result)
	if err != nil {
		t.Errorf("todayInTimezoneAt() returned invalid date: %s", result)
	}
	if result != "2026-03-19" {
		t.Errorf("todayInTimezoneAt() = %s, want 2026-03-19", result)
	}
}

func TestTomorrowInTimezone(t *testing.T) {
	parser := mustDateParser(t)
	loc, _ := time.LoadLocation("Asia/Dhaka")
	result := parser.tomorrowInTimezoneAt(time.Date(2026, 3, 19, 10, 0, 0, 0, loc))
	if len(result) != 10 {
		t.Errorf("tomorrowInTimezoneAt() returned invalid format: %s", result)
	}

	// Verify it's a valid date
	_, err := time.Parse("2006-01-02", result)
	if err != nil {
		t.Errorf("tomorrowInTimezoneAt() returned invalid date: %s", result)
	}
	if result != "2026-03-20" {
		t.Errorf("tomorrowInTimezoneAt() = %s, want 2026-03-20", result)
	}
}

func mustDateParser(t *testing.T) *DateParser {
	t.Helper()
	parser, err := NewDateParser("Asia/Dhaka")
	if err != nil {
		t.Fatalf("NewDateParser() returned unexpected error: %v", err)
	}
	return parser
}
