package service

import (
	"context"
	"testing"
	"time"

	"authmicro/configs"
	"authmicro/internal/domain"
	"authmicro/pkg/logger"
)

type fakeUsers struct{ byEmail map[string]domain.User }

func (f *fakeUsers) GetByID(_ context.Context, id int64) (domain.User, error) {
	for _, u := range f.byEmail {
		if u.ID == id {
			return u, nil
		}
	}
	return domain.User{}, domain.ErrNotFound
}
func (f *fakeUsers) GetByEmail(_ context.Context, email string) (domain.User, error) {
	u, ok := f.byEmail[email]
	if !ok {
		return u, domain.ErrNotFound
	}
	return u, nil
}
func (f *fakeUsers) NicknameExists(context.Context, string) (bool, error) { return false, nil }
func (f *fakeUsers) EmailExists(_ context.Context, email string) (bool, error) {
	_, ok := f.byEmail[email]
	return ok, nil
}

type fakeRoles struct{}

func (*fakeRoles) GetUserRoleNames(context.Context, int64) ([]string, error) {
	return []string{"user"}, nil
}
func (*fakeRoles) HasRole(context.Context, int64, string) (bool, error) { return false, nil }

type fakeSessions struct {
	registration domain.RegistrationSession
	completed    bool
	attempts     int
}

func (f *fakeSessions) CreateRegistrationSession(_ context.Context, s domain.RegistrationSession) (string, error) {
	s.ID = "registration-id"
	f.registration = s
	return s.ID, nil
}
func (f *fakeSessions) GetRegistrationSession(context.Context, string) (domain.RegistrationSession, error) {
	return f.registration, nil
}
func (f *fakeSessions) UpdateRegistrationCode(_ context.Context, _ string, hash string, expires, sent time.Time, _ time.Duration) error {
	f.registration.CodeHash = hash
	f.registration.CodeExpires = expires
	f.registration.LastSentAt = sent
	return nil
}
func (f *fakeSessions) IncrementRegistrationAttempts(context.Context, string) error {
	f.attempts++
	return nil
}
func (f *fakeSessions) CompleteRegistration(context.Context, string, domain.User, string) error {
	f.completed = true
	return nil
}
func (*fakeSessions) UpsertLoginSession(context.Context, domain.LoginSession, time.Duration) error {
	return nil
}
func (*fakeSessions) GetLoginSession(context.Context, string) (domain.LoginSession, error) {
	return domain.LoginSession{}, domain.ErrInvalidCode
}
func (*fakeSessions) IncrementLoginAttempts(context.Context, string) error { return nil }
func (*fakeSessions) DeleteLoginSession(context.Context, string) error     { return nil }

type fakeTokenIssuer struct{}

func (*fakeTokenIssuer) IssuePair(context.Context, domain.User, []string, string, string) (domain.TokenPair, error) {
	return domain.TokenPair{}, nil
}
func (*fakeTokenIssuer) ConsumeRefreshToken(context.Context, string) (domain.TokenSession, error) {
	return domain.TokenSession{}, domain.ErrInvalidToken
}
func (*fakeTokenIssuer) ValidateToken(string) (*domain.TokenClaims, error) {
	return nil, domain.ErrInvalidToken
}
func (*fakeTokenIssuer) RevokeAll(context.Context, int64) error { return nil }

type captureMailer struct {
	code string
	to   string
}

func (m *captureMailer) SendVerificationCode(to, code string) error {
	m.to = to
	m.code = code
	return nil
}

func newAuthForTest(sessions *fakeSessions, mail *captureMailer) *AuthService {
	otp := configs.OTPConfig{Pepper: "01234567890123456789012345678901", Expiration: 10 * time.Minute, ResendDelay: time.Minute, MaxAttempts: 5}
	return NewAuthService(&fakeUsers{byEmail: map[string]domain.User{}}, &fakeRoles{}, sessions, &fakeTokenIssuer{}, mail, logger.NewLogger(), otp)
}

func TestRegistrationStoresOnlyOTPHashAndConfirms(t *testing.T) {
	sessions := new(fakeSessions)
	mail := new(captureMailer)
	svc := newAuthForTest(sessions, mail)
	now := time.Unix(1_700_000_000, 0)
	svc.now = func() time.Time { return now }
	response, fields, err := svc.CreateRegistrationSession(context.Background(), domain.RegistrationRequest{
		FirstName: "Ada", LastName: "Lovelace", Nickname: "ada_dev", Email: "ADA@example.com", AcceptedPrivacyPolicy: true})
	if err != nil || len(fields) > 0 {
		t.Fatalf("registration failed: fields=%v err=%v", fields, err)
	}
	if response.RegistrationSessionID == "" || mail.to != "ada@example.com" {
		t.Fatalf("unexpected response/mail: %+v %q", response, mail.to)
	}
	if len(mail.code) != 6 {
		t.Fatalf("OTP must have six digits: %q", mail.code)
	}
	if sessions.registration.CodeHash == mail.code || len(sessions.registration.CodeHash) != 64 {
		t.Fatal("repository must receive only an HMAC hash")
	}
	if err = svc.ConfirmEmail(context.Background(), domain.ConfirmEmailRequest{RegistrationSessionID: response.RegistrationSessionID, Code: mail.code}); err != nil {
		t.Fatal(err)
	}
	if !sessions.completed {
		t.Fatal("registration transaction was not completed")
	}
}

func TestRegistrationRejectsWrongOTPAndCountsAttempt(t *testing.T) {
	sessions := new(fakeSessions)
	mail := new(captureMailer)
	svc := newAuthForTest(sessions, mail)
	response, _, err := svc.CreateRegistrationSession(context.Background(), domain.RegistrationRequest{
		FirstName: "Ada", LastName: "Lovelace", Nickname: "ada_dev", Email: "ada@example.com", AcceptedPrivacyPolicy: true})
	if err != nil {
		t.Fatal(err)
	}
	err = svc.ConfirmEmail(context.Background(), domain.ConfirmEmailRequest{RegistrationSessionID: response.RegistrationSessionID, Code: "000000"})
	if err != domain.ErrInvalidCode {
		t.Fatalf("expected invalid code, got %v", err)
	}
	if sessions.attempts != 1 {
		t.Fatalf("expected one failed attempt, got %d", sessions.attempts)
	}
}
