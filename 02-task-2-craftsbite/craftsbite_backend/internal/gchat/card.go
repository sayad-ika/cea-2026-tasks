package gchat

import (
	"encoding/json"
	"fmt"
	"html"
	"strings"

	"github.com/sayad-ika/craftsbite/internal/discord"
)

type CardResponse struct {
	CardsV2 []CardV2Wrapper `json:"cardsV2"`
}

type CardV2Wrapper struct {
	CardID string `json:"cardId"`
	Card   CardV2 `json:"card"`
}

type CardV2 struct {
	Header   *CardHeader   `json:"header,omitempty"`
	Sections []CardSection `json:"sections"`
}

type CardHeader struct {
	Title        string `json:"title"`
	Subtitle     string `json:"subtitle,omitempty"`
	ImageURL     string `json:"imageUrl,omitempty"`
	ImageType    string `json:"imageType,omitempty"`
	ImageAltText string `json:"imageAltText,omitempty"`
}

type CardSection struct {
	Widgets []CardWidget `json:"widgets"`
}

type CardWidget struct {
	TextParagraph *TextParagraph `json:"textParagraph,omitempty"`
	DecoratedText *DecoratedText `json:"decoratedText,omitempty"`
}

type TextParagraph struct {
	Text string `json:"text"`
}

type DecoratedText struct {
	TopLabel string `json:"topLabel"`
	Text     string `json:"text"`
}

type TeamRow struct {
	Label string
	Value string
}

type HeadcountTeamSection struct {
	TeamName    string
	MemberCount int
	Office      int
	WFH         int
	Meals       []TeamRow
}

type HelpCommand struct {
	Usage       string
	Description string
}

type HelpSection struct {
	Title    string
	Commands []HelpCommand
}

func NoticeCard(title, subtitle, text string, tone discord.NoticeTone) ([]byte, error) {
	return buildCardResponseWithTone(
		"response",
		title,
		subtitle,
		[]CardSection{textSection(text)},
		tone,
	)
}

func LocationCard(location, date, mealStatus string) ([]byte, error) {
	return buildCardResponse(
		"location-response",
		"Location Updated",
		date,
		[]CardSection{
			rowsSection([]TeamRow{{Label: "Work location", Value: location}}),
			textSection("<b>Meal status</b><br>" + paragraphText(mealStatus)),
		},
	)
}

func TeamSummaryCard(date string, rows []TeamRow) ([]byte, error) {
	sections := []CardSection{
		textSection("<b>Team overview</b>"),
		rowsSection(rows),
	}
	return buildCardResponse("team-summary-response", "Team Summary", date, sections)
}

func StatusCard(date, location string, meals []TeamRow) ([]byte, error) {
	sections := []CardSection{
		rowsSection([]TeamRow{{Label: "Location", Value: location}}),
	}
	if len(meals) == 0 {
		sections = append(sections, textSection("<b>Meals</b><br>No meals configured"))
	} else {
		sections = append(sections, textSection("<b>Meals</b>"), rowsSection(meals))
	}
	return buildCardResponse("status-response", "Status Snapshot", date, sections)
}

func MealStatusCard(date string, changedMeals []string, meals []TeamRow) ([]byte, error) {
	sections := []CardSection{}
	if len(changedMeals) > 0 {
		sections = append(sections, textSection("<b>Updated meals</b><br>"+paragraphText(strings.Join(changedMeals, ", "))))
	}
	if len(meals) == 0 {
		sections = append(sections, textSection("<b>Meals</b><br>No meals configured"))
	} else {
		sections = append(sections, textSection("<b>Current status</b>"), rowsSection(meals))
	}
	return buildCardResponse("meal-response", "Meal Status Updated", date, sections)
}

func HelpCard(title, subtitle string, sections []HelpSection) ([]byte, error) {
	cardSections := make([]CardSection, 0, len(sections))
	for _, section := range sections {
		cardSections = append(cardSections, textSection(helpSectionText(section)))
	}
	return buildCardResponse("help-response", title, subtitle, cardSections)
}

func ScheduleDayCard(date, dayStatusLabel, meals, reason string) ([]byte, error) {
	rows := []TeamRow{{Label: "Status", Value: dayStatusLabel}}
	if meals == "" {
		meals = "None"
	}
	rows = append(rows, TeamRow{Label: "Meals", Value: meals})
	if reason != "" {
		rows = append(rows, TeamRow{Label: "Reason", Value: reason})
	}
	return buildCardResponse(
		"schedule-day-response",
		"Schedule Updated",
		date,
		[]CardSection{rowsSection(rows)},
	)
}

func BulkScheduleDayCard(statusLabel string, dates []string) ([]byte, error) {
	sections := []CardSection{
		rowsSection([]TeamRow{
			{Label: "Day status", Value: statusLabel},
			{Label: "Weekdays updated", Value: fmt.Sprintf("%d", len(dates))},
		}),
		textSection("<b>Affected dates</b><br>" + paragraphText(strings.Join(dates, "\n"))),
	}
	return buildCardResponse("bulk-schedule-day-response", "Schedule Updated", "Bulk weekday update", sections)
}

func HeadcountCard(date, dayStatusLabel string, totalUsers, office, wfh int, overallMeals []TeamRow, teams []HeadcountTeamSection) ([]byte, error) {
	sections := []CardSection{
		rowsSection([]TeamRow{
			{Label: "Day status", Value: dayStatusLabel},
			{Label: "Total employees", Value: fmt.Sprintf("%d", totalUsers)},
			{Label: "Office / WFH", Value: fmt.Sprintf("%d / %d", office, wfh)},
		}),
	}

	if len(overallMeals) > 0 {
		sections = append(sections, textSection("<b>Overall meals</b>"), rowsSection(overallMeals))
	}

	for _, t := range teams {
		teamRows := []TeamRow{
			{Label: "Members", Value: fmt.Sprintf("%d", t.MemberCount)},
			{Label: "Office / WFH", Value: fmt.Sprintf("%d / %d", t.Office, t.WFH)},
		}
		teamRows = append(teamRows, t.Meals...)
		sections = append(
			sections,
			textSection(fmt.Sprintf("<b>%s</b>", t.TeamName)),
			rowsSection(teamRows),
		)
	}

	return buildCardResponse("headcount-response", "Headcount Snapshot", date, sections)
}

func CompactHeadcountCard(date, dayStatusLabel string, totalUsers, office, wfh int, overallMeals []TeamRow, note string) ([]byte, error) {
	sections := []CardSection{
		rowsSection([]TeamRow{
			{Label: "Day status", Value: dayStatusLabel},
			{Label: "Total headcount", Value: fmt.Sprintf("%d", totalUsers)},
			{Label: "Office / WFH", Value: fmt.Sprintf("%d / %d", office, wfh)},
		}),
	}

	if len(overallMeals) > 0 {
		sections = append(sections, textSection("<b>Meal summary</b>"), rowsSection(overallMeals))
	}
	if note != "" {
		sections = append(sections, textSection("<b>Note</b><br>"+paragraphText(note)))
	}

	return buildCardResponse("headcount-summary-response", "Daily Headcount Summary", date, sections)
}

func buildCardResponse(cardID, title, subtitle string, sections []CardSection) ([]byte, error) {
	return buildCardResponseWithTone(cardID, title, subtitle, sections, discord.NoticeToneInfo)
}

func buildCardResponseWithTone(cardID, title, subtitle string, sections []CardSection, tone discord.NoticeTone) ([]byte, error) {
	resp := CardResponse{
		CardsV2: []CardV2Wrapper{
			{
				CardID: cardID,
				Card: CardV2{
					Header:   brandedHeader(title, subtitle, tone),
					Sections: sections,
				},
			},
		},
	}
	return json.Marshal(resp)
}

func brandedHeader(title, subtitle string, tone discord.NoticeTone) *CardHeader {
	return &CardHeader{
		Title:        title,
		Subtitle:     subtitle,
		ImageURL:     placeholderBrandImageURL(tone),
		ImageType:    "SQUARE",
		ImageAltText: "CraftsBite",
	}
}

func placeholderBrandImageURL(tone discord.NoticeTone) string {
	color := "F47621"
	switch tone {
	case discord.NoticeToneSuccess:
		color = "2F9E44"
	case discord.NoticeToneWarning:
		color = "F08C00"
	case discord.NoticeToneError:
		color = "E03131"
	}
	return "https://placehold.co/96x96/" + color + "/FFFFFF.png?text=CB"
}

func textSection(text string) CardSection {
	return CardSection{
		Widgets: []CardWidget{{TextParagraph: &TextParagraph{Text: paragraphText(text)}}},
	}
}

func rowsSection(rows []TeamRow) CardSection {
	widgets := make([]CardWidget, 0, len(rows))
	for _, row := range rows {
		widgets = append(widgets, CardWidget{
			DecoratedText: &DecoratedText{
				TopLabel: row.Label,
				Text:     row.Value,
			},
		})
	}
	return CardSection{Widgets: widgets}
}

func paragraphText(text string) string {
	return strings.ReplaceAll(text, "\n", "<br>")
}

func helpSectionText(section HelpSection) string {
	parts := []string{fmt.Sprintf("<b>%s</b>", html.EscapeString(section.Title))}
	for _, command := range section.Commands {
		parts = append(parts,
			fmt.Sprintf("<b>%s</b><br>%s",
				html.EscapeString(command.Usage),
				html.EscapeString(command.Description),
			),
		)
	}
	return strings.Join(parts, "<br><br>")
}
