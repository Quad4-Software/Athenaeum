package server

import (
	"fmt"
	"net/smtp"
	"strings"

	"athenaeum/internal/models"
)

func buildMIMEText(from, to, subject, body string) []byte {
	var b strings.Builder
	b.WriteString("From: " + sanitizeHeaderValue(from) + "\r\n")
	b.WriteString("To: " + sanitizeHeaderValue(to) + "\r\n")
	b.WriteString("Subject: " + sanitizeHeaderValue(subject) + "\r\n")
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
	to = sanitizeHeaderValue(to)
	subject = sanitizeHeaderValue(subject)
	// Keep LF line breaks in the body. Strip CR/NUL so they cannot smuggle headers.
	body = strings.ReplaceAll(body, "\r", "")
	body = strings.ReplaceAll(body, "\x00", "")
	if to == "" {
		return fmt.Errorf("invalid recipient")
	}
	from := cfg.FromAddr
	if from == "" {
		from = cfg.Username
	}
	from = sanitizeHeaderValue(from)
	msg := buildMIMEText(from, to, subject, body)
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	var auth smtp.Auth
	if cfg.Username != "" {
		auth = smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
	}
	return smtp.SendMail(addr, auth, from, []string{to}, msg)
}
