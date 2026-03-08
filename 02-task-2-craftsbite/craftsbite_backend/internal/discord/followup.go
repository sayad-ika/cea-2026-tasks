package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

var followupClient = &http.Client{Timeout: 5 * time.Second}

func SendFollowup(applicationID, token, content string) error {
	url := fmt.Sprintf(
		"https://discord.com/api/v10/webhooks/%s/%s/messages/@original",
		applicationID,
		token,
	)

	payload := map[string]interface{}{
		"content": content,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("discord: failed to marshal followup payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPatch, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("discord: failed to build followup request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := followupClient.Do(req)
	if err != nil {
		return fmt.Errorf("discord: followup HTTP call failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("discord: followup returned non-2xx status %d: %s", resp.StatusCode, respBody)
	}

	return nil
}
