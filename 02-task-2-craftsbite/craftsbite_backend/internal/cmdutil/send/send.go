package send

import (
	"context"
	"fmt"

	"github.com/sayad-ika/craftsbite/internal/config"
	"github.com/sayad-ika/craftsbite/internal/discord"
	"github.com/sayad-ika/craftsbite/internal/gchat"
	gchatreply "github.com/sayad-ika/craftsbite/internal/gchat/reply"
	"github.com/sayad-ika/craftsbite/internal/payload"
)

func Reply(ctx context.Context, cfg *config.Config, event payload.CommandEvent, text string) error {
	if event.Source == "gchat" {
		card, err := gchat.NoticeCard(discord.DefaultNoticeTitle(discord.NoticeToneInfo), discord.DefaultNoticeSubtitle(discord.NoticeToneInfo), text, discord.NoticeToneInfo)
		if err != nil {
			return fmt.Errorf("cmdutil: build gchat text card: %w", err)
		}
		return gchatreply.CreatePrivateMessage(ctx, cfg.GChatServiceAccountJSON, event.GChatSpaceName, event.GChatViewerName, card)
	}
	return DiscordMessage(ctx, cfg, event, discord.NoticeMessage(discord.DefaultNoticeTitle(discord.NoticeToneInfo), text))
}

func GChatCard(ctx context.Context, cfg *config.Config, event payload.CommandEvent, card []byte) error {
	return gchatreply.CreatePrivateMessage(ctx, cfg.GChatServiceAccountJSON, event.GChatSpaceName, event.GChatViewerName, card)
}

func DiscordMessage(ctx context.Context, cfg *config.Config, event payload.CommandEvent, message discord.Message) error {
	_ = ctx
	_ = cfg
	return discord.SendFollowupMessage(event.ApplicationID, event.InteractionToken, message)
}
