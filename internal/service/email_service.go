package service

import (
	"fmt"
	"net/smtp"
	"strings"

	"authmicro/configs"
)

type EmailService struct{ config configs.SMTPConfig }

func NewEmailService(c configs.SMTPConfig) *EmailService { return &EmailService{config: c} }

func (s *EmailService) SendVerificationCode(to, code string) error {
	subject := "Your Authnode verification code"
	body := fmt.Sprintf("Your verification code is %s. It expires soon. If you did not request it, ignore this message.", code)
	message := strings.Join([]string{
		"From: " + s.config.From, "To: " + to, "Subject: " + subject,
		"MIME-Version: 1.0", "Content-Type: text/plain; charset=UTF-8", "", body,
	}, "\r\n")
	var auth smtp.Auth
	if s.config.Username != "" {
		auth = smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)
	}
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	if err := smtp.SendMail(addr, auth, s.config.From, []string{to}, []byte(message)); err != nil {
		return fmt.Errorf("smtp send: %w", err)
	}
	return nil
}
