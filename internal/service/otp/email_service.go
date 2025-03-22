package otp

import (
	"fmt"
	"net/smtp"
)

// Интерфейс email-сервиса
type EmailService interface {
	SendOTPEmail(email, otpCode string) error
}

// Реализация email-сервиса через Gmail
type GmailEmailService struct {
	smtpHost     string
	smtpPort     string
	senderEmail  string
	senderPasswd string
}

// Конструктор сервиса
func NewGmailEmailService(email, passwd string) *GmailEmailService {
	return &GmailEmailService{
		smtpHost:     "smtp.gmail.com",
		smtpPort:     "587",
		senderEmail:  email,
		senderPasswd: passwd,
	}
}

// Отправка OTP на email
func (s *GmailEmailService) SendOTPEmail(email, otpCode string) error {
	auth := smtp.PlainAuth("", s.senderEmail, s.senderPasswd, s.smtpHost)
	msg := fmt.Sprintf("Subject: Ваш OTP-код\n\nВаш OTP-код: %s", otpCode)

	err := smtp.SendMail(s.smtpHost+":"+s.smtpPort, auth, s.senderEmail, []string{email}, []byte(msg))
	if err != nil {
		return fmt.Errorf("ошибка отправки email: %w", err)
	}
	return nil
}
