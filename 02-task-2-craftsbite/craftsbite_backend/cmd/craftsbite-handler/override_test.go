package main

import (
	"context"
	"errors"
	"testing"

	"github.com/sayad-ika/craftsbite/internal/dateutil"
	"github.com/sayad-ika/craftsbite/internal/payload"
	"github.com/sayad-ika/craftsbite/internal/repository"
)

type overrideTestStore struct {
	users           []repository.User
	teams           []repository.Team
	membersByTeam   map[string][]repository.TeamMember
	day             *repository.DaySchedule
	meals           []string
	participations  []repository.MealParticipation
	workLocation    *repository.WorkLocation
	auditEntries    []repository.AuditEntry
	upserts         []repository.MealParticipation
	locationUpserts []repository.WorkLocation
	auditErr        error
}

func (s *overrideTestStore) ListActiveUsers(ctx context.Context) ([]repository.User, error) {
	return s.users, nil
}

func (s *overrideTestStore) FindTeamsByLeadID(ctx context.Context, leadUserID string) ([]repository.Team, error) {
	return s.teams, nil
}

func (s *overrideTestStore) GetTeamMembers(ctx context.Context, teamID string) ([]repository.TeamMember, error) {
	return s.membersByTeam[teamID], nil
}

func (s *overrideTestStore) GetDay(ctx context.Context, date string) (*repository.DaySchedule, error) {
	return s.day, nil
}

func (s *overrideTestStore) GetAvailableMeals(ctx context.Context, date string) ([]string, error) {
	return s.meals, nil
}

func (s *overrideTestStore) GetParticipationsByUserDate(ctx context.Context, userID, date string) ([]repository.MealParticipation, error) {
	return s.participations, nil
}

func (s *overrideTestStore) GetParticipationsByDate(ctx context.Context, date string) ([]repository.MealParticipation, error) {
	return s.participations, nil
}

func (s *overrideTestStore) UpsertParticipation(ctx context.Context, p repository.MealParticipation) error {
	s.upserts = append(s.upserts, p)
	return nil
}

func (s *overrideTestStore) GetWorkLocation(ctx context.Context, userID, date string) (*repository.WorkLocation, error) {
	return s.workLocation, nil
}

func (s *overrideTestStore) GetWorkLocationsByDate(ctx context.Context, date string) ([]repository.WorkLocation, error) {
	return nil, nil
}

func (s *overrideTestStore) UpsertWorkLocation(ctx context.Context, wl repository.WorkLocation) error {
	s.locationUpserts = append(s.locationUpserts, wl)
	return nil
}

func (s *overrideTestStore) WriteAuditEntry(ctx context.Context, entry repository.AuditEntry) error {
	s.auditEntries = append(s.auditEntries, entry)
	return s.auditErr
}

func overrideDateParser(t *testing.T) *dateutil.DateParser {
	t.Helper()
	parser, err := dateutil.NewDateParser("Asia/Dhaka")
	if err != nil {
		t.Fatalf("date parser: %v", err)
	}
	return parser
}

func TestExecuteOverrideCommand_AdminMealToggle(t *testing.T) {
	store := &overrideTestStore{
		users:          []repository.User{{ID: "u1", Email: "alice@example.com", Active: true}},
		meals:          []string{"lunch"},
		participations: []repository.MealParticipation{{UserID: "u1", Date: "2026-05-10", MealType: "lunch", IsParticipating: true}},
	}
	event := payload.CommandEvent{
		UserID:      "admin-1",
		Role:        "admin",
		CommandName: "override",
		Options:     []byte(`{"target":"alice@example.com","entry":"meal","date":"2026-05-10","meal":"lunch","reason":"Forgot to update"}`),
	}

	result, err := executeOverrideCommand(context.Background(), store, overrideDateParser(t), event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Summary != "Lunch -> Opted out" {
		t.Fatalf("summary = %q, want toggled meal summary", result.Summary)
	}
	if len(store.upserts) != 1 || store.upserts[0].IsParticipating {
		t.Fatal("expected meal override to toggle to opted out")
	}
	if store.upserts[0].OverrideBy != "admin-1" || store.upserts[0].OverrideReason != "Forgot to update" {
		t.Fatal("expected override metadata to be written")
	}
	if len(store.auditEntries) != 1 {
		t.Fatal("expected audit entry for meal override")
	}
	if store.auditEntries[0].Action != "UPDATE" {
		t.Fatalf("audit action = %q, want UPDATE", store.auditEntries[0].Action)
	}
	if store.auditEntries[0].OldValue == "" {
		t.Fatal("expected old value for update audit entry")
	}
}

func TestExecuteOverrideCommand_TeamLeadOutsideTeamDenied(t *testing.T) {
	store := &overrideTestStore{
		users: []repository.User{{ID: "u1", Email: "alice@example.com", Active: true}},
		teams: []repository.Team{{ID: "t1", TeamLeadID: "lead-1"}},
		membersByTeam: map[string][]repository.TeamMember{
			"t1": {{TeamID: "t1", UserID: "someone-else"}},
		},
	}
	event := payload.CommandEvent{
		UserID:      "lead-1",
		Role:        "team_lead",
		CommandName: "override",
		Options:     []byte(`{"target":"alice@example.com","entry":"location","date":"2026-05-10","reason":"Coverage update"}`),
	}

	_, err := executeOverrideCommand(context.Background(), store, overrideDateParser(t), event)
	if err == nil {
		t.Fatal("expected authorization error")
	}
}

func TestExecuteOverrideCommand_LocationToggleFromNotSetDefaultsToOffice(t *testing.T) {
	store := &overrideTestStore{
		users:        []repository.User{{ID: "u1", Email: "alice@example.com", Active: true}},
		workLocation: nil,
	}
	event := payload.CommandEvent{
		UserID:      "admin-1",
		Role:        "admin",
		CommandName: "override",
		Options:     []byte(`{"target":"alice@example.com","entry":"location","date":"2026-05-10","reason":"Set initial location"}`),
	}

	result, err := executeOverrideCommand(context.Background(), store, overrideDateParser(t), event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Summary != "Office" {
		t.Fatalf("summary = %q, want Office", result.Summary)
	}
	if len(store.locationUpserts) != 1 || store.locationUpserts[0].Location != "office" {
		t.Fatal("expected location override to default toggle to office")
	}
	if len(store.auditEntries) != 1 {
		t.Fatal("expected audit entry for location override")
	}
	if store.auditEntries[0].Action != "CREATE" {
		t.Fatalf("audit action = %q, want CREATE", store.auditEntries[0].Action)
	}
	if store.auditEntries[0].OldValue != "" {
		t.Fatal("expected empty old value for create audit entry")
	}
}

func TestExecuteOverrideCommand_NormalizesOverrideValueCase(t *testing.T) {
	store := &overrideTestStore{
		users:        []repository.User{{ID: "u1", Email: "alice@example.com", Active: true}},
		workLocation: &repository.WorkLocation{UserID: "u1", Date: "2026-05-10", Location: "office"},
	}
	event := payload.CommandEvent{
		UserID:      "admin-1",
		Role:        "admin",
		CommandName: "override",
		Options:     []byte(`{"target":" alice@example.com ","entry":"LOCATION","date":"2026-05-10","value":"WFH","reason":"Coverage"}`),
	}

	result, err := executeOverrideCommand(context.Background(), store, overrideDateParser(t), event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Summary != "WFH" {
		t.Fatalf("summary = %q, want WFH", result.Summary)
	}
	if len(store.locationUpserts) != 1 || store.locationUpserts[0].Location != "wfh" {
		t.Fatal("expected normalized location override value")
	}
}

func TestExecuteOverrideCommand_AuditFailureDoesNotFailOverride(t *testing.T) {
	store := &overrideTestStore{
		users:          []repository.User{{ID: "u1", Email: "alice@example.com", Active: true}},
		meals:          []string{"lunch"},
		participations: []repository.MealParticipation{{UserID: "u1", Date: "2026-05-10", MealType: "lunch", IsParticipating: true}},
		auditErr:       errors.New("audit unavailable"),
	}
	event := payload.CommandEvent{
		UserID:      "admin-1",
		Role:        "admin",
		CommandName: "override",
		Options:     []byte(`{"target":"alice@example.com","entry":"meal","date":"2026-05-10","meal":"lunch","reason":"Forgot to update"}`),
	}

	result, err := executeOverrideCommand(context.Background(), store, overrideDateParser(t), event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected successful override result")
	}
	if len(store.upserts) != 1 {
		t.Fatal("expected override write to succeed despite audit failure")
	}
}
