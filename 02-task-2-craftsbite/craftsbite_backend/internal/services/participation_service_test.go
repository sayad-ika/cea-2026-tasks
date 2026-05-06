package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/sayad-ika/craftsbite/internal/repository"
)

func TestGetUserMealStatus_Normal(t *testing.T) {
	dayRepo := &mockDayScheduleReader{
		getDayFn: func(_ context.Context, _ string) (*repository.DaySchedule, error) {
			return &repository.DaySchedule{DayStatus: "normal"}, nil
		},
		getAvailableMealsFn: func(_ context.Context, _ string) ([]string, error) {
			return []string{"lunch", "snacks"}, nil
		},
	}
	pRepo := &mockParticipationReader{
		getByUserDateFn: func(_ context.Context, _, _ string) ([]repository.MealParticipation, error) {
			return []repository.MealParticipation{
				{MealType: "lunch", IsParticipating: true},
				{MealType: "snacks", IsParticipating: false},
			}, nil
		},
		getByDateFn: func(_ context.Context, _ string) ([]repository.MealParticipation, error) {
			return nil, nil
		},
	}

	statuses, err := GetUserMealStatus(context.Background(), dayRepo, pRepo, "user-1", "2026-04-25")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(statuses) != 2 {
		t.Fatalf("expected 2 statuses, got %d", len(statuses))
	}
	if statuses[0].MealType != "lunch" || statuses[0].Status != "opted_in" {
		t.Errorf("lunch status = %q, want opted_in", statuses[0].Status)
	}
	if statuses[1].MealType != "snacks" || statuses[1].Status != "opted_out" {
		t.Errorf("snacks status = %q, want opted_out", statuses[1].Status)
	}
}

func TestGetUserMealStatus_NoMeals(t *testing.T) {
	dayRepo := &mockDayScheduleReader{
		getDayFn:            func(_ context.Context, _ string) (*repository.DaySchedule, error) { return nil, nil },
		getAvailableMealsFn: func(_ context.Context, _ string) ([]string, error) { return nil, nil },
	}
	pRepo := &mockParticipationReader{
		getByUserDateFn: func(_ context.Context, _, _ string) ([]repository.MealParticipation, error) { return nil, nil },
		getByDateFn:     func(_ context.Context, _ string) ([]repository.MealParticipation, error) { return nil, nil },
	}

	statuses, err := GetUserMealStatus(context.Background(), dayRepo, pRepo, "user-1", "2026-04-25")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(statuses) != 0 {
		t.Errorf("expected 0 statuses for no meals, got %d", len(statuses))
	}
}

func TestGetUserMealStatus_NoConfiguredMealsFallsBackToUserRecords(t *testing.T) {
	dayRepo := &mockDayScheduleReader{
		getDayFn:            func(_ context.Context, _ string) (*repository.DaySchedule, error) { return nil, nil },
		getAvailableMealsFn: func(_ context.Context, _ string) ([]string, error) { return nil, nil },
	}
	pRepo := &mockParticipationReader{
		getByUserDateFn: func(_ context.Context, _, _ string) ([]repository.MealParticipation, error) {
			return []repository.MealParticipation{
				{MealType: "snacks", IsParticipating: false},
				{MealType: "lunch", IsParticipating: true},
			}, nil
		},
		getByDateFn: func(_ context.Context, _ string) ([]repository.MealParticipation, error) { return nil, nil },
	}

	statuses, err := GetUserMealStatus(context.Background(), dayRepo, pRepo, "user-1", "2026-04-25")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(statuses) != 2 {
		t.Fatalf("expected 2 statuses from fallback records, got %d", len(statuses))
	}
	if statuses[0].MealType != "lunch" || statuses[0].Status != "opted_in" {
		t.Fatalf("statuses[0] = %+v, want lunch opted_in", statuses[0])
	}
	if statuses[1].MealType != "snacks" || statuses[1].Status != "opted_out" {
		t.Fatalf("statuses[1] = %+v, want snacks opted_out", statuses[1])
	}
}

func TestGetUserMealStatus_ScheduleError(t *testing.T) {
	dayRepo := &mockDayScheduleReader{
		getDayFn:            func(_ context.Context, _ string) (*repository.DaySchedule, error) { return nil, errors.New("db fail") },
		getAvailableMealsFn: func(_ context.Context, _ string) ([]string, error) { return nil, nil },
	}
	pRepo := &mockParticipationReader{
		getByUserDateFn: func(_ context.Context, _, _ string) ([]repository.MealParticipation, error) { return nil, nil },
		getByDateFn:     func(_ context.Context, _ string) ([]repository.MealParticipation, error) { return nil, nil },
	}

	_, err := GetUserMealStatus(context.Background(), dayRepo, pRepo, "user-1", "2026-04-25")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestUpdateParticipation_ValidOptIn(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Dhaka")
	tomorrow := time.Now().In(loc).AddDate(0, 0, 2).Format("2006-01-02")

	checker, _ := NewCutoffChecker(CutoffConfig{CutoffTime: "21:00", Timezone: "Asia/Dhaka"})

	var upserted []repository.MealParticipation
	dayRepo := &mockDayScheduleReader{
		getDayFn: func(_ context.Context, _ string) (*repository.DaySchedule, error) {
			return &repository.DaySchedule{DayStatus: "normal"}, nil
		},
		getAvailableMealsFn: func(_ context.Context, _ string) ([]string, error) {
			return []string{"lunch"}, nil
		},
	}
	pRepo := &mockParticipationWriter{
		mockParticipationReader: mockParticipationReader{
			getByUserDateFn: func(_ context.Context, _, _ string) ([]repository.MealParticipation, error) {
				return []repository.MealParticipation{{MealType: "lunch", IsParticipating: true}}, nil
			},
			getByDateFn: func(_ context.Context, _ string) ([]repository.MealParticipation, error) { return nil, nil },
		},
		upsertFn: func(_ context.Context, p repository.MealParticipation, _ time.Time) error {
			upserted = append(upserted, p)
			return nil
		},
	}

	statuses, err := UpdateParticipation(context.Background(), dayRepo, pRepo, "user-1", tomorrow, "lunch", true, checker)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(upserted) != 1 {
		t.Fatalf("expected 1 upsert, got %d", len(upserted))
	}
	if !upserted[0].IsParticipating {
		t.Error("expected IsParticipating = true")
	}
	if len(statuses) != 1 {
		t.Errorf("expected 1 status, got %d", len(statuses))
	}
}

func TestUpdateParticipation_PastDate(t *testing.T) {
	checker, _ := NewCutoffChecker(CutoffConfig{CutoffTime: "21:00", Timezone: "Asia/Dhaka"})

	dayRepo := &mockDayScheduleReader{
		getDayFn:            func(_ context.Context, _ string) (*repository.DaySchedule, error) { return nil, nil },
		getAvailableMealsFn: func(_ context.Context, _ string) ([]string, error) { return nil, nil },
	}
	pRepo := &mockParticipationWriter{
		mockParticipationReader: mockParticipationReader{
			getByUserDateFn: func(_ context.Context, _, _ string) ([]repository.MealParticipation, error) { return nil, nil },
			getByDateFn:     func(_ context.Context, _ string) ([]repository.MealParticipation, error) { return nil, nil },
		},
		upsertFn: func(_ context.Context, _ repository.MealParticipation, _ time.Time) error { return nil },
	}

	_, err := UpdateParticipation(context.Background(), dayRepo, pRepo, "user-1", "2020-01-01", "lunch", true, checker)
	if !errors.Is(err, ErrPastDate) {
		t.Fatalf("expected ErrPastDate, got: %v", err)
	}
}

func TestUpdateParticipation_NilCutoff(t *testing.T) {
	dayRepo := &mockDayScheduleReader{
		getDayFn:            func(_ context.Context, _ string) (*repository.DaySchedule, error) { return nil, nil },
		getAvailableMealsFn: func(_ context.Context, _ string) ([]string, error) { return nil, nil },
	}
	pRepo := &mockParticipationWriter{
		mockParticipationReader: mockParticipationReader{
			getByUserDateFn: func(_ context.Context, _, _ string) ([]repository.MealParticipation, error) { return nil, nil },
			getByDateFn:     func(_ context.Context, _ string) ([]repository.MealParticipation, error) { return nil, nil },
		},
		upsertFn: func(_ context.Context, _ repository.MealParticipation, _ time.Time) error { return nil },
	}

	_, err := UpdateParticipation(context.Background(), dayRepo, pRepo, "user-1", "2026-04-25", "lunch", true, nil)
	if err == nil {
		t.Fatal("expected error for nil cutoff, got nil")
	}
}

func TestUpdateParticipation_DayClosed(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Dhaka")
	tomorrow := time.Now().In(loc).AddDate(0, 0, 2).Format("2006-01-02")

	checker, _ := NewCutoffChecker(CutoffConfig{CutoffTime: "21:00", Timezone: "Asia/Dhaka"})

	dayRepo := &mockDayScheduleReader{
		getDayFn: func(_ context.Context, _ string) (*repository.DaySchedule, error) {
			return &repository.DaySchedule{DayStatus: "office_closed"}, nil
		},
		getAvailableMealsFn: func(_ context.Context, _ string) ([]string, error) { return []string{"lunch"}, nil },
	}
	pRepo := &mockParticipationWriter{
		mockParticipationReader: mockParticipationReader{
			getByUserDateFn: func(_ context.Context, _, _ string) ([]repository.MealParticipation, error) { return nil, nil },
			getByDateFn:     func(_ context.Context, _ string) ([]repository.MealParticipation, error) { return nil, nil },
		},
		upsertFn: func(_ context.Context, _ repository.MealParticipation, _ time.Time) error { return nil },
	}

	_, err := UpdateParticipation(context.Background(), dayRepo, pRepo, "user-1", tomorrow, "lunch", true, checker)
	if !errors.Is(err, ErrDayClosed) {
		t.Fatalf("expected ErrDayClosed, got: %v", err)
	}
}

func TestUpdateParticipation_NoMeals(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Dhaka")
	tomorrow := time.Now().In(loc).AddDate(0, 0, 2).Format("2006-01-02")

	checker, _ := NewCutoffChecker(CutoffConfig{CutoffTime: "21:00", Timezone: "Asia/Dhaka"})

	dayRepo := &mockDayScheduleReader{
		getDayFn:            func(_ context.Context, _ string) (*repository.DaySchedule, error) { return nil, nil },
		getAvailableMealsFn: func(_ context.Context, _ string) ([]string, error) { return []string{}, nil },
	}
	pRepo := &mockParticipationWriter{
		mockParticipationReader: mockParticipationReader{
			getByUserDateFn: func(_ context.Context, _, _ string) ([]repository.MealParticipation, error) { return nil, nil },
			getByDateFn:     func(_ context.Context, _ string) ([]repository.MealParticipation, error) { return nil, nil },
		},
		upsertFn: func(_ context.Context, _ repository.MealParticipation, _ time.Time) error { return nil },
	}

	_, err := UpdateParticipation(context.Background(), dayRepo, pRepo, "user-1", tomorrow, "lunch", true, checker)
	if !errors.Is(err, ErrNoMeals) {
		t.Fatalf("expected ErrNoMeals, got: %v", err)
	}
}

func TestUpdateParticipation_MealUnavailable(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Dhaka")
	tomorrow := time.Now().In(loc).AddDate(0, 0, 2).Format("2006-01-02")

	checker, _ := NewCutoffChecker(CutoffConfig{CutoffTime: "21:00", Timezone: "Asia/Dhaka"})

	dayRepo := &mockDayScheduleReader{
		getDayFn:            func(_ context.Context, _ string) (*repository.DaySchedule, error) { return nil, nil },
		getAvailableMealsFn: func(_ context.Context, _ string) ([]string, error) { return []string{"lunch"}, nil },
	}
	pRepo := &mockParticipationWriter{
		mockParticipationReader: mockParticipationReader{
			getByUserDateFn: func(_ context.Context, _, _ string) ([]repository.MealParticipation, error) { return nil, nil },
			getByDateFn:     func(_ context.Context, _ string) ([]repository.MealParticipation, error) { return nil, nil },
		},
		upsertFn: func(_ context.Context, _ repository.MealParticipation, _ time.Time) error { return nil },
	}

	_, err := UpdateParticipation(context.Background(), dayRepo, pRepo, "user-1", tomorrow, "snacks", true, checker)
	if !errors.Is(err, ErrMealUnavailable) {
		t.Fatalf("expected ErrMealUnavailable, got: %v", err)
	}
}

func TestUpdateParticipation_AllMeals(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Dhaka")
	tomorrow := time.Now().In(loc).AddDate(0, 0, 2).Format("2006-01-02")

	checker, _ := NewCutoffChecker(CutoffConfig{CutoffTime: "21:00", Timezone: "Asia/Dhaka"})

	var upsertCount int
	dayRepo := &mockDayScheduleReader{
		getDayFn: func(_ context.Context, _ string) (*repository.DaySchedule, error) {
			return &repository.DaySchedule{DayStatus: "normal"}, nil
		},
		getAvailableMealsFn: func(_ context.Context, _ string) ([]string, error) {
			return []string{"lunch", "snacks"}, nil
		},
	}
	pRepo := &mockParticipationWriter{
		mockParticipationReader: mockParticipationReader{
			getByUserDateFn: func(_ context.Context, _, _ string) ([]repository.MealParticipation, error) {
				return []repository.MealParticipation{
					{MealType: "lunch", IsParticipating: true},
					{MealType: "snacks", IsParticipating: true},
				}, nil
			},
			getByDateFn: func(_ context.Context, _ string) ([]repository.MealParticipation, error) { return nil, nil },
		},
		upsertFn: func(_ context.Context, _ repository.MealParticipation, _ time.Time) error {
			upsertCount++
			return nil
		},
	}

	statuses, err := UpdateParticipation(context.Background(), dayRepo, pRepo, "user-1", tomorrow, "all", true, checker)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if upsertCount != 2 {
		t.Errorf("expected 2 upserts, got %d", upsertCount)
	}
	if len(statuses) != 2 {
		t.Errorf("expected 2 statuses, got %d", len(statuses))
	}
}
