package gchat

import (
	"encoding/json"
	"fmt"
	"html"
	"strings"

	"github.com/sayad-ika/craftsbite/internal/discord"
)

const (
	InitCardFunctionSave   = "init_save"
	InitCardFunctionCancel = "init_cancel"
	InitCardFunctionEdit   = "init_edit"

	selectionTypeCheckbox = "CHECK_BOX"
	selectionTypeRadio    = "RADIO_BUTTON"
)

type CardResponse struct {
	CardsV2 []CardV2Wrapper `json:"cardsV2"`
}

type CardV2Wrapper struct {
	CardID string `json:"cardId"`
	Card   CardV2 `json:"card"`
}

type CardV2 struct {
	Header              *CardHeader      `json:"header,omitempty"`
	Sections            []CardSection    `json:"sections"`
	SectionDividerStyle string           `json:"sectionDividerStyle,omitempty"`
	FixedFooter         *CardFixedFooter `json:"fixedFooter,omitempty"`
}

type CardHeader struct {
	Title        string `json:"title"`
	Subtitle     string `json:"subtitle,omitempty"`
	ImageURL     string `json:"imageUrl,omitempty"`
	ImageType    string `json:"imageType,omitempty"`
	ImageAltText string `json:"imageAltText,omitempty"`
}

type CardSection struct {
	Header  string       `json:"header,omitempty"`
	Widgets []CardWidget `json:"widgets"`
}

type CardWidget struct {
	TextParagraph  *TextParagraph  `json:"textParagraph,omitempty"`
	DecoratedText  *DecoratedText  `json:"decoratedText,omitempty"`
	SelectionInput *SelectionInput `json:"selectionInput,omitempty"`
	ButtonList     *ButtonList     `json:"buttonList,omitempty"`
	Divider        *Divider        `json:"divider,omitempty"`
}

type TextParagraph struct {
	Text string `json:"text"`
}

type DecoratedText struct {
	TopLabel    string `json:"topLabel"`
	Text        string `json:"text"`
	BottomLabel string `json:"bottomLabel,omitempty"`
	StartIcon   *Icon  `json:"startIcon,omitempty"`
}

type SelectionInput struct {
	Name  string          `json:"name"`
	Label string          `json:"label,omitempty"`
	Type  string          `json:"type"`
	Items []SelectionItem `json:"items"`
}

type SelectionItem struct {
	Text     string `json:"text"`
	Value    string `json:"value"`
	Selected bool   `json:"selected,omitempty"`
}

type ButtonList struct {
	Buttons []Button `json:"buttons"`
}

type Button struct {
	Text    string   `json:"text"`
	Color   *Color   `json:"color,omitempty"`
	OnClick *OnClick `json:"onClick,omitempty"`
	Type    string   `json:"type,omitempty"`
	AltText string   `json:"altText,omitempty"`
}

type OnClick struct {
	Action *Action `json:"action,omitempty"`
}

type Action struct {
	Function        string            `json:"function,omitempty"`
	Parameters      []ActionParameter `json:"parameters,omitempty"`
	Interaction     string            `json:"interaction,omitempty"`
	LoadIndicator   string            `json:"loadIndicator,omitempty"`
	PersistValues   bool              `json:"persistValues,omitempty"`
	RequiredWidgets []string          `json:"requiredWidgets,omitempty"`
}

type ActionParameter struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Color struct {
	Red   float64 `json:"red"`
	Green float64 `json:"green"`
	Blue  float64 `json:"blue"`
}

type Divider struct{}

type Icon struct {
	KnownIcon string `json:"knownIcon,omitempty"`
	AltText   string `json:"altText,omitempty"`
}

type CardFixedFooter struct {
	PrimaryButton   *Button `json:"primaryButton,omitempty"`
	SecondaryButton *Button `json:"secondaryButton,omitempty"`
}

type InitCardInput struct {
	Title          string
	Subtitle       string
	Intro          string
	Note           string
	AnchorDate     string
	ActionFunction string
	Dates          []SelectionItem
	Locations      []SelectionItem
	Meals          []SelectionItem
	SummaryRows    []TeamRow
}

type InitSaveConfirmationInput struct {
	Title          string
	Subtitle       string
	Summary        string
	EditButtonText string
	AnchorDate     string
	ActionFunction string
	Tone           discord.NoticeTone
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

func InitCardResponse(input InitCardInput) ([]byte, error) {
	return json.Marshal(CardResponse{
		CardsV2: []CardV2Wrapper{{CardID: "init-setup", Card: InitCard(input)}},
	})
}

func InitSaveConfirmationCardResponse(input InitSaveConfirmationInput) ([]byte, error) {
	return json.Marshal(CardResponse{
		CardsV2: []CardV2Wrapper{{CardID: "init-saved", Card: InitSaveConfirmationCard(input)}},
	})
}

func InitCard(input InitCardInput) CardV2 {
	if input.Title == "" {
		input.Title = "CraftsBite Setup"
	}
	if input.Subtitle == "" {
		input.Subtitle = "Dates, location, and meals"
	}
	if input.Intro == "" {
		input.Intro = "Pick upcoming meal days, choose your work location, and select the meals to include."
	}
	saveFunction := input.ActionFunction
	cancelFunction := input.ActionFunction
	if saveFunction == "" {
		saveFunction = InitCardFunctionSave
	}
	if cancelFunction == "" {
		cancelFunction = InitCardFunctionCancel
	}

	sections := []CardSection{}
	if input.Note != "" {
		sections = append(sections, CardSection{
			Widgets: []CardWidget{{TextParagraph: &TextParagraph{Text: paragraphText(input.Note)}}},
		})
	}
	sections = append(sections, []CardSection{
		{
			Widgets: []CardWidget{{TextParagraph: &TextParagraph{Text: paragraphText(input.Intro)}}},
		},
		{
			Header: "Dates",
			Widgets: []CardWidget{{SelectionInput: &SelectionInput{
				Name:  "dates",
				Label: "Meal days",
				Type:  selectionTypeCheckbox,
				Items: input.Dates,
			}}},
		},
		{
			Header: "Work Location",
			Widgets: []CardWidget{{SelectionInput: &SelectionInput{
				Name:  "location",
				Label: "Where will you work?",
				Type:  selectionTypeRadio,
				Items: input.Locations,
			}}},
		},
		{
			Header: "Meals",
			Widgets: []CardWidget{{SelectionInput: &SelectionInput{
				Name:  "meals",
				Label: "Include meals",
				Type:  selectionTypeCheckbox,
				Items: input.Meals,
			}}},
		},
	}...)
	if len(input.SummaryRows) > 0 {
		sections = append(sections, CardSection{Header: "Preview", Widgets: rowWidgets(input.SummaryRows)})
	}
	sections = append(sections, CardSection{Widgets: []CardWidget{{ButtonList: &ButtonList{Buttons: []Button{
		{
			Text:  "Save",
			Color: &Color{Red: 0.18, Green: 0.62, Blue: 0.27},
			OnClick: &OnClick{Action: &Action{
				Function:      saveFunction,
				Parameters:    initCardActionParameters(InitCardFunctionSave, input.AnchorDate),
				LoadIndicator: "SPINNER",
			}},
		},
		{
			Text: "Cancel",
			OnClick: &OnClick{Action: &Action{
				Function:      cancelFunction,
				Parameters:    initCardActionParameters(InitCardFunctionCancel, input.AnchorDate),
				LoadIndicator: "NONE",
			}},
		},
	}}}}})

	return CardV2{
		Header:              brandedHeader(input.Title, input.Subtitle, discord.NoticeToneInfo),
		Sections:            sections,
		SectionDividerStyle: "SOLID_DIVIDER",
	}
}

func InitSaveConfirmationCard(input InitSaveConfirmationInput) CardV2 {
	if input.Title == "" {
		input.Title = "Setup saved"
	}
	if input.Subtitle == "" {
		input.Subtitle = "Your meal setup was updated"
	}
	if input.Summary == "" {
		input.Summary = "Your setup was saved."
	}
	if input.EditButtonText == "" {
		input.EditButtonText = "Edit setup"
	}
	if input.Tone == "" {
		input.Tone = discord.NoticeToneSuccess
	}
	actionFunction := input.ActionFunction
	if actionFunction == "" {
		actionFunction = InitCardFunctionEdit
	}

	return CardV2{
		Header: brandedHeader(input.Title, input.Subtitle, input.Tone),
		Sections: []CardSection{
			textSection(input.Summary),
			{Widgets: []CardWidget{{ButtonList: &ButtonList{Buttons: []Button{
				{
					Text: input.EditButtonText,
					OnClick: &OnClick{Action: &Action{
						Function:      actionFunction,
						Parameters:    initCardActionParameters(InitCardFunctionEdit, input.AnchorDate),
						LoadIndicator: "SPINNER",
					}},
				},
			}}}}},
		},
	}
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
	return CardSection{Widgets: rowWidgets(rows)}
}

func rowWidgets(rows []TeamRow) []CardWidget {
	widgets := make([]CardWidget, 0, len(rows))
	for _, row := range rows {
		widgets = append(widgets, CardWidget{
			DecoratedText: &DecoratedText{
				TopLabel: row.Label,
				Text:     row.Value,
			},
		})
	}
	return widgets
}

func initCardActionParameters(action, anchorDate string) []ActionParameter {
	return []ActionParameter{
		{Key: "action", Value: action},
		{Key: "date", Value: anchorDate},
	}
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
