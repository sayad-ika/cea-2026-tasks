package gchat

import (
	"context"
	"encoding/json"
	"fmt"

	"google.golang.org/api/chat/v1"
	"google.golang.org/api/option"
)

var newChatService = func(ctx context.Context, opts ...option.ClientOption) (*chat.Service, error) {
	return chat.NewService(ctx, opts...)
}

func chatServiceOptions(serviceAccountJSON string) []option.ClientOption {
	// Google Chat app-auth updates should request only the bot scope.
	return []option.ClientOption{
		option.WithAuthCredentialsJSON(option.ServiceAccount, []byte(serviceAccountJSON)),
		option.WithScopes(chat.ChatBotScope),
	}
}

func decodeMessageBody(body []byte) (*chat.Message, error) {
	var msg chat.Message
	if err := json.Unmarshal(body, &msg); err != nil {
		return nil, fmt.Errorf("json.Unmarshal: %w", err)
	}
	return &msg, nil
}

func CreatePrivateMessage(ctx context.Context, serviceAccountJSON, spaceName, viewerName string, body []byte) error {
	svc, err := newChatService(ctx, chatServiceOptions(serviceAccountJSON)...)
	if err != nil {
		return fmt.Errorf("chat.NewService: %w", err)
	}

	msg, err := decodeMessageBody(body)
	if err != nil {
		return err
	}

	if viewerName != "" {
		msg.PrivateMessageViewer = &chat.User{Name: viewerName}
	}

	_, err = svc.Spaces.Messages.Create(spaceName, msg).Context(ctx).Do()
	return err
}
