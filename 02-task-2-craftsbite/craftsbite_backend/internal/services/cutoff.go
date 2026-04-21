package services

import (
	"fmt"
	"time"
)

const (
	defaultCutoffTime = "21:00"
	defaultTimezone   = "Asia/Dhaka"
	defaultMaxDays    = 7
)

type CutoffConfig struct {
	CutoffTime   string
	Timezone     string
	MaxDaysAhead int
}

type CutoffChecker struct {
	loc        *time.Location
	cutoffHour int
	cutoffMin  int
	maxDays    int
}

func NewCutoffChecker(cfg CutoffConfig) (*CutoffChecker, error) {
	tz := cfg.Timezone
	if tz == "" {
		tz = defaultTimezone
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return nil, fmt.Errorf("cutoff: invalid TIMEZONE %q: %w", tz, err)
	}

	cutoffStr := cfg.CutoffTime
	if cutoffStr == "" {
		cutoffStr = defaultCutoffTime
	}

	var h, m int
	if _, err := fmt.Sscanf(cutoffStr, "%d:%d", &h, &m); err != nil {
		return nil, fmt.Errorf("cutoff: invalid CUTOFF_TIME %q: %w", cutoffStr, err)
	}

	maxDays := cfg.MaxDaysAhead
	if maxDays == 0 {
		maxDays = defaultMaxDays
	}

	return &CutoffChecker{loc: loc, cutoffHour: h, cutoffMin: m, maxDays: maxDays}, nil
}

func (c *CutoffChecker) IsBeforeCutoff(targetDate string) (bool, error) {
	return c.isBeforeCutoffAt(targetDate, time.Now())
}

func (c *CutoffChecker) isBeforeCutoffAt(targetDate string, now time.Time) (bool, error) {
	if c == nil || c.loc == nil {
		return false, fmt.Errorf("cutoff checker is not initialized")
	}

	nowLocal := now.In(c.loc)
	todayLocal := time.Date(nowLocal.Year(), nowLocal.Month(), nowLocal.Day(), 0, 0, 0, 0, c.loc)

	target, err := time.ParseInLocation("2006-01-02", targetDate, c.loc)
	if err != nil {
		return false, fmt.Errorf("cutoff: invalid targetDate %q: %w", targetDate, err)
	}

	if !todayLocal.Before(target) {
		return false, nil
	}

	daysAhead := int(target.Sub(todayLocal).Hours() / 24)
	if daysAhead > c.maxDays {
		return false, nil
	}

	dayBeforeTarget := target.AddDate(0, 0, -1)
	cutoff := time.Date(dayBeforeTarget.Year(), dayBeforeTarget.Month(), dayBeforeTarget.Day(), c.cutoffHour, c.cutoffMin, 0, 0, c.loc)
	return nowLocal.Before(cutoff), nil
}

func (c *CutoffChecker) Location() *time.Location {
	if c == nil {
		return nil
	}
	return c.loc
}

func (c *CutoffChecker) MaxDaysAhead() int {
	if c == nil {
		return 0
	}
	return c.maxDays
}
