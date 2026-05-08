package main

import (
	"context"
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

func TestHandler_AuthorizesValidDiscordShapedRequest(t *testing.T) {
	resp, err := handler(context.Background(), testRequest("POST", "/discord", map[string]string{
		"x-signature-ed25519":   validSignature(),
		"x-signature-timestamp": "1746610000",
	}))
	if err != nil {
		t.Fatalf("handler() error = %v", err)
	}
	if !resp.IsAuthorized {
		t.Fatal("expected request to be authorized")
	}
}

func TestHandler_DeniesMissingSignature(t *testing.T) {
	resp, err := handler(context.Background(), testRequest("POST", "/discord", map[string]string{
		"x-signature-timestamp": "1746610000",
	}))
	if err != nil {
		t.Fatalf("handler() error = %v", err)
	}
	if resp.IsAuthorized {
		t.Fatal("expected request to be denied")
	}
}

func TestHandler_DeniesMissingTimestamp(t *testing.T) {
	resp, err := handler(context.Background(), testRequest("POST", "/discord", map[string]string{
		"x-signature-ed25519": validSignature(),
	}))
	if err != nil {
		t.Fatalf("handler() error = %v", err)
	}
	if resp.IsAuthorized {
		t.Fatal("expected request to be denied")
	}
}

func TestHandler_DeniesWrongMethod(t *testing.T) {
	resp, err := handler(context.Background(), testRequest("GET", "/discord", map[string]string{
		"x-signature-ed25519":   validSignature(),
		"x-signature-timestamp": "1746610000",
	}))
	if err != nil {
		t.Fatalf("handler() error = %v", err)
	}
	if resp.IsAuthorized {
		t.Fatal("expected request to be denied")
	}
}

func TestHandler_DeniesMalformedSignature(t *testing.T) {
	resp, err := handler(context.Background(), testRequest("POST", "/discord", map[string]string{
		"x-signature-ed25519":   malformedSignature(),
		"x-signature-timestamp": "1746610000",
	}))
	if err != nil {
		t.Fatalf("handler() error = %v", err)
	}
	if resp.IsAuthorized {
		t.Fatal("expected request to be denied")
	}
}

func TestHandler_DeniesWrongSignatureLength(t *testing.T) {
	resp, err := handler(context.Background(), testRequest("POST", "/discord", map[string]string{
		"x-signature-ed25519":   "abcd",
		"x-signature-timestamp": "1746610000",
	}))
	if err != nil {
		t.Fatalf("handler() error = %v", err)
	}
	if resp.IsAuthorized {
		t.Fatal("expected request to be denied")
	}
}

func TestHandler_DeniesUnexpectedPath(t *testing.T) {
	resp, err := handler(context.Background(), testRequest("POST", "/gchat", map[string]string{
		"x-signature-ed25519":   validSignature(),
		"x-signature-timestamp": "1746610000",
	}))
	if err != nil {
		t.Fatalf("handler() error = %v", err)
	}
	if resp.IsAuthorized {
		t.Fatal("expected request to be denied")
	}
}

func testRequest(method, path string, headers map[string]string) events.APIGatewayV2CustomAuthorizerV2Request {
	return events.APIGatewayV2CustomAuthorizerV2Request{
		RawPath: path,
		Headers: headers,
		RequestContext: events.APIGatewayV2HTTPRequestContext{
			HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{
				Method: method,
				Path:   path,
			},
		},
	}
}

func validSignature() string {
	return stringsRepeat("ab", 64)
}

func malformedSignature() string {
	return stringsRepeat("zz", 64)
}

func stringsRepeat(value string, count int) string {
	result := ""
	for range count {
		result += value
	}
	return result
}
