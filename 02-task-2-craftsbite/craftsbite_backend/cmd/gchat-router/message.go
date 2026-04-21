package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	lambdaclient "github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	appconfig "github.com/sayad-ika/craftsbite/internal/config"
	"github.com/sayad-ika/craftsbite/internal/discord"
	"github.com/sayad-ika/craftsbite/internal/gchat"
	"github.com/sayad-ika/craftsbite/internal/repository"
)

func handleMessage(ctx context.Context, cfg *appconfig.Config, client *dynamodb.Client, lambdaClient *lambdaclient.Client, evt gchat.Event) (events.APIGatewayV2HTTPResponse, error) {
	viewerName := evt.Chat.User.Name

	p := evt.Chat.AppCommandPayload
	if p == nil {
		return gchatText("Only slash commands are supported.", viewerName), nil
	}

	userID, role, err := repository.GetUserByGChatEmail(ctx, client, cfg.DynamoDBTable, evt.Chat.User.Email)
	if err != nil {
		return events.APIGatewayV2HTTPResponse{StatusCode: 500}, fmt.Errorf("gchat-router: identity resolution: %w", err)
	}
	if userID == "" {
		return gchatText("Your Google Chat account is not linked to CraftsBite. Contact your admin.", viewerName), nil
	}

	cmdEvt, err := gchat.ToCommandEvent(evt, userID, role)
	if err != nil {
		return gchatText(err.Error(), viewerName), nil
	}

	if !discord.CheckPermission(cmdEvt.CommandName, role) {
		return gchatText(fmt.Sprintf("You do not have permission to use `/%s`.", cmdEvt.CommandName), viewerName), nil
	}

	targetFn, ok := discord.Dispatch(cfg, cmdEvt.CommandName)
	if !ok || targetFn == "" {
		log.Printf("gchat-router: no target function configured for command=%q", cmdEvt.CommandName)
		return gchatText(fmt.Sprintf("Command `/%s` is not configured.", cmdEvt.CommandName), viewerName), nil
	}

	payloadBytes, err := json.Marshal(cmdEvt)
	if err != nil {
		return events.APIGatewayV2HTTPResponse{StatusCode: 500}, fmt.Errorf("gchat-router: marshal payload: %w", err)
	}

	_, err = lambdaClient.Invoke(ctx, &lambdaclient.InvokeInput{
		FunctionName:   &targetFn,
		InvocationType: types.InvocationTypeEvent,
		Payload:        payloadBytes,
	})
	if err != nil {
		return events.APIGatewayV2HTTPResponse{StatusCode: 500}, fmt.Errorf("gchat-router: invoke %s: %w", targetFn, err)
	}

	return gchatText("Working on it\u2026", viewerName), nil
}

func gchatText(msg, viewerName string) events.APIGatewayV2HTTPResponse {
	message := map[string]interface{}{
		"text": msg,
	}
	if viewerName != "" {
		message["privateMessageViewer"] = map[string]string{"name": viewerName}
	}

	body, _ := json.Marshal(map[string]interface{}{
		"hostAppDataAction": map[string]interface{}{
			"chatDataAction": map[string]interface{}{
				"createMessageAction": map[string]interface{}{
					"message": message,
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
