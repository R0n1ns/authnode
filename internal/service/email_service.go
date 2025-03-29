package service

import (
	"fmt"
	"net/smtp"

	"github.com/authmicro/configs"
)

type EmailService struct {
	config configs.SMTPConfig
}

func NewEmailService(config configs.SMTPConfig) *EmailService {
	return &EmailService{
		config: config,
	}
}

// SendVerificationCode sends a verification code to the specified email
func (s *EmailService) SendVerificationCode(to, code string) error {
	// If SMTP is not configured, just return without error for development purposes
	if s.config.Username == "" || s.config.Password == "" {
		fmt.Printf("SMTP not configured, would send code %s to %s\n", code, to)
		return nil
	}

	// Set up authentication
	auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)

	// Compose message
	subject := "Your Verification Code"
	body := fmt.Sprintf("Your verification code is: %s\nThis code will expire in 15 minutes.", code)
	message := fmt.Sprintf("To: %s\r\nFrom: %s\r\nSubject: %s\r\n\r\n%s", to, s.config.From, subject, body)

	// Send email
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	return smtp.SendMail(addr, auth, s.config.From, []string{to}, []byte(message))
}

// SendWelcomeEmail sends a welcome email to a newly registered user
func (s *EmailService) SendWelcomeEmail(to, name string) error {
	// If SMTP is not configured, just return without error for development purposes
	if s.config.Username == "" || s.config.Password == "" {
		fmt.Printf("SMTP not configured, would send welcome email to %s\n", to)
		return nil
	}

	// Set up authentication
	auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)

	// Compose message
	subject := "Welcome to Our Service"
	body := fmt.Sprintf("Dear %s,\n\nWelcome to our service. Thank you for registering!\n\nBest regards,\nThe Team", name)
	message := fmt.Sprintf("To: %s\r\nFrom: %s\r\nSubject: %s\r\n\r\n%s", to, s.config.From, subject, body)

	// Send email
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	return smtp.SendMail(addr, auth, s.config.From, []string{to}, []byte(message))
}
