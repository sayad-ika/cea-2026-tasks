package main

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sayad-ika/craftsbite/internal/discord"
	"github.com/sayad-ika/craftsbite/internal/payload"
	"github.com/sayad-ika/craftsbite/internal/repository"
)

func TestHandleOverrideInitCommand_DiscordOpenPanel(t *testing.T) {
	store := &overrideTestStore{
		users: []repository.User{{ID: "u1", Email: "alice@example.com", Name: "Alice", TeamID: "t1", Active: true}},
		teams: []repository.Team{{ID: "t1", Name: "Platform", Active: true}},
		membersByTeam: map[string][]repository.TeamMember{
			"t1": {{TeamID: "t1", UserID: "u1"}},
		},
	}
	dateParser := overrideDateParser(t)
	date := nextInitTestDates(t, dateParser, 1)[0]
	ctx, recorder := withReplyRecorder(context.Background(), HandlerRequest{Platform: PlatformDiscord})
	err := handleOverrideInitCommand(ctx, store, nil, dateParser, payload.CommandEvent{
		UserID:  "admin-1",
		Role:    "admin",
		Source:  "discord",
		Options: json.RawMessage(`{"action":"open","date":"` + date + `"}`),
	})
	if err != nil {
		t.Fatalf("handleOverrideInitCommand() error = %v", err)
	}

	var body RouterResponse
	if err := json.Unmarshal([]byte(recorder.finalResponse().Body), &body); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if body.Type != 4 || body.Data == nil || body.Data.Flags != 64 {
		t.Fatalf("response = %#v, want ephemeral message", body)
	}
	if body.Data.Embeds[0].Title != overrideInitDiscordPanelTitle {
		t.Fatalf("title = %q", body.Data.Embeds[0].Title)
	}
	if len(body.Data.Components) == 0 || body.Data.Components[0].Components[0].Type != discord.ComponentTypeStringSelect {
		t.Fatalf("components = %#v, want team select", body.Data.Components)
	}
}

func TestDiscordComponentPayload_OverrideInitDetailsModalSubmit(t *testing.T) {
	state := overrideInitState{
		Draft: overrideInitDraft{Date: "2026-05-15", TeamID: "t1", TargetMode: overrideInitTargetWholeTeam},
		Teams: []repository.Team{{ID: "t1", Name: "Platform", Active: true}},
		Team:  &repository.Team{ID: "t1", Name: "Platform", Active: true},
	}
	_, raw, err := discordComponentPayload(interactionBody{Type: 5, Data: interactionData{CustomID: discordOverrideInitCustomID(overrideInitActionDetailsSave, state), Components: []discord.Component{
		{Type: discord.ComponentTypeActionRow, Components: []discord.Component{{CustomID: "date", Value: "2026-05-20"}}},
		{Type: discord.ComponentTypeActionRow, Components: []discord.Component{{CustomID: "entry", Value: "meal"}}},
		{Type: discord.ComponentTypeActionRow, Components: []discord.Component{{CustomID: "meal", Value: "lunch"}}},
		{Type: discord.ComponentTypeActionRow, Components: []discord.Component{{CustomID: "value", Value: "out"}}},
		{Type: discord.ComponentTypeActionRow, Components: []discord.Component{{CustomID: "reason", Value: "Late update"}}},
	}}})
	if err != nil {
		t.Fatalf("discordComponentPayload() error = %v", err)
	}
	var opts payload.OverrideInitOptions
	if err := json.Unmarshal(raw, &opts); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if opts.Action != overrideInitActionDetailsSave || opts.Date != "2026-05-20" || opts.Entry != "meal" || opts.Meal != "lunch" || opts.Value != "out" || opts.Reason != "Late update" {
		t.Fatalf("opts = %+v", opts)
	}
}
