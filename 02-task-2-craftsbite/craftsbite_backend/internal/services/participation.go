package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/sayad-ika/craftsbite/internal/repository"
)

type ResolvedStatus struct {
	MealType string
	Status   string
	Source   string
}

var (
	ErrPastDate        = errors.New("cannot update participation for today or past date")
	ErrCutoffPassed    = errors.New("cutoff has passed for this date")
	ErrTooFarAhead     = errors.New("cannot update participation more than 7 days in advance")
	ErrDayClosed       = errors.New("office is closed on this date")
	ErrMealUnavailable = errors.New("meal is not available on this date")
	ErrNoMeals         = errors.New("no meals are configured for this date")
)

func Resolve(schedule *repository.DaySchedule, availableMeals []string, record *repository.MealParticipation, mealType string) ResolvedStatus {
	if schedule != nil && (schedule.DayStatus == "office_closed" || schedule.DayStatus == "govt_holiday") {
		return ResolvedStatus{MealType: mealType, Status: "unavailable", Source: "day_schedule_unavailable"}
	}
	if !containsMeal(availableMeals, mealType) {
		return ResolvedStatus{MealType: mealType, Status: "unavailable", Source: "day_schedule_unavailable"}
	}

	if record != nil {
		status := "opted_out"
		if record.IsParticipating {
			status = "opted_in"
		}
		return ResolvedStatus{MealType: mealType, Status: status, Source: "explicit"}
	}

	return ResolvedStatus{MealType: mealType, Status: "opted_in", Source: "system_default"}
}

func GetUserMealStatus(ctx context.Context, client *dynamodb.Client, table, userID, date string) ([]ResolvedStatus, error) {
	schedule, err := repository.GetDay(ctx, client, table, date)
	if err != nil {
		return nil, fmt.Errorf("participation: GetUserMealStatus schedule: %w", err)
	}

	availableMeals, err := repository.GetAvailableMeals(ctx, client, table, date)
	if err != nil {
		return nil, fmt.Errorf("participation: GetUserMealStatus available meals: %w", err)
	}
	if len(availableMeals) == 0 {
		return []ResolvedStatus{}, nil
	}

	records, err := repository.GetParticipationsByUserDate(ctx, client, table, userID, date)
	if err != nil {
		return nil, fmt.Errorf("participation: GetUserMealStatus records: %w", err)
	}

	// Build a lookup from mealType → record.
	recordByMeal := make(map[string]*repository.MealParticipation, len(records))
	for i := range records {
		recordByMeal[records[i].MealType] = &records[i]
	}

	statuses := make([]ResolvedStatus, 0, len(availableMeals))
	for _, meal := range availableMeals {
		statuses = append(statuses, Resolve(schedule, availableMeals, recordByMeal[meal], meal))
	}
	return statuses, nil
}

func UpdateParticipation(ctx context.Context, client *dynamodb.Client, table, userID, date, mealType string, isParticipating bool) ([]ResolvedStatus, error) {
	loc, err := loadLocation()
	if err != nil {
		return nil, fmt.Errorf("participation: load timezone: %w", err)
	}
	nowLocal := time.Now().In(loc)
	today := nowLocal.Format("2006-01-02")

	if date <= today {
		return nil, ErrPastDate
	}
	if date > today {
		// Reject dates beyond the hard lookahead cap before hitting the cutoff check.
		target, parseErr := time.ParseInLocation("2006-01-02", date, loc)
		if parseErr != nil {
			return nil, fmt.Errorf("participation: invalid date %q: %w", date, parseErr)
		}
		todayMidnight := time.Date(nowLocal.Year(), nowLocal.Month(), nowLocal.Day(), 0, 0, 0, 0, loc)
		if int(target.Sub(todayMidnight).Hours()/24) > maxDaysAhead {
			return nil, ErrTooFarAhead
		}

		ok, err := IsBeforeCutoff(date)
		if err != nil {
			return nil, fmt.Errorf("participation: cutoff check: %w", err)
		}
		if !ok {
			return nil, ErrCutoffPassed
		}
	}

	schedule, err := repository.GetDay(ctx, client, table, date)
	if err != nil {
		return nil, fmt.Errorf("participation: UpdateParticipation schedule: %w", err)
	}

	if schedule != nil && (schedule.DayStatus == "office_closed" || schedule.DayStatus == "govt_holiday") {
		return nil, ErrDayClosed
	}

	availableMeals, err := repository.GetAvailableMeals(ctx, client, table, date)
	if err != nil {
		return nil, fmt.Errorf("participation: UpdateParticipation available meals: %w", err)
	}
	if len(availableMeals) == 0 {
		return nil, ErrNoMeals
	}

	if mealType == "" {
		mealType = "all"
	}

	var mealsToWrite []string

	if mealType == "all" {
		mealsToWrite = availableMeals
	} else {
		if !containsMeal(availableMeals, mealType) {
			return nil, ErrMealUnavailable
		}
		mealsToWrite = []string{mealType}
	}

	now := time.Now().UTC()
	for _, meal := range mealsToWrite {
		p := repository.MealParticipation{
			UserID:          userID,
			Date:            date,
			MealType:        meal,
			IsParticipating: isParticipating,
			CreatedAt:       now,
			UpdatedAt:       now,
		}
		if err := repository.UpsertParticipation(ctx, client, table, p); err != nil {
			return nil, fmt.Errorf("participation: upsert %s: %w", meal, err)
		}
	}

	return GetUserMealStatus(ctx, client, table, userID, date)
}

func containsMeal(available []string, meal string) bool {
	norm := strings.Map(unicode.ToLower, meal)
	for _, m := range available {
		if strings.Map(unicode.ToLower, m) == norm {
			return true
		}
	}
	return false
}
