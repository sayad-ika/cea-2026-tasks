package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/sayad-ika/craftsbite/internal/cmdutil"
	appconfig "github.com/sayad-ika/craftsbite/internal/config"
	"github.com/sayad-ika/craftsbite/internal/dateutil"
	"github.com/sayad-ika/craftsbite/internal/discord"
	"github.com/sayad-ika/craftsbite/internal/headcountreport"
	"github.com/sayad-ika/craftsbite/internal/payload"
	"github.com/sayad-ika/craftsbite/internal/repository"
	"github.com/sayad-ika/craftsbite/internal/services"
)

const (
	adminInitDiscordActionOpen       = "open"
	adminInitDiscordActionEdit       = "edit"
	adminInitDiscordActionModalSave  = "modal_save"
	adminInitDiscordActionSave       = "save"
	adminInitDiscordActionCancel     = "cancel"
	adminInitDiscordActionStatus     = "status:"
	adminInitDiscordActionMeal       = "meal:"
	adminInitDiscordCustomIDPrefix   = "ai"
	adminInitDiscordCustomIDSep      = "|"
	adminInitDiscordNoValue          = "-"
	adminInitDiscordPanelTitle       = "Admin Setup"
	adminInitDiscordPanelDescription = "Use Edit Details to change dates or reason. Leave End date empty for a single date; fill both dates for a range."
)

func handleDiscordAdminInitInteraction(ctx context.Context, cfg *appconfig.Config, store services.DayScheduleWriter, dateParser *dateutil.DateParser, event payload.CommandEvent) error {
	var opts payload.AdminInitOptions
	if len(event.Options) > 0 {
		if err := event.ParseOptions(&opts); err != nil {
			return sendDiscordAdminInitError(ctx, "Invalid `/admin-init` interaction payload.", false)
		}
	}
	if opts.Action == "" {
		opts.Action = adminInitDiscordActionOpen
	}

	switch opts.Action {
	case adminInitDiscordActionOpen:
		return openDiscordAdminInitPanel(ctx, store, dateParser, opts, false)
	case adminInitDiscordActionEdit:
		return sendDiscordAdminInitModal(ctx, adminInitDraftFromOptions(opts))
	case adminInitDiscordActionModalSave:
		draft := adminInitDraftFromOptions(opts)
		draft.UseRange = strings.TrimSpace(draft.EndDate) != ""
		return renderDiscordAdminInitPanel(ctx, draft, "Details updated. Save when ready.", discord.NoticeToneInfo, true)
	case adminInitActionScheduleApply, adminInitDiscordActionSave:
		return saveDiscordAdminInitSchedule(ctx, store, dateParser, event, opts)
	case adminInitActionScheduleCancel, adminInitDiscordActionCancel:
		return renderDiscordAdminInitResult(ctx, "Setup canceled", "No schedule changes were saved.", discord.NoticeToneInfo)
	default:
		if strings.HasPrefix(opts.Action, adminInitDiscordActionStatus) {
			draft := adminInitDraftFromOptions(opts)
			draft.Status = strings.TrimPrefix(opts.Action, adminInitDiscordActionStatus)
			if adminInitStatusBlocksMeals(draft.Status) {
				draft.Meals = nil
			}
			return renderDiscordAdminInitPanel(ctx, draft, "Day status updated. Save when ready.", discord.NoticeToneInfo, true)
		}
		if strings.HasPrefix(opts.Action, adminInitDiscordActionMeal) {
			draft := adminInitDraftFromOptions(opts)
			draft.Meals = toggleAdminInitMeal(draft.Meals, strings.TrimPrefix(opts.Action, adminInitDiscordActionMeal))
			return renderDiscordAdminInitPanel(ctx, draft, "Meals updated. Save when ready.", discord.NoticeToneInfo, true)
		}
		return renderDiscordAdminInitPanel(ctx, adminInitDraftFromOptions(opts), "This `/admin-init` action is not supported.", discord.NoticeToneWarning, true)
	}
}

func openDiscordAdminInitPanel(ctx context.Context, store services.DayScheduleWriter, dateParser *dateutil.DateParser, opts payload.AdminInitOptions, update bool) error {
	date, err := dateParser.ParseDateWithDefaults(opts.Date)
	if err != nil {
		return sendDiscordAdminInitError(ctx, fmt.Sprintf("Invalid date: %v", err), update)
	}
	schedule, err := store.GetDay(ctx, date)
	if err != nil {
		return sendDiscordAdminInitError(ctx, "Unable to load the current schedule. Please try again shortly.", update)
	}
	draft := adminInitDraftFromSchedule(date, schedule)
	draft.UseRange = false
	return renderDiscordAdminInitPanel(ctx, draft, "", discord.NoticeToneInfo, update)
}

func saveDiscordAdminInitSchedule(ctx context.Context, store services.DayScheduleWriter, dateParser *dateutil.DateParser, event payload.CommandEvent, opts payload.AdminInitOptions) error {
	draft := adminInitDraftFromOptions(opts)
	dates, err := discordAdminInitTargetDates(dateParser, draft)
	if err != nil {
		return renderDiscordAdminInitPanel(ctx, draft, fmt.Sprintf("Review needed: Invalid date: %v", err), discord.NoticeToneWarning, true)
	}
	draft.Date = dates[0]
	draft.UseRange = len(dates) > 1
	if draft.UseRange {
		draft.EndDate = dates[len(dates)-1]
	} else {
		draft.EndDate = ""
	}
	if draft.Status == "" {
		return renderDiscordAdminInitPanel(ctx, draft, "Review needed: Choose a day status before saving.", discord.NoticeToneWarning, true)
	}
	if adminInitStatusBlocksMeals(draft.Status) {
		draft.Meals = nil
	}

	input := services.SetDayScheduleInput{Date: draft.Date, DayStatus: draft.Status, AvailableMeals: draft.Meals, Reason: draft.Reason, SetBy: event.UserID}
	if len(dates) > 1 {
		result, err := services.BulkSetDaySchedule(ctx, store, dates, input)
		if err != nil {
			return renderDiscordAdminInitPanel(ctx, draft, "Review needed: "+formatBulkScheduleDayError(err), discord.NoticeToneWarning, true)
		}
		description := fmt.Sprintf("Status: %s\nMeals: %s\nWeekdays updated: %d\nAffected dates: %s%s", headcountreport.DisplayDayStatus(draft.Status), discordAdminInitMealSummary(draft.Meals), len(result.SuccessDates), strings.Join(result.SuccessDates, ", "), discordAdminInitReasonLine(draft.Reason))
		return renderDiscordAdminInitResult(ctx, "Schedule updated", description, discord.NoticeToneSuccess)
	}

	schedule, err := services.SetDaySchedule(ctx, store, input)
	if err != nil {
		return renderDiscordAdminInitPanel(ctx, draft, "Review needed: "+formatScheduleDayError(err), discord.NoticeToneWarning, true)
	}
	description := fmt.Sprintf("Date: %s\nStatus: %s\nMeals: %s%s", schedule.Date, headcountreport.DisplayDayStatus(schedule.DayStatus), discordAdminInitMealSummary(schedule.AvailableMeals), discordAdminInitReasonLine(schedule.Reason))
	return renderDiscordAdminInitResult(ctx, "Schedule updated", description, discord.NoticeToneSuccess)
}

func discordAdminInitTargetDates(dateParser *dateutil.DateParser, draft adminInitScheduleDraft) ([]string, error) {
	startDate, err := dateParser.ParseDateWithDefaults(draft.Date)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(draft.EndDate) == "" {
		return []string{startDate}, nil
	}
	endDate, err := dateParser.ParseDateWithDefaults(draft.EndDate)
	if err != nil {
		return nil, err
	}
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return nil, err
	}
	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return nil, err
	}
	if !end.After(start) {
		return nil, fmt.Errorf("end date must be after start date")
	}
	return dateParser.ParseDateRange(startDate + ".." + endDate)
}

func renderDiscordAdminInitPanel(ctx context.Context, draft adminInitScheduleDraft, note string, tone discord.NoticeTone, update bool) error {
	message := buildDiscordAdminInitMessage(draft, note, tone)
	if update {
		return sendDiscordInteractionResponse(ctx, updateMessage(message))
	}
	return sendDiscordInteractionResponse(ctx, ephemeralMessage(message))
}

func renderDiscordAdminInitResult(ctx context.Context, title, description string, tone discord.NoticeTone) error {
	message := discord.EmbedMessage(discord.ToneEmbed(title, description, nil, tone))
	return sendDiscordInteractionResponse(ctx, updateMessage(message))
}

func sendDiscordAdminInitError(ctx context.Context, text string, update bool) error {
	message := discord.ToneMessage(discord.DefaultNoticeTitle(discord.NoticeToneError), text, discord.NoticeToneError)
	if update {
		return sendDiscordInteractionResponse(ctx, updateMessage(message))
	}
	return sendDiscordInteractionResponse(ctx, ephemeralMessage(message))
}

func buildDiscordAdminInitMessage(draft adminInitScheduleDraft, note string, tone discord.NoticeTone) discord.Message {
	description := adminInitDiscordPanelDescription
	if note != "" {
		description = note + "\n" + description
	}
	fields := []discord.EmbedField{
		{Name: "Dates", Value: discordAdminInitDateSummary(draft), Inline: true},
		{Name: "Day status", Value: discordAdminInitDisplayStatus(draft.Status), Inline: true},
		{Name: "Meals", Value: discordAdminInitMealSummary(draft.Meals)},
		{Name: "Reason", Value: discordAdminInitValueOrNone(draft.Reason)},
	}
	embed := discord.BrandEmbed(adminInitDiscordPanelTitle, description, fields)
	if note != "" {
		embed = discord.ToneEmbed(adminInitDiscordPanelTitle, description, fields, tone)
	}
	message := discord.EmbedMessage(embed)
	message.Components = discordAdminInitComponents(draft)
	return message
}

func discordAdminInitComponents(draft adminInitScheduleDraft) []discord.Component {
	components := []discord.Component{
		{Type: discord.ComponentTypeActionRow, Components: []discord.Component{
			{Type: discord.ComponentTypeButton, Style: discord.ButtonStylePrimary, Label: "Edit Details", CustomID: discordAdminInitCustomID(adminInitDiscordActionEdit, draft)},
		}},
	}

	statusButtons := make([]discord.Component, 0, len(repository.ValidDayStatuses()))
	for _, status := range repository.ValidDayStatuses() {
		style := discord.ButtonStyleSecondary
		if draft.Status == status {
			style = discord.ButtonStylePrimary
		}
		statusButtons = append(statusButtons, discord.Component{Type: discord.ComponentTypeButton, Style: style, Label: shortAdminInitStatusLabel(status), CustomID: discordAdminInitCustomID(adminInitDiscordActionStatus+status, draft)})
	}
	for len(statusButtons) > 0 {
		rowSize := len(statusButtons)
		if rowSize > 5 {
			rowSize = 5
		}
		components = append(components, discord.Component{Type: discord.ComponentTypeActionRow, Components: statusButtons[:rowSize]})
		statusButtons = statusButtons[rowSize:]
	}

	mealButtons := make([]discord.Component, 0, len(repository.ValidMealTypes()))
	selected := make(map[string]struct{}, len(draft.Meals))
	for _, meal := range draft.Meals {
		selected[meal] = struct{}{}
	}
	blocked := adminInitStatusBlocksMeals(draft.Status)
	for _, meal := range repository.ValidMealTypes() {
		style := discord.ButtonStyleSecondary
		if _, ok := selected[meal]; ok && !blocked {
			style = discord.ButtonStylePrimary
		}
		mealButtons = append(mealButtons, discord.Component{Type: discord.ComponentTypeButton, Style: style, Label: cmdutil.DisplayMealName(meal), CustomID: discordAdminInitCustomID(adminInitDiscordActionMeal+meal, draft), Disabled: blocked})
	}
	components = append(components, discord.Component{Type: discord.ComponentTypeActionRow, Components: mealButtons})
	components = append(components, discord.Component{Type: discord.ComponentTypeActionRow, Components: []discord.Component{
		{Type: discord.ComponentTypeButton, Style: discord.ButtonStyleSuccess, Label: "Save", CustomID: discordAdminInitCustomID(adminInitDiscordActionSave, draft)},
		{Type: discord.ComponentTypeButton, Style: discord.ButtonStyleSecondary, Label: "Cancel", CustomID: discordAdminInitCustomID(adminInitDiscordActionCancel, draft)},
	}})
	return components
}

func sendDiscordAdminInitModal(ctx context.Context, draft adminInitScheduleDraft) error {
	required := true
	optional := false
	modal := discord.Message{Title: "Edit Admin Setup", CustomID: discordAdminInitCustomID(adminInitDiscordActionModalSave, draft), Components: []discord.Component{
		{Type: discord.ComponentTypeActionRow, Components: []discord.Component{{Type: discord.ComponentTypeTextInput, CustomID: "date", Label: "Start date", Style: discord.TextInputStyleShort, Value: draft.Date, Placeholder: "YYYY-MM-DD, today, tomorrow, or +N", Required: &required}}},
		{Type: discord.ComponentTypeActionRow, Components: []discord.Component{{Type: discord.ComponentTypeTextInput, CustomID: "end_date", Label: "End date", Style: discord.TextInputStyleShort, Value: draft.EndDate, Placeholder: "Optional; fill for a range", Required: &optional}}},
		{Type: discord.ComponentTypeActionRow, Components: []discord.Component{{Type: discord.ComponentTypeTextInput, CustomID: "reason", Label: "Reason", Style: discord.TextInputStyleParagraph, Value: draft.Reason, Placeholder: "Optional reason or context", Required: &optional}}},
	}}
	return sendDiscordInteractionResponse(ctx, RouterResponse{Type: 9, Data: &modal})
}

func discordAdminInitCustomID(action string, draft adminInitScheduleDraft) string {
	endDate := compactAdminInitDate(draft.EndDate)
	if endDate == "" {
		endDate = adminInitDiscordNoValue
	}
	mode := "s"
	if strings.TrimSpace(draft.EndDate) != "" {
		mode = "r"
	}
	status := compactAdminInitStatus(draft.Status)
	if status == "" {
		status = adminInitDiscordNoValue
	}
	meals := encodeInitMeals(draft.Meals)
	if meals == "" {
		meals = adminInitDiscordNoValue
	}
	return strings.Join([]string{adminInitDiscordCustomIDPrefix, encodeAdminInitAction(action), compactAdminInitDate(draft.Date), endDate, mode, status, meals}, adminInitDiscordCustomIDSep)
}

func parseDiscordAdminInitCustomID(id string) (string, adminInitScheduleDraft, error) {
	parts := strings.Split(id, adminInitDiscordCustomIDSep)
	if len(parts) != 7 || parts[0] != adminInitDiscordCustomIDPrefix {
		return "", adminInitScheduleDraft{}, fmt.Errorf("unrecognized admin setup control")
	}
	draft := adminInitScheduleDraft{Date: expandInitDate(parts[2]), EndDate: expandInitDate(parts[3])}
	if parts[3] == adminInitDiscordNoValue {
		draft.EndDate = ""
	}
	draft.UseRange = draft.EndDate != ""
	if parts[5] != adminInitDiscordNoValue {
		draft.Status = expandAdminInitStatus(parts[5])
	}
	if parts[6] != adminInitDiscordNoValue {
		draft.Meals = decodeInitMeals(parts[6])
	}
	return decodeAdminInitAction(parts[1]), draft, nil
}

func encodeAdminInitAction(action string) string {
	if strings.HasPrefix(action, adminInitDiscordActionStatus) {
		return adminInitDiscordActionStatus + compactAdminInitStatus(strings.TrimPrefix(action, adminInitDiscordActionStatus))
	}
	if strings.HasPrefix(action, adminInitDiscordActionMeal) {
		return adminInitDiscordActionMeal + compactInitMeal(strings.TrimPrefix(action, adminInitDiscordActionMeal))
	}
	return action
}

func decodeAdminInitAction(action string) string {
	if strings.HasPrefix(action, adminInitDiscordActionStatus) {
		return adminInitDiscordActionStatus + expandAdminInitStatus(strings.TrimPrefix(action, adminInitDiscordActionStatus))
	}
	if strings.HasPrefix(action, adminInitDiscordActionMeal) {
		return adminInitDiscordActionMeal + expandInitMeal(strings.TrimPrefix(action, adminInitDiscordActionMeal))
	}
	return action
}

func compactAdminInitDate(date string) string { return strings.ReplaceAll(date, "-", "") }

func compactAdminInitStatus(status string) string {
	switch status {
	case "normal":
		return "n"
	case "office_closed":
		return "c"
	case "govt_holiday":
		return "g"
	case "celebration":
		return "b"
	case "weekend":
		return "w"
	case "event_day":
		return "e"
	default:
		return status
	}
}

func expandAdminInitStatus(status string) string {
	switch status {
	case "n":
		return "normal"
	case "c":
		return "office_closed"
	case "g":
		return "govt_holiday"
	case "b":
		return "celebration"
	case "w":
		return "weekend"
	case "e":
		return "event_day"
	default:
		return status
	}
}

func discordAdminInitOptionsFromDraft(action string, draft adminInitScheduleDraft) payload.AdminInitOptions {
	draft.UseRange = strings.TrimSpace(draft.EndDate) != ""
	return payload.AdminInitOptions{Action: action, Date: draft.Date, EndDate: draft.EndDate, UseRange: draft.UseRange, Status: draft.Status, Meals: append([]string(nil), draft.Meals...), Reason: draft.Reason}
}

func discordAdminInitComponentPayload(interaction interactionBody) (json.RawMessage, error) {
	action, draft, err := parseDiscordAdminInitCustomID(interaction.Data.CustomID)
	if err != nil {
		return nil, fmt.Errorf("This interactive control is no longer recognized. Please run `/admin-init` again.")
	}
	draft.Reason = discordAdminInitReasonFromMessage(interaction.Message)
	if interaction.Type == 5 {
		for _, row := range interaction.Data.Components {
			for _, component := range row.Components {
				switch component.CustomID {
				case "date":
					draft.Date = strings.TrimSpace(component.Value)
				case "end_date":
					draft.EndDate = strings.TrimSpace(component.Value)
				case "reason":
					draft.Reason = strings.TrimSpace(component.Value)
				}
			}
		}
		draft.UseRange = strings.TrimSpace(draft.EndDate) != ""
	}
	raw, _ := json.Marshal(discordAdminInitOptionsFromDraft(action, draft))
	return raw, nil
}

func discordAdminInitReasonFromMessage(message discord.Message) string {
	if len(message.Embeds) == 0 {
		return ""
	}
	for _, field := range message.Embeds[0].Fields {
		if field.Name == "Reason" && field.Value != "None" {
			return field.Value
		}
	}
	return ""
}

func toggleAdminInitMeal(meals []string, meal string) []string { return toggleInitMeal(meals, meal) }

func discordAdminInitDateSummary(draft adminInitScheduleDraft) string {
	if strings.TrimSpace(draft.EndDate) != "" {
		return fmt.Sprintf("Range\nStart: %s\nEnd: %s", displayAdminInitDate(draft.Date), displayAdminInitDate(draft.EndDate))
	}
	return "Single date\n" + displayAdminInitDate(draft.Date)
}

func discordAdminInitDisplayStatus(status string) string {
	if status == "" {
		return "None"
	}
	return headcountreport.DisplayDayStatus(status)
}

func discordAdminInitMealSummary(meals []string) string {
	if len(meals) == 0 {
		return "None"
	}
	labels := make([]string, 0, len(meals))
	for _, meal := range meals {
		labels = append(labels, cmdutil.DisplayMealName(meal))
	}
	return strings.Join(labels, ", ")
}

func discordAdminInitValueOrNone(value string) string {
	if strings.TrimSpace(value) == "" {
		return "None"
	}
	return value
}

func discordAdminInitReasonLine(reason string) string {
	if strings.TrimSpace(reason) == "" {
		return ""
	}
	return "\nReason: " + reason
}

func displayAdminInitDate(date string) string {
	if strings.TrimSpace(date) == "" {
		return "None"
	}
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return date
	}
	return t.Format("Mon, Jan 2")
}

func shortAdminInitStatusLabel(status string) string {
	switch status {
	case "office_closed":
		return "Closed"
	case "govt_holiday":
		return "Holiday"
	case "event_day":
		return "Event"
	default:
		return headcountreport.DisplayDayStatus(status)
	}
}
