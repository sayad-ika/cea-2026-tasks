package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"sort"

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
			slog.Error("gchat request base64 decode failed", "error", err, "path", gchatRequestPath(req), "body_len", len(req.Body))
			resp := events.APIGatewayV2HTTPResponse{StatusCode: 400}
			return payload.CommandEvent{}, &resp, nil
		}
		body = string(decoded)
	}
	rawKeys := gchatRawTopLevelKeys(body)

	var evt gchat.Event
	if err := json.Unmarshal([]byte(body), &evt); err != nil {
		slog.Error("gchat request unmarshal failed", "error", err, "path", gchatRequestPath(req), "body_len", len(body), "top_level_keys", rawKeys)
		resp := events.APIGatewayV2HTTPResponse{StatusCode: 400}
		return payload.CommandEvent{}, &resp, nil
	}
	slog.Info("gchat request received",
		"path", gchatRequestPath(req),
		"base64", req.IsBase64Encoded,
		"body_len", len(body),
		"top_level_keys", rawKeys,
		"has_chat_user", evt.Chat.User.Name != "",
		"has_chat_user_email", evt.Chat.User.Email != "",
		"has_app_command_payload", evt.Chat.AppCommandPayload != nil,
		"has_button_clicked_payload", evt.Chat.ButtonClickedPayload != nil,
		"common_invoked_function", evt.CommonEventObject.InvokedFunction,
		"common_parameter_keys", gchatStringMapKeys(evt.CommonEventObject.Parameters),
		"common_form_input_keys", gchatFormInputKeys(evt.CommonEventObject.FormInputs),
	)
	if evt.Chat.AppCommandPayload != nil {
		slog.Info("gchat app command payload",
			"command_id", int64(evt.Chat.AppCommandPayload.AppCommandMetadata.AppCommandID),
			"command_type", evt.Chat.AppCommandPayload.AppCommandMetadata.AppCommandType,
			"space_present", evt.Chat.AppCommandPayload.Space.Name != "",
			"message_present", evt.Chat.AppCommandPayload.Message != nil,
		)
	}
	if evt.Chat.ButtonClickedPayload != nil {
		slog.Info("gchat button payload",
			"is_dialog_event", evt.Chat.ButtonClickedPayload.IsDialogEvent,
			"dialog_event_type", evt.Chat.ButtonClickedPayload.DialogEventType,
		)
	}

	viewerName := evt.Chat.User.Name
	isButtonEvent := evt.Chat.ButtonClickedPayload != nil
	if evt.Chat.AppCommandPayload == nil && !isButtonEvent {
		slog.Warn("gchat unsupported event shape", "top_level_keys", rawKeys, "has_common_event_object", evt.CommonEventObject.InvokedFunction != "" || len(evt.CommonEventObject.Parameters) > 0 || len(evt.CommonEventObject.FormInputs) > 0)
		resp := gchatNoticeText("Only slash commands and card actions are supported.", viewerName, discord.NoticeToneWarning)
		return payload.CommandEvent{}, &resp, nil
	}

	userID, role, err := store.GetUserByGChatEmail(ctx, evt.Chat.User.Email)
	if err != nil {
		slog.Error("gchat identity resolution failed", "error", err, "has_email", evt.Chat.User.Email != "")
		return payload.CommandEvent{}, nil, fmt.Errorf("gchat handler: identity resolution: %w", err)
	}
	if userID == "" {
		slog.Warn("gchat identity not linked", "has_email", evt.Chat.User.Email != "", "viewer_present", viewerName != "")
		resp := gchatNoticeText("Your Google Chat account is not linked to CraftsBite. Contact your admin.", viewerName, discord.NoticeToneWarning)
		return payload.CommandEvent{}, &resp, nil
	}
	slog.Info("gchat identity resolved", "role", role, "viewer_present", viewerName != "")

	var cmdEvt payload.CommandEvent
	if isButtonEvent {
		cmdEvt, err = gchat.ToCardActionCommandEvent(evt, userID, role)
	} else {
		cmdEvt, err = gchat.ToCommandEvent(evt, userID, role)
	}
	if err != nil {
		slog.Warn("gchat command normalization failed", "error", err, "is_button_event", isButtonEvent)
		resp := gchatNoticeText(err.Error(), viewerName, discord.NoticeToneWarning)
		return payload.CommandEvent{}, &resp, nil
	}
	slog.Info("gchat command normalized", "command", cmdEvt.CommandName, "is_button_event", isButtonEvent, "options_len", len(cmdEvt.Options), "space_present", cmdEvt.GChatSpaceName != "", "viewer_present", cmdEvt.GChatViewerName != "")

	if !discord.CheckPermission(cmdEvt.CommandName, role) {
		slog.Warn("gchat permission denied", "command", cmdEvt.CommandName, "role", role)
		resp := gchatNoticeText(fmt.Sprintf("You do not have permission to use `/%s`.", cmdEvt.CommandName), viewerName, discord.NoticeToneWarning)
		return payload.CommandEvent{}, &resp, nil
	}
	slog.Info("gchat permission granted", "command", cmdEvt.CommandName, "role", role)

	if !discord.IsKnownCommand(cmdEvt.CommandName) {
		slog.Warn("no target function configured for command", "command", cmdEvt.CommandName)
		resp := gchatNoticeText(fmt.Sprintf("Command `/%s` is not configured.", cmdEvt.CommandName), viewerName, discord.NoticeToneWarning)
		return payload.CommandEvent{}, &resp, nil
	}

	return cmdEvt, nil, nil
}

func gchatRequestPath(req events.APIGatewayV2HTTPRequest) string {
	if req.RawPath != "" {
		return req.RawPath
	}
	return req.RequestContext.HTTP.Path
}

func gchatRawTopLevelKeys(body string) []string {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal([]byte(body), &raw); err != nil {
		return nil
	}
	keys := make([]string, 0, len(raw))
	for key := range raw {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func gchatStringMapKeys(values map[string]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func gchatFormInputKeys(values map[string]gchat.FormInput) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func gchatNoticeText(msg, viewerName string, tone discord.NoticeTone) events.APIGatewayV2HTTPResponse {
	card, err := gchat.NoticeCard(discord.DefaultNoticeTitle(tone), discord.DefaultNoticeSubtitle(tone), msg, tone)
	if err == nil {
		resp, wrapErr := gchatCardResponse(card, viewerName)
		if wrapErr == nil {
			return resp
		}
	}

	message := map[string]interface{}{"text": msg}
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
