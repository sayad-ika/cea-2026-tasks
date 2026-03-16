package main

import (
	"context"
	"errors"
	"os"

	"google.golang.org/api/idtoken"
)

func verifyGChatToken(ctx context.Context, authHeader string) error {
	if len(authHeader) < 8 || authHeader[:7] != "Bearer " {
		return errors.New("missing or malformed Authorization header")
	}
	token := authHeader[7:]
	_, err := idtoken.Validate(ctx, token, os.Getenv("GCHAT_AUDIENCE"))
	return err
}
