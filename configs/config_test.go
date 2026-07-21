package configs

import (
	"strings"
	"testing"
	"time"
)

func TestLoadRejectsMissingSecrets(t *testing.T) {
	t.Setenv("JWT_SECRET", "")
	t.Setenv("OTP_PEPPER", "")
	if _, err := Load(); err == nil {
		t.Fatal("expected missing secrets to be rejected")
	}
}
func TestLoadParsesDurations(t *testing.T) {
	t.Setenv("JWT_SECRET", strings.Repeat("j", 32))
	t.Setenv("OTP_PEPPER", strings.Repeat("o", 32))
	t.Setenv("JWT_ACCESS_EXPIRATION", "20m")
	t.Setenv("OTP_RESEND_DELAY", "45s")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.JWT.AccessTokenExpiration != 20*time.Minute {
		t.Fatalf("unexpected access duration: %s", cfg.JWT.AccessTokenExpiration)
	}
	if cfg.OTP.ResendDelay != 45*time.Second {
		t.Fatalf("unexpected resend delay: %s", cfg.OTP.ResendDelay)
	}
}
