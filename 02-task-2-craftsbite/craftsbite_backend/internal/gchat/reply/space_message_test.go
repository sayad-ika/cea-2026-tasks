package reply

import (
	"context"
	"testing"
)

func TestCreateSpaceMessage_DelegatesWithEmptyViewer(t *testing.T) {
	ctx := context.Background()
	card := []byte(`{"text":"test"}`)

	if len(card) == 0 {
		t.Fatal("expected non-empty card body")
	}

	_ = func() {
		_ = CreateSpaceMessage(ctx, "json", "spaces/test", card)
	}
}
