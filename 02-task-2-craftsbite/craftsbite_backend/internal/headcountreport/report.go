package headcountreport

import (
	"fmt"
	"sort"
	"strings"

	"github.com/sayad-ika/craftsbite/internal/discord"
	"github.com/sayad-ika/craftsbite/internal/gchat"
	"github.com/sayad-ika/craftsbite/internal/services"
)

func BuildDiscordMessage(r *services.HeadcountResult) discord.Message {
	fields := []discord.EmbedField{
		{Name: "Total headcount", Value: fmt.Sprintf("%d", r.TotalUsers), Inline: true},
		{Name: "Office / WFH", Value: fmt.Sprintf("%d / %d", r.LocationCounts.Office, r.LocationCounts.WFH), Inline: true},
	}

	for _, mt := range SortedMealCountKeys(r.MealCounts) {
		mc := r.MealCounts[mt]
		fields = append(fields, discord.EmbedField{
			Name:   DisplayMealName(mt),
			Value:  fmt.Sprintf("%d confirmed", mc.OptedIn),
			Inline: true,
		})
	}
	if r.DayReason != "" {
		fields = append(fields, discord.EmbedField{Name: "Note", Value: r.DayReason})
	}

	return discord.EmbedMessage(discord.BrandEmbed("Daily Headcount Summary", r.Date, fields))
}

func BuildScheduledDiscordMessage(r *services.HeadcountResult) discord.Message {
	fields := []discord.EmbedField{
		{Name: "Total headcount", Value: fmt.Sprintf("%d", r.TotalUsers), Inline: true},
		{Name: "Office / WFH", Value: fmt.Sprintf("%d / %d", r.LocationCounts.Office, r.LocationCounts.WFH), Inline: true},
	}

	for _, mt := range SortedMealCountKeys(r.MealCounts) {
		mc := r.MealCounts[mt]
		fields = append(fields, discord.EmbedField{
			Name:   DisplayMealName(mt),
			Value:  fmt.Sprintf("%d confirmed", mc.OptedIn),
			Inline: true,
		})
	}
	if r.DayReason != "" {
		fields = append(fields, discord.EmbedField{Name: "Note", Value: r.DayReason})
	}

	return discord.EmbedMessage(discord.BrandEmbed("Daily Headcount Summary", r.Date, fields))
}

func BuildGChatCard(r *services.HeadcountResult) ([]byte, error) {
	statusLabel := DayStatusLabel(r.DayStatus, "")

	var overallMeals []gchat.TeamRow
	for _, mt := range SortedMealCountKeys(r.MealCounts) {
		mc := r.MealCounts[mt]
		overallMeals = append(overallMeals, gchat.TeamRow{
			Label: DisplayMealName(mt),
			Value: fmt.Sprintf("%d confirmed", mc.OptedIn),
		})
	}

	return gchat.CompactHeadcountCard(r.Date, statusLabel, r.TotalUsers, r.LocationCounts.Office, r.LocationCounts.WFH, overallMeals, r.DayReason)
}

func BuildScheduledGChatCard(r *services.HeadcountResult) ([]byte, error) {
	statusLabel := DayStatusLabel(r.DayStatus, "")

	var overallMeals []gchat.TeamRow
	for _, mt := range SortedMealCountKeys(r.MealCounts) {
		mc := r.MealCounts[mt]
		overallMeals = append(overallMeals, gchat.TeamRow{
			Label: DisplayMealName(mt),
			Value: fmt.Sprintf("%d confirmed", mc.OptedIn),
		})
	}

	return gchat.CompactHeadcountCard(r.Date, statusLabel, r.TotalUsers, r.LocationCounts.Office, r.LocationCounts.WFH, overallMeals, r.DayReason)
}

func DayStatusLabel(status, reason string) string {
	var label string
	switch status {
	case "", "normal":
		label = "📅 Normal Day"
	case "govt_holiday":
		label = "🎉 Government Holiday"
	case "holiday":
		label = "🎉 Holiday"
	case "office_closed":
		label = "🔒 Office Closed"
	case "celebration":
		label = "🎊 Celebration"
	case "event_day":
		label = "🎪 Event Day"
	case "wfh_day":
		label = "🏠 WFH Day"
	case "weekend":
		label = "🏖 Weekend"
	default:
		label = "📅 " + DisplayMealName(status)
	}
	if reason != "" {
		label += " — " + reason
	}
	return label
}

func DisplayDayStatus(status string) string {
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

func DisplayMealName(s string) string {
	words := strings.Split(strings.ReplaceAll(s, "_", " "), " ")
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, " ")
}

func FormatMealList(meals []string) string {
	var formatted []string
	for _, m := range meals {
		formatted = append(formatted, DisplayMealName(m))
	}
	return strings.Join(formatted, ", ")
}

func SortedMealCountKeys(m map[string]services.MealCount) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

