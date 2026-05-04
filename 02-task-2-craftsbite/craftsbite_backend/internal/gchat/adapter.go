package gchat

import (
	"encoding/json"
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
	9: "help",
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
	case 9:
		opts = map[string]interface{}{}
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

	optsJSON, _ := json.Marshal(opts)

	return payload.CommandEvent{
		UserID:           internalUserID,
		Role:             role,
		CommandName:      commandName,
		Options:          optsJSON,
		Source:           "gchat",
		GChatSpaceName:   p.Space.Name,
		GChatMessageName: msgName,
		GChatViewerName:  evt.Chat.User.Name,
	}, nil
}

func parseMealArgs(raw string) (map[string]interface{}, error) {
	tokens := strings.Fields(raw)
	status := ""
	mealType := "all"
	dateStr := ""
	mealSet := false

	for _, token := range tokens {
		lower := strings.ToLower(token)
		switch {
		case (lower == "in" || lower == "out") && status == "":
			status = lower
		case repository.IsValidMealOrAll(lower) && !mealSet:
			mealType = lower
			mealSet = true
		case dateStr == "" && looksLikeDateArg(lower):
			dateStr = token
		case !mealSet:
			return nil, fmt.Errorf("Invalid meal_type. Allowed: lunch, snacks, iftar, event_dinner, optional_dinner, all")
		default:
			return nil, fmt.Errorf("Usage: /meal [in|out] [meal_type] [date]\nIf status is omitted, CraftsBite toggles the current choice.")
		}
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
	dateStr := ""
	for _, token := range tokens {
		lower := strings.ToLower(token)
		switch {
		case (lower == "office" || lower == "wfh") && loc == "":
			loc = lower
		case dateStr == "" && looksLikeDateArg(lower):
			dateStr = token
		case loc == "":
			loc = lower
		}
	}

	// Pass raw date string to Lambda - let it handle parsing with proper timezone
	return map[string]interface{}{
		"location": loc,
		"date":     dateStr,
	}
}

func looksLikeDateArg(token string) bool {
	return token == "today" || token == "tomorrow" || token == "week" || strings.HasPrefix(token, "+") || strings.Contains(token, "..") || strings.ContainsAny(token, "0123456789")
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
