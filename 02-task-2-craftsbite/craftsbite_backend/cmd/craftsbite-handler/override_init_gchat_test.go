package main

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sayad-ika/craftsbite/internal/payload"
	"github.com/sayad-ika/craftsbite/internal/repository"
)

func TestHandleOverrideInitCommand_GChatOpenCard(t *testing.T) {
	store := &overrideTestStore{
		users: []repository.User{
			{ID: "u1", Email: "alice@example.com", Name: "Alice", TeamID: "t1", Active: true},
			{ID: "u2", Email: "bob@example.com", Name: "Bob", TeamID: "t1", Active: true},
		},
		teams: []repository.Team{{ID: "t1", Name: "Platform", Active: true}},
		membersByTeam: map[string][]repository.TeamMember{
			"t1": {{TeamID: "t1", UserID: "u1"}, {TeamID: "t1", UserID: "u2"}},
		},
	}
	dateParser := overrideDateParser(t)
	date := nextInitTestDates(t, dateParser, 1)[0]
	ctx, recorder := withReplyRecorder(context.Background(), HandlerRequest{Platform: PlatformGChat})
	err := handleOverrideInitCommand(ctx, store, nil, dateParser, payload.CommandEvent{
		UserID:  "admin-1",
		Role:    "admin",
		Source:  "gchat",
		Options: json.RawMessage(`{"action":"open","date":"` + date + `"}`),
	})
	if err != nil {
		t.Fatalf("handleOverrideInitCommand() error = %v", err)
	}

	cards := gchatCreatedCards(t, recorder.finalResponse().Body)
	card := &cards[0].Card
	if card.Header == nil || card.Header.Title != "Override Setup" {
		t.Fatalf("header = %#v, want Override Setup", card.Header)
	}
	teamInput := gchatSelectionInput(card, "team_id")
	if teamInput == nil || !gchatSelectionSelected(teamInput, "t1") {
		t.Fatalf("team input = %#v, want selected t1", teamInput)
	}
	memberInput := gchatSelectionInput(card, "members")
	if memberInput == nil || len(memberInput.Items) != 2 {
		t.Fatalf("member input = %#v, want two members", memberInput)
	}
	if save := gchatFirstButton(card, "Save override"); save == nil || save.OnClick == nil {
		t.Fatalf("missing save button: %#v", card)
	}
}
