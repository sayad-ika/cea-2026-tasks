package gchat

import (
	"encoding/json"
	"testing"

	"github.com/sayad-ika/craftsbite/internal/discord"
)

func TestSimpleTextCard(t *testing.T) {
	got, err := NoticeCard(discord.DefaultNoticeTitle(discord.NoticeToneInfo), discord.DefaultNoticeSubtitle(discord.NoticeToneInfo), "hello world", discord.NoticeToneInfo)
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
	if card.Card.Header == nil || card.Card.Header.Title != discord.DefaultNoticeTitle(discord.NoticeToneInfo) {
		t.Fatalf("header title = %#v, want %q", card.Card.Header, discord.DefaultNoticeTitle(discord.NoticeToneInfo))
	}
	if card.Card.Header.Subtitle != discord.DefaultNoticeSubtitle(discord.NoticeToneInfo) {
		t.Fatalf("subtitle = %q, want %q", card.Card.Header.Subtitle, discord.DefaultNoticeSubtitle(discord.NoticeToneInfo))
	}
	if card.Card.Header.ImageURL == "" {
		t.Error("expected placeholder brand image URL")
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
	if card.Card.Header == nil || card.Card.Header.Title != "Location Updated" {
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
	if len(card.Card.Sections) != 2 {
		t.Fatalf("expected 2 sections, got %d", len(card.Card.Sections))
	}
	if len(card.Card.Sections[1].Widgets) != 2 {
		t.Errorf("expected 2 widgets, got %d", len(card.Card.Sections[1].Widgets))
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
	if card.Card.Header == nil || card.Card.Header.Title != "Headcount Snapshot" {
		t.Error("header title mismatch")
	}
	// header section + meals section + team name + team stats = 4 sections
	if len(card.Card.Sections) < 3 {
		t.Errorf("expected at least 3 sections, got %d", len(card.Card.Sections))
	}
}

func TestCompactHeadcountCard(t *testing.T) {
	meals := []TeamRow{{Label: "Lunch", Value: "10 confirmed"}, {Label: "Iftar", Value: "8 confirmed"}}
	got, err := CompactHeadcountCard("2026-04-28", "Open", 20, 12, 8, meals, "Lunch moved to Level 12")
	if err != nil {
		t.Fatal(err)
	}
	var resp CardResponse
	if err := json.Unmarshal(got, &resp); err != nil {
		t.Fatal(err)
	}
	card := resp.CardsV2[0]
	if card.Card.Header == nil || card.Card.Header.Title != "Daily Headcount Summary" {
		t.Error("header title mismatch")
	}
	if len(card.Card.Sections) != 4 {
		t.Fatalf("expected 4 sections, got %d", len(card.Card.Sections))
	}
}

func TestStatusCard(t *testing.T) {
	got, err := StatusCard("2026-04-28", "Office", []TeamRow{{Label: "Lunch", Value: "Included"}})
	if err != nil {
		t.Fatal(err)
	}
	var resp CardResponse
	if err := json.Unmarshal(got, &resp); err != nil {
		t.Fatal(err)
	}
	card := resp.CardsV2[0]
	if card.Card.Header == nil || card.Card.Header.Title != "Status Snapshot" {
		t.Error("header title mismatch")
	}
}

func TestNoticeCard_WarningUsesWarningPlaceholder(t *testing.T) {
	got, err := NoticeCard(discord.DefaultNoticeTitle(discord.NoticeToneWarning), discord.DefaultNoticeSubtitle(discord.NoticeToneWarning), "warning text", discord.NoticeToneWarning)
	if err != nil {
		t.Fatal(err)
	}
	var resp CardResponse
	if err := json.Unmarshal(got, &resp); err != nil {
		t.Fatal(err)
	}
	card := resp.CardsV2[0]
	if card.Card.Header == nil {
		t.Fatal("expected header")
	}
	if card.Card.Header.Title != discord.DefaultNoticeTitle(discord.NoticeToneWarning) {
		t.Fatalf("Title = %q, want %q", card.Card.Header.Title, discord.DefaultNoticeTitle(discord.NoticeToneWarning))
	}
	if card.Card.Header.Subtitle != discord.DefaultNoticeSubtitle(discord.NoticeToneWarning) {
		t.Fatalf("Subtitle = %q, want %q", card.Card.Header.Subtitle, discord.DefaultNoticeSubtitle(discord.NoticeToneWarning))
	}
	if card.Card.Header.ImageURL != "https://placehold.co/96x96/F08C00/FFFFFF.png?text=CB" {
		t.Fatalf("ImageURL = %q, want warning placeholder", card.Card.Header.ImageURL)
	}
}

func TestInitCard(t *testing.T) {
	card := InitCard(InitCardInput{
		AnchorDate:     "2026-05-15",
		ActionFunction: "https://example.com/gchat",
		Dates: []SelectionItem{
			{Text: "Fri, May 15", Value: "2026-05-15", Selected: true},
		},
		Locations: []SelectionItem{
			{Text: "Office", Value: "office", Selected: true},
			{Text: "WFH", Value: "wfh"},
		},
		Meals: []SelectionItem{
			{Text: "Lunch", Value: "lunch", Selected: true},
		},
	})

	if card.Header == nil || card.Header.Title != "CraftsBite Setup" {
		t.Fatalf("header = %#v, want CraftsBite Setup", card.Header)
	}
	if card.FixedFooter != nil {
		t.Fatal("card messages must not use fixedFooter")
	}
	buttons := card.Sections[len(card.Sections)-1].Widgets[0].ButtonList
	if buttons == nil || len(buttons.Buttons) != 2 {
		t.Fatalf("buttons = %#v, want Save and Cancel button list", buttons)
	}
	if buttons.Buttons[0].OnClick == nil || buttons.Buttons[0].OnClick.Action.Function != "https://example.com/gchat" {
		t.Fatalf("save action = %#v", buttons.Buttons[0].OnClick)
	}
	if actionParameter(buttons.Buttons[0].OnClick.Action.Parameters, "action") != InitCardFunctionSave {
		t.Fatalf("save parameters = %#v, want save action parameter", buttons.Buttons[0].OnClick.Action.Parameters)
	}
	if buttons.Buttons[1].OnClick == nil || buttons.Buttons[1].OnClick.Action.Function != "https://example.com/gchat" {
		t.Fatalf("cancel action = %#v", buttons.Buttons[1].OnClick)
	}
	if len(card.Sections) < 4 {
		t.Fatalf("sections = %d, want at least 4", len(card.Sections))
	}
	if card.Sections[1].Widgets[0].SelectionInput == nil || card.Sections[1].Widgets[0].SelectionInput.Type != "CHECK_BOX" {
		t.Fatalf("date selection input = %#v", card.Sections[1].Widgets[0].SelectionInput)
	}
}

func actionParameter(params []ActionParameter, key string) string {
	for _, param := range params {
		if param.Key == key {
			return param.Value
		}
	}
	return ""
}

func TestInitCardResponse(t *testing.T) {
	body, err := InitCardResponse(InitCardInput{})
	if err != nil {
		t.Fatal(err)
	}
	var resp CardResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatal(err)
	}
	if len(resp.CardsV2) != 1 {
		t.Fatalf("cardsV2 = %d, want 1", len(resp.CardsV2))
	}
	if resp.CardsV2[0].CardID != "init-setup" {
		t.Fatalf("cardID = %q, want init-setup", resp.CardsV2[0].CardID)
	}
	if resp.CardsV2[0].Card.FixedFooter != nil {
		t.Fatal("init card response must be a card message, not a dialog card")
	}
}
