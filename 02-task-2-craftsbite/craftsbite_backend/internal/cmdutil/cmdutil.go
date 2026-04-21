package cmdutil

import (
	"context"
	"fmt"
	"strings"

	"github.com/sayad-ika/craftsbite/internal/config"
	"github.com/sayad-ika/craftsbite/internal/discord"
	"github.com/sayad-ika/craftsbite/internal/gchat"
	"github.com/sayad-ika/craftsbite/internal/payload"
)

func OptString(opts map[string]interface{}, key string) (string, bool) {
	v, ok := opts[key]
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

func DisplayMealName(s string) string {
	words := strings.Split(strings.ReplaceAll(s, "_", " "), " ")
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, " ")
}

func SendReply(ctx context.Context, cfg *config.Config, event payload.CommandEvent, text string) error {
	if event.Source == "gchat" {
		card, err := gchat.SimpleTextCard(text)
		if err != nil {
			return fmt.Errorf("cmdutil: build gchat text card: %w", err)
		}
		return gchat.CreatePrivateMessage(ctx, cfg.GChatServiceAccountJSON, event.GChatSpaceName, event.GChatViewerName, card)
	}
	return discord.SendFollowup(event.ApplicationID, event.InteractionToken, text)
}

func SendGChatCard(ctx context.Context, cfg *config.Config, event payload.CommandEvent, card []byte) error {
	return gchat.CreatePrivateMessage(ctx, cfg.GChatServiceAccountJSON, event.GChatSpaceName, event.GChatViewerName, card)
}
