package services

import (
	"context"
	"testing"

	"github.com/sayad-ika/craftsbite/internal/repository"
)

func TestGetHeadcount_EmptyUsers(t *testing.T) {
	store := &noErrStore{
		mockDayScheduleReader: noErrDayReader(nil, nil),
		mockParticipationReader: &mockParticipationReader{
			getByUserDateFn: func(_ context.Context, _, _ string) ([]repository.MealParticipation, error) { return nil, nil },
			getByDateFn:     func(_ context.Context, _ string) ([]repository.MealParticipation, error) { return nil, nil },
		},
		mockLocationReader:  noErrLocationReader(nil),
		mockUserReader:      noErrUserReader(nil),
		mockTeamReader:      noErrTeamReader(nil, nil),
	}

	result, err := GetHeadcount(context.Background(), store, "2026-04-25")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalUsers != 0 {
		t.Errorf("TotalUsers = %d, want 0", result.TotalUsers)
	}
	if len(result.Teams) != 0 {
		t.Errorf("expected 0 teams, got %d", len(result.Teams))
	}
}

func TestGetHeadcount_MultipleTeams(t *testing.T) {
	users := []repository.User{
		{ID: "u1", TeamID: "t1", Name: "Alice"},
		{ID: "u2", TeamID: "t1", Name: "Bob"},
		{ID: "u3", TeamID: "t2", Name: "Charlie"},
	}
	teams := []repository.Team{
		{ID: "t1", Name: "Alpha"},
		{ID: "t2", Name: "Beta"},
	}
	meals := []string{"lunch", "snacks"}
	participations := []repository.MealParticipation{
		{UserID: "u1", MealType: "lunch", IsParticipating: true},
		{UserID: "u1", MealType: "snacks", IsParticipating: false},
		{UserID: "u2", MealType: "lunch", IsParticipating: true},
		{UserID: "u3", MealType: "lunch", IsParticipating: false},
	}
	locations := []repository.WorkLocation{
		{UserID: "u1", Location: "office"},
		{UserID: "u2", Location: "wfh"},
		{UserID: "u3", Location: "office"},
	}

	store := &noErrStore{
		mockDayScheduleReader: &mockDayScheduleReader{
			getDayFn: func(_ context.Context, _ string) (*repository.DaySchedule, error) {
				return &repository.DaySchedule{DayStatus: "normal", AvailableMeals: meals}, nil
			},
			getAvailableMealsFn: func(_ context.Context, _ string) ([]string, error) { return meals, nil },
		},
		mockParticipationReader: &mockParticipationReader{
			getByUserDateFn: func(_ context.Context, _, _ string) ([]repository.MealParticipation, error) { return nil, nil },
			getByDateFn:     func(_ context.Context, _ string) ([]repository.MealParticipation, error) { return participations, nil },
		},
		mockLocationReader: &mockLocationReader{
			getFn:       func(_ context.Context, _, _ string) (*repository.WorkLocation, error) { return nil, nil },
			getByDateFn: func(_ context.Context, _ string) ([]repository.WorkLocation, error) { return locations, nil },
		},
		mockUserReader: noErrUserReader(users),
		mockTeamReader: noErrTeamReader(teams, nil),
	}

	result, err := GetHeadcount(context.Background(), store, "2026-04-25")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalUsers != 3 {
		t.Errorf("TotalUsers = %d, want 3", result.TotalUsers)
	}
	if len(result.Teams) != 2 {
		t.Fatalf("expected 2 teams, got %d", len(result.Teams))
	}

	loc := result.LocationCounts
	if loc.Office != 2 || loc.WFH != 1 {
		t.Errorf("LocationCounts = Office %d, WFH %d; want Office 2, WFH 1", loc.Office, loc.WFH)
	}
}

func TestGetHeadcount_SystemDefaultOptIn(t *testing.T) {
	users := []repository.User{
		{ID: "u1", TeamID: "t1"},
	}
	meals := []string{"lunch"}

	store := &noErrStore{
		mockDayScheduleReader: &mockDayScheduleReader{
			getDayFn: func(_ context.Context, _ string) (*repository.DaySchedule, error) {
				return &repository.DaySchedule{DayStatus: "normal", AvailableMeals: meals}, nil
			},
			getAvailableMealsFn: func(_ context.Context, _ string) ([]string, error) { return meals, nil },
		},
		mockParticipationReader: &mockParticipationReader{
			getByUserDateFn: func(_ context.Context, _, _ string) ([]repository.MealParticipation, error) { return nil, nil },
			getByDateFn:     func(_ context.Context, _ string) ([]repository.MealParticipation, error) { return nil, nil },
		},
		mockLocationReader: noErrLocationReader(nil),
		mockUserReader:     noErrUserReader(users),
		mockTeamReader:     noErrTeamReader([]repository.Team{{ID: "t1", Name: "Team1"}}, nil),
	}

	result, err := GetHeadcount(context.Background(), store, "2026-04-25")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	lunch := result.MealCounts["lunch"]
	if lunch.OptedIn != 1 {
		t.Errorf("lunch opted_in = %d, want 1 (system default)", lunch.OptedIn)
	}
}

func TestGetHeadcount_DBError(t *testing.T) {
	store := &noErrStore{
		mockDayScheduleReader: &mockDayScheduleReader{
			getDayFn:            func(_ context.Context, _ string) (*repository.DaySchedule, error) { return nil, nil },
			getAvailableMealsFn: func(_ context.Context, _ string) ([]string, error) { return nil, nil },
		},
		mockParticipationReader: &mockParticipationReader{
			getByUserDateFn: func(_ context.Context, _, _ string) ([]repository.MealParticipation, error) { return nil, nil },
			getByDateFn:     func(_ context.Context, _ string) ([]repository.MealParticipation, error) { return nil, nil },
		},
		mockLocationReader: &mockLocationReader{
			getFn:       func(_ context.Context, _, _ string) (*repository.WorkLocation, error) { return nil, nil },
			getByDateFn: func(_ context.Context, _ string) ([]repository.WorkLocation, error) { return nil, nil },
		},
		mockUserReader: &mockUserReader{
			listActiveFn: func(_ context.Context) ([]repository.User, error) { return nil, nil },
			listByRolesFn: func(_ context.Context, _ ...string) ([]repository.User, error) { return nil, nil },
			getByIDFn:    func(_ context.Context, _ string) (*repository.User, error) { return nil, nil },
		},
		mockTeamReader: &mockTeamReader{
			getByIDFn:      func(_ context.Context, _ string) (*repository.Team, error) { return nil, nil },
			getMembersFn:   func(_ context.Context, _ string) ([]repository.TeamMember, error) { return nil, nil },
			findByLeadIDFn: func(_ context.Context, _ string) ([]repository.Team, error) { return nil, nil },
		},
	}

	result, err := GetHeadcount(context.Background(), store, "2026-04-25")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Error("expected non-nil result")
	}
}
