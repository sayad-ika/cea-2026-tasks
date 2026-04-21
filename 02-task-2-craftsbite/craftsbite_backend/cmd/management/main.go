package main

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/sayad-ika/craftsbite/internal/cmdutil"
	appconfig "github.com/sayad-ika/craftsbite/internal/config"
	"github.com/sayad-ika/craftsbite/internal/dateutil"
	"github.com/sayad-ika/craftsbite/internal/dynamo"
	"github.com/sayad-ika/craftsbite/internal/gchat"
	"github.com/sayad-ika/craftsbite/internal/payload"
	"github.com/sayad-ika/craftsbite/internal/repository"
	"github.com/sayad-ika/craftsbite/internal/services"
)

func handler(ctx context.Context, client *dynamodb.Client, cfg *appconfig.Config, dateParser *dateutil.DateParser, event payload.CommandEvent) error {
	switch event.CommandName {
	case "team-summary":
		return handleTeamSummaryCommand(ctx, client, cfg, dateParser, event)
	case "override":
		return cmdutil.SendReply(ctx, cfg, event, "This feature is coming soon.")
	default:
		return cmdutil.SendReply(ctx, cfg, event, fmt.Sprintf("Unknown command: /%s", event.CommandName))
	}
}

func handleTeamSummaryCommand(ctx context.Context, client *dynamodb.Client, cfg *appconfig.Config, dateParser *dateutil.DateParser, event payload.CommandEvent) error {
	if event.Role != "team_lead" && event.Role != "admin" {
		return cmdutil.SendReply(ctx, cfg, event, "You do not have permission to use `/team-summary`.")
	}

	dateStr, _ := cmdutil.OptString(event.Options, "date")
	date, err := dateParser.ParseDateWithDefaults(dateStr)
	if err != nil {
		return cmdutil.SendReply(ctx, cfg, event, fmt.Sprintf("Invalid date: %v\nUse: tomorrow (default), +N, or YYYY-MM-DD", err))
	}

	teams, err := repository.FindTeamsByLeadID(ctx, client, cfg.DynamoDBTable, event.UserID)
	if err != nil {
		return cmdutil.SendReply(ctx, cfg, event, "Something went wrong fetching your team. Please try again later.")
	}
	if len(teams) == 0 {
		return cmdutil.SendReply(ctx, cfg, event, "You are not assigned as a team lead to any team.")
	}

	team, err := repository.GetTeamByID(ctx, client, cfg.DynamoDBTable, teams[0].ID)
	if err != nil || team == nil {
		return cmdutil.SendReply(ctx, cfg, event, "Team not found. Please try again later.")
	}

	summary, err := services.GetTeamSummary(ctx, client, cfg.DynamoDBTable, teams[0].ID, date)
	if err != nil {
		return cmdutil.SendReply(ctx, cfg, event, "Something went wrong fetching the team summary. Please try again later.")
	}

	if event.Source == "gchat" {
		rows := buildTeamSummaryRows(team, summary)
		card, _ := gchat.TeamSummaryCard(date, rows)
		return cmdutil.SendGChatCard(ctx, cfg, event, card)
	}

	return cmdutil.SendReply(ctx, cfg, event, formatTeamSummary(team, date, summary))
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

func main() {
	cfg := appconfig.MustLoad()
	client, err := dynamo.NewClient(cfg)
	if err != nil {
		log.Fatalf("management: %v", err)
	}
	dateParser, err := dateutil.NewDateParser(cfg.Timezone)
	if err != nil {
		log.Fatalf("management: %v", err)
	}

	lambda.Start(func(ctx context.Context, event payload.CommandEvent) error {
		return handler(ctx, client, cfg, dateParser, event)
	})
}
