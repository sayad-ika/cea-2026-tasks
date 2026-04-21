package gchat

import (
	"fmt"
	"strings"

	"github.com/sayad-ika/craftsbite/internal/payload"
	"github.com/sayad-ika/craftsbite/internal/repository"
)

var gchatCommandNames = map[int64]string{
	1: "meal",
	2: "location",
	3: "team-summary",
	4: "headcount",
	5: "status",
	6: "schedule-day",
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
	case 5:
		opts = parseDateArg(argText)
	case 6:
		opts = parseScheduleDayArgs(argText)
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

func parseMealArgs(raw string) (map[string]interface{}, error) {
	tokens := strings.Fields(raw)
	if len(tokens) == 0 {
		return nil, fmt.Errorf("Usage: /meal <in|out> [meal_type] [date]\nDate can be: tomorrow (default), +N, or YYYY-MM-DD")
	}

	status := strings.ToLower(tokens[0])
	if status != "in" && status != "out" {
		return nil, fmt.Errorf("Usage: /meal <in|out> [meal_type] [date]\nDate can be: tomorrow (default), +N, or YYYY-MM-DD")
	}

	mealType := "all"
	if len(tokens) > 1 {
		mealType = strings.ToLower(tokens[1])
	}
	if !repository.IsValidMealOrAll(mealType) {
		return nil, fmt.Errorf("Invalid meal_type. Allowed: lunch, snacks, iftar, event_dinner, optional_dinner, all")
	}

	dateStr := ""
	if len(tokens) > 2 {
		dateStr = tokens[2]
	}

	// Pass raw date string to Lambda - let it handle parsing with proper timezone
	return map[string]interface{}{
		"status": status,
		"meal":   mealType,
		"date":   dateStr,
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

	// Pass raw date string to Lambda - let it handle parsing with proper timezone
	return map[string]interface{}{
		"location": loc,
		"date":     dateStr,
	}
}

func parseDateArg(raw string) map[string]interface{} {
	tokens := strings.Fields(raw)

	dateStr := ""
	if len(tokens) > 0 {
		dateStr = tokens[0]
	}

	// Pass raw date string to Lambda - let it handle parsing with proper timezone
	return map[string]interface{}{
		"date": dateStr,
	}
}

func parseHeadcountArgs(raw string) (map[string]interface{}, error) {
	tokens := strings.Fields(raw)

	dateStr := ""
	if len(tokens) > 0 {
		dateStr = tokens[0]
		if dateStr != "today" && dateStr != "tomorrow" && !strings.HasPrefix(dateStr, "+") && !strings.ContainsAny(dateStr, "0123456789") {
			dateStr = ""
		}
	}

	// Pass raw date string to Lambda - let it handle parsing with proper timezone
	return map[string]interface{}{
		"date": dateStr,
	}, nil
}

// parseScheduleDayArgs parses "/schedule-day <date> <status> [meals] [reason]"
// Format: /schedule-day 2026-03-25 normal lunch,snacks Optional reason text
func parseScheduleDayArgs(raw string) map[string]interface{} {
	parts := strings.Fields(raw)
	opts := make(map[string]interface{})

	if len(parts) >= 1 {
		opts["date"] = parts[0]
	}
	if len(parts) >= 2 {
		opts["status"] = parts[1]
	}
	if len(parts) >= 3 {
		opts["meals"] = parts[2]
	}
	if len(parts) >= 4 {
		opts["reason"] = strings.Join(parts[3:], " ")
	}

	return opts
}
