package main

import (
	"context"

	"github.com/sayad-ika/craftsbite/internal/payload"
	"github.com/sayad-ika/craftsbite/internal/ratelimit"
)

func applyRateLimit(ctx context.Context, limiter *ratelimit.Limiter, event payload.CommandEvent) (bool, string, error) {
	if event.Role == "admin" {
		return true, "", nil
	}
	allowed, err := limiter.Allow(ctx, event.UserID, event.CommandName)
	if err != nil {
		return false, "An internal error occurred. Please try again.", err
	}
	if !allowed {
		return false, "Rate limit exceeded. Please slow down and try again after some time.", nil
	}
	return true, "", nil
}
