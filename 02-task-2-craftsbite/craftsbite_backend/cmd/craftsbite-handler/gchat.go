package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/aws/aws-lambda-go/events"
	appconfig "github.com/sayad-ika/craftsbite/internal/config"
	"github.com/sayad-ika/craftsbite/internal/discord"
	"github.com/sayad-ika/craftsbite/internal/gchat"
	"github.com/sayad-ika/craftsbite/internal/payload"
	"github.com/sayad-ika/craftsbite/internal/repository"
)

func gchatCommandEvent(ctx context.Context, cfg *appconfig.Config, store *repository.Store, req events.APIGatewayV2HTTPRequest) (payload.CommandEvent, *events.APIGatewayV2HTTPResponse, error) {
	body := req.Body
	if req.IsBase64Encoded {
		decoded, err := base64.StdEncoding.DecodeString(body)
		if err != nil {
			slog.Error("base64 decode failed", "error", err)
			resp := events.APIGatewayV2HTTPResponse{StatusCode: 400}
			return payload.CommandEvent{}, &resp, nil
		}
		body = string(decoded)
	}

	var evt gchat.Event
	if err := json.Unmarshal([]byte(body), &evt); err != nil {
		slog.Error("failed to unmarshal request body", "error", err)
		resp := events.APIGatewayV2HTTPResponse{StatusCode: 400}
		return payload.CommandEvent{}, &resp, nil
	}

	viewerName := evt.Chat.User.Name
	if evt.Chat.AppCommandPayload == nil {
		resp := gchatText("Only slash commands are supported.", viewerName)
		return payload.CommandEvent{}, &resp, nil
	}

	userID, role, err := store.GetUserByGChatEmail(ctx, evt.Chat.User.Email)
	if err != nil {
		return payload.CommandEvent{}, nil, fmt.Errorf("gchat handler: identity resolution: %w", err)
	}
	if userID == "" {
		resp := gchatText("Your Google Chat account is not linked to CraftsBite. Contact your admin.", viewerName)
		return payload.CommandEvent{}, &resp, nil
	}

	cmdEvt, err := gchat.ToCommandEvent(evt, userID, role)
	if err != nil {
		resp := gchatText(err.Error(), viewerName)
		return payload.CommandEvent{}, &resp, nil
	}

	if !discord.CheckPermission(cmdEvt.CommandName, role) {
		resp := gchatText(fmt.Sprintf("You do not have permission to use `/%s`.", cmdEvt.CommandName), viewerName)
		return payload.CommandEvent{}, &resp, nil
	}

	if _, ok := discord.Dispatch(cfg, cmdEvt.CommandName); !ok {
		slog.Warn("no target function configured for command", "command", cmdEvt.CommandName)
		resp := gchatText(fmt.Sprintf("Command `/%s` is not configured.", cmdEvt.CommandName), viewerName)
		return payload.CommandEvent{}, &resp, nil
	}

	return cmdEvt, nil, nil
}

func gchatText(msg, viewerName string) events.APIGatewayV2HTTPResponse {
	message := map[string]interface{}{
		"text": msg,
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
	}
}
