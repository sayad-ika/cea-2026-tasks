package gchat

import (
	"encoding/json"
	"testing"
)

func TestSimpleTextCard(t *testing.T) {
	got, err := SimpleTextCard("hello world")
	if err != nil {
		t.Fatal(err)
	}
	var resp CardResponse
	if err := json.Unmarshal(got, &resp); err != nil {
		t.Fatal(err)
	}
	if len(resp.CardsV2) != 1 {
		t.Fatalf("expected 1 card, got %d", len(resp.CardsV2))
	}
	card := resp.CardsV2[0]
	if card.CardID != "response" {
		t.Errorf("cardID = %q, want %q", card.CardID, "response")
	}
	if len(card.Card.Sections) == 0 || len(card.Card.Sections[0].Widgets) == 0 {
		t.Fatal("expected at least one widget")
	}
	w := card.Card.Sections[0].Widgets[0]
	if w.TextParagraph == nil || w.TextParagraph.Text != "hello world" {
		t.Errorf("text = %q, want %q", w.TextParagraph.Text, "hello world")
	}
}

func TestLocationCard(t *testing.T) {
	got, err := LocationCard("Office", "2026-04-28", "Lunch ordered")
	if err != nil {
		t.Fatal(err)
	}
	var resp CardResponse
	if err := json.Unmarshal(got, &resp); err != nil {
		t.Fatal(err)
	}
	card := resp.CardsV2[0]
	if card.Card.Header == nil || card.Card.Header.Title != "Location Set" {
		t.Error("header title mismatch")
	}
	if len(card.Card.Sections) != 2 {
		t.Fatalf("expected 2 sections, got %d", len(card.Card.Sections))
	}
}

func TestTeamSummaryCard(t *testing.T) {
	rows := []TeamRow{{Label: "Office", Value: "5"}, {Label: "WFH", Value: "3"}}
	got, err := TeamSummaryCard("2026-04-28", rows)
	if err != nil {
		t.Fatal(err)
	}
	var resp CardResponse
	if err := json.Unmarshal(got, &resp); err != nil {
		t.Fatal(err)
	}
	card := resp.CardsV2[0]
	if card.Card.Header == nil || card.Card.Header.Title != "Team Summary" {
		t.Error("header title mismatch")
	}
	if len(card.Card.Sections[0].Widgets) != 2 {
		t.Errorf("expected 2 widgets, got %d", len(card.Card.Sections[0].Widgets))
	}
}

func TestHeadcountCard(t *testing.T) {
	meals := []TeamRow{{Label: "Lunch", Value: "10"}, {Label: "Dinner", Value: "4"}}
	teams := []HeadcountTeamSection{
		{TeamName: "Engineering", MemberCount: 8, Office: 5, WFH: 3, Meals: []TeamRow{{Label: "Lunch", Value: "6"}}},
	}
	got, err := HeadcountCard("2026-04-28", "Open", 20, 12, 8, meals, teams)
	if err != nil {
		t.Fatal(err)
	}
	var resp CardResponse
	if err := json.Unmarshal(got, &resp); err != nil {
		t.Fatal(err)
	}
	card := resp.CardsV2[0]
	if card.Card.Header == nil || card.Card.Header.Title != "Headcount" {
		t.Error("header title mismatch")
	}
	// header section + meals section + team name + team stats = 4 sections
	if len(card.Card.Sections) < 3 {
		t.Errorf("expected at least 3 sections, got %d", len(card.Card.Sections))
	}
}
