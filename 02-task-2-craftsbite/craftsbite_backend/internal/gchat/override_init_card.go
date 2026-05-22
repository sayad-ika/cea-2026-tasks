package gchat

import (
	"encoding/json"
	"html"
	"strings"

	"github.com/sayad-ika/craftsbite/internal/discord"
)

const (
	OverrideInitFunctionSave       = "override_init_save"
	OverrideInitFunctionCancel     = "override_init_cancel"
	OverrideInitFunctionTeamChange = "override_init_team_change"
	OverrideInitFunctionEdit       = "override_init_edit"

	selectionTypeDropdown = "DROPDOWN"
)

type OverrideInitCardInput struct {
	Title          string
	Subtitle       string
	Intro          string
	Note           string
	Date           string
	Reason         string
	ActionFunction string
	Teams          []SelectionItem
	TargetModes    []SelectionItem
	Members        []SelectionItem
	Entries        []SelectionItem
	Meals          []SelectionItem
	MealValues     []SelectionItem
	LocationValues []SelectionItem
	SummaryRows    []TeamRow
}

type OverrideInitConfirmationInput struct {
	Title          string
	Subtitle       string
	Summary        string
	EditButtonText string
	ActionFunction string
	Tone           discord.NoticeTone
}

func OverrideInitCardResponse(input OverrideInitCardInput) ([]byte, error) {
	return json.Marshal(CardResponse{
		CardsV2: []CardV2Wrapper{{CardID: "override-init", Card: OverrideInitCard(input)}},
	})
}

func OverrideInitConfirmationCardResponse(input OverrideInitConfirmationInput) ([]byte, error) {
	return json.Marshal(CardResponse{
		CardsV2: []CardV2Wrapper{{CardID: "override-init-result", Card: OverrideInitConfirmationCard(input)}},
	})
}

func OverrideInitCard(input OverrideInitCardInput) CardV2 {
	if input.Title == "" {
		input.Title = "Override Setup"
	}
	if input.Subtitle == "" {
		input.Subtitle = "Team meal and location override"
	}
	if input.Intro == "" {
		input.Intro = "Choose a team, decide whether to update the whole team or selected members, then set the override details."
	}
	saveFunction := input.ActionFunction
	cancelFunction := input.ActionFunction
	teamChangeFunction := input.ActionFunction
	if saveFunction == "" {
		saveFunction = OverrideInitFunctionSave
	}
	if cancelFunction == "" {
		cancelFunction = OverrideInitFunctionCancel
	}
	if teamChangeFunction == "" {
		teamChangeFunction = OverrideInitFunctionTeamChange
	}

	sections := []CardSection{}
	if input.Note != "" {
		sections = append(sections, CardSection{Widgets: []CardWidget{{TextParagraph: &TextParagraph{Text: paragraphText(input.Note)}}}})
	}
	sections = append(sections, CardSection{Widgets: []CardWidget{{TextParagraph: &TextParagraph{Text: paragraphText(input.Intro)}}}})

	teamWidgets := []CardWidget{}
	if len(input.Teams) > 0 {
		teamWidgets = append(teamWidgets, CardWidget{SelectionInput: &SelectionInput{
			Name:  "team_id",
			Label: "Team",
			Type:  selectionTypeDropdown,
			Items: input.Teams,
			OnChangeAction: &Action{
				Function:      teamChangeFunction,
				Parameters:    overrideInitActionParameters(OverrideInitFunctionTeamChange),
				LoadIndicator: "NONE",
				PersistValues: true,
			},
		}})
	} else {
		teamWidgets = append(teamWidgets, CardWidget{TextParagraph: &TextParagraph{Text: "No teams are available for your role."}})
	}
	teamWidgets = append(teamWidgets, CardWidget{SelectionInput: &SelectionInput{Name: "target_mode", Label: "Target", Type: selectionTypeRadio, Items: input.TargetModes}})
	if len(input.Members) > 0 {
		teamWidgets = append(teamWidgets, CardWidget{SelectionInput: &SelectionInput{Name: "members", Label: "Specific members", Type: selectionTypeCheckbox, Items: input.Members}})
	}
	sections = append(sections, CardSection{Header: "Target", Widgets: teamWidgets})

	sections = append(sections, CardSection{Header: "Override Details", Widgets: []CardWidget{
		{TextInput: &TextInput{Name: "date", Label: "Date", Type: textInputSingleLine, HintText: "YYYY-MM-DD, today, tomorrow, or +N", Value: input.Date}},
		{SelectionInput: &SelectionInput{Name: "entry", Label: "Entry", Type: selectionTypeRadio, Items: input.Entries}},
		{SelectionInput: &SelectionInput{Name: "meal", Label: "Meal type", Type: selectionTypeRadio, Items: input.Meals}},
		{SelectionInput: &SelectionInput{Name: "meal_value", Label: "Meal value", Type: selectionTypeRadio, Items: input.MealValues}},
		{SelectionInput: &SelectionInput{Name: "location_value", Label: "Location value", Type: selectionTypeRadio, Items: input.LocationValues}},
		{TextInput: &TextInput{Name: "reason", Label: "Reason", Type: textInputMultipleLine, HintText: "Reason for the override", Value: input.Reason}},
	}})

	if len(input.SummaryRows) > 0 {
		sections = append(sections, CardSection{Header: "Preview", Widgets: rowWidgets(input.SummaryRows)})
	}
	sections = append(sections, CardSection{Widgets: []CardWidget{{ButtonList: &ButtonList{Buttons: []Button{
		{
			Text:  "Save override",
			Color: &Color{Red: 0.18, Green: 0.62, Blue: 0.27},
			OnClick: &OnClick{Action: &Action{
				Function:        saveFunction,
				Parameters:      overrideInitActionParameters(OverrideInitFunctionSave),
				LoadIndicator:   "SPINNER",
				RequiredWidgets: []string{"team_id", "target_mode", "date", "entry", "reason"},
			}},
		},
		{
			Text: "Cancel",
			OnClick: &OnClick{Action: &Action{
				Function:      cancelFunction,
				Parameters:    overrideInitActionParameters(OverrideInitFunctionCancel),
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

func OverrideInitConfirmationCard(input OverrideInitConfirmationInput) CardV2 {
	if input.Title == "" {
		input.Title = "Override saved"
	}
	if input.Subtitle == "" {
		input.Subtitle = "Team override result"
	}
	if input.Summary == "" {
		input.Summary = "The override request was processed."
	}
	if input.EditButtonText == "" {
		input.EditButtonText = "Open override panel"
	}
	if input.Tone == "" {
		input.Tone = discord.NoticeToneSuccess
	}
	actionFunction := input.ActionFunction
	if actionFunction == "" {
		actionFunction = OverrideInitFunctionEdit
	}

	return CardV2{
		Header: brandedHeader(input.Title, input.Subtitle, input.Tone),
		Sections: []CardSection{
			textSection(input.Summary),
			{Widgets: []CardWidget{{ButtonList: &ButtonList{Buttons: []Button{
				{Text: input.EditButtonText, OnClick: &OnClick{Action: &Action{Function: actionFunction, Parameters: overrideInitActionParameters(OverrideInitFunctionEdit), LoadIndicator: "SPINNER"}}},
			}}}}},
		},
	}
}

func overrideInitActionParameters(action string) []ActionParameter {
	return []ActionParameter{{Key: "action", Value: action}}
}

func EscapedOverrideInitSummary(lines []string) string {
	escaped := make([]string, 0, len(lines))
	for _, line := range lines {
		escaped = append(escaped, html.EscapeString(line))
	}
	return paragraphText(strings.Join(escaped, "\n"))
}
