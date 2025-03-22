package config

import "time"

// Конфигурация JWT
var JWTSettings = struct {
	AccessSecret  string
	RefreshSecret string
	AccessTTL     time.Duration
	RefreshTTL    time.Duration
}{
	AccessSecret:  "super_secret_access_key",
	RefreshSecret: "super_secret_refresh_key",
	AccessTTL:     time.Minute * 15,
	RefreshTTL:    time.Hour * 24 * 7, // 7 дней
}
