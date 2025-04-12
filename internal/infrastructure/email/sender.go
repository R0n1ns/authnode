package email

import (
	"fmt"
	"net/smtp"

	"authmicro/internal/config"
)

// Sender implements the EmailSender interface
type Sender struct {
	config config.EmailConfig
}

// NewSender creates a new email sender
func NewSender(config config.EmailConfig) *Sender {
	return &Sender{
		config: config,
	}
}

// Send sends an email
func (s *Sender) Send(to, subject, body string) error {
	// Set up SMTP authentication
	auth := smtp.PlainAuth("", s.config.SMTPUsername, s.config.SMTPPassword, s.config.SMTPHost)

	// Prepare email content
	from := fmt.Sprintf("%s <%s>", s.config.FromName, s.config.FromEmail)
	headers := make(map[string]string)
	headers["From"] = from
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/plain; charset=\"utf-8\""

	// Build message
	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body

	// Send email
	addr := fmt.Sprintf("%s:%d", s.config.SMTPHost, s.config.SMTPPort)
	err := smtp.SendMail(addr, auth, s.config.FromEmail, []string{to}, []byte(message))
	if err != nil {
		return fmt.Errorf("error sending email: %w", err)
	}

	return nil
}
