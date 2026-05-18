package gchat

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"sort"
	"strings"

	"github.com/sayad-ika/craftsbite/internal/payload"
	"github.com/sayad-ika/craftsbite/internal/repository"
)

var gchatCommandNames = map[int64]string{
	1:  "meal",
	2:  "location",
	3:  "team-summary",
	4:  "headcount",
	5:  "status",
	6:  "schedule-day",
	7:  "override",
	9:  "help",
	10: "init",
}

func ToCommandEvent(evt Event, internalUserID, role string) (payload.CommandEvent, error) {
	p := evt.Chat.AppCommandPayload
	if p == nil {
		return payload.CommandEvent{}, fmt.Errorf("appCommandPayload is nil")
	}

	commandID := int64(p.AppCommandMetadata.AppCommandID)
	commandName, known := gchatCommandNames[commandID]
	if !known {
		slog.Warn("gchat unknown command id", "command_id", commandID, "command_type", p.AppCommandMetadata.AppCommandType)
		return payload.CommandEvent{}, fmt.Errorf("unknown command ID %d", commandID)
	}
	slog.Info("gchat slash command mapping", "command_id", commandID, "command_type", p.AppCommandMetadata.AppCommandType, "command", commandName, "has_message", p.Message != nil)

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
	case 10:
		opts = parseDateArg(argText)
		opts["action"] = "open"
	case 7:
		opts, err = parseOverrideArgs(argText)
		if err != nil {
			return payload.CommandEvent{}, err
		}
	case 1:
		opts, err = parseMealArgs(argText)
		if err != nil {
			return payload.CommandEvent{}, err
		}
	case 2:
		opts = parseLocationArgs(argText)
	case 3:
		opts = parseTeamSummaryArgs(argText)
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
	slog.Info("gchat slash command options parsed", "command", commandName, "command_id", commandID, "options_keys", gchatInterfaceMapKeys(opts), "options_len", len(optsJSON))

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

func ToCardActionCommandEvent(evt Event, internalUserID, role string) (payload.CommandEvent, error) {
	action := dialogAction(evt.CommonEventObject)
	if action == "" {
		slog.Warn("gchat card action missing", "parameter_keys", gchatStringMapKeys(evt.CommonEventObject.Parameters), "form_input_keys", gchatFormInputMapKeys(evt.CommonEventObject.FormInputs), "invoked_function", evt.CommonEventObject.InvokedFunction)
		return payload.CommandEvent{}, fmt.Errorf("card action is missing")
	}
	slog.Info("gchat card action received", "action", action, "parameter_keys", gchatStringMapKeys(evt.CommonEventObject.Parameters), "form_input_keys", gchatFormInputMapKeys(evt.CommonEventObject.FormInputs), "is_dialog_event", evt.Chat.ButtonClickedPayload != nil && evt.Chat.ButtonClickedPayload.IsDialogEvent)

	var opts map[string]interface{}
	switch action {
	case InitCardFunctionSave:
		opts = initCardOptions(evt.CommonEventObject)
	case InitCardFunctionCancel:
		opts = map[string]interface{}{"action": "cancel"}
	default:
		slog.Warn("gchat unsupported card action", "action", action)
		return payload.CommandEvent{}, fmt.Errorf("unsupported card action %q", action)
	}

	optsJSON, _ := json.Marshal(opts)
	slog.Info("gchat card action options parsed", "action", action, "options_keys", gchatInterfaceMapKeys(opts), "options_len", len(optsJSON))
	return payload.CommandEvent{
		UserID:          internalUserID,
		Role:            role,
		CommandName:     "init",
		Options:         optsJSON,
		Source:          "gchat",
		GChatSpaceName:  evt.Chat.Space.Name,
		GChatViewerName: evt.Chat.User.Name,
	}, nil
}

func dialogAction(common CommonEventObject) string {
	if common.Parameters != nil {
		if action := common.Parameters["action"]; action != "" {
			return action
		}
	}
	return common.InvokedFunction
}

func initCardOptions(common CommonEventObject) map[string]interface{} {
	date := ""
	if common.Parameters != nil {
		date = common.Parameters["date"]
	}
	return map[string]interface{}{
		"action":   "apply",
		"date":     date,
		"dates":    formStringValues(common.FormInputs, "dates"),
		"location": firstFormStringValue(common.FormInputs, "location"),
		"meals":    formStringValues(common.FormInputs, "meals"),
	}
}

func formStringValues(inputs map[string]FormInput, name string) []string {
	if inputs == nil {
		return []string{}
	}
	input := inputs[name]
	if input.StringInputs == nil || len(input.StringInputs.Value) == 0 {
		return []string{}
	}
	values := make([]string, 0, len(input.StringInputs.Value))
	seen := make(map[string]struct{}, len(input.StringInputs.Value))
	for _, value := range input.StringInputs.Value {
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		values = append(values, value)
	}
	return values
}

func firstFormStringValue(inputs map[string]FormInput, name string) string {
	values := formStringValues(inputs, name)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func gchatInterfaceMapKeys(values map[string]interface{}) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func gchatStringMapKeys(values map[string]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func gchatFormInputMapKeys(values map[string]FormInput) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
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

func parseOverrideArgs(raw string) (map[string]interface{}, error) {
	tokens := strings.Fields(raw)
	if len(tokens) < 4 {
		return nil, fmt.Errorf("Usage: /override <target_email> <meal|location> <date> [meal] [value] <reason>")
	}

	target := tokens[0]
	entry := strings.ToLower(tokens[1])
	dateStr := tokens[2]
	rest := tokens[3:]

	if entry != "meal" && entry != "location" {
		return nil, fmt.Errorf("Entry must be `meal` or `location`.")
	}

	opts := map[string]interface{}{
		"target": target,
		"entry":  entry,
		"date":   dateStr,
		"meal":   "",
		"value":  "",
		"reason": "",
	}

	sepIndex := -1
	for i, token := range rest {
		if token == "--" {
			sepIndex = i
			break
		}
	}

	if entry == "meal" {
		if sepIndex == -1 {
			if len(rest) == 0 {
				return nil, fmt.Errorf("Override reason is required.")
			}
			first := strings.ToLower(rest[0])
			if repository.IsValidMealOrAll(first) || first == "in" || first == "out" {
				return nil, fmt.Errorf("Ambiguous override syntax. In Google Chat, use `--` before the reason when specifying meal or value.")
			}
			opts["reason"] = strings.TrimSpace(strings.Join(rest, " "))
			return opts, nil
		}

		beforeReason := rest[:sepIndex]
		reason := strings.TrimSpace(strings.Join(rest[sepIndex+1:], " "))
		if reason == "" {
			return nil, fmt.Errorf("Override reason is required.")
		}
		meal := ""
		value := ""
		idx := 0
		if idx < len(beforeReason) && repository.IsValidMealOrAll(strings.ToLower(beforeReason[idx])) {
			meal = strings.ToLower(beforeReason[idx])
			idx++
		}
		if idx < len(beforeReason) {
			candidate := strings.ToLower(beforeReason[idx])
			if candidate == "in" || candidate == "out" {
				value = candidate
				idx++
			}
		}
		if idx != len(beforeReason) {
			return nil, fmt.Errorf("Usage: /override <target_email> meal <date> [meal] [value] -- <reason>")
		}
		opts["meal"] = meal
		opts["value"] = value
		opts["reason"] = reason
		return opts, nil
	}

	if sepIndex == -1 {
		if len(rest) == 0 {
			return nil, fmt.Errorf("Override reason is required.")
		}
		first := strings.ToLower(rest[0])
		if first == "office" || first == "wfh" {
			return nil, fmt.Errorf("Ambiguous override syntax. In Google Chat, use `--` before the reason when specifying location value.")
		}
		opts["reason"] = strings.TrimSpace(strings.Join(rest, " "))
		return opts, nil
	}

	beforeReason := rest[:sepIndex]
	reason := strings.TrimSpace(strings.Join(rest[sepIndex+1:], " "))
	if reason == "" {
		return nil, fmt.Errorf("Override reason is required.")
	}
	value := ""
	idx := 0
	if idx < len(beforeReason) {
		candidate := strings.ToLower(beforeReason[idx])
		if candidate == "office" || candidate == "wfh" {
			value = candidate
			idx++
		}
	}
	if idx != len(beforeReason) {
		return nil, fmt.Errorf("Usage: /override <target_email> location <date> [value] -- <reason>")
	}
	opts["value"] = value
	opts["reason"] = reason
	return opts, nil
}

func parseTeamSummaryArgs(raw string) map[string]interface{} {
	parts := strings.Fields(raw)
	opts := map[string]interface{}{"date": "", "team_id": ""}
	for _, part := range parts {
		if opts["date"] == "" && looksLikeDateArg(part) {
			opts["date"] = part
		} else if opts["team_id"] == "" {
			opts["team_id"] = part
		}
	}
	return opts
}

func looksLikeDateArg(token string) bool {
	if token == "today" || token == "tomorrow" || token == "week" {
		return true
	}
	if strings.HasPrefix(token, "+") || strings.Contains(token, "..") {
		return true
	}
	return len(token) >= 8 && token[0] >= '0' && token[0] <= '9' && strings.Contains(token, "-")
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
		if dateStr != "today" && dateStr != "tomorrow" && !looksLikeDateArg(dateStr) {
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
	if len(parts) < 3 {
		return opts
	}

	if scheduleDayStatusBlocksMeals(parts[1]) {
		opts["reason"] = strings.Join(parts[2:], " ")
		return opts
	}

	opts["meals"] = parts[2]
	if len(parts) >= 4 {
		opts["reason"] = strings.Join(parts[3:], " ")
	}

	return opts
}

func scheduleDayStatusBlocksMeals(status string) bool {
	switch strings.ToLower(status) {
	case "office_closed", "govt_holiday":
		return true
	default:
		return false
	}
}
