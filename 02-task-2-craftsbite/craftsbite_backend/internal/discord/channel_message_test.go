package discord

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type urlRewriter struct {
	target string
}

func (u *urlRewriter) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Scheme = "http"
	req.URL.Host = u.target
	return http.DefaultTransport.RoundTrip(req)
}

func withTestServer(t *testing.T, handler http.HandlerFunc, fn func()) {
	t.Helper()
	server := httptest.NewServer(handler)
	defer server.Close()

	orig := channelHTTPClient
	defer func() { channelHTTPClient = orig }()

	channelHTTPClient = &http.Client{
		Transport: &urlRewriter{target: server.Listener.Addr().String()},
	}

	fn()
}

func TestCreateChannelMessage_EmptyToken(t *testing.T) {
	err := CreateChannelMessage("", "123", "hello")
	if err != nil {
		t.Errorf("expected nil for empty token, got %v", err)
	}
}

func TestCreateChannelMessage_EmptyChannelID(t *testing.T) {
	err := CreateChannelMessage("token", "", "hello")
	if err != nil {
		t.Errorf("expected nil for empty channel ID, got %v", err)
	}
}

func TestCreateChannelMessage_Success(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bot test-token" {
			t.Errorf("expected Bot test-token auth header")
		}
		w.WriteHeader(http.StatusOK)
	}, func() {
		err := CreateChannelMessage("test-token", "chan123", "hello world")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestCreateChannelMessage_Non2xx(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"message": "Missing Access"}`))
	}, func() {
		err := CreateChannelMessage("test-token", "chan123", "hello")
		if err == nil {
			t.Fatal("expected error for 403 response")
		}
		if !strings.Contains(err.Error(), "403") {
			t.Errorf("expected error to contain status 403, got: %v", err)
		}
	})
}

func TestCreateChannelMessage_Truncation(t *testing.T) {
	var receivedContent string
	withTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		receivedContent = body["content"]
		w.WriteHeader(http.StatusOK)
	}, func() {
		longContent := strings.Repeat("a", 2500)
		err := CreateChannelMessage("test-token", "chan123", longContent)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(receivedContent) > 2000 {
			t.Errorf("expected content <= 2000 chars, got %d", len(receivedContent))
		}
		if !strings.Contains(receivedContent, "truncated") {
			t.Error("expected truncated notice in content")
		}
	})
}
