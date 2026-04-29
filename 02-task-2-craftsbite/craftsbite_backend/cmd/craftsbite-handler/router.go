package main

import (
	"context"
	"fmt"
)

func route(ctx context.Context, deps handlerDeps, req HandlerRequest) error {
	switch req.Platform {
	case PlatformDiscord, PlatformGChat:
		switch req.Command.CommandName {
		case "meal", "location", "status":
			return handleSelfCommand(ctx, deps, req.Command)
		case "override", "team-summary":
			return handleManagementCommand(ctx, deps, req.Command)
		case "headcount", "schedule-day", "admin":
			return handleOpsCommand(ctx, deps, req.Command)
		default:
			return sendReply(ctx, deps.cfg, req.Command, fmt.Sprintf("Unknown command: /%s", req.Command.CommandName))
		}
	default:
		return fmt.Errorf("unsupported platform: %s", req.Platform)
	}
}
