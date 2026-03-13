package config_test

import (
	"strings"
	"testing"

	"github.com/sayad-ika/craftsbite/internal/config"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name        string
		env         map[string]string
		wantErr     bool
		wantErrStr  string
		checkConfig func(t *testing.T, cfg *config.Config)
	}{
		{
			name: "all required vars set — returns valid config",
			env: map[string]string{
				"DISCORD_PUBLIC_KEY": "deadbeef",
				"AWS_REGION":         "ap-southeast-1",
				"DYNAMODB_ENDPOINT":  "http://localhost:8000",
				"DYNAMODB_TABLE":     "myapp",
			},
			wantErr: false,
			checkConfig: func(t *testing.T, cfg *config.Config) {
				t.Helper()
				if cfg.DiscordPublicKey != "deadbeef" {
					t.Errorf("DiscordPublicKey = %q, want %q", cfg.DiscordPublicKey, "deadbeef")
				}
				if cfg.AWSRegion != "ap-southeast-1" {
					t.Errorf("AWSRegion = %q, want %q", cfg.AWSRegion, "ap-southeast-1")
				}
				if cfg.DynamoDBEndpoint != "http://localhost:8000" {
					t.Errorf("DynamoDBEndpoint = %q, want %q", cfg.DynamoDBEndpoint, "http://localhost:8000")
				}
				if cfg.DynamoDBTable != "myapp" {
					t.Errorf("DynamoDBTable = %q, want %q", cfg.DynamoDBTable, "myapp")
				}
			},
		},
		{
			name: "AWS_REGION unset — defaults to ap-southeast-1",
			env: map[string]string{
				"DISCORD_PUBLIC_KEY": "deadbeef",
			},
			wantErr: false,
			checkConfig: func(t *testing.T, cfg *config.Config) {
				t.Helper()
				if cfg.AWSRegion != "ap-southeast-1" {
					t.Errorf("AWSRegion = %q, want default %q", cfg.AWSRegion, "ap-southeast-1")
				}
			},
		},
		{
			name: "DYNAMODB_ENDPOINT unset — defaults to empty string",
			env: map[string]string{
				"DISCORD_PUBLIC_KEY": "deadbeef",
			},
			wantErr: false,
			checkConfig: func(t *testing.T, cfg *config.Config) {
				t.Helper()
				if cfg.DynamoDBEndpoint != "" {
					t.Errorf("DynamoDBEndpoint = %q, want empty string", cfg.DynamoDBEndpoint)
				}
			},
		},
		{
			name: "DYNAMODB_TABLE unset — defaults to craftsbite",
			env: map[string]string{
				"DISCORD_PUBLIC_KEY": "deadbeef",
			},
			wantErr: false,
			checkConfig: func(t *testing.T, cfg *config.Config) {
				t.Helper()
				if cfg.DynamoDBTable != "craftsbite" {
					t.Errorf("DynamoDBTable = %q, want default %q", cfg.DynamoDBTable, "craftsbite")
				}
			},
		},
		{
			name:       "DISCORD_PUBLIC_KEY missing — error names the var",
			env:        map[string]string{},
			wantErr:    true,
			wantErrStr: "DISCORD_PUBLIC_KEY",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for _, key := range []string{
				"DISCORD_PUBLIC_KEY", "AWS_REGION", "DYNAMODB_ENDPOINT", "DYNAMODB_TABLE",
			} {
				t.Setenv(key, "")
			}
			for k, v := range tc.env {
				t.Setenv(k, v)
			}

			cfg, err := config.Load()

			if tc.wantErr {
				if err == nil {
					t.Fatal("Load() returned nil error; expected an error")
				}
				if tc.wantErrStr != "" && !strings.Contains(err.Error(), tc.wantErrStr) {
					t.Errorf("error message %q does not contain %q", err.Error(), tc.wantErrStr)
				}
				if cfg != nil {
					t.Errorf("Load() returned non-nil *Config on error; want nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Load() returned unexpected error: %v", err)
			}
			if cfg == nil {
				t.Fatal("Load() returned nil *Config with no error")
			}
			if tc.checkConfig != nil {
				tc.checkConfig(t, cfg)
			}
		})
	}
}

func TestMustLoad_Panics_OnMissingRequired(t *testing.T) {
	for _, key := range []string{
		"DISCORD_PUBLIC_KEY", "AWS_REGION", "DYNAMODB_ENDPOINT", "DYNAMODB_TABLE",
	} {
		t.Setenv(key, "")
	}

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("MustLoad() did not panic; expected a panic")
		}
		msg, ok := r.(string)
		if !ok {
			t.Fatalf("panic value is not a string: %v", r)
		}
		if !strings.Contains(msg, "config:") {
			t.Errorf("panic message %q does not start with \"config:\"", msg)
		}
	}()

	config.MustLoad()
}
