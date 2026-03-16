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
	"github.com/sayad-ika/craftsbite/internal/discord"
	"github.com/sayad-ika/craftsbite/internal/dynamo"
	"github.com/sayad-ika/craftsbite/internal/repository"
	"github.com/sayad-ika/craftsbite/internal/services"
)

type CommandEvent struct {
	UserID           string                 `json:"userID"`
	Role             string                 `json:"role"`
	DiscordID        string                 `json:"discordId"`
	CommandName      string                 `json:"commandName"`
	Options          map[string]interface{} `json:"options"`
	InteractionToken string                 `json:"interactionToken"`
	ApplicationID    string                 `json:"applicationId"`
}

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

func handler(ctx context.Context, event CommandEvent) error {
	c := getConfig()
	client := dynamo.GetClient(c)

	var replyContent string

	switch event.CommandName {
	case "team-summary":
		replyContent = handleTeamSummary(ctx, client, c.DynamoDBTable, event)
	case "override":
		replyContent = "This feature is coming soon."
	default:
		replyContent = fmt.Sprintf("Unknown command: /%s", event.CommandName)
	}

	return discord.SendFollowup(event.ApplicationID, event.InteractionToken, replyContent)
}

func handleTeamSummary(ctx context.Context, client *dynamodb.Client, table string, event CommandEvent) string {
	if event.Role != "team_lead" && event.Role != "admin" {
		return "You do not have permission to use `/team-summary`."
	}

	date, ok := optString(event.Options, "date")
	if !ok || date == "" {
		return "Please provide a date. Example: `/team-summary date:2026-03-10`"
	}

	teams, err := repository.FindTeamsByLeadID(ctx, client, table, event.UserID)
	if err != nil {
		return "Something went wrong fetching your team. Please try again later."
	}
	if len(teams) == 0 {
		return "You are not assigned as a team lead to any team."
	}

	team, err := repository.GetTeamByID(ctx, client, table, teams[0].ID)
	if err != nil || team == nil {
		return "Team not found. Please try again later."
	}

	summary, err := services.GetTeamSummary(ctx, client, table, teams[0].ID, date)
	if err != nil {
		return "Something went wrong fetching the team summary. Please try again later."
	}

	return formatTeamSummary(team, date, summary)
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
