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
		card, err := gchat.NoticeCard(discord.DefaultNoticeTitle(discord.NoticeToneInfo), discord.DefaultNoticeSubtitle(discord.NoticeToneInfo), text, discord.NoticeToneInfo)
		if err != nil {
			return fmt.Errorf("cmdutil: build gchat text card: %w", err)
		}
		return gchat.CreatePrivateMessage(ctx, cfg.GChatServiceAccountJSON, event.GChatSpaceName, event.GChatViewerName, card)
	}
	return SendDiscordMessage(ctx, cfg, event, discord.NoticeMessage(discord.DefaultNoticeTitle(discord.NoticeToneInfo), text))
}

func SendGChatCard(ctx context.Context, cfg *config.Config, event payload.CommandEvent, card []byte) error {
	return gchat.CreatePrivateMessage(ctx, cfg.GChatServiceAccountJSON, event.GChatSpaceName, event.GChatViewerName, card)
}

func SendDiscordMessage(ctx context.Context, cfg *config.Config, event payload.CommandEvent, message discord.Message) error {
	_ = ctx
	_ = cfg
	return discord.SendFollowupMessage(event.ApplicationID, event.InteractionToken, message)
}
