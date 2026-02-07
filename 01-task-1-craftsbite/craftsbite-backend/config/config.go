package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
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
	Port string
	Env  string
}

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	DBName          string
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

// Load reads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists (development)
	// In production, environment variables should be set directly
	if err := godotenv.Load(); err != nil {
		// Don't fail if .env doesn't exist (production scenario)
		fmt.Println("Warning: .env file not found, using environment variables")
	}

	config := &Config{
		Server: ServerConfig{
			Host: getEnv("SERVER_HOST", "localhost"),
			Port: getEnv("SERVER_PORT", "8080"),
			Env:  getEnv("ENV", "development"),
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnv("DB_PORT", "5432"),
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", "postgres123"),
			DBName:          getEnv("DB_NAME", "craftsbite_db"),
			SSLMode:         getEnv("DB_SSLMODE", "disable"),
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvAsDuration("DB_CONN_MAX_LIFETIME", "5m"),
		},
		JWT: JWTConfig{
			Secret:     getEnv("JWT_SECRET", ""),
			Expiration: getEnvAsDuration("JWT_EXPIRATION", "24h"),
		},
		CORS: CORSConfig{
			AllowedOrigins: getEnvAsSlice("CORS_ALLOWED_ORIGINS", []string{"http://localhost:3000"}),
		},
		Logging: LoggingConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
		Meal: MealConfig{
			CutoffTime:     getEnv("MEAL_CUTOFF_TIME", "11:00"),
			CutoffTimezone: getEnv("MEAL_CUTOFF_TIMEZONE", "Asia/Dhaka"),
		},
		Cleanup: CleanupConfig{
			RetentionMonths: getEnvAsInt("HISTORY_RETENTION_MONTHS", 3),
			CronSchedule:    getEnv("CLEANUP_CRON", "0 0 * * *"),
		},
		RateLimit: RateLimitConfig{
			Enabled:           getEnvAsBool("RATE_LIMIT_ENABLED", true),
			RequestsPerMinute: getEnvAsInt("RATE_LIMIT_REQUESTS_PER_MINUTE", 100),
		},
	}

	// Validate required fields
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}

// Validate checks if all required configuration fields are set
func (c *Config) Validate() error {
	if c.Database.Password == "" {
		return fmt.Errorf("DB_PASSWORD is required")
	}

	if c.JWT.Secret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}

	if len(c.JWT.Secret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters long")
	}

	if c.Server.Env != "development" && c.Server.Env != "production" && c.Server.Env != "test" {
		return fmt.Errorf("ENV must be one of: development, production, test")
	}

	return nil
}

// GetDSN returns the database connection string
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host,
		c.Port,
		c.User,
		c.Password,
		c.DBName,
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

// GetServerAddress returns the full server address
func (c *ServerConfig) GetServerAddress() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

// Helper functions to read environment variables with defaults

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

func getEnvAsDuration(key, defaultValue string) time.Duration {
	valueStr := getEnv(key, defaultValue)

	duration, err := time.ParseDuration(valueStr)
	if err != nil {
		// If parsing fails, try to parse as duration string
		// Default to the defaultValue
		duration, _ = time.ParseDuration(defaultValue)
	}

	return duration
}

func getEnvAsSlice(key string, defaultValue []string) []string {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	// Split by comma and trim spaces
	var result []string
	for _, v := range splitAndTrim(valueStr, ',') {
		if v != "" {
			result = append(result, v)
		}
	}

	if len(result) == 0 {
		return defaultValue
	}

	return result
}

func splitAndTrim(s string, sep rune) []string {
	var result []string
	current := ""

	for _, char := range s {
		if char == sep {
			if trimmed := trim(current); trimmed != "" {
				result = append(result, trimmed)
			}
			current = ""
		} else {
			current += string(char)
		}
	}

	if trimmed := trim(current); trimmed != "" {
		result = append(result, trimmed)
	}

	return result
}

func trim(s string) string {
	// Simple trim implementation
	start := 0
	end := len(s)

	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n') {
		start++
	}

	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n') {
		end--
	}

	return s[start:end]
}
