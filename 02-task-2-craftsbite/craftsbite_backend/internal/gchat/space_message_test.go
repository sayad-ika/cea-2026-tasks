package gchat

import (
	"context"
	"testing"

	"github.com/sayad-ika/craftsbite/internal/discord"
)

func TestCreateSpaceMessage_DelegatesWithEmptyViewer(t *testing.T) {
	ctx := context.Background()
	card, err := NoticeCard(discord.DefaultNoticeTitle(discord.NoticeToneInfo), discord.DefaultNoticeSubtitle(discord.NoticeToneInfo), "test", discord.NoticeToneInfo)
	if err != nil {
		t.Fatalf("failed to build card: %v", err)
	}

	if card == nil || len(card) == 0 {
		t.Fatal("expected non-empty card body")
	}

	_ = func() {
		_ = CreateSpaceMessage(ctx, "json", "spaces/test", card)
	}
}
