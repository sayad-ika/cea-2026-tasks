package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	appconfig "github.com/sayad-ika/craftsbite/internal/config"
	"github.com/sayad-ika/craftsbite/internal/discord"
	"github.com/sayad-ika/craftsbite/internal/payload"
	"github.com/sayad-ika/craftsbite/internal/repository"
)

// Type 1 = PONG, Type 4 = immediate message, Type 5 = deferred ("thinking"), Type 7 = update message.
type RouterResponse struct {
	Type int              `json:"type"`
	Data *discord.Message `json:"data,omitempty"`
}

func ephemeralNotice(msg string, tone discord.NoticeTone) RouterResponse {
	return ephemeralMessage(discord.ToneMessage(discord.DefaultNoticeTitle(tone), msg, tone))
}

func ephemeralMessage(message discord.Message) RouterResponse {
	message.Flags = 64
	message = discord.NormalizeMessage(message)
	return RouterResponse{
		Type: 4,
		Data: &message,
	}
}

func updateMessage(message discord.Message) RouterResponse {
	message.Flags = 0
	message = discord.NormalizeMessage(message)
	return RouterResponse{
		Type: 7,
		Data: &message,
	}
}

type interactionOption struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
}

type interactionData struct {
	Name          string              `json:"name,omitempty"`
	Options       []interactionOption `json:"options,omitempty"`
	CustomID      string              `json:"custom_id,omitempty"`
	Values        []string            `json:"values,omitempty"`
	ComponentType int                 `json:"component_type,omitempty"`
}

type interactionBody struct {
	Type          int             `json:"type"`
	Data          interactionData `json:"data"`
	Message       discord.Message `json:"message"`
	Token         string          `json:"token"`
	ApplicationID string          `json:"application_id"`
	Member        struct {
		User struct {
			ID string `json:"id"`
		} `json:"user"`
	} `json:"member"`
	User struct {
		ID string `json:"id"`
	} `json:"user"`
}

func discordCommandEvent(ctx context.Context, cfg *appconfig.Config, store *repository.Store, event events.APIGatewayV2HTTPRequest) (payload.CommandEvent, *RouterResponse, error) {
	body, err := decodeAPIGatewayBody(event)
	if err != nil {
		return payload.CommandEvent{}, nil, fmt.Errorf("400: malformed request body: %w", err)
	}

	timestamp := getHeader(event.Headers, "x-signature-timestamp")
	signature := getHeader(event.Headers, "x-signature-ed25519")
	if !discord.VerifySignature(cfg.DiscordPublicKey, timestamp, body, signature) {
		return payload.CommandEvent{}, nil, fmt.Errorf("401: invalid request signature")
	}

	var interaction interactionBody
	if err := json.Unmarshal([]byte(body), &interaction); err != nil {
		return payload.CommandEvent{}, nil, fmt.Errorf("400: malformed JSON body: %w", err)
	}

	if interaction.Type == 1 {
		pong := RouterResponse{Type: 1}
		return payload.CommandEvent{}, &pong, nil
	}

	discordID := interaction.Member.User.ID
	if discordID == "" {
		discordID = interaction.User.ID
	}

	userID, role, err := store.GetUserByDiscordID(ctx, discordID)
	if err != nil {
		return payload.CommandEvent{}, nil, fmt.Errorf("identity resolution failed: %w", err)
	}

	if userID == "" {
		resp := ephemeralNotice("You are not registered. Please contact an administrator.", discord.NoticeToneWarning)
		return payload.CommandEvent{}, &resp, nil
	}

	switch interaction.Type {
	case 2:
		commandName := interaction.Data.Name
		if !knownCommand(commandName) {
			resp := ephemeralNotice(fmt.Sprintf("Unknown command: /%s", commandName), discord.NoticeToneWarning)
			return payload.CommandEvent{}, &resp, nil
		}

		if !discord.CheckPermission(commandName, role) {
			resp := ephemeralNotice(fmt.Sprintf("You do not have permission to use `/%s`.", commandName), discord.NoticeToneWarning)
			return payload.CommandEvent{}, &resp, nil
		}

		optionsMap := make(map[string]interface{}, len(interaction.Data.Options))
		for _, opt := range interaction.Data.Options {
			optionsMap[opt.Name] = opt.Value
		}

		optionsJSON, _ := json.Marshal(optionsMap)

		return payload.CommandEvent{
			UserID:           userID,
			Role:             role,
			DiscordID:        discordID,
			CommandName:      commandName,
			Options:          optionsJSON,
			InteractionToken: interaction.Token,
			ApplicationID:    interaction.ApplicationID,
			Source:           "discord",
		}, nil, nil
	case 3:
		commandName, optionsJSON, err := discordComponentPayload(interaction)
		if err != nil {
			resp := ephemeralNotice(err.Error(), discord.NoticeToneWarning)
			return payload.CommandEvent{}, &resp, nil
		}
		if !discord.CheckPermission(commandName, role) {
			resp := ephemeralNotice(fmt.Sprintf("You do not have permission to use `/%s`.", commandName), discord.NoticeToneWarning)
			return payload.CommandEvent{}, &resp, nil
		}
		return payload.CommandEvent{
			UserID:           userID,
			Role:             role,
			DiscordID:        discordID,
			CommandName:      commandName,
			Options:          optionsJSON,
			InteractionToken: interaction.Token,
			ApplicationID:    interaction.ApplicationID,
			Source:           "discord",
		}, nil, nil
	default:
		resp := ephemeralNotice("This Discord interaction type is not supported yet.", discord.NoticeToneWarning)
		return payload.CommandEvent{}, &resp, nil
	}
}

func decodeAPIGatewayBody(event events.APIGatewayV2HTTPRequest) (string, error) {
	body := event.Body
	if event.IsBase64Encoded {
		decoded, err := base64.StdEncoding.DecodeString(body)
		if err != nil {
			return "", err
		}
		body = string(decoded)
	}
	return body, nil
}

func getHeader(headers map[string]string, name string) string {
	if headers == nil {
		return ""
	}
	if v := headers[name]; v != "" {
		return v
	}
	for k, v := range headers {
		if strings.EqualFold(k, name) {
			return v
		}
	}
	return ""
}

func knownCommand(commandName string) bool {
	switch commandName {
	case "help", "init", "meal", "location", "status", "override", "team-summary", "headcount", "schedule-day", "admin":
		return true
	default:
		return false
	}
}
