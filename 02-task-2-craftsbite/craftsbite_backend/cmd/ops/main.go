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
	case "headcount":
		return handleHeadcountCommand(ctx, client, c, event)
	case "schedule-day":
		return handleScheduleDayCommand(ctx, client, c, event)
	case "admin":
		return sendOpsReply(ctx, c, event, "This feature is coming soon.")
	default:
		return sendOpsReply(ctx, c, event, fmt.Sprintf("Unknown command: /%s", event.CommandName))
	}
}

func sendOpsReply(ctx context.Context, c *appconfig.Config, event payload.CommandEvent, text string) error {
	if event.Source == "gchat" {
		card, _ := gchat.SimpleTextCard(text)
		return gchat.CreatePrivateMessage(ctx, c.GChatServiceAccountJSON, event.GChatSpaceName, event.GChatViewerName, card)
	}
	return discord.SendFollowup(event.ApplicationID, event.InteractionToken, text)
}

func handleHeadcountCommand(ctx context.Context, client *dynamodb.Client, c *appconfig.Config, event payload.CommandEvent) error {
	if event.Role != "admin" && event.Role != "logistics" {
		return sendOpsReply(ctx, c, event, "You do not have permission to use `/headcount`.")
	}

	dateStr, _ := optString(event.Options, "date")
	date, err := dateutil.ParseDateWithDefaults(dateStr)
	if err != nil {
		return sendOpsReply(ctx, c, event, fmt.Sprintf("Invalid date: %v\nUse: tomorrow (default), +N, or YYYY-MM-DD", err))
	}

	result, err := services.GetHeadcount(ctx, client, c.DynamoDBTable, date)
	if err != nil {
		return sendOpsReply(ctx, c, event, "Something went wrong fetching headcount. Please try again later.")
	}

	if event.Source == "gchat" {
		card, _ := buildHeadcountCard(result)
		return gchat.CreatePrivateMessage(ctx, c.GChatServiceAccountJSON, event.GChatSpaceName, event.GChatViewerName, card)
	}

	return discord.SendFollowup(event.ApplicationID, event.InteractionToken, formatHeadcount(result))
}

func buildHeadcountCard(r *services.HeadcountResult) ([]byte, error) {
	statusLabel := dayStatusLabel(r.DayStatus, r.DayReason)

	var overallMeals []gchat.TeamRow
	for _, mt := range sortedKeys(r.MealCounts) {
		mc := r.MealCounts[mt]
		overallMeals = append(overallMeals, gchat.TeamRow{
			Label: displayMealName(mt),
			Value: fmt.Sprintf("%d in / %d out", mc.OptedIn, mc.OptedOut),
		})
	}

	var teams []gchat.HeadcountTeamSection
	for _, t := range r.Teams {
		var meals []gchat.TeamRow
		for _, mt := range sortedKeys(t.MealCounts) {
			mc := t.MealCounts[mt]
			meals = append(meals, gchat.TeamRow{
				Label: displayMealName(mt),
				Value: fmt.Sprintf("%d in / %d out", mc.OptedIn, mc.OptedOut),
			})
		}
		teams = append(teams, gchat.HeadcountTeamSection{
			TeamName:    t.TeamName,
			MemberCount: t.MemberCount,
			Office:      t.LocationCounts.Office,
			WFH:         t.LocationCounts.WFH,
			Meals:       meals,
		})
	}

	return gchat.HeadcountCard(r.Date, statusLabel, r.TotalUsers, r.LocationCounts.Office, r.LocationCounts.WFH, overallMeals, teams)
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

func handleScheduleDayCommand(ctx context.Context, client *dynamodb.Client, c *appconfig.Config, event payload.CommandEvent) error {
	if event.Role != "admin" {
		return sendOpsReply(ctx, c, event, "You do not have permission to use `/schedule-day`.")
	}

	dateStr, ok := optString(event.Options, "date")
	if !ok || dateStr == "" {
		return sendOpsReply(ctx, c, event, "Usage: /schedule-day <date> <status> [meals] [reason]\nExample: /schedule-day 2026-03-25 normal lunch,snacks")
	}

	statusStr, ok := optString(event.Options, "status")
	if !ok || statusStr == "" {
		return sendOpsReply(ctx, c, event, fmt.Sprintf("Day status is required. Valid values: %v", repository.ValidDayStatuses()))
	}

	mealsStr, _ := optString(event.Options, "meals")
	var meals []string
	if mealsStr != "" {
		meals = strings.Split(strings.ReplaceAll(mealsStr, " ", ""), ",")
	}

	reason, _ := optString(event.Options, "reason")

	input := services.SetDayScheduleInput{
		Date:           dateStr,
		DayStatus:      statusStr,
		AvailableMeals: meals,
		Reason:         reason,
		SetBy:          event.UserID,
	}

	schedule, err := services.SetDaySchedule(ctx, client, c.DynamoDBTable, input)
	if err != nil {
		return sendOpsReply(ctx, c, event, formatScheduleDayError(err))
	}

	return sendOpsReply(ctx, c, event, formatScheduleDaySuccess(schedule))
}

func formatScheduleDayError(err error) string {
	switch {
	case strings.Contains(err.Error(), "invalid date format"):
		return "Invalid date format. Use YYYY-MM-DD (e.g., 2026-03-25)"
	case strings.Contains(err.Error(), "invalid day status"):
		return err.Error()
	case strings.Contains(err.Error(), "invalid meal type"):
		return err.Error()
	case strings.Contains(err.Error(), "cannot have meals"):
		return "Office closed and government holiday days cannot have meals."
	default:
		return "Failed to set day schedule. Please try again."
	}
}

func formatScheduleDaySuccess(schedule *repository.DaySchedule) string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "✓ Day schedule set for %s\n\n", schedule.Date)
	fmt.Fprintf(&sb, "**Status**: %s\n", displayDayStatus(schedule.DayStatus))

	if len(schedule.AvailableMeals) > 0 {
		fmt.Fprintf(&sb, "**Meals**: %s\n", formatMealList(schedule.AvailableMeals))
	} else {
		fmt.Fprintf(&sb, "**Meals**: None\n")
	}

	if schedule.Reason != "" {
		fmt.Fprintf(&sb, "**Reason**: %s\n", schedule.Reason)
	}

	return sb.String()
}

func displayDayStatus(status string) string {
	switch status {
	case "normal":
		return "📅 Normal Day"
	case "office_closed":
		return "🔒 Office Closed"
	case "govt_holiday":
		return "🎉 Government Holiday"
	case "celebration":
		return "🎊 Celebration"
	case "weekend":
		return "🏖 Weekend"
	case "event_day":
		return "🎪 Event Day"
	default:
		return status
	}
}

func formatMealList(meals []string) string {
	var formatted []string
	for _, m := range meals {
		formatted = append(formatted, displayMealName(m))
	}
	return strings.Join(formatted, ", ")
}

func main() {
	lambda.Start(handler)
}
