package email

import (
	"fmt"
	"net/smtp"
	"strings"
)

type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	BaseURL  string
}

type Service struct {
	config *Config
}

func NewService(config *Config) *Service {
	return &Service{config: config}
}

func (s *Service) SendVerificationEmail(to, displayName, token string) error {
	verifyURL := fmt.Sprintf("%s/verify-email?token=%s", s.config.BaseURL, token)

	subject := "Verify your Aihub email address"
	body := fmt.Sprintf(`Hello %s,

Welcome to Aihub! Please verify your email address by clicking the link below:

%s

This link will expire in 24 hours.

If you did not create an account, please ignore this email.

- The Aihub Team`, displayName, verifyURL)

	return s.sendEmail(to, subject, body)
}

func (s *Service) SendInvitationEmail(to, orgName, inviterName, role, token string) error {
	acceptURL := fmt.Sprintf("%s/invitations/%s", s.config.BaseURL, token)

	subject := fmt.Sprintf("You've been invited to %s on Aihub", orgName)
	body := fmt.Sprintf(`Hello,

%s has invited you to join the organization "%s" on Aihub as a %s.

Click the link below to accept the invitation:

%s

This invitation will expire in 7 days.

- The Aihub Team`, inviterName, orgName, role, acceptURL)

	return s.sendEmail(to, subject, body)
}

func (s *Service) sendEmail(to, subject, body string) error {
	if s.config.Host == "" {
		// In development, just log the email
		fmt.Printf("[EMAIL] To: %s, Subject: %s\n%s\n", to, subject, body)
		return nil
	}

	msg := strings.Join([]string{
		"From: " + s.config.From,
		"To: " + to,
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		body,
	}, "\r\n")

	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)

	var auth smtp.Auth
	if s.config.Username != "" {
		auth = smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)
	}

	return smtp.SendMail(addr, auth, s.config.From, []string{to}, []byte(msg))
}
