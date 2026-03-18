package gchat

import (
	"encoding/json"
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
