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
	embeds := []discord.Embed{
		discord.BrandEmbed("Headcount Snapshot", r.Date, []discord.EmbedField{
			{Name: "Day status", Value: DayStatusLabel(r.DayStatus, r.DayReason), Inline: false},
			{Name: "Total employees", Value: fmt.Sprintf("%d", r.TotalUsers), Inline: true},
			{Name: "Office", Value: fmt.Sprintf("%d", r.LocationCounts.Office), Inline: true},
			{Name: "WFH", Value: fmt.Sprintf("%d", r.LocationCounts.WFH), Inline: true},
		}),
	}

	if len(r.MealCounts) > 0 {
		mealFields := make([]discord.EmbedField, 0, len(r.MealCounts))
		for _, mt := range SortedMealCountKeys(r.MealCounts) {
			mc := r.MealCounts[mt]
			mealFields = append(mealFields, discord.EmbedField{
				Name:   DisplayMealName(mt),
				Value:  fmt.Sprintf("%d in / %d out", mc.OptedIn, mc.OptedOut),
				Inline: true,
			})
		}
		embeds = append(embeds, discord.BrandEmbed("Meal Participation", "Overall totals", mealFields))
	}

	teamFields := make([]discord.EmbedField, 0, len(r.Teams))
	for _, team := range r.Teams {
		teamFields = append(teamFields, discord.EmbedField{
			Name:  team.TeamName,
			Value: formatTeamValue(team),
		})
	}
	for i := 0; i < len(teamFields); i += 6 {
		end := i + 6
		if end > len(teamFields) {
			end = len(teamFields)
		}
		title := "Team Breakdown"
		if len(teamFields) > 6 {
			title = fmt.Sprintf("Team Breakdown (%d-%d)", i+1, end)
		}
		embeds = append(embeds, discord.BrandEmbed(title, "Operations by team", teamFields[i:end]))
	}

	return discord.EmbedMessage(embeds...)
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
	statusLabel := DayStatusLabel(r.DayStatus, r.DayReason)

	var overallMeals []gchat.TeamRow
	for _, mt := range SortedMealCountKeys(r.MealCounts) {
		mc := r.MealCounts[mt]
		overallMeals = append(overallMeals, gchat.TeamRow{
			Label: DisplayMealName(mt),
			Value: fmt.Sprintf("%d in / %d out", mc.OptedIn, mc.OptedOut),
		})
	}

	var teams []gchat.HeadcountTeamSection
	for _, t := range r.Teams {
		var meals []gchat.TeamRow
		for _, mt := range SortedMealCountKeys(t.MealCounts) {
			mc := t.MealCounts[mt]
			meals = append(meals, gchat.TeamRow{
				Label: DisplayMealName(mt),
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

func formatTeamValue(team services.TeamHeadcount) string {
	parts := []string{
		fmt.Sprintf("Members: %d", team.MemberCount),
		fmt.Sprintf("Office / WFH: %d / %d", team.LocationCounts.Office, team.LocationCounts.WFH),
	}
	if len(team.MealCounts) == 0 {
		parts = append(parts, "Meals: no meal records")
		return strings.Join(parts, "\n")
	}

	mealParts := make([]string, 0, len(team.MealCounts))
	for _, mealType := range SortedMealCountKeys(team.MealCounts) {
		meal := team.MealCounts[mealType]
		mealParts = append(mealParts, fmt.Sprintf("%s %d/%d", DisplayMealName(mealType), meal.OptedIn, meal.OptedOut))
	}
	parts = append(parts, "Meals: "+strings.Join(mealParts, " | "))
	return strings.Join(parts, "\n")
}
