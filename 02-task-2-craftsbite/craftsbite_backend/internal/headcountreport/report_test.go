package headcountreport

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/sayad-ika/craftsbite/internal/discord"
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

	if len(msg.Embeds) != 1 {
		t.Fatalf("expected 1 compact embed, got %d", len(msg.Embeds))
	}
	embed := msg.Embeds[0]
	if embed.Title != "Daily Headcount Summary" {
		t.Errorf("embed title = %q, want %q", embed.Title, "Daily Headcount Summary")
	}
	if embed.Color != discord.BrandColor {
		t.Errorf("embed color = %d, want %d", embed.Color, discord.BrandColor)
	}
	if embed.Description != "2026-04-21" {
		t.Errorf("description = %q, want date", embed.Description)
	}
	if hasEmbedField(embed.Fields, "Engineering", "") {
		t.Error("did not expect team-specific field in compact summary")
	}
	if !hasEmbedField(embed.Fields, "Office / WFH", "7 / 3") {
		t.Error("expected compact Office / WFH field")
	}
	if !hasEmbedField(embed.Fields, "Total headcount", "10") {
		t.Error("expected Total headcount field")
	}
	if !hasEmbedField(embed.Fields, "Lunch", "6 confirmed") {
		t.Error("expected compact meal summary field")
	}
}

func TestBuildDiscordMessage_EmptyResult(t *testing.T) {
	result := &services.HeadcountResult{
		Date:       "2026-04-21",
		DayStatus:  "",
		TotalUsers: 0,
	}

	msg := BuildDiscordMessage(result)

	if len(msg.Embeds) != 1 {
		t.Fatalf("expected exactly 1 embed for empty result, got %d", len(msg.Embeds))
	}
	if msg.Embeds[0].Description != "2026-04-21" {
		t.Errorf("description = %q, want date", msg.Embeds[0].Description)
	}
	if got := len(msg.Embeds[0].Fields); got != 2 {
		t.Errorf("expected 2 summary fields (Total headcount + Office / WFH), got %d", got)
	}
}

func TestBuildGChatCard(t *testing.T) {
	result := &services.HeadcountResult{
		Date:       "2026-04-21",
		DayStatus:  "normal",
		DayReason:  "Lunch moved to Level 12",
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
	cardStr := string(card)
	if strings.Contains(cardStr, "Design") {
		t.Fatal("did not expect team-specific content in compact card")
	}
	if !strings.Contains(cardStr, "Daily Headcount Summary") {
		t.Fatal("expected compact title")
	}
	if !strings.Contains(cardStr, "Office / WFH") {
		t.Fatal("expected Office / WFH row")
	}
	if !strings.Contains(cardStr, "Lunch moved to Level 12") {
		t.Fatal("expected day reason note in compact card")
	}
}

func TestBuildScheduledDiscordMessage(t *testing.T) {
	result := &services.HeadcountResult{
		Date:       "2026-04-21",
		DayStatus:  "normal",
		DayReason:  "Lunch moved to Level 12",
		TotalUsers: 8,
		LocationCounts: services.LocationCount{
			Office: 5,
			WFH:    3,
		},
		MealCounts: map[string]services.MealCount{
			"lunch": {MealType: "lunch", OptedIn: 5, OptedOut: 3},
		},
		Teams: []services.TeamHeadcount{{TeamName: "Design", MemberCount: 3}},
	}

	msg := BuildScheduledDiscordMessage(result)
	if len(msg.Embeds) != 1 {
		t.Fatalf("expected 1 compact embed, got %d", len(msg.Embeds))
	}
	embed := msg.Embeds[0]
	if embed.Title != "Daily Headcount Summary" {
		t.Fatalf("title = %q, want %q", embed.Title, "Daily Headcount Summary")
	}
	if embed.Description != "2026-04-21" {
		t.Fatalf("description = %q, want date", embed.Description)
	}
	for _, field := range embed.Fields {
		if field.Name == "Design" {
			t.Fatal("did not expect team-specific field in scheduled summary")
		}
	}
	if !hasEmbedField(embed.Fields, "Office / WFH", "5 / 3") {
		t.Fatal("expected compact Office / WFH field")
	}
	if !hasEmbedField(embed.Fields, "Lunch", "5 confirmed") {
		t.Fatal("expected compact meal summary field")
	}
	if !hasEmbedField(embed.Fields, "Note", "Lunch moved to Level 12") {
		t.Fatal("expected day note in scheduled summary")
	}
}

func TestBuildScheduledGChatCard(t *testing.T) {
	result := &services.HeadcountResult{
		Date:       "2026-04-21",
		DayStatus:  "normal",
		DayReason:  "Lunch moved to Level 12",
		TotalUsers: 8,
		LocationCounts: services.LocationCount{
			Office: 5,
			WFH:    3,
		},
		MealCounts: map[string]services.MealCount{
			"lunch": {MealType: "lunch", OptedIn: 5, OptedOut: 3},
		},
		Teams: []services.TeamHeadcount{{TeamName: "Design", MemberCount: 3}},
	}

	card, err := BuildScheduledGChatCard(result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !json.Valid(card) {
		t.Fatal("expected valid JSON output")
	}
	if strings.Contains(string(card), "Design") {
		t.Fatal("did not expect team-specific content in scheduled gchat summary")
	}
	if !strings.Contains(string(card), "Daily Headcount Summary") {
		t.Fatal("expected compact scheduled title")
	}
	if !strings.Contains(string(card), "Office / WFH") {
		t.Fatal("expected Office / WFH row")
	}
	if !strings.Contains(string(card), "Lunch moved to Level 12") {
		t.Fatal("expected note content in scheduled gchat summary")
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
		"dinner":    {MealType: "dinner"},
		"breakfast": {MealType: "breakfast"},
		"lunch":     {MealType: "lunch"},
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

func hasEmbedField(fields []discord.EmbedField, name, value string) bool {
	for _, field := range fields {
		if field.Name == name && field.Value == value {
			return true
		}
	}
	return false
}
