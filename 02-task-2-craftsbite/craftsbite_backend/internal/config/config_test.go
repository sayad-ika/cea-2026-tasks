package config

import (
	"os"
	"testing"
)

func setEnv(t *testing.T, pairs map[string]string) {
	t.Helper()
	for k, v := range pairs {
		t.Setenv(k, v)
	}
}

func requiredEnv() map[string]string {
	return map[string]string{
		"USERS_TABLE":            "craftsbite-users",
		"MEALS_TABLE":            "craftsbite-meals",
		"WORK_TABLE":             "craftsbite-work",
		"DISCORD_APPLICATION_ID": "app-id",
		"DISCORD_PUBLIC_KEY":     "pub-key",
		"DISCORD_BOT_TOKEN":      "bot-token",
	}
}

func TestLoad_AllFieldsPopulated(t *testing.T) {
	env := requiredEnv()
	env["PORT"] = "9090"
	env["AWS_REGION"] = "us-east-1"
	env["DYNAMODB_ENDPOINT"] = "http://localhost:8000"
	env["CUTOFF_TIME"] = "21:00"
	env["TIMEZONE"] = "Asia/Dhaka"
	setEnv(t, env)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Server.Port != 9090 {
		t.Errorf("expected port 9090, got %d", cfg.Server.Port)
	}
	if cfg.AWS.Region != "us-east-1" {
		t.Errorf("expected region us-east-1, got %s", cfg.AWS.Region)
	}
	if cfg.AWS.DynamoDBEndpoint != "http://localhost:8000" {
		t.Errorf("unexpected DynamoDBEndpoint: %s", cfg.AWS.DynamoDBEndpoint)
	}
	if cfg.Tables.UsersTable != "craftsbite-users" {
		t.Errorf("unexpected UsersTable: %s", cfg.Tables.UsersTable)
	}
	if cfg.Cutoff.CutoffTime != "21:00" {
		t.Errorf("unexpected CutoffTime: %s", cfg.Cutoff.CutoffTime)
	}
	if cfg.Discord.BotToken != "bot-token" {
		t.Errorf("unexpected BotToken: %s", cfg.Discord.BotToken)
	}
}

func TestLoad_PortDefaultsTo8080(t *testing.T) {
	setEnv(t, requiredEnv())
	os.Unsetenv("PORT")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("expected default port 8080, got %d", cfg.Server.Port)
	}
}

func TestLoad_MissingDiscordVar_ReturnsError(t *testing.T) {
	env := requiredEnv()
	delete(env, "DISCORD_BOT_TOKEN")
	setEnv(t, env)
	os.Unsetenv("DISCORD_BOT_TOKEN")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error when DISCORD_BOT_TOKEN is missing, got nil")
	}
}

func TestLoad_MissingTableVar_ReturnsError(t *testing.T) {
	env := requiredEnv()
	delete(env, "MEALS_TABLE")
	setEnv(t, env)
	os.Unsetenv("MEALS_TABLE")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error when MEALS_TABLE is missing, got nil")
	}
}
