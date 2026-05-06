package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/sayad-ika/craftsbite/internal/repository"
)

var (
	ErrLocationPastDate     = errors.New("cannot set work location for a past date")
	ErrLocationCutoffPassed = errors.New("cutoff has passed for this date")
	ErrLocationTooFarAhead  = errors.New("cannot set work location more than 7 days in advance")
)

func GetLocation(ctx context.Context, repo LocationReader, userID, date string) (*repository.WorkLocation, error) {
	wl, err := repo.GetWorkLocation(ctx, userID, date)
	if err != nil {
		return nil, fmt.Errorf("location: GetLocation: %w", err)
	}
	if wl == nil {
		return notSetLocation(userID, date), nil
	}
	return wl, nil
}

func notSetLocation(userID, date string) *repository.WorkLocation {
	return &repository.WorkLocation{
		UserID:   userID,
		Date:     date,
		Location: "not_set",
	}
}

func SetLocation(ctx context.Context, repo LocationWriter, userID, date, location string, cutoff *CutoffChecker) (*repository.WorkLocation, error) {
	if cutoff == nil {
		return nil, fmt.Errorf("location: cutoff checker is required")
	}

	loc := cutoff.Location()
	nowLocal := time.Now().In(loc)
	today := nowLocal.Format("2006-01-02")

	if date < today {
		return nil, ErrLocationPastDate
	}
	if date > today {
		target, parseErr := time.ParseInLocation("2006-01-02", date, loc)
		if parseErr != nil {
			return nil, fmt.Errorf("location: invalid date %q: %w", date, parseErr)
		}
		todayMidnight := time.Date(nowLocal.Year(), nowLocal.Month(), nowLocal.Day(), 0, 0, 0, 0, loc)
		if int(target.Sub(todayMidnight).Hours()/24) > cutoff.MaxDaysAhead() {
			return nil, ErrLocationTooFarAhead
		}

		ok, err := cutoff.IsBeforeCutoff(date)
		if err != nil {
			return nil, fmt.Errorf("location: cutoff check: %w", err)
		}
		if !ok {
			return nil, ErrLocationCutoffPassed
		}
	}

	wl := repository.WorkLocation{
		UserID:   userID,
		Date:     date,
		Location: location,
	}
	if err := repo.UpsertWorkLocation(ctx, wl, time.Time{}); err != nil {
		return nil, fmt.Errorf("location: SetLocation upsert: %w", err)
	}

	return GetLocation(ctx, repo, userID, date)
}
