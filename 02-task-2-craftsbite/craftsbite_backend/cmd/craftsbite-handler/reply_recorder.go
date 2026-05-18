package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/aws/aws-lambda-go/events"
	appconfig "github.com/sayad-ika/craftsbite/internal/config"
	"github.com/sayad-ika/craftsbite/internal/discord"
	"github.com/sayad-ika/craftsbite/internal/gchat"
	"github.com/sayad-ika/craftsbite/internal/payload"
)

type replyRecorderKey struct{}

type replyRecorder struct {
	req      HandlerRequest
	response *events.APIGatewayV2HTTPResponse
}

func withReplyRecorder(ctx context.Context, req HandlerRequest) (context.Context, *replyRecorder) {
	recorder := &replyRecorder{req: req}
	return context.WithValue(ctx, replyRecorderKey{}, recorder), recorder
}

func recorderFromContext(ctx context.Context) *replyRecorder {
	recorder, _ := ctx.Value(replyRecorderKey{}).(*replyRecorder)
	return recorder
}

func sendReply(ctx context.Context, cfg *appconfig.Config, event payload.CommandEvent, text string) error {
	return sendNotice(ctx, cfg, event, discord.NoticeToneInfo, text)
}

func sendWarningReply(ctx context.Context, cfg *appconfig.Config, event payload.CommandEvent, text string) error {
	return sendNotice(ctx, cfg, event, discord.NoticeToneWarning, text)
}

func sendErrorReply(ctx context.Context, cfg *appconfig.Config, event payload.CommandEvent, text string) error {
	return sendNotice(ctx, cfg, event, discord.NoticeToneError, text)
}

func sendNotice(ctx context.Context, cfg *appconfig.Config, event payload.CommandEvent, tone discord.NoticeTone, text string) error {
	if event.Source == "discord" {
		message := discord.ToneMessage(discord.DefaultNoticeTitle(tone), text, tone)
		return sendDiscordMessage(ctx, cfg, event, message)
	}

	card, err := gchat.NoticeCard(discord.DefaultNoticeTitle(tone), discord.DefaultNoticeSubtitle(tone), text, tone)
	if err != nil {
		if recorder := recorderFromContext(ctx); recorder != nil {
			resp := gchatNoticeText(text, event.GChatViewerName, tone)
			recorder.response = &resp
			return nil
		}
		return fmt.Errorf("build gchat notice card: %w", err)
	}
	return sendGChatCard(ctx, cfg, event, card)
}

func sendDiscordMessage(ctx context.Context, cfg *appconfig.Config, event payload.CommandEvent, message discord.Message) error {
	return sendDiscordInteractionResponse(ctx, ephemeralMessage(message))
}

func sendDiscordInteractionResponse(ctx context.Context, response RouterResponse) error {
	if recorder := recorderFromContext(ctx); recorder != nil {
		resp := discordJSON(response)
		recorder.response = &resp
		return nil
	}
	return fmt.Errorf("no reply recorder in context")
}

func sendGChatCard(ctx context.Context, cfg *appconfig.Config, event payload.CommandEvent, card []byte) error {
	if recorder := recorderFromContext(ctx); recorder != nil {
		resp, err := gchatCardResponse(card, event.GChatViewerName)
		if err != nil {
			slog.Error("gchat card response build failed", "error", err, "card_len", len(card))
			return err
		}
		slog.Info("gchat card response recorded", "command", event.CommandName, "status", resp.StatusCode, "body_len", len(resp.Body), "viewer_present", event.GChatViewerName != "")
		recorder.response = &resp
		return nil
	}
	return fmt.Errorf("no reply recorder in context")
}

func sendGChatJSON(ctx context.Context, body []byte) error {
	if recorder := recorderFromContext(ctx); recorder != nil {
		resp := gchatJSONResponse(body)
		slog.Info("gchat json response recorded", "status", resp.StatusCode, "body_len", len(resp.Body))
		recorder.response = &resp
		return nil
	}
	return fmt.Errorf("no reply recorder in context")
}

func (r *replyRecorder) finalResponse() events.APIGatewayV2HTTPResponse {
	if r.response != nil {
		slog.Info("reply recorder returning recorded response", "platform", r.req.Platform, "command", r.req.Command.CommandName, "status", r.response.StatusCode, "body_len", len(r.response.Body))
		return *r.response
	}
	slog.Warn("reply recorder using default response", "platform", r.req.Platform, "command", r.req.Command.CommandName)
	if r.req.Platform == PlatformDiscord {
		return discordJSON(ephemeralMessage(discord.NoticeMessage(discord.DefaultNoticeTitle(discord.NoticeToneInfo), "Done.")))
	}
	return gchatNoticeText("Done.", r.req.Command.GChatViewerName, discord.NoticeToneInfo)
}

func gchatCardResponse(card []byte, viewerName string) (events.APIGatewayV2HTTPResponse, error) {
	var message map[string]interface{}
	if err := json.Unmarshal(card, &message); err != nil {
		return events.APIGatewayV2HTTPResponse{}, fmt.Errorf("decode gchat card: %w", err)
	}
	messageShape := gchatMessageShape(message)
	if viewerName != "" {
		message["privateMessageViewer"] = map[string]string{"name": viewerName}
	}

	body, _ := json.Marshal(map[string]interface{}{
		"hostAppDataAction": map[string]interface{}{
			"chatDataAction": map[string]interface{}{
				"createMessageAction": map[string]interface{}{
					"message": message,
				},
			},
		},
	})
	slog.Info("gchat create message response built",
		"card_len", len(card),
		"body_len", len(body),
		"has_cards_v2", messageShape.hasCardsV2,
		"cards_v2_count", messageShape.cardsV2Count,
		"has_text", messageShape.hasText,
		"viewer_present", viewerName != "",
	)
	return events.APIGatewayV2HTTPResponse{
		StatusCode: 200,
		Body:       string(body),
		Headers:    map[string]string{"Content-Type": "application/json"},
	}, nil
}

func gchatJSONResponse(body []byte) events.APIGatewayV2HTTPResponse {
	slog.Info("gchat raw json response built", "body_len", len(body))
	return events.APIGatewayV2HTTPResponse{
		StatusCode: 200,
		Body:       string(body),
		Headers:    map[string]string{"Content-Type": "application/json"},
	}
}

type gchatResponseShape struct {
	hasCardsV2   bool
	cardsV2Count int
	hasText      bool
}

func gchatMessageShape(message map[string]interface{}) gchatResponseShape {
	shape := gchatResponseShape{}
	if cards, ok := message["cardsV2"].([]interface{}); ok {
		shape.hasCardsV2 = true
		shape.cardsV2Count = len(cards)
	}
	if text, ok := message["text"].(string); ok && text != "" {
		shape.hasText = true
	}
	return shape
}
