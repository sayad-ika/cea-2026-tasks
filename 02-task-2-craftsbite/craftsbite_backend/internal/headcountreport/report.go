package headcountreport

import (
	"fmt"
	"sort"
	"strings"

	"github.com/sayad-ika/craftsbite/internal/gchat"
	"github.com/sayad-ika/craftsbite/internal/services"
)

func BuildDiscordMessage(r *services.HeadcountResult) string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "**📊 Headcount — %s**\n", r.Date)

	statusLine := DayStatusLabel(r.DayStatus, r.DayReason)
	fmt.Fprintf(&sb, "> %s  ·  **%d** employees  ·  🏢 **%d** office  |  🏠 **%d** WFH\n",
		statusLine, r.TotalUsers, r.LocationCounts.Office, r.LocationCounts.WFH)

	if len(r.MealCounts) > 0 {
		sb.WriteString("\n**🍽 Overall Meals**\n")
		for _, mt := range SortedMealCountKeys(r.MealCounts) {
			mc := r.MealCounts[mt]
			fmt.Fprintf(&sb, "> %-14s %d opted in  /  %d opted out\n",
				DisplayMealName(mt)+":", mc.OptedIn, mc.OptedOut)
		}
	}

	if len(r.Teams) > 0 {
		sb.WriteString("\n**🏢 By Team**\n")
		for _, t := range r.Teams {
			fmt.Fprintf(&sb, "\n> **%s** · %d members · 🏢 %d office  |  🏠 %d WFH\n",
				t.TeamName, t.MemberCount, t.LocationCounts.Office, t.LocationCounts.WFH)
			if len(t.MealCounts) > 0 {
				for _, mt := range SortedMealCountKeys(t.MealCounts) {
					mc := t.MealCounts[mt]
					fmt.Fprintf(&sb, "> 🍽 %-10s %d in / %d out\n",
						DisplayMealName(mt)+":", mc.OptedIn, mc.OptedOut)
				}
			} else {
				sb.WriteString("> _(no meal records)_\n")
			}
		}
	}

	return strings.TrimRight(sb.String(), "\n")
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

func DayStatusLabel(status, reason string) string {
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
