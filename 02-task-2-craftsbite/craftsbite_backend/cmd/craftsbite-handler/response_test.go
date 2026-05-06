package main

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/sayad-ika/craftsbite/internal/discord"
	"github.com/sayad-ika/craftsbite/internal/payload"
)

func TestEphemeralMessageUsesEmbedPayload(t *testing.T) {
	resp := ephemeralMessage(discord.NoticeMessage(discord.DefaultNoticeTitle(discord.NoticeToneInfo), "hello world"))
	if resp.Type != 4 {
		t.Fatalf("Type = %d, want 4", resp.Type)
	}
	if resp.Data == nil {
		t.Fatal("expected response data")
	}
	if resp.Data.Flags != 64 {
		t.Fatalf("Flags = %d, want 64", resp.Data.Flags)
	}
	if len(resp.Data.Embeds) != 1 {
		t.Fatalf("expected 1 embed, got %d", len(resp.Data.Embeds))
	}
	if resp.Data.Embeds[0].Color != discord.BrandColor {
		t.Fatalf("embed color = %d, want %d", resp.Data.Embeds[0].Color, discord.BrandColor)
	}
}

func TestEphemeralNoticeUsesWarningColor(t *testing.T) {
	resp := ephemeralNotice("slow down", discord.NoticeToneWarning)
	if resp.Data == nil || len(resp.Data.Embeds) != 1 {
		t.Fatal("expected warning embed payload")
	}
	if resp.Data.Embeds[0].Color != discord.WarningColor {
		t.Fatalf("embed color = %d, want %d", resp.Data.Embeds[0].Color, discord.WarningColor)
	}
	if resp.Data.Embeds[0].Title != discord.DefaultNoticeTitle(discord.NoticeToneWarning) {
		t.Fatalf("title = %q, want %q", resp.Data.Embeds[0].Title, discord.DefaultNoticeTitle(discord.NoticeToneWarning))
	}
}

func TestGChatTextBuildsCardEnvelope(t *testing.T) {
	resp := gchatText("hello world", "users/123")
	if resp.StatusCode != 200 {
		t.Fatalf("StatusCode = %d, want 200", resp.StatusCode)
	}

	var body map[string]interface{}
	if err := json.Unmarshal([]byte(resp.Body), &body); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	message := body["hostAppDataAction"].(map[string]interface{})["chatDataAction"].(map[string]interface{})["createMessageAction"].(map[string]interface{})["message"].(map[string]interface{})
	if _, ok := message["cardsV2"].([]interface{}); !ok {
		t.Fatal("expected cardsV2 payload")
	}
	viewer := message["privateMessageViewer"].(map[string]interface{})
	if viewer["name"] != "users/123" {
		t.Fatalf("viewer name = %v, want users/123", viewer["name"])
	}
}

func TestPlatformTextResponseUsesRichDiscordPayload(t *testing.T) {
	resp := platformTextResponse(HandlerRequest{Platform: PlatformDiscord, Command: payload.CommandEvent{}}, "rate limit")
	var body RouterResponse
	if err := json.Unmarshal([]byte(resp.Body), &body); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if body.Data == nil || len(body.Data.Embeds) == 0 {
		t.Fatal("expected discord embed payload")
	}
}

func TestGChatNoticeTextUsesWarningCardHeader(t *testing.T) {
	resp := gchatNoticeText("warning text", "users/123", discord.NoticeToneWarning)
	var body map[string]interface{}
	if err := json.Unmarshal([]byte(resp.Body), &body); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	message := body["hostAppDataAction"].(map[string]interface{})["chatDataAction"].(map[string]interface{})["createMessageAction"].(map[string]interface{})["message"].(map[string]interface{})
	cards := message["cardsV2"].([]interface{})
	header := cards[0].(map[string]interface{})["card"].(map[string]interface{})["header"].(map[string]interface{})
	if header["title"] != discord.DefaultNoticeTitle(discord.NoticeToneWarning) {
		t.Fatalf("header title = %v, want %q", header["title"], discord.DefaultNoticeTitle(discord.NoticeToneWarning))
	}
	if header["subtitle"] != discord.DefaultNoticeSubtitle(discord.NoticeToneWarning) {
		t.Fatalf("header subtitle = %v, want %q", header["subtitle"], discord.DefaultNoticeSubtitle(discord.NoticeToneWarning))
	}
	if header["imageUrl"] != "https://placehold.co/96x96/F08C00/FFFFFF.png?text=CB" {
		t.Fatalf("imageUrl = %v, want warning placeholder", header["imageUrl"])
	}
}

func TestSendWarningReply_DiscordUsesInteractionPayload(t *testing.T) {
	ctx, recorder := withReplyRecorder(context.Background(), HandlerRequest{Platform: PlatformDiscord})
	err := sendWarningReply(ctx, nil, payload.CommandEvent{Source: "discord"}, "slow down")
	if err != nil {
		t.Fatalf("sendWarningReply() error = %v", err)
	}

	resp := recorder.finalResponse()
	if strings.Contains(resp.Body, "hostAppDataAction") {
		t.Fatalf("expected discord interaction payload, got gchat envelope: %s", resp.Body)
	}

	var body RouterResponse
	if err := json.Unmarshal([]byte(resp.Body), &body); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if body.Data == nil || len(body.Data.Embeds) != 1 {
		t.Fatal("expected discord embed payload")
	}
	if body.Data.Embeds[0].Color != discord.WarningColor {
		t.Fatalf("embed color = %d, want %d", body.Data.Embeds[0].Color, discord.WarningColor)
	}
}
