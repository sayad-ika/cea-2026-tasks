package main

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/sayad-ika/craftsbite/internal/cmdutil"
	appconfig "github.com/sayad-ika/craftsbite/internal/config"
	"github.com/sayad-ika/craftsbite/internal/dateutil"
	"github.com/sayad-ika/craftsbite/internal/discord"
	"github.com/sayad-ika/craftsbite/internal/gchat"
	"github.com/sayad-ika/craftsbite/internal/payload"
	"github.com/sayad-ika/craftsbite/internal/repository"
	"github.com/sayad-ika/craftsbite/internal/services"
)

func handleManagementCommand(ctx context.Context, deps handlerDeps, event payload.CommandEvent) error {
	switch event.CommandName {
	case "team-summary":
		return handleTeamSummaryCommand(ctx, deps.store, deps.cfg, deps.dateParser, event)
	case "override":
		return sendReply(ctx, deps.cfg, event, "This feature is coming soon.")
	default:
		return sendWarningReply(ctx, deps.cfg, event, fmt.Sprintf("Unknown command: /%s", event.CommandName))
	}
}

func handleTeamSummaryCommand(ctx context.Context, store *repository.Store, cfg *appconfig.Config, dateParser *dateutil.DateParser, event payload.CommandEvent) error {
	if event.Role != "team_lead" && event.Role != "admin" {
		return sendWarningReply(ctx, cfg, event, "You do not have permission to use `/team-summary`.")
	}

	var opts payload.TeamSummaryOptions
	if err := event.ParseOptions(&opts); err != nil {
		return sendWarningReply(ctx, cfg, event, "Invalid command options.")
	}
	date, err := dateParser.ParseDateWithDefaults(opts.Date)
	if err != nil {
		return sendWarningReply(ctx, cfg, event, fmt.Sprintf("Invalid date: %v\nUse: tomorrow (default), +N, or YYYY-MM-DD", err))
	}

	teams, err := store.FindTeamsByLeadID(ctx, event.UserID)
	if err != nil {
		return sendErrorReply(ctx, cfg, event, "Something went wrong fetching your team. Please try again later.")
	}
	if len(teams) == 0 {
		return sendWarningReply(ctx, cfg, event, "You are not assigned as a team lead to any team.")
	}

	team, err := store.GetTeamByID(ctx, teams[0].ID)
	if err != nil || team == nil {
		return sendErrorReply(ctx, cfg, event, "Team not found. Please try again later.")
	}

	summary, err := services.GetTeamSummary(ctx, store, teams[0].ID, date)
	if err != nil {
		return sendErrorReply(ctx, cfg, event, "Something went wrong fetching the team summary. Please try again later.")
	}

	if event.Source == "gchat" {
		rows := buildTeamSummaryRows(team, summary)
		card, _ := gchat.TeamSummaryCard(date, rows)
		return sendGChatCard(ctx, cfg, event, card)
	}

	return sendDiscordMessage(ctx, cfg, event, buildDiscordTeamSummaryMessage(team, date, summary))
}

func buildTeamSummaryRows(team *repository.Team, summary *services.TeamSummary) []gchat.TeamRow {
	rows := []gchat.TeamRow{
		{Label: "Team", Value: team.Name},
		{Label: "Members", Value: fmt.Sprintf("%d", summary.MemberCount)},
	}

	mealTypes := make([]string, 0, len(summary.MealCounts))
	for mt := range summary.MealCounts {
		mealTypes = append(mealTypes, mt)
	}
	sort.Strings(mealTypes)

	for _, mt := range mealTypes {
		rows = append(rows, gchat.TeamRow{
			Label: cmdutil.DisplayMealName(mt),
			Value: fmt.Sprintf("%d / %d", summary.MealCounts[mt], summary.MemberCount),
		})
	}

	officeCount := summary.MemberCount - summary.WFHCount
	rows = append(rows, gchat.TeamRow{
		Label: "Location",
		Value: fmt.Sprintf("Office %d / WFH %d", officeCount, summary.WFHCount),
	})

	return rows
}

func formatTeamSummary(team *repository.Team, date string, summary *services.TeamSummary) string {
	if summary.MemberCount == 0 {
		return fmt.Sprintf("Team %s — %s has no members.", team.Name, date)
	}

	mealTypes := make([]string, 0, len(summary.MealCounts))
	for mt := range summary.MealCounts {
		mealTypes = append(mealTypes, mt)
	}
	sort.Strings(mealTypes)

	var sb strings.Builder
	fmt.Fprintf(&sb, "Team %s — %s (%d members)\n", team.Name, date, summary.MemberCount)

	if len(mealTypes) > 0 {
		parts := make([]string, 0, len(mealTypes))
		for _, mt := range mealTypes {
			parts = append(parts, fmt.Sprintf("%s %d/%d", cmdutil.DisplayMealName(mt), summary.MealCounts[mt], summary.MemberCount))
		}
		fmt.Fprintf(&sb, "Meals:    %s\n", strings.Join(parts, "  │  "))
	}

	officeCount := summary.MemberCount - summary.WFHCount
	fmt.Fprintf(&sb, "Location: Office %d/%d  │  WFH %d/%d", officeCount, summary.MemberCount, summary.WFHCount, summary.MemberCount)

	return sb.String()
}

func buildDiscordTeamSummaryMessage(team *repository.Team, date string, summary *services.TeamSummary) discord.Message {
	fields := []discord.EmbedField{
		{Name: "Team", Value: team.Name, Inline: true},
		{Name: "Members", Value: fmt.Sprintf("%d", summary.MemberCount), Inline: true},
		{Name: "Office / WFH", Value: fmt.Sprintf("%d / %d", summary.MemberCount-summary.WFHCount, summary.WFHCount), Inline: true},
	}

	mealTypes := make([]string, 0, len(summary.MealCounts))
	for mealType := range summary.MealCounts {
		mealTypes = append(mealTypes, mealType)
	}
	sort.Strings(mealTypes)

	for _, mealType := range mealTypes {
		fields = append(fields, discord.EmbedField{
			Name:   cmdutil.DisplayMealName(mealType),
			Value:  fmt.Sprintf("%d / %d included", summary.MealCounts[mealType], summary.MemberCount),
			Inline: true,
		})
	}

	return discord.EmbedMessage(discord.BrandEmbed("Team Summary", date, fields))
}
