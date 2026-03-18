package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/events"
	lambdaclient "github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/sayad-ika/craftsbite/internal/discord"
	"github.com/sayad-ika/craftsbite/internal/dynamo"
	"github.com/sayad-ika/craftsbite/internal/gchat"
	"github.com/sayad-ika/craftsbite/internal/repository"
)

type CommandPayload struct {
	UserID           string                 `json:"userID"`
	Role             string                 `json:"role"`
	CommandName      string                 `json:"commandName"`
	Options          map[string]interface{} `json:"options"`
	Source           string                 `json:"source"`
	GChatSpaceName   string                 `json:"gchatSpaceName,omitempty"`
	GChatMessageName string                 `json:"gchatMessageName,omitempty"`
}

var gchatCommandNames = map[int64]string{
	1: "meal",
	2: "location",
	3: "team-summary",
	4: "headcount",
}

func handleMessage(ctx context.Context, evt gchat.Event) (events.APIGatewayV2HTTPResponse, error) {
	c := getConfig()

	p := evt.Chat.AppCommandPayload
	if p == nil {
		return gchatText("Only slash commands are supported."), nil
	}

	commandID := int64(p.AppCommandMetadata.AppCommandID)
	commandName, known := gchatCommandNames[commandID]
	if !known {
		return gchatText(fmt.Sprintf("Unknown command ID %d.", commandID)), nil
	}

	userID, role, err := repository.GetUserByGChatEmail(ctx, dynamo.GetClient(c), c.DynamoDBTable, evt.Chat.User.Email)
	if err != nil {
		return events.APIGatewayV2HTTPResponse{StatusCode: 500}, fmt.Errorf("gchat-router: identity resolution: %w", err)
	}
	if userID == "" {
		return gchatText("Your Google Chat account is not linked to CraftsBite. Contact your admin."), nil
	}

	if !discord.CheckPermission(commandName, role) {
		return gchatText(fmt.Sprintf("You do not have permission to use `/%s`.", commandName)), nil
	}

	targetFn, ok := discord.Dispatch(c, commandName)
	if !ok || targetFn == "" {
		log.Printf("gchat-router: no target function configured for command=%q", commandName)
		return gchatText(fmt.Sprintf("Command `/%s` is not configured.", commandName)), nil
	}

	var msgName string
	if p.Message != nil {
		msgName = p.Message.Name
	}

	payload := CommandPayload{
		UserID:           userID,
		Role:             role,
		CommandName:      commandName,
		Options:          map[string]interface{}{},
		Source:           "gchat",
		GChatSpaceName:   p.Space.Name,
		GChatMessageName: msgName,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return events.APIGatewayV2HTTPResponse{StatusCode: 500}, fmt.Errorf("gchat-router: marshal payload: %w", err)
	}

	_, err = getLambdaClient().Invoke(ctx, &lambdaclient.InvokeInput{
		FunctionName:   &targetFn,
		InvocationType: types.InvocationTypeEvent,
		Payload:        payloadBytes,
	})
	if err != nil {
		return events.APIGatewayV2HTTPResponse{StatusCode: 500}, fmt.Errorf("gchat-router: invoke %s: %w", targetFn, err)
	}

	return gchatText("Working on it\u2026"), nil
}

func gchatText(msg string) events.APIGatewayV2HTTPResponse {
	body, _ := json.Marshal(map[string]interface{}{
		"hostAppDataAction": map[string]interface{}{
			"chatDataAction": map[string]interface{}{
				"createMessageAction": map[string]interface{}{
					"message": map[string]string{"text": msg},
				},
			},
		},
	})
	return events.APIGatewayV2HTTPResponse{
		StatusCode: 200,
		Body:       string(body),
		Headers:    map[string]string{"Content-Type": "application/json"},
	}
}
