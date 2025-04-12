package configs

import (
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the service
type Config struct {
	DB                DBConfig
	JWT               JWTConfig
	SMTP              SMTPConfig
	HTTPServerAddress string
	GRPCServerAddress string
}

// DBConfig holds database configuration
type DBConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	DBName   string
	SSLMode  string
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	AccessTokenExpiration  time.Duration
	RefreshTokenExpiration time.Duration
	Secret                 string
}

// SMTPConfig holds email configuration
type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

// NewConfig initializes and returns a new Config
func NewConfig() *Config {
	return &Config{
		DB: DBConfig{
			Host:     getEnv("PGHOST", "localhost"),
			Port:     getEnv("PGPORT", "5432"),
			Username: getEnv("PGUSER", "postgres"),
			Password: getEnv("PGPASSWORD", "postgres"),
			DBName:   getEnv("PGDATABASE", "auth_db"),
			SSLMode:  getEnv("PGSSLMODE", "require"),
		},
		JWT: JWTConfig{
			AccessTokenExpiration:  time.Duration(getEnvAsInt("JWT_ACCESS_EXPIRATION", 15)) * time.Minute,
			RefreshTokenExpiration: time.Duration(getEnvAsInt("JWT_REFRESH_EXPIRATION", 24*7)) * time.Hour,
			Secret:                 getEnv("JWT_SECRET", "my-super-secret-key"),
		},
		SMTP: SMTPConfig{
			Host:     getEnv("SMTP_HOST", "smtp.gmail.com"),
			Port:     getEnvAsInt("SMTP_PORT", 587),
			Username: getEnv("SMTP_USERNAME", ""),
			Password: getEnv("SMTP_PASSWORD", ""),
			From:     getEnv("SMTP_FROM", "no-reply@example.com"),
		},
		HTTPServerAddress: getEnv("HTTP_SERVER_ADDRESS", "0.0.0.0:8000"),
		GRPCServerAddress: getEnv("GRPC_SERVER_ADDRESS", "0.0.0.0:9000"),
	}
}

// getEnv retrieves the value of the environment variable named by the key
// If the variable is not present, it returns the fallback value
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// getEnvAsInt retrieves the value of the environment variable named by the key as an int
// If the variable is not present or cannot be parsed as an int, it returns the fallback value
func getEnvAsInt(key string, fallback int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return fallback
}
