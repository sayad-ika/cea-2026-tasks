package services

import (
	"context"
	"testing"

	"github.com/sayad-ika/craftsbite/internal/repository"
)

func TestGetTeamSummary_NoMembers(t *testing.T) {
	store := &noErrStore{
		mockDayScheduleReader: noErrDayReader(nil, nil),
		mockParticipationReader: &mockParticipationReader{
			getByUserDateFn: func(_ context.Context, _, _ string) ([]repository.MealParticipation, error) { return nil, nil },
			getByDateFn:     func(_ context.Context, _ string) ([]repository.MealParticipation, error) { return nil, nil },
		},
		mockLocationReader: noErrLocationReader(nil),
		mockUserReader:     noErrUserReader(nil),
		mockTeamReader: noErrTeamReader(nil, nil),
	}

	summary, err := GetTeamSummary(context.Background(), store, "t1", "2026-04-25")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.MemberCount != 0 {
		t.Errorf("MemberCount = %d, want 0", summary.MemberCount)
	}
}

func TestGetTeamSummary_WithMembers(t *testing.T) {
	users := []repository.User{
		{ID: "u1", Name: "Alice"},
		{ID: "u2", Name: "Bob"},
	}
	members := []repository.TeamMember{
		{TeamID: "t1", UserID: "u1"},
		{TeamID: "t1", UserID: "u2"},
	}
	meals := []string{"lunch", "snacks"}
	participations := []repository.MealParticipation{
		{UserID: "u1", MealType: "lunch", IsParticipating: true},
		{UserID: "u1", MealType: "snacks", IsParticipating: false},
		{UserID: "u2", MealType: "lunch", IsParticipating: true},
		{UserID: "u2", MealType: "snacks", IsParticipating: true},
	}

	store := &noErrStore{
		mockDayScheduleReader: &mockDayScheduleReader{
			getDayFn: func(_ context.Context, _ string) (*repository.DaySchedule, error) {
				return &repository.DaySchedule{DayStatus: "normal", AvailableMeals: meals}, nil
			},
			getAvailableMealsFn: func(_ context.Context, _ string) ([]string, error) { return meals, nil },
		},
		mockParticipationReader: &mockParticipationReader{
			getByUserDateFn: func(_ context.Context, userID, _ string) ([]repository.MealParticipation, error) {
				var result []repository.MealParticipation
				for _, p := range participations {
					if p.UserID == userID {
						result = append(result, p)
					}
				}
				return result, nil
			},
			getByDateFn: func(_ context.Context, _ string) ([]repository.MealParticipation, error) { return participations, nil },
		},
		mockLocationReader: &mockLocationReader{
			getFn: func(_ context.Context, userID, _ string) (*repository.WorkLocation, error) {
				if userID == "u2" {
					return &repository.WorkLocation{Location: "wfh"}, nil
				}
				return &repository.WorkLocation{Location: "office"}, nil
			},
			getByDateFn: func(_ context.Context, _ string) ([]repository.WorkLocation, error) { return nil, nil },
		},
		mockUserReader: noErrUserReader(users),
		mockTeamReader: noErrTeamReader([]repository.Team{{ID: "t1", Name: "Team1"}}, members),
	}

	summary, err := GetTeamSummary(context.Background(), store, "t1", "2026-04-25")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.MemberCount != 2 {
		t.Errorf("MemberCount = %d, want 2", summary.MemberCount)
	}
	if summary.WFHCount != 1 {
		t.Errorf("WFHCount = %d, want 1", summary.WFHCount)
	}
	if summary.MealCounts["lunch"] != 2 {
		t.Errorf("lunch count = %d, want 2", summary.MealCounts["lunch"])
	}
	if summary.MealCounts["snacks"] != 1 {
		t.Errorf("snacks count = %d, want 1 (u2 opted in, u1 opted out default)", summary.MealCounts["snacks"])
	}
}

func TestGetTeamSummary_NoAvailableMeals_UsesParticipation(t *testing.T) {
	users := []repository.User{
		{ID: "u1", Name: "Alice"},
	}
	members := []repository.TeamMember{
		{TeamID: "t1", UserID: "u1"},
	}
	participations := []repository.MealParticipation{
		{UserID: "u1", MealType: "iftar", IsParticipating: true},
	}

	store := &noErrStore{
		mockDayScheduleReader: &mockDayScheduleReader{
			getDayFn:            func(_ context.Context, _ string) (*repository.DaySchedule, error) { return nil, nil },
			getAvailableMealsFn: func(_ context.Context, _ string) ([]string, error) { return nil, nil },
		},
		mockParticipationReader: &mockParticipationReader{
			getByUserDateFn: func(_ context.Context, _, _ string) ([]repository.MealParticipation, error) {
				return participations, nil
			},
			getByDateFn: func(_ context.Context, _ string) ([]repository.MealParticipation, error) {
				return participations, nil
			},
		},
		mockLocationReader: noErrLocationReader(nil),
		mockUserReader:     noErrUserReader(users),
		mockTeamReader:     noErrTeamReader(nil, members),
	}

	summary, err := GetTeamSummary(context.Background(), store, "t1", "2026-04-25")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.MealCounts["iftar"] != 1 {
		t.Errorf("iftar count = %d, want 1", summary.MealCounts["iftar"])
	}
}

func TestGetTeamSummary_GetMembersError(t *testing.T) {
	store := &noErrStore{
		mockDayScheduleReader: noErrDayReader(nil, nil),
		mockParticipationReader: &mockParticipationReader{
			getByUserDateFn: func(_ context.Context, _, _ string) ([]repository.MealParticipation, error) { return nil, nil },
			getByDateFn:     func(_ context.Context, _ string) ([]repository.MealParticipation, error) { return nil, nil },
		},
		mockLocationReader: noErrLocationReader(nil),
		mockUserReader:     noErrUserReader(nil),
		mockTeamReader: &mockTeamReader{
			getByIDFn:      func(_ context.Context, _ string) (*repository.Team, error) { return nil, nil },
			getMembersFn:   func(_ context.Context, _ string) ([]repository.TeamMember, error) { return nil, nil },
			findByLeadIDFn: func(_ context.Context, _ string) ([]repository.Team, error) { return nil, nil },
		},
	}

	_, err := GetTeamSummary(context.Background(), store, "t1", "2026-04-25")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
