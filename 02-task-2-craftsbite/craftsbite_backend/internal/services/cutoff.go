package services

import (
	"fmt"
	"os"
	"time"
)

const (
	defaultCutoffTime = "21:00"
	defaultTimezone   = "Asia/Dhaka"

	maxDaysAhead = 7
)

func IsBeforeCutoff(targetDate string) (bool, error) {
	return isBeforeCutoffAt(targetDate, time.Now())
}

func isBeforeCutoffAt(targetDate string, now time.Time) (bool, error) {
	loc, err := loadLocation()
	if err != nil {
		return false, err
	}

	cutoffStr := os.Getenv("CUTOFF_TIME")
	if cutoffStr == "" {
		cutoffStr = defaultCutoffTime
	}

	var cutoffHour, cutoffMin int
	if _, err := fmt.Sscanf(cutoffStr, "%d:%d", &cutoffHour, &cutoffMin); err != nil {
		return false, fmt.Errorf("cutoff: invalid CUTOFF_TIME %q: %w", cutoffStr, err)
	}

	target, err := time.ParseInLocation("2006-01-02", targetDate, loc)
	if err != nil {
		return false, fmt.Errorf("cutoff: invalid targetDate %q: %w", targetDate, err)
	}

	nowLocal := now.In(loc)
	todayLocal := time.Date(nowLocal.Year(), nowLocal.Month(), nowLocal.Day(), 0, 0, 0, 0, loc)

	if !todayLocal.Before(target) {
		return false, nil
	}

	daysAhead := int(target.Sub(todayLocal).Hours() / 24)
	if daysAhead > maxDaysAhead {
		return false, nil
	}

	dayBeforeTarget := target.AddDate(0, 0, -1)
	cutoff := time.Date(dayBeforeTarget.Year(), dayBeforeTarget.Month(), dayBeforeTarget.Day(), cutoffHour, cutoffMin, 0, 0, loc)
	return nowLocal.Before(cutoff), nil
}

func loadLocation() (*time.Location, error) {
	tz := os.Getenv("TIMEZONE")
	if tz == "" {
		tz = defaultTimezone
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return nil, fmt.Errorf("cutoff: invalid TIMEZONE %q: %w", tz, err)
	}
	return loc, nil
}
