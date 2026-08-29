package server

import (
	"fmt"
	"net/smtp"
	"strings"

	"athenaeum/internal/models"
)

func buildMIMEText(from, to, subject, body string) []byte {
	var b strings.Builder
	b.WriteString("From: " + from + "\r\n")
	b.WriteString("To: " + to + "\r\n")
	b.WriteString("Subject: " + subject + "\r\n")
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/plain; charset=utf-8\r\n\r\n")
	b.WriteString(body)
	b.WriteString("\r\n")
	return []byte(b.String())
}

func sendSMTPText(cfg models.SMTPSettings, to, subject, body string) error {
	if !cfg.Enabled || cfg.Host == "" {
		return fmt.Errorf("smtp is not configured")
	}
	from := cfg.FromAddr
	if from == "" {
		from = cfg.Username
	}
	msg := buildMIMEText(from, to, subject, body)
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	var auth smtp.Auth
	if cfg.Username != "" {
		auth = smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
	}
	return smtp.SendMail(addr, auth, from, []string{to}, msg)
}
