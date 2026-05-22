package main

import (
	"context"
	"fmt"
	"html"
	"log/slog"
	"strings"

	"github.com/sayad-ika/craftsbite/internal/cmdutil"
	appconfig "github.com/sayad-ika/craftsbite/internal/config"
	"github.com/sayad-ika/craftsbite/internal/dateutil"
	"github.com/sayad-ika/craftsbite/internal/discord"
	"github.com/sayad-ika/craftsbite/internal/gchat"
	"github.com/sayad-ika/craftsbite/internal/payload"
	"github.com/sayad-ika/craftsbite/internal/repository"
)

func handleGChatOverrideInitInteraction(ctx context.Context, cfg *appconfig.Config, store overrideInitStore, dateParser *dateutil.DateParser, event payload.CommandEvent) error {
	var opts payload.OverrideInitOptions
	if len(event.Options) > 0 {
		if err := event.ParseOptions(&opts); err != nil {
			return sendWarningReply(ctx, cfg, event, "Invalid `/override-init` interaction payload.")
		}
	}
	draft := overrideInitDraftFromOptions(opts)
	if draft.Date == "" {
		if date, err := dateParser.ParseDateWithDefaults(""); err == nil {
			draft.Date = date
		}
	}

	switch draft.Action {
	case overrideInitActionApply:
		return saveGChatOverrideInit(ctx, cfg, store, dateParser, event, draft)
	case overrideInitActionCancel:
		return sendGChatOverrideInitResultCard(ctx, cfg, event, "Override canceled", "No override changes were saved.", discord.NoticeToneInfo)
	case overrideInitActionOpen, overrideInitActionTeamChange, overrideInitActionEdit:
		state, err := loadOverrideInitState(ctx, store, event, draft)
		if err != nil {
			slog.Error("override init state load failed", "error", err)
			return sendErrorReply(ctx, cfg, event, "Unable to load override setup. Please try again shortly.")
		}
		note := ""
		if draft.Action == overrideInitActionTeamChange {
			note = "Team updated. Review the member list before saving."
		}
		if draft.Action == overrideInitActionOpen {
			return sendGChatOverrideInitCard(ctx, cfg, event, gchatOverrideInitCardInput(state, note))
		}
		return sendGChatOverrideInitCardUpdate(ctx, cfg, event, gchatOverrideInitCardInput(state, note))
	default:
		state, err := loadOverrideInitState(ctx, store, event, draft)
		if err != nil {
			slog.Error("override init state load failed", "error", err)
			return sendErrorReply(ctx, cfg, event, "Unable to load override setup. Please try again shortly.")
		}
		return sendGChatOverrideInitCardUpdate(ctx, cfg, event, gchatOverrideInitCardInput(state, "Unknown override setup action."))
	}
}

func saveGChatOverrideInit(ctx context.Context, cfg *appconfig.Config, store overrideInitStore, dateParser *dateutil.DateParser, event payload.CommandEvent, draft overrideInitDraft) error {
	result, err := applyOverrideInitDraft(ctx, store, dateParser, event, draft)
	if err != nil {
		state, stateErr := loadOverrideInitState(ctx, store, event, draft)
		if stateErr != nil {
			slog.Error("override init validation state load failed", "error", stateErr)
			return sendErrorReply(ctx, cfg, event, "Unable to load override setup. Please try again shortly.")
		}
		return sendGChatOverrideInitCardUpdate(ctx, cfg, event, gchatOverrideInitCardInput(state, "<b>Review needed</b><br>"+html.EscapeString(overrideInitFailureMessage(err))))
	}

	tone := discord.NoticeToneSuccess
	if result.Succeeded == 0 || result.Failed > 0 {
		tone = discord.NoticeToneWarning
	}
	return sendGChatOverrideInitResultCard(ctx, cfg, event, overrideInitResultTitle(result), gchat.EscapedOverrideInitSummary(overrideInitResultLines(result)), tone)
}

func gchatOverrideInitCardInput(state overrideInitState, note string) gchat.OverrideInitCardInput {
	return gchat.OverrideInitCardInput{
		Title:          "Override Setup",
		Subtitle:       "Team meal and location override",
		Intro:          "Choose a team, select whole-team or specific-member target scope, then set an explicit override value.",
		Note:           note,
		Date:           state.Draft.Date,
		Reason:         state.Draft.Reason,
		Teams:          gchatOverrideInitTeamItems(state),
		TargetModes:    gchatOverrideInitTargetModeItems(state.Draft.TargetMode),
		Members:        gchatOverrideInitMemberItems(state),
		Entries:        gchatOverrideInitEntryItems(state.Draft.Entry),
		Meals:          gchatOverrideInitMealItems(state.Draft.Meal),
		MealValues:     gchatOverrideInitMealValueItems(state.Draft.Value),
		LocationValues: gchatOverrideInitLocationValueItems(state.Draft.Value),
		SummaryRows:    gchatOverrideInitSummaryRows(state),
	}
}

func gchatOverrideInitTeamItems(state overrideInitState) []gchat.SelectionItem {
	items := make([]gchat.SelectionItem, 0, len(state.Teams))
	selectedTeamID := ""
	if state.Team != nil {
		selectedTeamID = state.Team.ID
	}
	for _, team := range state.Teams {
		items = append(items, gchat.SelectionItem{Text: overrideInitTeamLabel(team), Value: team.ID, Selected: team.ID == selectedTeamID})
	}
	return items
}

func gchatOverrideInitTargetModeItems(selected string) []gchat.SelectionItem {
	return []gchat.SelectionItem{
		{Text: "Whole team", Value: overrideInitTargetWholeTeam, Selected: selected == overrideInitTargetWholeTeam},
		{Text: "Specific members", Value: overrideInitTargetMembers, Selected: selected == overrideInitTargetMembers},
	}
}

func gchatOverrideInitMemberItems(state overrideInitState) []gchat.SelectionItem {
	selected := make(map[string]bool, len(state.Draft.Members))
	for _, ref := range state.Draft.Members {
		member := resolveOverrideInitMemberRef(state.Members, ref)
		if member != nil {
			selected[member.ID] = true
		}
	}
	items := make([]gchat.SelectionItem, 0, len(state.Members))
	for _, member := range state.Members {
		items = append(items, gchat.SelectionItem{Text: overrideInitUserLabel(member), Value: member.ID, Selected: selected[member.ID]})
	}
	return items
}

func gchatOverrideInitEntryItems(selected string) []gchat.SelectionItem {
	return []gchat.SelectionItem{
		{Text: "Meal", Value: "meal", Selected: selected == "meal"},
		{Text: "Location", Value: "location", Selected: selected == "location"},
	}
}

func gchatOverrideInitMealItems(selected string) []gchat.SelectionItem {
	if selected == "" {
		selected = "all"
	}
	items := []gchat.SelectionItem{{Text: "All meals", Value: "all", Selected: selected == "all"}}
	for _, meal := range repository.ValidMealTypes() {
		items = append(items, gchat.SelectionItem{Text: cmdutil.DisplayMealName(meal), Value: meal, Selected: selected == meal})
	}
	return items
}

func gchatOverrideInitMealValueItems(selected string) []gchat.SelectionItem {
	return []gchat.SelectionItem{
		{Text: "Include", Value: "in", Selected: selected == "in"},
		{Text: "Opt out", Value: "out", Selected: selected == "out"},
	}
}

func gchatOverrideInitLocationValueItems(selected string) []gchat.SelectionItem {
	return []gchat.SelectionItem{
		{Text: "Office", Value: "office", Selected: selected == "office"},
		{Text: "WFH", Value: "wfh", Selected: selected == "wfh"},
	}
}

func gchatOverrideInitSummaryRows(state overrideInitState) []gchat.TeamRow {
	team := "None"
	if state.Team != nil {
		team = html.EscapeString(overrideInitTeamLabel(*state.Team))
	}
	target := "None"
	switch state.Draft.TargetMode {
	case overrideInitTargetWholeTeam:
		target = fmt.Sprintf("Whole team (%d active members)", len(state.Members))
	case overrideInitTargetMembers:
		target = fmt.Sprintf("Specific members (%d selected)", len(state.Draft.Members))
	}
	entry := "None"
	if state.Draft.Entry == "meal" {
		meal := state.Draft.Meal
		if meal == "" {
			meal = "all"
		}
		entry = "Meal: " + html.EscapeString(discordOverrideInitMealLabel(meal))
	} else if state.Draft.Entry == "location" {
		entry = "Location"
	}
	value := discordOverrideInitValueLabel(state.Draft.Entry, state.Draft.Value)
	if value == "" {
		value = "None"
	}
	return []gchat.TeamRow{
		{Label: "Team", Value: team},
		{Label: "Target", Value: html.EscapeString(target)},
		{Label: "Date", Value: html.EscapeString(state.Draft.Date)},
		{Label: "Entry", Value: entry},
		{Label: "Value", Value: html.EscapeString(value)},
		{Label: "Reason", Value: overrideInitEscapedValueOrNone(state.Draft.Reason)},
	}
}

func overrideInitEscapedValueOrNone(value string) string {
	if strings.TrimSpace(value) == "" {
		return "None"
	}
	return html.EscapeString(value)
}

func sendGChatOverrideInitCard(ctx context.Context, cfg *appconfig.Config, event payload.CommandEvent, input gchat.OverrideInitCardInput) error {
	if input.ActionFunction == "" {
		input.ActionFunction = gchatActionFunctionFromContext(ctx)
	}
	body, err := gchat.OverrideInitCardResponse(input)
	if err != nil {
		slog.Error("override init card build failed", "error", err)
		return err
	}
	return sendGChatCard(ctx, cfg, event, body)
}

func sendGChatOverrideInitCardUpdate(ctx context.Context, cfg *appconfig.Config, event payload.CommandEvent, input gchat.OverrideInitCardInput) error {
	if input.ActionFunction == "" {
		input.ActionFunction = gchatActionFunctionFromContext(ctx)
	}
	body, err := gchat.OverrideInitCardResponse(input)
	if err != nil {
		slog.Error("override init card build failed", "error", err)
		return err
	}
	return sendGChatCardUpdate(ctx, cfg, event, body)
}

func sendGChatOverrideInitResultCard(ctx context.Context, cfg *appconfig.Config, event payload.CommandEvent, title, summary string, tone discord.NoticeTone) error {
	body, err := gchat.OverrideInitConfirmationCardResponse(gchat.OverrideInitConfirmationInput{
		Title:          title,
		Subtitle:       "Team override result",
		Summary:        summary,
		EditButtonText: "Open override panel",
		ActionFunction: gchatActionFunctionFromContext(ctx),
		Tone:           tone,
	})
	if err != nil {
		slog.Error("override init result card build failed", "error", err)
		return err
	}
	return sendGChatCardUpdate(ctx, cfg, event, body)
}
