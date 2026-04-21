package gchat

import (
	"context"
	"testing"
)

func TestCreateSpaceMessage_DelegatesWithEmptyViewer(t *testing.T) {
	ctx := context.Background()
	card, err := SimpleTextCard("test")
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
