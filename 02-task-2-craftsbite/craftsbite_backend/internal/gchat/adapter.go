package gchat

import (
	"fmt"
	"strings"

	"github.com/sayad-ika/craftsbite/internal/dateutil"
	"github.com/sayad-ika/craftsbite/internal/payload"
)

var gchatCommandNames = map[int64]string{
	1: "meal",
	2: "location",
	3: "team-summary",
	4: "headcount",
}

func ToCommandEvent(evt Event, internalUserID, role string) (payload.CommandEvent, error) {
	p := evt.Chat.AppCommandPayload
	if p == nil {
		return payload.CommandEvent{}, fmt.Errorf("appCommandPayload is nil")
	}

	commandID := int64(p.AppCommandMetadata.AppCommandID)
	commandName, known := gchatCommandNames[commandID]
	if !known {
		return payload.CommandEvent{}, fmt.Errorf("unknown command ID %d", commandID)
	}

	var argText string
	var msgName string
	if p.Message != nil {
		argText = p.Message.ArgumentText
		msgName = p.Message.Name
	}

	var opts map[string]interface{}
	var err error

	switch commandID {
	case 1:
		opts, err = parseMealArgs(argText)
		if err != nil {
			return payload.CommandEvent{}, err
		}
	case 2:
		opts = parseLocationArgs(argText)
	case 3:
		opts = parseDateArg(argText)
	case 4:
		opts, err = parseHeadcountArgs(argText)
		if err != nil {
			return payload.CommandEvent{}, err
		}
	}

	return payload.CommandEvent{
		UserID:           internalUserID,
		Role:             role,
		CommandName:      commandName,
		Options:          opts,
		Source:           "gchat",
		GChatSpaceName:   p.Space.Name,
		GChatMessageName: msgName,
		GChatViewerName:  evt.Chat.User.Name,
	}, nil
}

var validMealTypes = map[string]bool{
	"lunch":           true,
	"snacks":          true,
	"event_dinner":    true,
	"optional_dinner": true,
	"iftar":           true,
	"all":             true,
}

func parseMealArgs(raw string) (map[string]interface{}, error) {
	tokens := strings.Fields(raw)
	if len(tokens) == 0 {
		return nil, fmt.Errorf("Usage: /meal <in|out> [meal_type] [date]\nDate can be: tomorrow (default), today, +N, or YYYY-MM-DD")
	}

	status := strings.ToLower(tokens[0])
	if status != "in" && status != "out" {
		return nil, fmt.Errorf("Usage: /meal <in|out> [meal_type] [date]\nDate can be: tomorrow (default), today, +N, or YYYY-MM-DD")
	}

	mealType := "all"
	if len(tokens) > 1 {
		mealType = strings.ToLower(tokens[1])
	}
	if !validMealTypes[mealType] {
		return nil, fmt.Errorf("Invalid meal_type. Allowed: lunch, snacks, event_dinner, optional_dinner, all")
	}

	dateStr := ""
	if len(tokens) > 2 {
		dateStr = tokens[2]
	}

	date, err := dateutil.ParseDateWithDefaults(dateStr)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"status": status,
		"meal":   mealType,
		"date":   date,
	}, nil
}

func parseLocationArgs(raw string) map[string]interface{} {
	tokens := strings.Fields(raw)

	loc := ""
	if len(tokens) > 0 {
		loc = strings.ToLower(tokens[0])
	}

	dateStr := ""
	if len(tokens) > 1 {
		dateStr = tokens[1]
	}

	date, err := dateutil.ParseDateWithDefaults(dateStr)
	if err != nil {
		// Fallback to tomorrow if parsing fails (shouldn't happen with valid input)
		date = dateutil.TomorrowInTimezone()
	}

	return map[string]interface{}{
		"location": loc,
		"date":     date,
	}
}

func parseDateArg(raw string) map[string]interface{} {
	tokens := strings.Fields(raw)

	dateStr := ""
	if len(tokens) > 0 {
		dateStr = tokens[0]
	}

	date, err := dateutil.ParseDateWithDefaults(dateStr)
	if err != nil {
		// Fallback to tomorrow if parsing fails
		date = dateutil.TomorrowInTimezone()
	}

	return map[string]interface{}{
		"date": date,
	}
}

func parseHeadcountArgs(raw string) (map[string]interface{}, error) {
	tokens := strings.Fields(raw)

	dateStr := ""
	if len(tokens) > 0 {
		dateStr = tokens[0]
	}

	date, err := dateutil.ParseDateWithDefaults(dateStr)
	if err != nil {
		return nil, fmt.Errorf("Usage: /headcount [date]\nDate can be: tomorrow (default), today, +N, or YYYY-MM-DD\nError: %v", err)
	}

	return map[string]interface{}{
		"date": date,
	}, nil
}

