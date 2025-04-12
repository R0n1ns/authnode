package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config represents the application configuration
type Config struct {
	HTTP     HTTPConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Email    EmailConfig
}

// HTTPConfig represents the HTTP server configuration
type HTTPConfig struct {
	Port int
}

// DatabaseConfig represents the database configuration
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

// JWTConfig represents the JWT configuration
type JWTConfig struct {
	Secret                   string
	AccessTokenExpiryMinutes int
	RefreshTokenExpiryDays   int
}

// EmailConfig represents the email configuration
type EmailConfig struct {
	FromEmail    string
	FromName     string
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	Debug        bool
}

// LoadConfig loads the application configuration from environment variables
func LoadConfig() (*Config, error) {
	// Load HTTP config
	httpPort, err := strconv.Atoi(getEnv("HTTP_PORT", "5000"))
	if err != nil {
		return nil, fmt.Errorf("invalid HTTP_PORT: %w", err)
	}

	// Load database config
	dbPort, err := strconv.Atoi(getEnv("PGPORT", "5432"))
	if err != nil {
		return nil, fmt.Errorf("invalid PGPORT: %w", err)
	}

	// Load JWT config
	accessTokenExpiryMinutes, err := strconv.Atoi(getEnv("JWT_ACCESS_TOKEN_EXPIRY_MINUTES", "15"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_ACCESS_TOKEN_EXPIRY_MINUTES: %w", err)
	}

	refreshTokenExpiryDays, err := strconv.Atoi(getEnv("JWT_REFRESH_TOKEN_EXPIRY_DAYS", "7"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_REFRESH_TOKEN_EXPIRY_DAYS: %w", err)
	}

	// Load email config
	smtpPort, err := strconv.Atoi(getEnv("SMTP_PORT", "587"))
	if err != nil {
		return nil, fmt.Errorf("invalid SMTP_PORT: %w", err)
	}

	emailDebug, err := strconv.ParseBool(getEnv("EMAIL_DEBUG", "true"))
	if err != nil {
		return nil, fmt.Errorf("invalid EMAIL_DEBUG: %w", err)
	}

	// Create config
	config := &Config{
		HTTP: HTTPConfig{
			Port: httpPort,
		},
		Database: DatabaseConfig{
			Host:     getEnv("PGHOST", "localhost"),
			Port:     dbPort,
			User:     getEnv("PGUSER", "postgres"),
			Password: getEnv("PGPASSWORD", "postgres"),
			DBName:   getEnv("PGDATABASE", "auth"),
		},
		JWT: JWTConfig{
			Secret:                   getEnv("JWT_SECRET", "your-secret-key"),
			AccessTokenExpiryMinutes: accessTokenExpiryMinutes,
			RefreshTokenExpiryDays:   refreshTokenExpiryDays,
		},
		Email: EmailConfig{
			FromEmail:    getEnv("EMAIL_FROM", "noreply@example.com"),
			FromName:     getEnv("EMAIL_FROM_NAME", "Auth Service"),
			SMTPHost:     getEnv("SMTP_HOST", "smtp.example.com"),
			SMTPPort:     smtpPort,
			SMTPUsername: getEnv("SMTP_USERNAME", ""),
			SMTPPassword: getEnv("SMTP_PASSWORD", ""),
			Debug:        emailDebug,
		},
	}

	return config, nil
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
