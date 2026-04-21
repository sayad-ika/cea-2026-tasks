package dateutil

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	defaultTimezone = "Asia/Dhaka"
	dateFormat      = "2006-01-02"
)

type DateParser struct {
	loc *time.Location
}

func NewDateParser(timezone string) (*DateParser, error) {
	tz := timezone
	if tz == "" {
		tz = defaultTimezone
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return nil, fmt.Errorf("invalid TIMEZONE %q: %w", tz, err)
	}
	return &DateParser{loc: loc}, nil
}

// ParseDateWithDefaults parses a date string with support for shortcuts and defaults.
// If dateStr is empty, returns tomorrow's date.
func (p *DateParser) ParseDateWithDefaults(dateStr string) (string, error) {
	return p.parseDateWithDefaultsAt(dateStr, time.Now())
}

func (p *DateParser) parseDateWithDefaultsAt(dateStr string, now time.Time) (string, error) {
	if p == nil || p.loc == nil {
		return "", fmt.Errorf("date parser is not initialized")
	}

	now = now.In(p.loc)

	// Empty string defaults to tomorrow
	if dateStr == "" {
		return now.AddDate(0, 0, 1).Format(dateFormat), nil
	}

	// Handle shortcuts
	switch strings.ToLower(dateStr) {
	case "today":
		return now.Format(dateFormat), nil
	case "tomorrow":
		return now.AddDate(0, 0, 1).Format(dateFormat), nil
	}

	// Handle relative dates: +1, +2, etc.
	if strings.HasPrefix(dateStr, "+") {
		days, err := strconv.Atoi(dateStr[1:])
		if err != nil {
			return "", fmt.Errorf("invalid relative date format: %s (use +1, +2, etc.)", dateStr)
		}
		if days < 0 {
			return "", fmt.Errorf("relative days must be positive: %s", dateStr)
		}
		return now.AddDate(0, 0, days).Format(dateFormat), nil
	}

	// Validate YYYY-MM-DD format
	parsed, err := time.Parse(dateFormat, dateStr)
	if err != nil {
		return "", fmt.Errorf("invalid date format: %s (use YYYY-MM-DD, 'today', 'tomorrow', or '+N')", dateStr)
	}

	return parsed.Format(dateFormat), nil
}

// TodayInTimezone returns today's date in the configured timezone
func (p *DateParser) TodayInTimezone() string {
	return p.todayInTimezoneAt(time.Now())
}

// TomorrowInTimezone returns tomorrow's date in the configured timezone
func (p *DateParser) TomorrowInTimezone() string {
	return p.tomorrowInTimezoneAt(time.Now())
}

func (p *DateParser) todayInTimezoneAt(now time.Time) string {
	if p == nil || p.loc == nil {
		return now.UTC().Format(dateFormat)
	}
	return now.In(p.loc).Format(dateFormat)
}

func (p *DateParser) tomorrowInTimezoneAt(now time.Time) string {
	if p == nil || p.loc == nil {
		return now.UTC().AddDate(0, 0, 1).Format(dateFormat)
	}
	return now.In(p.loc).AddDate(0, 0, 1).Format(dateFormat)
}
