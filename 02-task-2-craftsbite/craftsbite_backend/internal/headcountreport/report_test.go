package headcountreport

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/sayad-ika/craftsbite/internal/services"
)

func TestBuildDiscordMessage(t *testing.T) {
	result := &services.HeadcountResult{
		Date:       "2026-04-21",
		DayStatus:  "normal",
		TotalUsers: 10,
		LocationCounts: services.LocationCount{
			Office: 7,
			WFH:    3,
		},
		MealCounts: map[string]services.MealCount{
			"lunch": {MealType: "lunch", OptedIn: 6, OptedOut: 4},
		},
		Teams: []services.TeamHeadcount{
			{
				TeamName:    "Engineering",
				MemberCount: 5,
				LocationCounts: services.LocationCount{
					Office: 4,
					WFH:    1,
				},
				MealCounts: map[string]services.MealCount{
					"lunch": {MealType: "lunch", OptedIn: 3, OptedOut: 2},
				},
			},
		},
	}

	msg := BuildDiscordMessage(result)

	if !strings.Contains(msg, "Headcount") {
		t.Error("expected message to contain 'Headcount'")
	}
	if !strings.Contains(msg, "2026-04-21") {
		t.Error("expected message to contain date")
	}
	if !strings.Contains(msg, "10") {
		t.Error("expected message to contain total users")
	}
	if !strings.Contains(msg, "Engineering") {
		t.Error("expected message to contain team name")
	}
}

func TestBuildDiscordMessage_EmptyResult(t *testing.T) {
	result := &services.HeadcountResult{
		Date:       "2026-04-21",
		DayStatus:  "",
		TotalUsers: 0,
	}

	msg := BuildDiscordMessage(result)

	if !strings.Contains(msg, "2026-04-21") {
		t.Error("expected message to contain date")
	}
	if strings.Contains(msg, "Overall Meals") {
		t.Error("expected no meals section for empty meal counts")
	}
	if strings.Contains(msg, "By Team") {
		t.Error("expected no teams section for empty teams")
	}
}

func TestBuildGChatCard(t *testing.T) {
	result := &services.HeadcountResult{
		Date:       "2026-04-21",
		DayStatus:  "normal",
		TotalUsers: 8,
		LocationCounts: services.LocationCount{
			Office: 5,
			WFH:    3,
		},
		MealCounts: map[string]services.MealCount{
			"lunch": {MealType: "lunch", OptedIn: 5, OptedOut: 3},
		},
		Teams: []services.TeamHeadcount{
			{
				TeamName:    "Design",
				MemberCount: 3,
				LocationCounts: services.LocationCount{
					Office: 2,
					WFH:    1,
				},
				MealCounts: map[string]services.MealCount{
					"lunch": {MealType: "lunch", OptedIn: 2, OptedOut: 1},
				},
			},
		},
	}

	card, err := BuildGChatCard(result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !json.Valid(card) {
		t.Error("expected valid JSON output")
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(card, &raw); err != nil {
		t.Fatalf("failed to unmarshal card: %v", err)
	}
}

func TestDayStatusLabel(t *testing.T) {
	tests := []struct {
		status string
		reason string
		want   string
	}{
		{"", "", "📅 Normal Day"},
		{"normal", "", "📅 Normal Day"},
		{"holiday", "", "🎉 Holiday"},
		{"office_closed", "", "🔒 Office Closed"},
		{"event_day", "", "🎪 Event Day"},
		{"wfh_day", "", "🏠 WFH Day"},
		{"normal", "Team outing", "📅 Normal Day — Team outing"},
		{"unknown_status", "", "📅 Unknown Status"},
	}

	for _, tt := range tests {
		got := DayStatusLabel(tt.status, tt.reason)
		if got != tt.want {
			t.Errorf("DayStatusLabel(%q, %q) = %q, want %q", tt.status, tt.reason, got, tt.want)
		}
	}
}

func TestDisplayDayStatus(t *testing.T) {
	tests := []struct {
		status string
		want   string
	}{
		{"normal", "📅 Normal Day"},
		{"office_closed", "🔒 Office Closed"},
		{"govt_holiday", "🎉 Government Holiday"},
		{"celebration", "🎊 Celebration"},
		{"weekend", "🏖 Weekend"},
		{"event_day", "🎪 Event Day"},
		{"something_else", "something_else"},
	}

	for _, tt := range tests {
		got := DisplayDayStatus(tt.status)
		if got != tt.want {
			t.Errorf("DisplayDayStatus(%q) = %q, want %q", tt.status, got, tt.want)
		}
	}
}

func TestDisplayMealName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"lunch", "Lunch"},
		{"afternoon_snacks", "Afternoon Snacks"},
		{"breakfast", "Breakfast"},
		{"", ""},
	}

	for _, tt := range tests {
		got := DisplayMealName(tt.input)
		if got != tt.want {
			t.Errorf("DisplayMealName(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestFormatMealList(t *testing.T) {
	got := FormatMealList([]string{"lunch", "afternoon_snacks"})
	want := "Lunch, Afternoon Snacks"
	if got != want {
		t.Errorf("FormatMealList() = %q, want %q", got, want)
	}
}

func TestSortedMealCountKeys(t *testing.T) {
	m := map[string]services.MealCount{
		"dinner":   {MealType: "dinner"},
		"breakfast": {MealType: "breakfast"},
		"lunch":    {MealType: "lunch"},
	}

	keys := SortedMealCountKeys(m)
	if len(keys) != 3 {
		t.Fatalf("expected 3 keys, got %d", len(keys))
	}
	for i := 1; i < len(keys); i++ {
		if keys[i] < keys[i-1] {
			t.Errorf("keys not sorted: %v", keys)
		}
	}
}
