package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
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

// SetDayScheduleInput represents input for setting a day schedule
type SetDayScheduleInput struct {
	Date           string
	DayStatus      string
	AvailableMeals []string
	Reason         string
	SetBy          string // user ID of admin setting this
}

// SetDaySchedule creates or updates a day schedule
func SetDaySchedule(ctx context.Context, client *dynamodb.Client, table string, input SetDayScheduleInput) (*repository.DaySchedule, error) {
	// Validate date format
	parsedDate, err := time.Parse("2006-01-02", input.Date)
	if err != nil {
		return nil, ErrInvalidDate
	}

	// Validate day status
	if !repository.IsValidDayStatus(input.DayStatus) {
		return nil, fmt.Errorf("%w: %s (valid: %v)",
			ErrInvalidDayStatus, input.DayStatus, repository.ValidDayStatuses())
	}

	// Validate meal types
	for _, meal := range input.AvailableMeals {
		if !repository.IsValidMealType(meal) {
			return nil, fmt.Errorf("%w: %s (valid: %v)",
				ErrInvalidMealType, meal, repository.ValidMealTypes())
		}
	}

	// Business rule: office_closed and govt_holiday should have no meals
	if (input.DayStatus == string(repository.DayStatusOfficeClosed) ||
		input.DayStatus == string(repository.DayStatusGovtHoliday)) &&
		len(input.AvailableMeals) > 0 {
		return nil, errors.New("office_closed and govt_holiday days cannot have meals")
	}

	// Create schedule
	schedule := repository.DaySchedule{
		Date:           input.Date,
		DayStatus:      input.DayStatus,
		AvailableMeals: input.AvailableMeals,
		Reason:         input.Reason,
		CreatedBy:      input.SetBy,
		CreatedAt:      parsedDate,
		UpdatedAt:      time.Now().UTC(),
	}

	// Save to repository
	if err := repository.UpsertDaySchedule(ctx, client, table, schedule); err != nil {
		return nil, fmt.Errorf("failed to save day schedule: %w", err)
	}

	return &schedule, nil
}

// GetDaySchedule retrieves a day schedule (wrapper for repository function)
func GetDaySchedule(ctx context.Context, client *dynamodb.Client, table, date string) (*repository.DaySchedule, error) {
	return repository.GetDay(ctx, client, table, date)
}

// DeleteDaySchedule removes a day schedule (resets to default)
func DeleteDaySchedule(ctx context.Context, client *dynamodb.Client, table, date string) error {
	return repository.DeleteDaySchedule(ctx, client, table, date)
}

func BulkSetDaySchedule(ctx context.Context, client *dynamodb.Client, table string, dates []string, input SetDayScheduleInput) (*BulkSetDayScheduleResult, error) {
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
		_, err := SetDaySchedule(ctx, client, table, singleInput)
		if err != nil {
			for _, prev := range written {
				_ = repository.DeleteDaySchedule(ctx, client, table, prev)
			}
			return nil, fmt.Errorf("failed on %s: %w", date, err)
		}
		written = append(written, date)
	}

	return &BulkSetDayScheduleResult{SuccessDates: written}, nil
}
