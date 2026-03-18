package dateutil

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultTimezone = "Asia/Dhaka"
	dateFormat      = "2006-01-02"
)

// ParseDateWithDefaults parses a date string with support for shortcuts and defaults.
// If dateStr is empty, returns tomorrow's date.
// Supports: "today", "tomorrow", "+N" (days from today), or "YYYY-MM-DD"
func ParseDateWithDefaults(dateStr string) (string, error) {
	loc, err := loadLocation()
	if err != nil {
		return "", err
	}

	now := time.Now().In(loc)

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
func TodayInTimezone() string {
	loc, err := loadLocation()
	if err != nil {
		loc = time.UTC
	}
	return time.Now().In(loc).Format(dateFormat)
}

// TomorrowInTimezone returns tomorrow's date in the configured timezone
func TomorrowInTimezone() string {
	loc, err := loadLocation()
	if err != nil {
		loc = time.UTC
	}
	return time.Now().In(loc).AddDate(0, 0, 1).Format(dateFormat)
}

func loadLocation() (*time.Location, error) {
	tz := os.Getenv("TIMEZONE")
	if tz == "" {
		tz = defaultTimezone
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return nil, fmt.Errorf("invalid TIMEZONE %q: %w", tz, err)
	}
	return loc, nil
}
