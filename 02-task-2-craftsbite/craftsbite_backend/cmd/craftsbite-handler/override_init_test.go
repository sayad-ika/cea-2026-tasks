package main

import (
	"context"
	"strings"
	"testing"

	"github.com/sayad-ika/craftsbite/internal/payload"
	"github.com/sayad-ika/craftsbite/internal/repository"
)

func TestApplyOverrideInitDraft_AdminWholeTeamMeal(t *testing.T) {
	store := &overrideTestStore{
		users: []repository.User{
			{ID: "u1", Email: "alice@example.com", Name: "Alice", TeamID: "t1", Active: true},
			{ID: "u2", Email: "bob@example.com", Name: "Bob", TeamID: "t1", Active: true},
		},
		teams: []repository.Team{{ID: "t1", Name: "Platform", Active: true}},
		membersByTeam: map[string][]repository.TeamMember{
			"t1": {{TeamID: "t1", UserID: "u1"}, {TeamID: "t1", UserID: "u2"}},
		},
		meals: []string{"lunch"},
	}
	event := payload.CommandEvent{UserID: "admin-1", Role: "admin", CommandName: "override-init"}
	draft := overrideInitDraft{Date: "2026-05-10", TeamID: "t1", TargetMode: overrideInitTargetWholeTeam, Entry: "meal", Meal: "lunch", Value: "out", Reason: "Team offsite"}

	result, err := applyOverrideInitDraft(context.Background(), store, overrideDateParser(t), event, draft)
	if err != nil {
		t.Fatalf("applyOverrideInitDraft() error = %v", err)
	}
	if result.Succeeded != 2 || result.Failed != 0 || result.Requested != 2 {
		t.Fatalf("result = %+v, want 2 successful overrides", result)
	}
	if len(store.upserts) != 2 {
		t.Fatalf("upserts = %d, want 2", len(store.upserts))
	}
	for _, upsert := range store.upserts {
		if upsert.IsParticipating {
			t.Fatalf("upsert = %+v, want opted out", upsert)
		}
		if upsert.OverrideBy != "admin-1" || upsert.OverrideReason != "Team offsite" {
			t.Fatalf("upsert metadata = %+v", upsert)
		}
	}
}

func TestApplyOverrideInitDraft_TeamLeadOutsideScopedTeamDenied(t *testing.T) {
	store := &overrideTestStore{
		users: []repository.User{{ID: "u1", Email: "alice@example.com", TeamID: "t2", Active: true}},
		teams: []repository.Team{{ID: "t1", Name: "Owned", TeamLeadID: "lead-1", Active: true}},
		membersByTeam: map[string][]repository.TeamMember{
			"t1": {{TeamID: "t1", UserID: "u1"}},
		},
		meals: []string{"lunch"},
	}
	event := payload.CommandEvent{UserID: "lead-1", Role: "team_lead", CommandName: "override-init"}
	draft := overrideInitDraft{Date: "2026-05-10", TeamID: "t2", TargetMode: overrideInitTargetWholeTeam, Entry: "meal", Meal: "lunch", Value: "in", Reason: "Coverage"}

	_, err := applyOverrideInitDraft(context.Background(), store, overrideDateParser(t), event, draft)
	if err == nil || !strings.Contains(err.Error(), "Selected team is not available") {
		t.Fatalf("error = %v, want scoped team denial", err)
	}
}

func TestApplyOverrideInitDraft_SpecificMemberLocation(t *testing.T) {
	store := &overrideTestStore{
		users: []repository.User{
			{ID: "u1", Email: "alice@example.com", Name: "Alice", TeamID: "t1", Active: true},
			{ID: "u2", Email: "bob@example.com", Name: "Bob", TeamID: "t1", Active: true},
		},
		teams: []repository.Team{{ID: "t1", Name: "Platform", Active: true}},
		membersByTeam: map[string][]repository.TeamMember{
			"t1": {{TeamID: "t1", UserID: "u1"}, {TeamID: "t1", UserID: "u2"}},
		},
	}
	event := payload.CommandEvent{UserID: "admin-1", Role: "admin", CommandName: "override-init"}
	draft := overrideInitDraft{Date: "2026-05-10", TeamID: "t1", TargetMode: overrideInitTargetMembers, Members: []string{"u2"}, Entry: "location", Value: "wfh", Reason: "Remote workshop"}

	result, err := applyOverrideInitDraft(context.Background(), store, overrideDateParser(t), event, draft)
	if err != nil {
		t.Fatalf("applyOverrideInitDraft() error = %v", err)
	}
	if result.Succeeded != 1 || result.Requested != 1 {
		t.Fatalf("result = %+v, want one target", result)
	}
	if len(store.locationUpserts) != 1 || store.locationUpserts[0].UserID != "u2" || store.locationUpserts[0].Location != "wfh" {
		t.Fatalf("location upserts = %+v, want u2 wfh", store.locationUpserts)
	}
}

func TestApplyOverrideInitDraft_BulkRequiresExplicitValue(t *testing.T) {
	store := &overrideTestStore{
		users: []repository.User{{ID: "u1", Email: "alice@example.com", TeamID: "t1", Active: true}},
		teams: []repository.Team{{ID: "t1", Name: "Platform", Active: true}},
		membersByTeam: map[string][]repository.TeamMember{
			"t1": {{TeamID: "t1", UserID: "u1"}},
		},
		meals: []string{"lunch"},
	}
	event := payload.CommandEvent{UserID: "admin-1", Role: "admin", CommandName: "override-init"}
	draft := overrideInitDraft{Date: "2026-05-10", TeamID: "t1", TargetMode: overrideInitTargetWholeTeam, Entry: "meal", Meal: "lunch", Reason: "Missing value"}

	_, err := applyOverrideInitDraft(context.Background(), store, overrideDateParser(t), event, draft)
	if err == nil || !strings.Contains(err.Error(), "Choose whether the meal") {
		t.Fatalf("error = %v, want explicit value validation", err)
	}
	if len(store.upserts) != 0 {
		t.Fatalf("upserts = %+v, want no writes", store.upserts)
	}
}
