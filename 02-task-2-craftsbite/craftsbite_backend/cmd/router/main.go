package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	lambdaclient "github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	appconfig "github.com/sayad-ika/craftsbite/internal/config"
	"github.com/sayad-ika/craftsbite/internal/discord"
	"github.com/sayad-ika/craftsbite/internal/dynamo"
	"github.com/sayad-ika/craftsbite/internal/repository"
)

// Type 1 = PONG, Type 4 = immediate message, Type 5 = deferred ("thinking").
type RouterResponse struct {
	Type int              `json:"type"`
	Data *ResponseData   `json:"data,omitempty"`
}

type ResponseData struct {
	Content string `json:"content"`
	Flags   int    `json:"flags"`
}

func ephemeral(msg string) RouterResponse {
	return RouterResponse{
		Type: 4,
		Data: &ResponseData{Content: msg, Flags: 64},
	}
}

type interactionOption struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
}

type interactionBody struct {
	Type int `json:"type"`
	Data struct {
		Name    string              `json:"name"`
		Options []interactionOption `json:"options"`
	} `json:"data"`
	Token         string `json:"token"`
	ApplicationID string `json:"application_id"`
	Member        struct {
		User struct {
			ID string `json:"id"`
		} `json:"user"`
	} `json:"member"`
	User struct {
		ID string `json:"id"`
	} `json:"user"`
}

type CommandPayload struct {
	UserID           string                 `json:"userID"`
	Role             string                 `json:"role"`
	DiscordID        string                 `json:"discordId"`
	CommandName      string                 `json:"commandName"`
	Options          map[string]interface{} `json:"options"`
	InteractionToken string                 `json:"interactionToken"`
	ApplicationID    string                 `json:"applicationId"`
}

var (
	cfgOnce    sync.Once
	cfg        *appconfig.Config
	lambdaOnce sync.Once
	lc         *lambdaclient.Client
)

func getConfig() *appconfig.Config {
	cfgOnce.Do(func() {
		cfg = appconfig.MustLoad()
	})
	return cfg
}

func newLambdaClient(c *appconfig.Config) *lambdaclient.Client {
	awscfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion(c.AWSRegion),
	)
	if err != nil {
		panic(fmt.Sprintf("router: failed to load AWS config for Lambda client: %v", err))
	}
	return lambdaclient.NewFromConfig(awscfg)
}

func getLambdaClient() *lambdaclient.Client {
	lambdaOnce.Do(func() {
		lc = newLambdaClient(getConfig())
	})
	return lc
}

func handler(ctx context.Context, event events.APIGatewayV2HTTPRequest) (RouterResponse, error) {
	c := getConfig()

	timestamp := event.Headers["x-signature-timestamp"]
	signature := event.Headers["x-signature-ed25519"]
	if !discord.VerifySignature(c.DiscordPublicKey, timestamp, event.Body, signature) {
		return RouterResponse{}, fmt.Errorf("401: invalid request signature")
	}

	var interaction interactionBody
	if err := json.Unmarshal([]byte(event.Body), &interaction); err != nil {
		return RouterResponse{}, fmt.Errorf("400: malformed JSON body: %w", err)
	}

	if interaction.Type == 1 {
		return RouterResponse{Type: 1}, nil
	}

	discordID := interaction.Member.User.ID
	if discordID == "" {
		discordID = interaction.User.ID
	}

	userID, role, err := repository.GetUserByDiscordID(ctx, dynamo.GetClient(c), c.DynamoDBTable, discordID)
	if err != nil {
		return RouterResponse{}, fmt.Errorf("identity resolution failed: %w", err)
	}

	if userID == "" {
		return ephemeral("You are not registered. Please contact an administrator."), nil
	}

	commandName := interaction.Data.Name

	targetFn, ok := discord.Dispatch(c, commandName)
	if !ok {
		return ephemeral(fmt.Sprintf("Unknown command: /%s", commandName)), nil
	}

	if targetFn == "" {
		return RouterResponse{}, fmt.Errorf("500: function name for command %q is not configured", commandName)
	}

	if !discord.CheckPermission(commandName, role) {
		return ephemeral(fmt.Sprintf("You do not have permission to use `/%s`.", commandName)), nil
	}

	optionsMap := make(map[string]interface{}, len(interaction.Data.Options))
	for _, opt := range interaction.Data.Options {
		optionsMap[opt.Name] = opt.Value
	}

	payload := CommandPayload{
		UserID:           userID,
		Role:             role,
		DiscordID:        discordID,
		CommandName:      commandName,
		Options:          optionsMap,
		InteractionToken: interaction.Token,
		ApplicationID:    interaction.ApplicationID,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return RouterResponse{}, fmt.Errorf("failed to marshal command payload: %w", err)
	}

	_, err = getLambdaClient().Invoke(ctx, &lambdaclient.InvokeInput{
		FunctionName:   &targetFn,
		InvocationType: types.InvocationTypeEvent,
		Payload:        payloadBytes,
	})
	if err != nil {
		return RouterResponse{}, fmt.Errorf("failed to invoke %s: %w", targetFn, err)
	}

	return RouterResponse{
		Type: 5,
		Data: &ResponseData{Flags: 64},
	}, nil
}

func main() {
	lambda.Start(handler)
}
