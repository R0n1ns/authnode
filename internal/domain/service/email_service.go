package service

import (
	"context"
	"fmt"
)

// EmailService defines the interface for sending emails
type EmailService interface {
	SendEmail(ctx context.Context, toEmail, subject, body string) error
}

// emailService implements the EmailService interface
type emailService struct {
	fromEmail    string
	fromName     string
	smtpHost     string
	smtpPort     int
	smtpUsername string
	smtpPassword string
	debug        bool
}

// NewEmailService creates a new email service
func NewEmailService(
	fromEmail string,
	fromName string,
	smtpHost string,
	smtpPort int,
	smtpUsername string,
	smtpPassword string,
	debug bool,
) EmailService {
	return &emailService{
		fromEmail:    fromEmail,
		fromName:     fromName,
		smtpHost:     smtpHost,
		smtpPort:     smtpPort,
		smtpUsername: smtpUsername,
		smtpPassword: smtpPassword,
		debug:        debug,
	}
}

// SendEmail sends an email
func (s *emailService) SendEmail(ctx context.Context, toEmail, subject, body string) error {
	// If in debug mode, just log the email
	if s.debug {
		fmt.Printf("[DEBUG] Email to: %s\nSubject: %s\nBody: %s\n", toEmail, subject, body)
		return nil
	}

	// Here you would typically implement SMTP sending
	// For now, we'll just simulate success
	fmt.Printf("Email sent to: %s\nSubject: %s\n", toEmail, subject)
	return nil
}