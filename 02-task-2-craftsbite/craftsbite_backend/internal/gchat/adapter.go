package gchat

import (
	"fmt"
	"strings"
	"time"

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
		opts = parseDateArg(argText)
	}

	return payload.CommandEvent{
		UserID:           internalUserID,
		Role:             role,
		CommandName:      commandName,
		Options:          opts,
		Source:           "gchat",
		GChatSpaceName:   p.Space.Name,
		GChatMessageName: msgName,
	}, nil
}

func parseMealArgs(raw string) (map[string]interface{}, error) {
	tokens := strings.Fields(raw)
	if len(tokens) == 0 {
		return nil, fmt.Errorf("usage: /meal <in|out> [meal_type] [YYYY-MM-DD]")
	}

	status := strings.ToLower(tokens[0])
	if status != "in" && status != "out" {
		return nil, fmt.Errorf("usage: /meal <in|out> [meal_type] [YYYY-MM-DD]")
	}

	mealType := "all"
	if len(tokens) > 1 {
		mealType = strings.ToLower(tokens[1])
	}

	date := todayDhaka()
	if len(tokens) > 2 {
		date = tokens[2]
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

	date := todayDhaka()
	if len(tokens) > 1 {
		date = tokens[1]
	}

	return map[string]interface{}{
		"location": loc,
		"date":     date,
	}
}

func parseDateArg(raw string) map[string]interface{} {
	tokens := strings.Fields(raw)

	date := todayDhaka()
	if len(tokens) > 0 {
		date = tokens[0]
	}

	return map[string]interface{}{
		"date": date,
	}
}

func todayDhaka() string {
	loc, err := time.LoadLocation("Asia/Dhaka")
	if err != nil {
		loc = time.UTC
	}
	return time.Now().In(loc).Format("2006-01-02")
}
