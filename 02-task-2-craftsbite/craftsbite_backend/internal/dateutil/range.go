package dateutil

import (
	"fmt"
	"strings"
	"time"
)

// ParseDateRange parses a date parameter and returns a slice of dates.
// Supports:
// - Single date: "2026-03-20" or shortcuts (today, tomorrow, +N)
// - Date range: "2026-03-20..2026-03-22"
// - Week keyword: "week" (next 5 business days from tomorrow)
func ParseDateRange(dateParam string) ([]string, error) {
	// Handle date range syntax: "2026-03-20..2026-03-22"
	if strings.Contains(dateParam, "..") {
		return parseDateRangeSyntax(dateParam)
	}

	// Handle "week" keyword
	if strings.ToLower(dateParam) == "week" {
		return getNextBusinessWeek()
	}

	// Single date - use existing parser
	date, err := ParseDateWithDefaults(dateParam)
	if err != nil {
		return nil, err
	}
	return []string{date}, nil
}

func parseDateRangeSyntax(rangeStr string) ([]string, error) {
	parts := strings.Split(rangeStr, "..")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid date range format: %s (use YYYY-MM-DD..YYYY-MM-DD)", rangeStr)
	}

	startStr := strings.TrimSpace(parts[0])
	endStr := strings.TrimSpace(parts[1])

	// Parse start date (supports shortcuts)
	startDate, err := ParseDateWithDefaults(startStr)
	if err != nil {
		return nil, fmt.Errorf("invalid start date: %w", err)
	}

	// Parse end date (supports shortcuts)
	endDate, err := ParseDateWithDefaults(endStr)
	if err != nil {
		return nil, fmt.Errorf("invalid end date: %w", err)
	}

	start, err := time.Parse(dateFormat, startDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start date format: %w", err)
	}

	end, err := time.Parse(dateFormat, endDate)
	if err != nil {
		return nil, fmt.Errorf("invalid end date format: %w", err)
	}

	if end.Before(start) {
		return nil, fmt.Errorf("end date %s is before start date %s", endDate, startDate)
	}

	// Limit range to prevent abuse (max 14 days)
	const maxRangeDays = 14
	daysDiff := int(end.Sub(start).Hours() / 24)
	if daysDiff > maxRangeDays {
		return nil, fmt.Errorf("date range too large: %d days (max %d days)", daysDiff+1, maxRangeDays)
	}

	// Generate all dates in range
	var dates []string
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		dates = append(dates, d.Format(dateFormat))
	}

	return dates, nil
}

func getNextBusinessWeek() ([]string, error) {
	loc, err := loadLocation()
	if err != nil {
		return nil, err
	}

	now := time.Now().In(loc)
	tomorrow := now.AddDate(0, 0, 1)

	var dates []string
	current := tomorrow

	// Get next 5 business days (Monday-Friday)
	for len(dates) < 5 {
		weekday := current.Weekday()
		// Skip Saturday (6) and Sunday (0)
		if weekday != time.Saturday && weekday != time.Sunday {
			dates = append(dates, current.Format(dateFormat))
		}
		current = current.AddDate(0, 0, 1)
	}

	return dates, nil
}
