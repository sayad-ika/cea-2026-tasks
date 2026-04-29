package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	appconfig "github.com/sayad-ika/craftsbite/internal/config"
	"github.com/sayad-ika/craftsbite/internal/payload"
	"github.com/sayad-ika/craftsbite/internal/repository"
)

type Platform string

const (
	PlatformDiscord Platform = "discord"
	PlatformGChat   Platform = "gchat"
)

type HandlerRequest struct {
	Platform Platform
	Command  payload.CommandEvent
	Raw      events.APIGatewayV2HTTPRequest
}

func normalizeRequest(ctx context.Context, cfg *appconfig.Config, store *repository.Store, req events.APIGatewayV2HTTPRequest) (HandlerRequest, *events.APIGatewayV2HTTPResponse, *RouterResponse, error) {
	platform, err := detectPlatform(req)
	if err != nil {
		return HandlerRequest{}, nil, nil, err
	}

	switch platform {
	case PlatformDiscord:
		cmdEvt, immediate, err := discordCommandEvent(ctx, cfg, store, req)
		if immediate != nil || err != nil {
			return HandlerRequest{}, nil, immediate, err
		}
		return HandlerRequest{Platform: platform, Command: cmdEvt, Raw: req}, nil, nil, nil
	case PlatformGChat:
		cmdEvt, immediate, err := gchatCommandEvent(ctx, cfg, store, req)
		if immediate != nil || err != nil {
			return HandlerRequest{}, immediate, nil, err
		}
		return HandlerRequest{Platform: platform, Command: cmdEvt, Raw: req}, nil, nil, nil
	default:
		return HandlerRequest{}, nil, nil, fmt.Errorf("unsupported platform: %s", platform)
	}
}

func detectPlatform(req events.APIGatewayV2HTTPRequest) (Platform, error) {
	path := req.RawPath
	if path == "" {
		path = req.RequestContext.HTTP.Path
	}
	path = strings.ToLower(path)

	if strings.Contains(path, "/discord") || path == "/interactions" || hasHeader(req.Headers, "x-signature-ed25519") {
		return PlatformDiscord, nil
	}
	if strings.Contains(path, "/gchat") {
		return PlatformGChat, nil
	}
	return "", fmt.Errorf("unknown request platform for path %q", path)
}

func hasHeader(headers map[string]string, name string) bool {
	if headers == nil {
		return false
	}
	if headers[name] != "" {
		return true
	}
	for k, v := range headers {
		if v != "" && strings.EqualFold(k, name) {
			return true
		}
	}
	return false
}
