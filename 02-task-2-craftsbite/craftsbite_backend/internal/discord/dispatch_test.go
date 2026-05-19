package discord_test

import (
	"testing"

	"github.com/sayad-ika/craftsbite/internal/discord"
)

func TestCheckPermission(t *testing.T) {
	tests := []struct {
		command string
		role    string
		want    bool
	}{
		{"meal", "employee", true},
		{"help", "employee", true},
		{"help", "admin", true},
		{"meal", "admin", true},
		{"meal", "logistics", true},
		{"meal", "team_lead", true},
		{"headcount", "admin", true},
		{"headcount", "logistics", true},
		{"headcount", "employee", false},
		{"headcount", "team_lead", false},
		{"schedule-day", "admin", true},
		{"schedule-day", "employee", false},
		{"admin-init", "admin", true},
		{"admin-init", "team_lead", false},
		{"unknown-cmd", "admin", false},
		{"meal", "", false},
	}

	for _, tc := range tests {
		got := discord.CheckPermission(tc.command, tc.role)
		if got != tc.want {
			t.Errorf("CheckPermission(%q, %q) = %v, want %v", tc.command, tc.role, got, tc.want)
		}
	}
}

func TestIsKnownCommand(t *testing.T) {
	tests := []struct {
		command string
		want    bool
	}{
		{"meal", true},
		{"help", true},
		{"location", true},
		{"status", true},
		{"override", true},
		{"team-summary", true},
		{"headcount", true},
		{"schedule-day", true},
		{"admin", true},
		{"admin-init", true},
		{"unknown-cmd", false},
	}

	for _, tc := range tests {
		got := discord.IsKnownCommand(tc.command)
		if got != tc.want {
			t.Errorf("IsKnownCommand(%q) = %v, want %v", tc.command, got, tc.want)
		}
	}
}
