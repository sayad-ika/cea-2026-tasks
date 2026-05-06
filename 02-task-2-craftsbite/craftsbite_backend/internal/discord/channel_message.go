package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

var channelHTTPClient = &http.Client{Timeout: 10 * time.Second}

func CreateChannelMessageObject(botToken, channelID string, message Message) error {
	if botToken == "" || channelID == "" {
		return nil
	}

	url := fmt.Sprintf("https://discord.com/api/v10/channels/%s/messages", channelID)
	body, err := json.Marshal(NormalizeMessage(message))
	if err != nil {
		return fmt.Errorf("discord: marshal channel message: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("discord: build channel message request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bot "+botToken)

	resp, err := channelHTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("discord: channel message HTTP call failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("discord: channel message status %d: %s", resp.StatusCode, respBody)
	}
	return nil
}
