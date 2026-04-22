package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/sayad-ika/craftsbite/internal/repository"
)

var (
	ErrInvalidDate      = errors.New("invalid date format (use YYYY-MM-DD)")
	ErrInvalidDayStatus = errors.New("invalid day status")
	ErrInvalidMealType  = errors.New("invalid meal type")
	ErrPastDateSchedule = errors.New("cannot set schedule for past dates")
	ErrAllWeekend       = errors.New("all dates in the range fall on weekends — no schedule was set")
)

type BulkSetDayScheduleResult struct {
	SuccessDates []string
}

type SetDayScheduleInput struct {
	Date           string
	DayStatus      string
	AvailableMeals []string
	Reason         string
	SetBy          string
}

func SetDaySchedule(ctx context.Context, repo DayScheduleWriter, input SetDayScheduleInput) (*repository.DaySchedule, error) {
	parsedDate, err := time.Parse("2006-01-02", input.Date)
	if err != nil {
		return nil, ErrInvalidDate
	}

	if !repository.IsValidDayStatus(input.DayStatus) {
		return nil, fmt.Errorf("%w: %s (valid: %v)",
			ErrInvalidDayStatus, input.DayStatus, repository.ValidDayStatuses())
	}

	for _, meal := range input.AvailableMeals {
		if !repository.IsValidMealType(meal) {
			return nil, fmt.Errorf("%w: %s (valid: %v)",
				ErrInvalidMealType, meal, repository.ValidMealTypes())
		}
	}

	if (input.DayStatus == string(repository.DayStatusOfficeClosed) ||
		input.DayStatus == string(repository.DayStatusGovtHoliday)) &&
		len(input.AvailableMeals) > 0 {
		return nil, errors.New("office_closed and govt_holiday days cannot have meals")
	}

	schedule := repository.DaySchedule{
		Date:           input.Date,
		DayStatus:      input.DayStatus,
		AvailableMeals: input.AvailableMeals,
		Reason:         input.Reason,
		CreatedBy:      input.SetBy,
		CreatedAt:      parsedDate,
		UpdatedAt:      time.Now().UTC(),
	}

	if err := repo.UpsertDaySchedule(ctx, schedule); err != nil {
		return nil, fmt.Errorf("failed to save day schedule: %w", err)
	}

	return &schedule, nil
}

func GetDaySchedule(ctx context.Context, repo DayScheduleReader, date string) (*repository.DaySchedule, error) {
	return repo.GetDay(ctx, date)
}

func DeleteDaySchedule(ctx context.Context, repo DayScheduleWriter, date string) error {
	return repo.DeleteDaySchedule(ctx, date)
}

func BulkSetDaySchedule(ctx context.Context, repo DayScheduleWriter, dates []string, input SetDayScheduleInput) (*BulkSetDayScheduleResult, error) {
	var weekdays []string
	for _, d := range dates {
		t, err := time.Parse("2006-01-02", d)
		if err != nil {
			return nil, fmt.Errorf("%w: %s", ErrInvalidDate, d)
		}
		if t.Weekday() == time.Saturday || t.Weekday() == time.Sunday {
			continue
		}
		weekdays = append(weekdays, d)
	}

	if len(weekdays) == 0 {
		return nil, ErrAllWeekend
	}

	var written []string
	for _, date := range weekdays {
		singleInput := input
		singleInput.Date = date
		_, err := SetDaySchedule(ctx, repo, singleInput)
		if err != nil {
			for _, prev := range written {
				_ = repo.DeleteDaySchedule(ctx, prev)
			}
			return nil, fmt.Errorf("failed on %s: %w", date, err)
		}
		written = append(written, date)
	}

	return &BulkSetDayScheduleResult{SuccessDates: written}, nil
}
