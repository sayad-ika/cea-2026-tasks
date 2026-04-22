package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/sayad-ika/craftsbite/internal/repository"
)

func TestGetLocation_Set(t *testing.T) {
	repo := &mockLocationReader{
		getFn: func(_ context.Context, _, _ string) (*repository.WorkLocation, error) {
			return &repository.WorkLocation{Location: "office"}, nil
		},
		getByDateFn: func(_ context.Context, _ string) ([]repository.WorkLocation, error) { return nil, nil },
	}

	wl, err := GetLocation(context.Background(), repo, "user-1", "2026-04-25")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if wl.Location != "office" {
		t.Errorf("Location = %q, want office", wl.Location)
	}
}

func TestGetLocation_NotSet(t *testing.T) {
	repo := &mockLocationReader{
		getFn: func(_ context.Context, _, _ string) (*repository.WorkLocation, error) {
			return nil, nil
		},
		getByDateFn: func(_ context.Context, _ string) ([]repository.WorkLocation, error) { return nil, nil },
	}

	wl, err := GetLocation(context.Background(), repo, "user-1", "2026-04-25")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if wl.Location != "not_set" {
		t.Errorf("Location = %q, want not_set", wl.Location)
	}
}

func TestGetLocation_DBError(t *testing.T) {
	repo := &mockLocationReader{
		getFn: func(_ context.Context, _, _ string) (*repository.WorkLocation, error) {
			return nil, errors.New("db fail")
		},
		getByDateFn: func(_ context.Context, _ string) ([]repository.WorkLocation, error) { return nil, nil },
	}

	_, err := GetLocation(context.Background(), repo, "user-1", "2026-04-25")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestSetLocation_Valid(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Dhaka")
	tomorrow := time.Now().In(loc).AddDate(0, 0, 2).Format("2006-01-02")

	checker, _ := NewCutoffChecker(CutoffConfig{CutoffTime: "21:00", Timezone: "Asia/Dhaka"})

	var upsertCalled bool
	repo := &mockLocationWriter{
		mockLocationReader: mockLocationReader{
			getFn: func(_ context.Context, _, _ string) (*repository.WorkLocation, error) {
				return &repository.WorkLocation{Location: "office"}, nil
			},
			getByDateFn: func(_ context.Context, _ string) ([]repository.WorkLocation, error) { return nil, nil },
		},
		upsertFn: func(_ context.Context, _ repository.WorkLocation) error {
			upsertCalled = true
			return nil
		},
	}

	wl, err := SetLocation(context.Background(), repo, "user-1", tomorrow, "office", checker)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !upsertCalled {
		t.Error("expected upsert to be called")
	}
	if wl.Location != "office" {
		t.Errorf("Location = %q, want office", wl.Location)
	}
}

func TestSetLocation_PastDate(t *testing.T) {
	checker, _ := NewCutoffChecker(CutoffConfig{CutoffTime: "21:00", Timezone: "Asia/Dhaka"})

	repo := &mockLocationWriter{
		mockLocationReader: mockLocationReader{
			getFn: func(_ context.Context, _, _ string) (*repository.WorkLocation, error) { return nil, nil },
			getByDateFn: func(_ context.Context, _ string) ([]repository.WorkLocation, error) { return nil, nil },
		},
		upsertFn: func(_ context.Context, _ repository.WorkLocation) error { return nil },
	}

	_, err := SetLocation(context.Background(), repo, "user-1", "2020-01-01", "office", checker)
	if !errors.Is(err, ErrLocationPastDate) {
		t.Fatalf("expected ErrLocationPastDate, got: %v", err)
	}
}

func TestSetLocation_NilCutoff(t *testing.T) {
	repo := &mockLocationWriter{
		mockLocationReader: mockLocationReader{
			getFn: func(_ context.Context, _, _ string) (*repository.WorkLocation, error) { return nil, nil },
			getByDateFn: func(_ context.Context, _ string) ([]repository.WorkLocation, error) { return nil, nil },
		},
		upsertFn: func(_ context.Context, _ repository.WorkLocation) error { return nil },
	}

	_, err := SetLocation(context.Background(), repo, "user-1", "2026-04-25", "office", nil)
	if err == nil {
		t.Fatal("expected error for nil cutoff, got nil")
	}
}

func TestSetLocation_TooFarAhead(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Dhaka")
	farFuture := time.Now().In(loc).AddDate(0, 0, 30).Format("2006-01-02")

	checker, _ := NewCutoffChecker(CutoffConfig{CutoffTime: "21:00", Timezone: "Asia/Dhaka"})

	repo := &mockLocationWriter{
		mockLocationReader: mockLocationReader{
			getFn: func(_ context.Context, _, _ string) (*repository.WorkLocation, error) { return nil, nil },
			getByDateFn: func(_ context.Context, _ string) ([]repository.WorkLocation, error) { return nil, nil },
		},
		upsertFn: func(_ context.Context, _ repository.WorkLocation) error { return nil },
	}

	_, err := SetLocation(context.Background(), repo, "user-1", farFuture, "office", checker)
	if !errors.Is(err, ErrLocationTooFarAhead) {
		t.Fatalf("expected ErrLocationTooFarAhead, got: %v", err)
	}
}

func TestSetLocation_UpsertError(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Dhaka")
	tomorrow := time.Now().In(loc).AddDate(0, 0, 2).Format("2006-01-02")

	checker, _ := NewCutoffChecker(CutoffConfig{CutoffTime: "21:00", Timezone: "Asia/Dhaka"})

	repo := &mockLocationWriter{
		mockLocationReader: mockLocationReader{
			getFn: func(_ context.Context, _, _ string) (*repository.WorkLocation, error) { return nil, nil },
			getByDateFn: func(_ context.Context, _ string) ([]repository.WorkLocation, error) { return nil, nil },
		},
		upsertFn: func(_ context.Context, _ repository.WorkLocation) error { return errors.New("db fail") },
	}

	_, err := SetLocation(context.Background(), repo, "user-1", tomorrow, "office", checker)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
