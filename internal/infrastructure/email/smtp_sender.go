package email

import (
	"fmt"
	"net/smtp"

	"loviary.app/backend/pkg/logger"
)

type smtpSender struct {
	cfg *Config
	log *logger.Logger
}

// SendVerificationEmail sends a verification email with the given code
func (s *smtpSender) SendVerificationEmail(to, code string) error {
	subject := "Xác thực tài khoản Loviary"
	body := fmt.Sprintf(
		"Chào bạn,\n\nMã xác thực tài khoản Loviary của bạn là: %s\n\nMã này có hiệu lực trong 15 phút.\n\nTrân trọng,\nĐội ngũ Loviary",
		code,
	)

	// Log attempt
	if s.log != nil {
		s.log.Info("Attempting to send verification email", map[string]interface{}{
			"to":      to,
			"subject": subject,
			"enabled": s.cfg.Enabled,
		})
	} else {
		fmt.Printf("[EMAIL] Attempting to send verification email to: %s, enabled: %v\n", to, s.cfg.Enabled)
	}

	// If SMTP is not enabled, just log to console (dev mode)
	if !s.cfg.Enabled {
		// Try to get logger from context if available, otherwise just print
		if s.log != nil {
			s.log.Info("DEV MODE: Verification email (SMTP disabled)", map[string]interface{}{
				"to":      to,
				"code":    code,
				"subject": subject,
				"body":    body,
			})
		} else {
			fmt.Printf("[DEV] Verification Email\nTo: %s\nSubject: %s\nBody: %s\n\n", to, subject, body)
		}
		return nil
	}

	// Build email headers
	from := s.cfg.From
	toList := []string{to}
	msg := []byte(
		fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s",
			from, to, subject, body),
	)

	// Connect to SMTP server
	auth := smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.Host)
	addr := fmt.Sprintf("%s:%s", s.cfg.Host, s.cfg.Port)

	if s.log != nil {
		s.log.Info("Connecting to SMTP server", map[string]interface{}{
			"host":     s.cfg.Host,
			"port":     s.cfg.Port,
			"userName": s.cfg.Username,
			"password": s.cfg.Password,
			"from":     from,
			"to":       to,
		})
	}

	// Attempt to send
	err := smtp.SendMail(addr, auth, from, toList, msg)
	if err != nil {
		if s.log != nil {
			s.log.Error("Failed to send verification email", err, map[string]interface{}{
				"to":        to,
				"code":      code,
				"smtp_host": s.cfg.Host,
				"smtp_port": s.cfg.Port,
				"smtp_user": s.cfg.Username,
			})
		}
		return fmt.Errorf("failed to send email via SMTP: %w", err)
	}

	if s.log != nil {
		s.log.Info("Verification email sent successfully", map[string]interface{}{
			"to":        to,
			"subject":   subject,
			"smtp_host": s.cfg.Host,
		})
	}

	return nil
}
