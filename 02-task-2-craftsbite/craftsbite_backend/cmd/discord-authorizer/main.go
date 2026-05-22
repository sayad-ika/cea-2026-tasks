package main

import (
	"context"
	"encoding/hex"
	"log"
	"strconv"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

const discordSignatureHexLength = 128

func handler(_ context.Context, req events.APIGatewayV2CustomAuthorizerV2Request) (events.APIGatewayV2CustomAuthorizerSimpleResponse, error) {
	// Log all headers for debugging
	log.Printf("=== AUTHORIZER REQUEST ===")
	log.Printf("Method: %s", req.RequestContext.HTTP.Method)
	log.Printf("Path: %s", req.RequestContext.HTTP.Path)
	log.Printf("RawPath: %s", req.RawPath)

	authorized := isAuthorized(req)
	log.Printf("=== AUTHORIZER RESULT: %v ===", authorized)
	return events.APIGatewayV2CustomAuthorizerSimpleResponse{IsAuthorized: authorized}, nil
}

func isAuthorized(req events.APIGatewayV2CustomAuthorizerV2Request) bool {
	method := req.RequestContext.HTTP.Method
	log.Printf("Checking method: %s", method)
	if !strings.EqualFold(method, "POST") {
		log.Printf("Method check failed: not POST")
		return false
	}

	path := strings.ToLower(strings.TrimSpace(req.RawPath))
	if path == "" {
		path = strings.ToLower(strings.TrimSpace(req.RequestContext.HTTP.Path))
	}
	log.Printf("Checking path: %s (RawPath: %s)", path, req.RawPath)
	if !isDiscordPath(path) {
		log.Printf("Path check failed: not a Discord path")
		return false
	}

	timestamp := header(req.Headers, "x-signature-timestamp")
	log.Printf("Timestamp header: '%s'", timestamp)
	if timestamp == "" {
		log.Printf("Timestamp header missing")
		return false
	}
	if _, err := strconv.ParseInt(timestamp, 10, 64); err != nil {
		log.Printf("Timestamp parse failed: %v", err)
		return false
	}

	signature := header(req.Headers, "x-signature-ed25519")
	log.Printf("Signature header length: %d", len(signature))
	result := isHexOfLength(signature, discordSignatureHexLength)
	log.Printf("Signature valid hex: %v", result)
	return result
}

func isDiscordPath(path string) bool {
	return strings.Contains(path, "/discord") || strings.Contains(path, "/interactions")
}

func header(headers map[string]string, name string) string {
	if headers == nil {
		return ""
	}
	if value := headers[name]; value != "" {
		return value
	}
	for key, value := range headers {
		if value != "" && strings.EqualFold(key, name) {
			return value
		}
	}
	return ""
}

func isHexOfLength(value string, expected int) bool {
	if len(value) != expected {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func main() {
	lambda.Start(handler)
}
