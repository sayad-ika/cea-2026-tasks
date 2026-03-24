package main

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	appconfig "github.com/sayad-ika/craftsbite/internal/config"
	"github.com/sayad-ika/craftsbite/internal/dateutil"
	"github.com/sayad-ika/craftsbite/internal/discord"
	"github.com/sayad-ika/craftsbite/internal/dynamo"
	"github.com/sayad-ika/craftsbite/internal/gchat"
	"github.com/sayad-ika/craftsbite/internal/payload"
	"github.com/sayad-ika/craftsbite/internal/repository"
	"github.com/sayad-ika/craftsbite/internal/services"
)

var (
	cfgOnce sync.Once
	cfg     *appconfig.Config
)

func getConfig() *appconfig.Config {
	cfgOnce.Do(func() {
		cfg = appconfig.MustLoad()
	})
	return cfg
}

func handler(ctx context.Context, event payload.CommandEvent) error {
	c := getConfig()
	client := dynamo.GetClient(c)

	switch event.CommandName {
	case "team-summary":
		return handleTeamSummaryCommand(ctx, client, c, event)
	case "override":
		return sendMgmtReply(ctx, c, event, "This feature is coming soon.")
	default:
		return sendMgmtReply(ctx, c, event, fmt.Sprintf("Unknown command: /%s", event.CommandName))
	}
}

func sendMgmtReply(ctx context.Context, c *appconfig.Config, event payload.CommandEvent, text string) error {
	if event.Source == "gchat" {
		card, _ := gchat.SimpleTextCard(text)
		return gchat.CreatePrivateMessage(ctx, c.GChatServiceAccountJSON, event.GChatSpaceName, event.GChatViewerName, card)
	}
	return discord.SendFollowup(event.ApplicationID, event.InteractionToken, text)
}

func handleTeamSummaryCommand(ctx context.Context, client *dynamodb.Client, c *appconfig.Config, event payload.CommandEvent) error {
	if event.Role != "team_lead" && event.Role != "admin" {
		return sendMgmtReply(ctx, c, event, "You do not have permission to use `/team-summary`.")
	}

	dateStr, _ := optString(event.Options, "date")
	date, err := dateutil.ParseDateWithDefaults(dateStr)
	if err != nil {
		return sendMgmtReply(ctx, c, event, fmt.Sprintf("Invalid date: %v\nUse: tomorrow (default), +N, or YYYY-MM-DD", err))
	}

	teams, err := repository.FindTeamsByLeadID(ctx, client, c.DynamoDBTable, event.UserID)
	if err != nil {
		return sendMgmtReply(ctx, c, event, "Something went wrong fetching your team. Please try again later.")
	}
	if len(teams) == 0 {
		return sendMgmtReply(ctx, c, event, "You are not assigned as a team lead to any team.")
	}

	team, err := repository.GetTeamByID(ctx, client, c.DynamoDBTable, teams[0].ID)
	if err != nil || team == nil {
		return sendMgmtReply(ctx, c, event, "Team not found. Please try again later.")
	}

	summary, err := services.GetTeamSummary(ctx, client, c.DynamoDBTable, teams[0].ID, date)
	if err != nil {
		return sendMgmtReply(ctx, c, event, "Something went wrong fetching the team summary. Please try again later.")
	}

	if event.Source == "gchat" {
		rows := buildTeamSummaryRows(team, summary)
		card, _ := gchat.TeamSummaryCard(date, rows)
		return gchat.CreatePrivateMessage(ctx, c.GChatServiceAccountJSON, event.GChatSpaceName, event.GChatViewerName, card)
	}

	return discord.SendFollowup(event.ApplicationID, event.InteractionToken, formatTeamSummary(team, date, summary))
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
			Label: displayMealName(mt),
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

func optString(opts map[string]interface{}, key string) (string, bool) {
	v, ok := opts[key]
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
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
			parts = append(parts, fmt.Sprintf("%s %d/%d", displayMealName(mt), summary.MealCounts[mt], summary.MemberCount))
		}
		fmt.Fprintf(&sb, "Meals:    %s\n", strings.Join(parts, "  │  "))
	}

	officeCount := summary.MemberCount - summary.WFHCount
	fmt.Fprintf(&sb, "Location: Office %d/%d  │  WFH %d/%d", officeCount, summary.MemberCount, summary.WFHCount, summary.MemberCount)

	return sb.String()
}

func displayMealName(s string) string {
	words := strings.Split(strings.ReplaceAll(s, "_", " "), " ")
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, " ")
}

func main() {
	lambda.Start(handler)
}
