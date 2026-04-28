package discord

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func withFollowupServer(t *testing.T, handler http.HandlerFunc, fn func()) {
	t.Helper()
	server := httptest.NewServer(handler)
	defer server.Close()

	orig := followupClient
	defer func() { followupClient = orig }()

	followupClient = &http.Client{
		Timeout: orig.Timeout,
		Transport: &urlRewriter{target: server.Listener.Addr().String()},
	}

	fn()
}

func TestSendFollowup_Success(t *testing.T) {
	withFollowupServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}, func() {
		err := SendFollowup("app-id", "token-abc", "response text")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestSendFollowup_Non2xx(t *testing.T) {
	withFollowupServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"message": "Missing Access"}`))
	}, func() {
		err := SendFollowup("app-id", "token-abc", "response text")
		if err == nil {
			t.Fatal("expected error for 403 response")
		}
		if !strings.Contains(err.Error(), "403") {
			t.Errorf("expected error to contain status 403, got: %v", err)
		}
	})
}

func TestSendFollowup_NetworkError(t *testing.T) {
	orig := followupClient
	defer func() { followupClient = orig }()

	followupClient = &http.Client{
		Transport: &http.Transport{},
		Timeout:   followupClient.Timeout,
	}

	err := SendFollowup("", "", "")
	if err == nil {
		t.Fatal("expected error from invalid URL")
	}
}
