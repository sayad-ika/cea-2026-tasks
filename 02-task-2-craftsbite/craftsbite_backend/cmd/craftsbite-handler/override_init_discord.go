package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/sayad-ika/craftsbite/internal/cmdutil"
	appconfig "github.com/sayad-ika/craftsbite/internal/config"
	"github.com/sayad-ika/craftsbite/internal/dateutil"
	"github.com/sayad-ika/craftsbite/internal/discord"
	"github.com/sayad-ika/craftsbite/internal/payload"
)

const (
	overrideInitDiscordCustomIDPrefix = "oi"
	overrideInitDiscordCustomIDSep    = "|"
	overrideInitDiscordNoValue        = "-"
	overrideInitDiscordPanelTitle     = "Override Setup"
	overrideInitDiscordPanelDesc      = "Choose a team and target scope, then edit details and save. Bulk overrides require an explicit value."

	overrideInitDiscordActionTeam    = "team"
	overrideInitDiscordActionTarget  = "target"
	overrideInitDiscordActionMembers = "members"
)

func handleDiscordOverrideInitInteraction(ctx context.Context, cfg *appconfig.Config, store overrideInitStore, dateParser *dateutil.DateParser, event payload.CommandEvent) error {
	var opts payload.OverrideInitOptions
	if len(event.Options) > 0 {
		if err := event.ParseOptions(&opts); err != nil {
			return sendDiscordOverrideInitError(ctx, "Invalid `/override-init` interaction payload.", false)
		}
	}
	draft := overrideInitDraftFromOptions(opts)
	if draft.Date == "" {
		if date, err := dateParser.ParseDateWithDefaults(""); err == nil {
			draft.Date = date
		}
	}

	switch draft.Action {
	case overrideInitActionOpen, overrideInitActionTeamChange, overrideInitActionDetailsSave, overrideInitDiscordActionTarget, overrideInitDiscordActionMembers:
		note := ""
		if draft.Action == overrideInitActionTeamChange {
			note = "Team updated. Choose target scope and details, then save."
		} else if draft.Action == overrideInitActionDetailsSave {
			note = "Details updated. Save when ready."
		}
		return renderDiscordOverrideInitPanel(ctx, store, event, draft, note, discord.NoticeToneInfo, draft.Action != overrideInitActionOpen)
	case overrideInitActionDetailsEdit:
		state, err := loadOverrideInitState(ctx, store, event, draft)
		if err != nil {
			return sendDiscordOverrideInitError(ctx, "Unable to load override setup. Please try again shortly.", true)
		}
		return sendDiscordOverrideInitModal(ctx, state)
	case overrideInitActionApply:
		result, err := applyOverrideInitDraft(ctx, store, dateParser, event, draft)
		if err != nil {
			return renderDiscordOverrideInitValidation(ctx, store, event, draft, overrideInitFailureMessage(err))
		}
		return renderDiscordOverrideInitResult(ctx, result)
	case overrideInitActionCancel:
		message := discord.ToneMessage("Override canceled", "No override changes were saved.", discord.NoticeToneInfo)
		return sendDiscordInteractionResponse(ctx, updateMessage(message))
	default:
		return renderDiscordOverrideInitPanel(ctx, store, event, draft, "This `/override-init` action is not supported.", discord.NoticeToneWarning, true)
	}
}

func renderDiscordOverrideInitValidation(ctx context.Context, store overrideInitStore, event payload.CommandEvent, draft overrideInitDraft, note string) error {
	return renderDiscordOverrideInitPanel(ctx, store, event, draft, "Review needed: "+note, discord.NoticeToneWarning, true)
}

func renderDiscordOverrideInitPanel(ctx context.Context, store overrideInitStore, event payload.CommandEvent, draft overrideInitDraft, note string, tone discord.NoticeTone, update bool) error {
	state, err := loadOverrideInitState(ctx, store, event, draft)
	if err != nil {
		return sendDiscordOverrideInitError(ctx, "Unable to load override setup. Please try again shortly.", update)
	}
	message := buildDiscordOverrideInitMessage(state, note, tone)
	if update {
		return sendDiscordInteractionResponse(ctx, updateMessage(message))
	}
	return sendDiscordInteractionResponse(ctx, ephemeralMessage(message))
}

func renderDiscordOverrideInitResult(ctx context.Context, result *overrideInitBulkResult) error {
	tone := discord.NoticeToneSuccess
	if result == nil || result.Succeeded == 0 {
		tone = discord.NoticeToneWarning
	} else if result.Failed > 0 {
		tone = discord.NoticeToneWarning
	}
	message := discord.ToneMessage(overrideInitResultTitle(result), strings.Join(overrideInitResultLines(result), "\n"), tone)
	return sendDiscordInteractionResponse(ctx, updateMessage(message))
}

func sendDiscordOverrideInitError(ctx context.Context, text string, update bool) error {
	message := discord.ToneMessage(discord.DefaultNoticeTitle(discord.NoticeToneError), text, discord.NoticeToneError)
	if update {
		return sendDiscordInteractionResponse(ctx, updateMessage(message))
	}
	return sendDiscordInteractionResponse(ctx, ephemeralMessage(message))
}

func buildDiscordOverrideInitMessage(state overrideInitState, note string, tone discord.NoticeTone) discord.Message {
	description := overrideInitDiscordPanelDesc
	if note != "" {
		description = note + "\n" + description
	}
	embed := discord.BrandEmbed(overrideInitDiscordPanelTitle, description, discordOverrideInitFields(state))
	if note != "" {
		embed = discord.ToneEmbed(overrideInitDiscordPanelTitle, description, discordOverrideInitFields(state), tone)
	}
	message := discord.EmbedMessage(embed)
	message.Components = discordOverrideInitComponents(state)
	return message
}

func discordOverrideInitFields(state overrideInitState) []discord.EmbedField {
	team := "Choose a team"
	if state.Team != nil {
		team = overrideInitTeamLabel(*state.Team)
	}
	target := "Choose target scope"
	switch state.Draft.TargetMode {
	case overrideInitTargetWholeTeam:
		target = fmt.Sprintf("Whole team (%d active members)", len(state.Members))
	case overrideInitTargetMembers:
		target = discordOverrideInitSelectedMembersSummary(state)
	}
	entry := "Choose entry"
	if state.Draft.Entry == "meal" {
		meal := state.Draft.Meal
		if meal == "" {
			meal = "all"
		}
		entry = "Meal: " + discordOverrideInitMealLabel(meal)
	} else if state.Draft.Entry == "location" {
		entry = "Location"
	}
	value := discordOverrideInitValueLabel(state.Draft.Entry, state.Draft.Value)
	if value == "" {
		value = "Choose explicit value"
	}
	reason := state.Draft.Reason
	if reason == "" {
		reason = "None"
	}
	return []discord.EmbedField{
		{Name: "Team", Value: team, Inline: true},
		{Name: "Target", Value: target, Inline: true},
		{Name: "Date", Value: displayAdminInitDate(state.Draft.Date), Inline: true},
		{Name: "Entry", Value: entry, Inline: true},
		{Name: "Value", Value: value, Inline: true},
		{Name: "Reason", Value: reason},
	}
}

func discordOverrideInitSelectedMembersSummary(state overrideInitState) string {
	if len(state.Draft.Members) == 0 {
		return "Specific members: none selected"
	}
	labels := make([]string, 0, len(state.Draft.Members))
	for _, ref := range state.Draft.Members {
		member := resolveOverrideInitMemberRef(state.Members, ref)
		if member == nil {
			continue
		}
		labels = append(labels, overrideInitUserLabel(*member))
	}
	if len(labels) == 0 {
		return "Specific members: none selected"
	}
	return "Specific members:\n" + strings.Join(labels, "\n")
}

func discordOverrideInitComponents(state overrideInitState) []discord.Component {
	components := []discord.Component{}
	if len(state.Teams) > 0 {
		components = append(components, discord.Component{Type: discord.ComponentTypeActionRow, Components: []discord.Component{discordOverrideInitTeamSelect(state)}})
	}
	components = append(components, discord.Component{Type: discord.ComponentTypeActionRow, Components: []discord.Component{discordOverrideInitTargetSelect(state)}})
	if state.Team != nil && state.Draft.TargetMode == overrideInitTargetMembers {
		components = append(components, discord.Component{Type: discord.ComponentTypeActionRow, Components: []discord.Component{discordOverrideInitMemberSelect(state)}})
	}
	components = append(components, discord.Component{Type: discord.ComponentTypeActionRow, Components: []discord.Component{
		{Type: discord.ComponentTypeButton, Style: discord.ButtonStylePrimary, Label: "Edit Details", CustomID: discordOverrideInitCustomID(overrideInitActionDetailsEdit, state)},
		{Type: discord.ComponentTypeButton, Style: discord.ButtonStyleSuccess, Label: "Save", CustomID: discordOverrideInitCustomID(overrideInitActionApply, state)},
		{Type: discord.ComponentTypeButton, Style: discord.ButtonStyleSecondary, Label: "Cancel", CustomID: discordOverrideInitCustomID(overrideInitActionCancel, state)},
	}})
	return components
}

func discordOverrideInitTeamSelect(state overrideInitState) discord.Component {
	minValues := 1
	maxValues := 1
	options := make([]discord.SelectOption, 0, len(state.Teams))
	selectedIndex := discordOverrideInitTeamIndex(state)
	for i, team := range state.Teams {
		if i >= 25 {
			break
		}
		options = append(options, discord.SelectOption{Label: truncateDiscordOptionLabel(overrideInitTeamLabel(team)), Value: overrideInitIndexRef(i), Default: i == selectedIndex})
	}
	return discord.Component{Type: discord.ComponentTypeStringSelect, CustomID: discordOverrideInitCustomID(overrideInitDiscordActionTeam, state), Placeholder: "Choose team", Options: options, MinValues: &minValues, MaxValues: &maxValues}
}

func discordOverrideInitTargetSelect(state overrideInitState) discord.Component {
	minValues := 1
	maxValues := 1
	options := []discord.SelectOption{
		{Label: "Whole team", Value: overrideInitTargetWholeTeam, Default: state.Draft.TargetMode == overrideInitTargetWholeTeam},
		{Label: "Specific members", Value: overrideInitTargetMembers, Default: state.Draft.TargetMode == overrideInitTargetMembers},
	}
	return discord.Component{Type: discord.ComponentTypeStringSelect, CustomID: discordOverrideInitCustomID(overrideInitDiscordActionTarget, state), Placeholder: "Choose target scope", Options: options, MinValues: &minValues, MaxValues: &maxValues}
}

func discordOverrideInitMemberSelect(state overrideInitState) discord.Component {
	minValues := 0
	maxValues := len(state.Members)
	if maxValues > 25 {
		maxValues = 25
	}
	selected := discordOverrideInitSelectedMemberIDs(state)
	options := make([]discord.SelectOption, 0, maxValues)
	for i, member := range state.Members {
		if i >= 25 {
			break
		}
		options = append(options, discord.SelectOption{Label: truncateDiscordOptionLabel(overrideInitUserLabel(member)), Value: overrideInitIndexRef(i), Description: truncateDiscordOptionLabel(member.Email), Default: selected[member.ID]})
	}
	return discord.Component{Type: discord.ComponentTypeStringSelect, CustomID: discordOverrideInitCustomID(overrideInitDiscordActionMembers, state), Placeholder: "Choose member(s)", Options: options, MinValues: &minValues, MaxValues: &maxValues}
}

func discordOverrideInitSelectedMemberIDs(state overrideInitState) map[string]bool {
	selected := make(map[string]bool, len(state.Draft.Members))
	for _, ref := range state.Draft.Members {
		member := resolveOverrideInitMemberRef(state.Members, ref)
		if member != nil {
			selected[member.ID] = true
		}
	}
	return selected
}

func discordOverrideInitTeamIndex(state overrideInitState) int {
	if state.Team == nil {
		return -1
	}
	for i := range state.Teams {
		if state.Teams[i].ID == state.Team.ID {
			return i
		}
	}
	return -1
}

func sendDiscordOverrideInitModal(ctx context.Context, state overrideInitState) error {
	required := true
	optional := false
	meal := state.Draft.Meal
	if meal == "" {
		meal = "all"
	}
	modal := discord.Message{Title: "Edit Override", CustomID: discordOverrideInitCustomID(overrideInitActionDetailsSave, state), Components: []discord.Component{
		{Type: discord.ComponentTypeActionRow, Components: []discord.Component{{Type: discord.ComponentTypeTextInput, CustomID: "date", Label: "Date", Style: discord.TextInputStyleShort, Value: state.Draft.Date, Placeholder: "YYYY-MM-DD, today, tomorrow, or +N", Required: &required}}},
		{Type: discord.ComponentTypeActionRow, Components: []discord.Component{{Type: discord.ComponentTypeTextInput, CustomID: "entry", Label: "Entry", Style: discord.TextInputStyleShort, Value: state.Draft.Entry, Placeholder: "meal or location", Required: &required}}},
		{Type: discord.ComponentTypeActionRow, Components: []discord.Component{{Type: discord.ComponentTypeTextInput, CustomID: "meal", Label: "Meal", Style: discord.TextInputStyleShort, Value: meal, Placeholder: "all, lunch, snacks, iftar, event_dinner, optional_dinner", Required: &optional}}},
		{Type: discord.ComponentTypeActionRow, Components: []discord.Component{{Type: discord.ComponentTypeTextInput, CustomID: "value", Label: "Value", Style: discord.TextInputStyleShort, Value: state.Draft.Value, Placeholder: "meal: in/out; location: office/wfh", Required: &required}}},
		{Type: discord.ComponentTypeActionRow, Components: []discord.Component{{Type: discord.ComponentTypeTextInput, CustomID: "reason", Label: "Reason", Style: discord.TextInputStyleParagraph, Value: state.Draft.Reason, Placeholder: "Reason for the override", Required: &required}}},
	}}
	return sendDiscordInteractionResponse(ctx, RouterResponse{Type: 9, Data: &modal})
}

func discordOverrideInitComponentPayload(interaction interactionBody) (json.RawMessage, error) {
	action, draft, err := parseDiscordOverrideInitCustomID(interaction.Data.CustomID)
	if err != nil {
		return nil, fmt.Errorf("This interactive control is no longer recognized. Please run `/override-init` again.")
	}
	draft.Reason = discordOverrideInitReasonFromMessage(interaction.Message)
	if interaction.Type == 5 {
		for _, row := range interaction.Data.Components {
			for _, component := range row.Components {
				switch component.CustomID {
				case "date":
					draft.Date = strings.TrimSpace(component.Value)
				case "entry":
					draft.Entry = strings.ToLower(strings.TrimSpace(component.Value))
				case "meal":
					draft.Meal = strings.ToLower(strings.TrimSpace(component.Value))
				case "value":
					draft.Value = strings.ToLower(strings.TrimSpace(component.Value))
				case "reason":
					draft.Reason = strings.TrimSpace(component.Value)
				}
			}
		}
	}

	switch action {
	case overrideInitDiscordActionTeam:
		draft.Action = overrideInitActionTeamChange
		if len(interaction.Data.Values) == 1 {
			draft.TeamID = interaction.Data.Values[0]
			draft.Members = nil
		}
	case overrideInitDiscordActionTarget:
		draft.Action = overrideInitDiscordActionTarget
		if len(interaction.Data.Values) == 1 {
			draft.TargetMode = interaction.Data.Values[0]
			if draft.TargetMode != overrideInitTargetMembers {
				draft.Members = nil
			}
		}
	case overrideInitDiscordActionMembers:
		draft.Action = overrideInitDiscordActionMembers
		draft.Members = append([]string(nil), interaction.Data.Values...)
	case overrideInitActionDetailsSave:
		draft.Action = overrideInitActionDetailsSave
	default:
		draft.Action = action
	}

	raw, _ := json.Marshal(overrideInitOptionsFromDraft(draft))
	return raw, nil
}

func discordOverrideInitReasonFromMessage(message discord.Message) string {
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

func discordOverrideInitCustomID(action string, state overrideInitState) string {
	date := compactAdminInitDate(state.Draft.Date)
	if date == "" {
		date = overrideInitDiscordNoValue
	}
	teamIndex := overrideInitDiscordNoValue
	if idx := discordOverrideInitTeamIndex(state); idx >= 0 {
		teamIndex = strconv.FormatInt(int64(idx), 36)
	}
	members := encodeDiscordOverrideInitMembers(state)
	if members == "" {
		members = overrideInitDiscordNoValue
	}
	return strings.Join([]string{
		overrideInitDiscordCustomIDPrefix,
		encodeDiscordOverrideInitAction(action),
		date,
		teamIndex,
		compactDiscordOverrideInitTargetMode(state.Draft.TargetMode),
		compactDiscordOverrideInitEntry(state.Draft.Entry),
		compactDiscordOverrideInitMeal(state.Draft.Meal),
		compactDiscordOverrideInitValue(state.Draft.Value),
		members,
	}, overrideInitDiscordCustomIDSep)
}

func parseDiscordOverrideInitCustomID(id string) (string, overrideInitDraft, error) {
	parts := strings.Split(id, overrideInitDiscordCustomIDSep)
	if len(parts) != 9 || parts[0] != overrideInitDiscordCustomIDPrefix {
		return "", overrideInitDraft{}, fmt.Errorf("unrecognized override setup control")
	}
	draft := overrideInitDraft{Date: expandInitDate(parts[2])}
	if parts[2] == overrideInitDiscordNoValue {
		draft.Date = ""
	}
	if parts[3] != overrideInitDiscordNoValue {
		idx, err := strconv.ParseInt(parts[3], 36, 0)
		if err != nil {
			return "", overrideInitDraft{}, err
		}
		draft.TeamID = overrideInitIndexRef(int(idx))
	}
	draft.TargetMode = expandDiscordOverrideInitTargetMode(parts[4])
	draft.Entry = expandDiscordOverrideInitEntry(parts[5])
	draft.Meal = expandDiscordOverrideInitMeal(parts[6])
	draft.Value = expandDiscordOverrideInitValue(parts[7])
	if parts[8] != overrideInitDiscordNoValue {
		draft.Members = decodeDiscordOverrideInitMembers(parts[8])
	}
	return decodeDiscordOverrideInitAction(parts[1]), draft, nil
}

func encodeDiscordOverrideInitMembers(state overrideInitState) string {
	if len(state.Draft.Members) == 0 || len(state.Members) == 0 {
		return ""
	}
	parts := make([]string, 0, len(state.Draft.Members))
	seen := make(map[int]struct{}, len(state.Draft.Members))
	for _, ref := range state.Draft.Members {
		member := resolveOverrideInitMemberRef(state.Members, ref)
		if member == nil {
			continue
		}
		for i := range state.Members {
			if state.Members[i].ID == member.ID {
				if _, ok := seen[i]; ok {
					break
				}
				seen[i] = struct{}{}
				parts = append(parts, strconv.FormatInt(int64(i), 36))
				break
			}
		}
	}
	return strings.Join(parts, ".")
}

func decodeDiscordOverrideInitMembers(encoded string) []string {
	parts := strings.Split(encoded, ".")
	members := make([]string, 0, len(parts))
	for _, part := range parts {
		if part == "" {
			continue
		}
		idx, err := strconv.ParseInt(part, 36, 0)
		if err != nil {
			continue
		}
		members = append(members, overrideInitIndexRef(int(idx)))
	}
	return members
}

func encodeDiscordOverrideInitAction(action string) string {
	switch action {
	case overrideInitDiscordActionTeam:
		return "t"
	case overrideInitDiscordActionTarget:
		return "g"
	case overrideInitDiscordActionMembers:
		return "m"
	case overrideInitActionDetailsEdit:
		return "e"
	case overrideInitActionDetailsSave:
		return "d"
	case overrideInitActionApply:
		return "a"
	case overrideInitActionCancel:
		return "c"
	default:
		return action
	}
}

func decodeDiscordOverrideInitAction(action string) string {
	switch action {
	case "t":
		return overrideInitDiscordActionTeam
	case "g":
		return overrideInitDiscordActionTarget
	case "m":
		return overrideInitDiscordActionMembers
	case "e":
		return overrideInitActionDetailsEdit
	case "d":
		return overrideInitActionDetailsSave
	case "a":
		return overrideInitActionApply
	case "c":
		return overrideInitActionCancel
	default:
		return action
	}
}

func compactDiscordOverrideInitTargetMode(mode string) string {
	switch mode {
	case overrideInitTargetWholeTeam:
		return "w"
	case overrideInitTargetMembers:
		return "m"
	default:
		return overrideInitDiscordNoValue
	}
}

func expandDiscordOverrideInitTargetMode(mode string) string {
	switch mode {
	case "w":
		return overrideInitTargetWholeTeam
	case "m":
		return overrideInitTargetMembers
	default:
		return ""
	}
}

func compactDiscordOverrideInitEntry(entry string) string {
	switch entry {
	case "meal":
		return "m"
	case "location":
		return "l"
	default:
		return overrideInitDiscordNoValue
	}
}

func expandDiscordOverrideInitEntry(entry string) string {
	switch entry {
	case "m":
		return "meal"
	case "l":
		return "location"
	default:
		return ""
	}
}

func compactDiscordOverrideInitMeal(meal string) string {
	if meal == "" {
		return overrideInitDiscordNoValue
	}
	if meal == "all" {
		return "a"
	}
	return compactInitMeal(meal)
}

func expandDiscordOverrideInitMeal(meal string) string {
	if meal == overrideInitDiscordNoValue {
		return ""
	}
	if meal == "a" {
		return "all"
	}
	return expandInitMeal(meal)
}

func compactDiscordOverrideInitValue(value string) string {
	switch value {
	case "in":
		return "i"
	case "out":
		return "o"
	case "office":
		return "f"
	case "wfh":
		return "w"
	default:
		return overrideInitDiscordNoValue
	}
}

func expandDiscordOverrideInitValue(value string) string {
	switch value {
	case "i":
		return "in"
	case "o":
		return "out"
	case "f":
		return "office"
	case "w":
		return "wfh"
	default:
		return ""
	}
}

func discordOverrideInitMealLabel(meal string) string {
	if meal == "" || meal == "all" {
		return "All meals"
	}
	return cmdutil.DisplayMealName(meal)
}

func discordOverrideInitValueLabel(entry, value string) string {
	if entry == "meal" {
		switch value {
		case "in":
			return "Included"
		case "out":
			return "Opted out"
		}
	}
	if entry == "location" {
		return displayLocationLabel(value)
	}
	return ""
}

func truncateDiscordOptionLabel(value string) string {
	if len(value) <= 100 {
		return value
	}
	return value[:97] + "..."
}
