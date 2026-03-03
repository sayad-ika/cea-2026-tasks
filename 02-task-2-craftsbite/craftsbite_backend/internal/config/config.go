package config

import (
	"fmt"
	"os"
	"strconv"
)

type ServerConfig struct {
	Port int
}

type AWSConfig struct {
	Region           string
	DynamoDBEndpoint string
}

type TablesConfig struct {
	UsersTable string
	MealsTable string
	WorkTable  string
}

type CutoffConfig struct {
	CutoffTime string
	Timezone   string
}

type DiscordConfig struct {
	ApplicationID string
	PublicKey     string
	BotToken      string
}

type Config struct {
	Server  ServerConfig
	AWS     AWSConfig
	Tables  TablesConfig
	Cutoff  CutoffConfig
	Discord DiscordConfig
}

func Load() (*Config, error) {
	port := 8080
	if p := os.Getenv("PORT"); p != "" {
		parsed, err := strconv.Atoi(p)
		if err != nil {
			return nil, fmt.Errorf("invalid PORT value %q: %w", p, err)
		}
		port = parsed
	}

	cfg := &Config{
		Server: ServerConfig{Port: port},
		AWS: AWSConfig{
			Region:           os.Getenv("AWS_REGION"),
			DynamoDBEndpoint: os.Getenv("DYNAMODB_ENDPOINT"),
		},
		Tables: TablesConfig{
			UsersTable: os.Getenv("USERS_TABLE"),
			MealsTable: os.Getenv("MEALS_TABLE"),
			WorkTable:  os.Getenv("WORK_TABLE"),
		},
		Cutoff: CutoffConfig{
			CutoffTime: os.Getenv("CUTOFF_TIME"),
			Timezone:   os.Getenv("TIMEZONE"),
		},
		Discord: DiscordConfig{
			ApplicationID: os.Getenv("DISCORD_APPLICATION_ID"),
			PublicKey:     os.Getenv("DISCORD_PUBLIC_KEY"),
			BotToken:      os.Getenv("DISCORD_BOT_TOKEN"),
		},
	}

	required := []struct {
		name  string
		value string
	}{
		{"USERS_TABLE", cfg.Tables.UsersTable},
		{"MEALS_TABLE", cfg.Tables.MealsTable},
		{"WORK_TABLE", cfg.Tables.WorkTable},
		{"DISCORD_APPLICATION_ID", cfg.Discord.ApplicationID},
		{"DISCORD_PUBLIC_KEY", cfg.Discord.PublicKey},
		{"DISCORD_BOT_TOKEN", cfg.Discord.BotToken},
	}

	for _, r := range required {
		if r.value == "" {
			return nil, fmt.Errorf("%s is required but not set", r.name)
		}
	}

	return cfg, nil
}
