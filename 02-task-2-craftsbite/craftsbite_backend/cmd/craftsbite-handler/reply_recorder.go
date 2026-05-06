package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/sayad-ika/craftsbite/internal/cmdutil"
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
		return cmdutil.SendReply(ctx, cfg, event, text)
	}
	return sendGChatCard(ctx, cfg, event, card)
}

func sendDiscordMessage(ctx context.Context, cfg *appconfig.Config, event payload.CommandEvent, message discord.Message) error {
	if recorder := recorderFromContext(ctx); recorder != nil {
		resp := discordJSON(ephemeralMessage(message))
		recorder.response = &resp
		return nil
	}
	return cmdutil.SendDiscordMessage(ctx, cfg, event, message)
}

func sendGChatCard(ctx context.Context, cfg *appconfig.Config, event payload.CommandEvent, card []byte) error {
	if recorder := recorderFromContext(ctx); recorder != nil {
		resp, err := gchatCardResponse(card, event.GChatViewerName)
		if err != nil {
			return err
		}
		recorder.response = &resp
		return nil
	}
	return cmdutil.SendGChatCard(ctx, cfg, event, card)
}

func (r *replyRecorder) finalResponse() events.APIGatewayV2HTTPResponse {
	if r.response != nil {
		return *r.response
	}
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
	return events.APIGatewayV2HTTPResponse{
		StatusCode: 200,
		Body:       string(body),
		Headers:    map[string]string{"Content-Type": "application/json"},
	}, nil
}
