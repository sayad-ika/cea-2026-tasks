package services

import (
	"context"
	"errors"
	"testing"

	"github.com/sayad-ika/craftsbite/internal/repository"
)

func TestSetDaySchedule_ValidCreate(t *testing.T) {
	repo := &mockDayScheduleWriter{
		mockDayScheduleReader: mockDayScheduleReader{
			getDayFn:            func(_ context.Context, _ string) (*repository.DaySchedule, error) { return nil, nil },
			getAvailableMealsFn: func(_ context.Context, _ string) ([]string, error) { return nil, nil },
		},
		upsertFn: func(_ context.Context, _ repository.DaySchedule) error { return nil },
		deleteFn: func(_ context.Context, _ string) error { return nil },
	}

	input := SetDayScheduleInput{
		Date:           "2026-04-25",
		DayStatus:      "normal",
		AvailableMeals: []string{"lunch", "snacks"},
		Reason:         "test",
		SetBy:          "admin-1",
	}

	schedule, err := SetDaySchedule(context.Background(), repo, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if schedule.Date != input.Date {
		t.Errorf("Date = %q, want %q", schedule.Date, input.Date)
	}
	if schedule.DayStatus != input.DayStatus {
		t.Errorf("DayStatus = %q, want %q", schedule.DayStatus, input.DayStatus)
	}
}

func TestSetDaySchedule_InvalidDate(t *testing.T) {
	repo := &mockDayScheduleWriter{
		mockDayScheduleReader: mockDayScheduleReader{
			getDayFn:            func(_ context.Context, _ string) (*repository.DaySchedule, error) { return nil, nil },
			getAvailableMealsFn: func(_ context.Context, _ string) ([]string, error) { return nil, nil },
		},
		upsertFn: func(_ context.Context, _ repository.DaySchedule) error { return nil },
		deleteFn: func(_ context.Context, _ string) error { return nil },
	}

	_, err := SetDaySchedule(context.Background(), repo, SetDayScheduleInput{
		Date:      "not-a-date",
		DayStatus: "normal",
	})
	if !errors.Is(err, ErrInvalidDate) {
		t.Fatalf("expected ErrInvalidDate, got: %v", err)
	}
}

func TestSetDaySchedule_InvalidDayStatus(t *testing.T) {
	repo := &mockDayScheduleWriter{
		mockDayScheduleReader: mockDayScheduleReader{
			getDayFn:            func(_ context.Context, _ string) (*repository.DaySchedule, error) { return nil, nil },
			getAvailableMealsFn: func(_ context.Context, _ string) ([]string, error) { return nil, nil },
		},
		upsertFn: func(_ context.Context, _ repository.DaySchedule) error { return nil },
		deleteFn: func(_ context.Context, _ string) error { return nil },
	}

	_, err := SetDaySchedule(context.Background(), repo, SetDayScheduleInput{
		Date:      "2026-04-25",
		DayStatus: "invalid_status",
	})
	if !errors.Is(err, ErrInvalidDayStatus) {
		t.Fatalf("expected ErrInvalidDayStatus, got: %v", err)
	}
}

func TestSetDaySchedule_ClosedDayWithMeals(t *testing.T) {
	repo := &mockDayScheduleWriter{
		mockDayScheduleReader: mockDayScheduleReader{
			getDayFn:            func(_ context.Context, _ string) (*repository.DaySchedule, error) { return nil, nil },
			getAvailableMealsFn: func(_ context.Context, _ string) ([]string, error) { return nil, nil },
		},
		upsertFn: func(_ context.Context, _ repository.DaySchedule) error { return nil },
		deleteFn: func(_ context.Context, _ string) error { return nil },
	}

	_, err := SetDaySchedule(context.Background(), repo, SetDayScheduleInput{
		Date:           "2026-04-25",
		DayStatus:      "office_closed",
		AvailableMeals: []string{"lunch"},
	})
	if err == nil {
		t.Fatal("expected error for closed day with meals, got nil")
	}
}

func TestSetDaySchedule_InvalidMealType(t *testing.T) {
	repo := &mockDayScheduleWriter{
		mockDayScheduleReader: mockDayScheduleReader{
			getDayFn:            func(_ context.Context, _ string) (*repository.DaySchedule, error) { return nil, nil },
			getAvailableMealsFn: func(_ context.Context, _ string) ([]string, error) { return nil, nil },
		},
		upsertFn: func(_ context.Context, _ repository.DaySchedule) error { return nil },
		deleteFn: func(_ context.Context, _ string) error { return nil },
	}

	_, err := SetDaySchedule(context.Background(), repo, SetDayScheduleInput{
		Date:           "2026-04-25",
		DayStatus:      "normal",
		AvailableMeals: []string{"invalid_meal"},
	})
	if !errors.Is(err, ErrInvalidMealType) {
		t.Fatalf("expected ErrInvalidMealType, got: %v", err)
	}
}

func TestSetDaySchedule_UpsertError(t *testing.T) {
	repo := &mockDayScheduleWriter{
		mockDayScheduleReader: mockDayScheduleReader{
			getDayFn:            func(_ context.Context, _ string) (*repository.DaySchedule, error) { return nil, nil },
			getAvailableMealsFn: func(_ context.Context, _ string) ([]string, error) { return nil, nil },
		},
		upsertFn: func(_ context.Context, _ repository.DaySchedule) error { return errors.New("db error") },
		deleteFn: func(_ context.Context, _ string) error { return nil },
	}

	_, err := SetDaySchedule(context.Background(), repo, SetDayScheduleInput{
		Date:      "2026-04-25",
		DayStatus: "normal",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetDaySchedule(t *testing.T) {
	expected := &repository.DaySchedule{Date: "2026-04-25", DayStatus: "normal"}
	repo := &mockDayScheduleReader{
		getDayFn:            func(_ context.Context, date string) (*repository.DaySchedule, error) { return expected, nil },
		getAvailableMealsFn: func(_ context.Context, _ string) ([]string, error) { return nil, nil },
	}

	got, err := GetDaySchedule(context.Background(), repo, "2026-04-25")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Date != "2026-04-25" {
		t.Errorf("Date = %q, want %q", got.Date, "2026-04-25")
	}
}

func TestDeleteDaySchedule(t *testing.T) {
	called := false
	repo := &mockDayScheduleWriter{
		mockDayScheduleReader: mockDayScheduleReader{
			getDayFn:            func(_ context.Context, _ string) (*repository.DaySchedule, error) { return nil, nil },
			getAvailableMealsFn: func(_ context.Context, _ string) ([]string, error) { return nil, nil },
		},
		upsertFn: func(_ context.Context, _ repository.DaySchedule) error { return nil },
		deleteFn: func(_ context.Context, _ string) error { called = true; return nil },
	}

	err := DeleteDaySchedule(context.Background(), repo, "2026-04-25")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("DeleteDaySchedule was not called")
	}
}

func TestBulkSetDaySchedule_AllWeekends_Unit(t *testing.T) {
	repo := &mockDayScheduleWriter{
		mockDayScheduleReader: mockDayScheduleReader{
			getDayFn:            func(_ context.Context, _ string) (*repository.DaySchedule, error) { return nil, nil },
			getAvailableMealsFn: func(_ context.Context, _ string) ([]string, error) { return nil, nil },
		},
		upsertFn: func(_ context.Context, _ repository.DaySchedule) error { return nil },
		deleteFn: func(_ context.Context, _ string) error { return nil },
	}

	_, err := BulkSetDaySchedule(context.Background(), repo, []string{"2026-04-25", "2026-04-26"}, SetDayScheduleInput{DayStatus: "normal", SetBy: "admin"})
	if !errors.Is(err, ErrAllWeekend) {
		t.Fatalf("expected ErrAllWeekend, got: %v", err)
	}
}

func TestBulkSetDaySchedule_InvalidDate(t *testing.T) {
	repo := &mockDayScheduleWriter{
		mockDayScheduleReader: mockDayScheduleReader{
			getDayFn:            func(_ context.Context, _ string) (*repository.DaySchedule, error) { return nil, nil },
			getAvailableMealsFn: func(_ context.Context, _ string) ([]string, error) { return nil, nil },
		},
		upsertFn: func(_ context.Context, _ repository.DaySchedule) error { return nil },
		deleteFn: func(_ context.Context, _ string) error { return nil },
	}

	_, err := BulkSetDaySchedule(context.Background(), repo, []string{"not-a-date"}, SetDayScheduleInput{DayStatus: "normal"})
	if !errors.Is(err, ErrInvalidDate) {
		t.Fatalf("expected ErrInvalidDate, got: %v", err)
	}
}

func TestBulkSetDaySchedule_RollbackRestoresPreviousSchedule(t *testing.T) {
	state := map[string]*repository.DaySchedule{
		"2026-04-20": {
			Date:           "2026-04-20",
			DayStatus:      "celebration",
			AvailableMeals: []string{"lunch"},
			Reason:         "Original reason",
			CreatedBy:      "admin-0",
		},
	}

	repo := &mockDayScheduleWriter{
		mockDayScheduleReader: mockDayScheduleReader{
			getDayFn: func(_ context.Context, date string) (*repository.DaySchedule, error) {
				schedule, ok := state[date]
				if !ok {
					return nil, nil
				}
				clone := *schedule
				clone.AvailableMeals = append([]string(nil), schedule.AvailableMeals...)
				return &clone, nil
			},
			getAvailableMealsFn: func(_ context.Context, _ string) ([]string, error) { return nil, nil },
		},
		upsertFn: func(_ context.Context, schedule repository.DaySchedule) error {
			if schedule.Date == "2026-04-21" {
				return errors.New("db error")
			}
			clone := schedule
			clone.AvailableMeals = append([]string(nil), schedule.AvailableMeals...)
			state[schedule.Date] = &clone
			return nil
		},
		deleteFn: func(_ context.Context, date string) error {
			delete(state, date)
			return nil
		},
	}

	_, err := BulkSetDaySchedule(context.Background(), repo, []string{"2026-04-20", "2026-04-21"}, SetDayScheduleInput{
		DayStatus:      "normal",
		AvailableMeals: []string{"snacks"},
		Reason:         "Updated reason",
		SetBy:          "admin-1",
	})
	if err == nil {
		t.Fatal("expected rollback error, got nil")
	}

	restored := state["2026-04-20"]
	if restored == nil {
		t.Fatal("expected original schedule to be restored")
	}
	if restored.DayStatus != "celebration" {
		t.Fatalf("DayStatus = %q, want celebration", restored.DayStatus)
	}
	if len(restored.AvailableMeals) != 1 || restored.AvailableMeals[0] != "lunch" {
		t.Fatalf("AvailableMeals = %v, want [lunch]", restored.AvailableMeals)
	}
	if restored.Reason != "Original reason" {
		t.Fatalf("Reason = %q, want Original reason", restored.Reason)
	}
}

func TestBulkSetDaySchedule_RollbackDeletesNewSchedules(t *testing.T) {
	state := map[string]*repository.DaySchedule{}

	repo := &mockDayScheduleWriter{
		mockDayScheduleReader: mockDayScheduleReader{
			getDayFn: func(_ context.Context, date string) (*repository.DaySchedule, error) {
				schedule, ok := state[date]
				if !ok {
					return nil, nil
				}
				clone := *schedule
				clone.AvailableMeals = append([]string(nil), schedule.AvailableMeals...)
				return &clone, nil
			},
			getAvailableMealsFn: func(_ context.Context, _ string) ([]string, error) { return nil, nil },
		},
		upsertFn: func(_ context.Context, schedule repository.DaySchedule) error {
			if schedule.Date == "2026-04-21" {
				return errors.New("db error")
			}
			clone := schedule
			clone.AvailableMeals = append([]string(nil), schedule.AvailableMeals...)
			state[schedule.Date] = &clone
			return nil
		},
		deleteFn: func(_ context.Context, date string) error {
			delete(state, date)
			return nil
		},
	}

	_, err := BulkSetDaySchedule(context.Background(), repo, []string{"2026-04-20", "2026-04-21"}, SetDayScheduleInput{
		DayStatus:      "normal",
		AvailableMeals: []string{"snacks"},
		SetBy:          "admin-1",
	})
	if err == nil {
		t.Fatal("expected rollback error, got nil")
	}
	if _, ok := state["2026-04-20"]; ok {
		t.Fatal("expected newly created schedule to be removed during rollback")
	}
}
