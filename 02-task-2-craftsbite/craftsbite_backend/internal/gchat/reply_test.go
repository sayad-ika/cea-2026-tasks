package gchat

import (
	"testing"
)

func TestDecodeMessageBody_Valid(t *testing.T) {
	msg, err := decodeMessageBody([]byte(`{"text":"hello"}`))
	if err != nil {
		t.Fatal(err)
	}
	if msg == nil {
		t.Fatal("expected non-nil message")
	}
	if msg.Text != "hello" {
		t.Errorf("expected text 'hello', got %q", msg.Text)
	}
}

func TestDecodeMessageBody_Invalid(t *testing.T) {
	_, err := decodeMessageBody([]byte(`{not json`))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestChatServiceOptions(t *testing.T) {
	opts := chatServiceOptions("fake-service-account-json")
	if opts == nil {
		t.Fatal("expected non-nil options")
	}
	if len(opts) != 2 {
		t.Errorf("expected 2 options, got %d", len(opts))
	}
}
