package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/sayad-ika/craftsbite/internal/config"
	"github.com/sayad-ika/craftsbite/internal/discord"
)

type RouterEvent struct {
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

type RouterResponse struct {
	Type int `json:"type"`
}

type interactionBody struct {
	Type int `json:"type"`
}

var (
	cfgOnce sync.Once
	cfg     *config.Config
)

func getConfig() *config.Config {
	cfgOnce.Do(func() {
		cfg = config.MustLoad()
	})
	return cfg
}

func handler(ctx context.Context, event RouterEvent) (RouterResponse, error) {
	c := getConfig()

	timestamp := event.Headers["x-signature-timestamp"]
	signature := event.Headers["x-signature-ed25519"]

	if !discord.VerifySignature(c.DiscordPublicKey, timestamp, event.Body, signature) {
		return RouterResponse{}, errors.New("401: invalid request signature")
	}

	var interaction interactionBody
	if err := json.Unmarshal([]byte(event.Body), &interaction); err != nil {
		return RouterResponse{}, fmt.Errorf("400: malformed JSON body: %w", err)
	}

	if interaction.Type == 1 {
		return RouterResponse{Type: 1}, nil
	}

	return RouterResponse{Type: 5}, nil
}

func main() {
	lambda.Start(handler)
}
