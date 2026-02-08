package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	Server    ServerConfig
	Database  DatabaseConfig
	JWT       JWTConfig
	CORS      CORSConfig
	Logging   LoggingConfig
	Meal      MealConfig
	Cleanup   CleanupConfig
	RateLimit RateLimitConfig
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Host string
	Port int
	Env  string
}

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	Name            string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// JWTConfig holds JWT authentication configuration
type JWTConfig struct {
	Secret     string
	Expiration time.Duration
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins []string
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string
	Format string
}

// MealConfig holds meal-specific business rules
type MealConfig struct {
	CutoffTime     string
	CutoffTimezone string
}

// CleanupConfig holds history cleanup configuration
type CleanupConfig struct {
	RetentionMonths int
	CronSchedule    string
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Enabled           bool
	RequestsPerMinute int
}

// LoadConfig reads configuration from .env file and environment variables using Viper
func LoadConfig() (*Config, error) {
	// Set config file name and type
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./")

	// Enable automatic environment variable binding
	viper.AutomaticEnv()

	// Read config file (don't fail if it doesn't exist - use env vars)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		// Config file not found; using environment variables only
		fmt.Println("Warning: .env file not found, using environment variables")
	}

	// Set defaults
	setDefaults()

	// Build config struct
	config := &Config{
		Server: ServerConfig{
			Host: viper.GetString("SERVER_HOST"),
			Port: viper.GetInt("SERVER_PORT"),
			Env:  viper.GetString("ENV"),
		},
		Database: DatabaseConfig{
			Host:            viper.GetString("DB_HOST"),
			Port:            viper.GetString("DB_PORT"),
			User:            viper.GetString("DB_USER"),
			Password:        viper.GetString("DB_PASSWORD"),
			Name:            viper.GetString("DB_NAME"),
			SSLMode:         viper.GetString("DB_SSLMODE"),
			MaxOpenConns:    viper.GetInt("DB_MAX_OPEN_CONNS"),
			MaxIdleConns:    viper.GetInt("DB_MAX_IDLE_CONNS"),
			ConnMaxLifetime: viper.GetDuration("DB_CONN_MAX_LIFETIME"),
		},
		JWT: JWTConfig{
			Secret:     viper.GetString("JWT_SECRET"),
			Expiration: viper.GetDuration("JWT_EXPIRATION"),
		},
		CORS: CORSConfig{
			AllowedOrigins: parseCommaSeparated(viper.GetString("CORS_ALLOWED_ORIGINS")),
		},
		Logging: LoggingConfig{
			Level:  viper.GetString("LOG_LEVEL"),
			Format: viper.GetString("LOG_FORMAT"),
		},
		Meal: MealConfig{
			CutoffTime:     viper.GetString("MEAL_CUTOFF_TIME"),
			CutoffTimezone: viper.GetString("MEAL_CUTOFF_TIMEZONE"),
		},
		Cleanup: CleanupConfig{
			RetentionMonths: viper.GetInt("HISTORY_RETENTION_MONTHS"),
			CronSchedule:    viper.GetString("CLEANUP_CRON"),
		},
		RateLimit: RateLimitConfig{
			Enabled:           viper.GetBool("RATE_LIMIT_ENABLED"),
			RequestsPerMinute: viper.GetInt("RATE_LIMIT_REQUESTS_PER_MINUTE"),
		},
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

// setDefaults sets default values for configuration
func setDefaults() {
	// Server defaults
	viper.SetDefault("SERVER_HOST", "localhost")
	viper.SetDefault("SERVER_PORT", 8080)
	viper.SetDefault("ENV", "development")

	// Database defaults
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", "5432")
	viper.SetDefault("DB_USER", "craftsbite")
	viper.SetDefault("DB_NAME", "craftsbite_db")
	viper.SetDefault("DB_SSLMODE", "disable")
	viper.SetDefault("DB_MAX_OPEN_CONNS", 25)
	viper.SetDefault("DB_MAX_IDLE_CONNS", 5)
	viper.SetDefault("DB_CONN_MAX_LIFETIME", "5m")

	// JWT defaults
	viper.SetDefault("JWT_EXPIRATION", "24h")

	// CORS defaults
	viper.SetDefault("CORS_ALLOWED_ORIGINS", "http://localhost:3000")

	// Logging defaults
	viper.SetDefault("LOG_LEVEL", "debug")
	viper.SetDefault("LOG_FORMAT", "json")

	// Meal defaults
	viper.SetDefault("MEAL_CUTOFF_TIME", "11:00")
	viper.SetDefault("MEAL_CUTOFF_TIMEZONE", "Asia/Dhaka")

	// Cleanup defaults
	viper.SetDefault("HISTORY_RETENTION_MONTHS", 3)
	viper.SetDefault("CLEANUP_CRON", "0 2 * * *")

	// Rate limit defaults
	viper.SetDefault("RATE_LIMIT_ENABLED", true)
	viper.SetDefault("RATE_LIMIT_REQUESTS_PER_MINUTE", 100)
}

// Validate checks if all required configuration fields are set correctly
func (c *Config) Validate() error {
	// Validate JWT Secret
	if c.JWT.Secret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}
	if len(c.JWT.Secret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters long")
	}

	// Validate Database Password
	if c.Database.Password == "" {
		return fmt.Errorf("DB_PASSWORD is required")
	}

	// Validate Environment
	validEnvs := map[string]bool{
		"development": true,
		"staging":     true,
		"production":  true,
	}
	if !validEnvs[c.Server.Env] {
		return fmt.Errorf("ENV must be one of: development, staging, production")
	}

	return nil
}

// GetDSN returns the PostgreSQL connection string
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host,
		c.Port,
		c.User,
		c.Password,
		c.Name,
		c.SSLMode,
	)
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.Server.Env == "production"
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Server.Env == "development"
}

// GetServerAddress returns the full server address (host:port)
func (c *ServerConfig) GetServerAddress() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// parseCommaSeparated splits a comma-separated string into a slice
func parseCommaSeparated(s string) []string {
	if s == "" {
		return []string{}
	}

	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}
