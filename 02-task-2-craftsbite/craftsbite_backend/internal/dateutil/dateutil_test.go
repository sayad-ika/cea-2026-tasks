package dateutil

import (
	"os"
	"testing"
	"time"
)

func TestParseDateWithDefaults(t *testing.T) {
	// Set timezone for consistent testing
	os.Setenv("TIMEZONE", "Asia/Dhaka")
	defer os.Unsetenv("TIMEZONE")

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
			got, err := ParseDateWithDefaults(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDateWithDefaults() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ParseDateWithDefaults() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTodayInTimezone(t *testing.T) {
	os.Setenv("TIMEZONE", "Asia/Dhaka")
	defer os.Unsetenv("TIMEZONE")

	result := TodayInTimezone()
	if len(result) != 10 {
		t.Errorf("TodayInTimezone() returned invalid format: %s", result)
	}

	// Verify it's a valid date
	_, err := time.Parse("2006-01-02", result)
	if err != nil {
		t.Errorf("TodayInTimezone() returned invalid date: %s", result)
	}
}

func TestTomorrowInTimezone(t *testing.T) {
	os.Setenv("TIMEZONE", "Asia/Dhaka")
	defer os.Unsetenv("TIMEZONE")

	result := TomorrowInTimezone()
	if len(result) != 10 {
		t.Errorf("TomorrowInTimezone() returned invalid format: %s", result)
	}

	// Verify it's a valid date
	_, err := time.Parse("2006-01-02", result)
	if err != nil {
		t.Errorf("TomorrowInTimezone() returned invalid date: %s", result)
	}
}
