package gchat

import (
	"encoding/json"
	"fmt"
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
	Title    string `json:"title"`
	Subtitle string `json:"subtitle,omitempty"`
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

func SimpleTextCard(text string) ([]byte, error) {
	resp := CardResponse{
		CardsV2: []CardV2Wrapper{
			{
				CardID: "response",
				Card: CardV2{
					Sections: []CardSection{
						{
							Widgets: []CardWidget{
								{
									TextParagraph: &TextParagraph{Text: text},
								},
							},
						},
					},
				},
			},
		},
	}
	return json.Marshal(resp)
}

func LocationCard(location, date, mealStatus string) ([]byte, error) {
	resp := CardResponse{
		CardsV2: []CardV2Wrapper{
			{
				CardID: "location-response",
				Card: CardV2{
					Header: &CardHeader{
						Title:    "Location Set",
						Subtitle: date,
					},
					Sections: []CardSection{
						{
							Widgets: []CardWidget{
								{
									DecoratedText: &DecoratedText{
										TopLabel: "Work location",
										Text:     location,
									},
								},
							},
						},
						{
							Widgets: []CardWidget{
								{
									DecoratedText: &DecoratedText{
										TopLabel: "Meal status",
										Text:     mealStatus,
									},
								},
							},
						},
					},
				},
			},
		},
	}
	return json.Marshal(resp)
}

type TeamRow struct {
	Label string
	Value string
}

func TeamSummaryCard(date string, rows []TeamRow) ([]byte, error) {
	widgets := make([]CardWidget, len(rows))
	for i, r := range rows {
		widgets[i] = CardWidget{
			DecoratedText: &DecoratedText{
				TopLabel: r.Label,
				Text:     r.Value,
			},
		}
	}

	resp := CardResponse{
		CardsV2: []CardV2Wrapper{
			{
				CardID: "team-summary-response",
				Card: CardV2{
					Header: &CardHeader{
						Title:    "Team Summary",
						Subtitle: date,
					},
					Sections: []CardSection{
						{
							Widgets: widgets,
						},
					},
				},
			},
		},
	}
	return json.Marshal(resp)
}

func HeadcountCard(date, dayStatusLabel string, totalUsers, office, wfh int, overallMeals []TeamRow, teams []HeadcountTeamSection) ([]byte, error) {
	headerSection := CardSection{
		Widgets: []CardWidget{
			{DecoratedText: &DecoratedText{TopLabel: "Day Status", Text: dayStatusLabel}},
			{DecoratedText: &DecoratedText{TopLabel: "Total Employees", Text: fmt.Sprintf("%d", totalUsers)}},
			{DecoratedText: &DecoratedText{TopLabel: "Office", Text: fmt.Sprintf("%d", office)}},
			{DecoratedText: &DecoratedText{TopLabel: "WFH", Text: fmt.Sprintf("%d", wfh)}},
		},
	}

	sections := []CardSection{headerSection}

	if len(overallMeals) > 0 {
		mealWidgets := make([]CardWidget, len(overallMeals))
		for i, m := range overallMeals {
			mealWidgets[i] = CardWidget{
				DecoratedText: &DecoratedText{
					TopLabel: m.Label,
					Text:     m.Value,
				},
			}
		}
		sections = append(sections, CardSection{Widgets: mealWidgets})
	}

	for _, t := range teams {
		teamWidgets := []CardWidget{
			{DecoratedText: &DecoratedText{TopLabel: "Members", Text: fmt.Sprintf("%d", t.MemberCount)}},
			{DecoratedText: &DecoratedText{TopLabel: "Office / WFH", Text: fmt.Sprintf("%d / %d", t.Office, t.WFH)}},
		}
		for _, m := range t.Meals {
			teamWidgets = append(teamWidgets, CardWidget{
				DecoratedText: &DecoratedText{
					TopLabel: m.Label,
					Text:     m.Value,
				},
			})
		}
		sections = append(sections, CardSection{
			Widgets: []CardWidget{
				{TextParagraph: &TextParagraph{Text: fmt.Sprintf("<b>%s</b>", t.TeamName)}},
			},
		})
		sections = append(sections, CardSection{Widgets: teamWidgets})
	}

	resp := CardResponse{
		CardsV2: []CardV2Wrapper{
			{
				CardID: "headcount-response",
				Card: CardV2{
					Header: &CardHeader{
						Title:    "Headcount",
						Subtitle: date,
					},
					Sections: sections,
				},
			},
		},
	}
	return json.Marshal(resp)
}

type HeadcountTeamSection struct {
	TeamName    string
	MemberCount int
	Office      int
	WFH         int
	Meals       []TeamRow
}
