package configs

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Environment       string
	DB                DBConfig
	JWT               JWTConfig
	OTP               OTPConfig
	SMTP              SMTPConfig
	HTTPServerAddress string
	GRPCServerAddress string
	ShutdownTimeout   time.Duration
}

type DBConfig struct {
	Host            string
	Port            string
	Username        string
	Password        string
	DBName          string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnectTimeout  time.Duration
}

type JWTConfig struct {
	AccessTokenExpiration  time.Duration
	RefreshTokenExpiration time.Duration
	Secret                 string
	Issuer                 string
	Audience               string
}

type OTPConfig struct {
	Pepper      string
	Expiration  time.Duration
	ResendDelay time.Duration
	MaxAttempts int
}

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

func Load() (*Config, error) {
	cfg := &Config{
		Environment: env("APP_ENV", "development"),
		DB: DBConfig{
			Host: env("PGHOST", "localhost"), Port: env("PGPORT", "5432"),
			Username: env("PGUSER", "postgres"), Password: env("PGPASSWORD", "postgres"),
			DBName: env("PGDATABASE", "auth"), SSLMode: env("PGSSLMODE", "disable"),
			MaxOpenConns: envInt("DB_MAX_OPEN_CONNS", 20), MaxIdleConns: envInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: envDuration("DB_CONN_MAX_LIFETIME", time.Hour),
			ConnectTimeout:  envDuration("DB_CONNECT_TIMEOUT", 10*time.Second),
		},
		JWT: JWTConfig{
			AccessTokenExpiration:  envDuration("JWT_ACCESS_EXPIRATION", 15*time.Minute),
			RefreshTokenExpiration: envDuration("JWT_REFRESH_EXPIRATION", 7*24*time.Hour),
			Secret:                 os.Getenv("JWT_SECRET"), Issuer: env("JWT_ISSUER", "authnode"),
			Audience: env("JWT_AUDIENCE", "authnode-api"),
		},
		OTP: OTPConfig{
			Pepper: os.Getenv("OTP_PEPPER"), Expiration: envDuration("OTP_EXPIRATION", 10*time.Minute),
			ResendDelay: envDuration("OTP_RESEND_DELAY", time.Minute), MaxAttempts: envInt("OTP_MAX_ATTEMPTS", 5),
		},
		SMTP: SMTPConfig{
			Host: env("SMTP_HOST", "localhost"), Port: envInt("SMTP_PORT", 1025),
			Username: os.Getenv("SMTP_USERNAME"), Password: os.Getenv("SMTP_PASSWORD"),
			From: env("SMTP_FROM", "no-reply@authnode.local"),
		},
		HTTPServerAddress: env("HTTP_SERVER_ADDRESS", "0.0.0.0:8000"),
		GRPCServerAddress: env("GRPC_SERVER_ADDRESS", "0.0.0.0:9000"),
		ShutdownTimeout:   envDuration("SHUTDOWN_TIMEOUT", 10*time.Second),
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) Validate() error {
	var errs []error
	if len(c.JWT.Secret) < 32 {
		errs = append(errs, fmt.Errorf("JWT_SECRET must be at least 32 characters"))
	}
	if len(c.OTP.Pepper) < 32 {
		errs = append(errs, fmt.Errorf("OTP_PEPPER must be at least 32 characters"))
	}
	if c.OTP.MaxAttempts < 1 {
		errs = append(errs, errors.New("OTP_MAX_ATTEMPTS must be positive"))
	}
	if c.JWT.AccessTokenExpiration <= 0 || c.JWT.RefreshTokenExpiration <= 0 {
		errs = append(errs, errors.New("token durations must be positive"))
	}
	if c.DB.MaxOpenConns < 1 || c.DB.MaxIdleConns < 0 {
		errs = append(errs, errors.New("invalid database pool limits"))
	}
	return errors.Join(errs...)
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func envInt(key string, fallback int) int {
	value, err := strconv.Atoi(env(key, strconv.Itoa(fallback)))
	if err != nil {
		return fallback
	}
	return value
}

func envDuration(key string, fallback time.Duration) time.Duration {
	value, err := time.ParseDuration(env(key, fallback.String()))
	if err != nil {
		return fallback
	}
	return value
}
