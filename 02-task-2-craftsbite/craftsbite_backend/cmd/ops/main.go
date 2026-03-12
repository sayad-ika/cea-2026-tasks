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
	case "headcount":
		replyContent = handleHeadcount(ctx, client, c.DynamoDBTable, event)
	case "set-day":
		replyContent = "This feature is coming soon."
	case "admin":
		replyContent = "This feature is coming soon."
	default:
		replyContent = fmt.Sprintf("Unknown command: /%s", event.CommandName)
	}

	return discord.SendFollowup(event.ApplicationID, event.InteractionToken, replyContent)
}

func handleHeadcount(ctx context.Context, client *dynamodb.Client, table string, event CommandEvent) string {
	if event.Role != "admin" && event.Role != "logistics" {
		return "You do not have permission to use `/headcount`."
	}

	date, ok := optString(event.Options, "date")
	if !ok || date == "" {
		return "Please provide a date. Example: `/headcount date:2026-03-10`"
	}

	result, err := services.GetHeadcount(ctx, client, table, date)
	if err != nil {
		return "Something went wrong fetching headcount. Please try again later."
	}

	return formatHeadcount(result)
}

func optString(opts map[string]interface{}, key string) (string, bool) {
	v, ok := opts[key]
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

func formatHeadcount(r *services.HeadcountResult) string {
	var sb strings.Builder

	// Header
	fmt.Fprintf(&sb, "**📊 Headcount — %s**\n", r.Date)

	// Day status + overall totals
	statusLine := dayStatusLabel(r.DayStatus, r.DayReason)
	fmt.Fprintf(&sb, "> %s  ·  **%d** employees  ·  🏢 **%d** office  |  🏠 **%d** WFH\n",
		statusLine, r.TotalUsers, r.LocationCounts.Office, r.LocationCounts.WFH)

	// Overall meals
	if len(r.MealCounts) > 0 {
		sb.WriteString("\n**🍽 Overall Meals**\n")
		for _, mt := range sortedKeys(r.MealCounts) {
			mc := r.MealCounts[mt]
			fmt.Fprintf(&sb, "> %-14s %d opted in  /  %d opted out\n",
				displayMealName(mt)+":", mc.OptedIn, mc.OptedOut)
		}
	}

	// Per-team breakdown
	if len(r.Teams) > 0 {
		sb.WriteString("\n**🏢 By Team**\n")
		for _, t := range r.Teams {
			fmt.Fprintf(&sb, "\n> **%s** · %d members · 🏢 %d office  |  🏠 %d WFH\n",
				t.TeamName, t.MemberCount, t.LocationCounts.Office, t.LocationCounts.WFH)
			if len(t.MealCounts) > 0 {
				for _, mt := range sortedKeys(t.MealCounts) {
					mc := t.MealCounts[mt]
					fmt.Fprintf(&sb, "> 🍽 %-10s %d in / %d out\n",
						displayMealName(mt)+":", mc.OptedIn, mc.OptedOut)
				}
			} else {
				sb.WriteString("> _(no meal records)_\n")
			}
		}
	}

	return strings.TrimRight(sb.String(), "\n")
}

func dayStatusLabel(status, reason string) string {
	var label string
	switch status {
	case "", "normal":
		label = "📅 Normal Day"
	case "holiday":
		label = "🎉 Holiday"
	case "office_closed":
		label = "🔒 Office Closed"
	case "event_day":
		label = "🎪 Event Day"
	case "wfh_day":
		label = "🏠 WFH Day"
	default:
		label = "📅 " + displayMealName(status)
	}
	if reason != "" {
		label += " — " + reason
	}
	return label
}

func sortedKeys(m map[string]services.MealCount) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
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
