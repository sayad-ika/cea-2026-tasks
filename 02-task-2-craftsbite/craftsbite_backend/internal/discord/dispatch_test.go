package discord_test

import (
	"testing"

	"github.com/sayad-ika/craftsbite/internal/config"
	"github.com/sayad-ika/craftsbite/internal/discord"
)

func TestCheckPermission(t *testing.T) {
	tests := []struct {
		command string
		role    string
		want    bool
	}{
		{"meal", "employee", true},
		{"meal", "admin", true},
		{"meal", "logistics", true},
		{"meal", "team_lead", true},
		{"headcount", "admin", true},
		{"headcount", "logistics", true},
		{"headcount", "employee", false},
		{"headcount", "team_lead", false},
		{"schedule-day", "admin", true},
		{"schedule-day", "employee", false},
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

func TestDispatch(t *testing.T) {
	cfg := &config.Config{
		LambdaSelfFunctionName:       "self-fn",
		LambdaManagementFunctionName: "mgmt-fn",
		LambdaOpsFunctionName:        "ops-fn",
	}

	tests := []struct {
		command string
		wantARN string
		wantOK  bool
	}{
		{"meal", "self-fn", true},
		{"location", "self-fn", true},
		{"status", "self-fn", true},
		{"override", "mgmt-fn", true},
		{"team-summary", "mgmt-fn", true},
		{"headcount", "ops-fn", true},
		{"schedule-day", "ops-fn", true},
		{"admin", "ops-fn", true},
		{"unknown-cmd", "", false},
	}

	for _, tc := range tests {
		arn, ok := discord.Dispatch(cfg, tc.command)
		if ok != tc.wantOK || arn != tc.wantARN {
			t.Errorf("Dispatch(%q) = (%q, %v), want (%q, %v)", tc.command, arn, ok, tc.wantARN, tc.wantOK)
		}
	}
}
