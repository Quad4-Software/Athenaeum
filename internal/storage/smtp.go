package storage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"athenaeum/internal/models"
)

// GetSMTPSettings returns the singleton outbound mail configuration.
func (s *Store) GetSMTPSettings(ctx context.Context) (models.SMTPSettings, error) {
	var c models.SMTPSettings
	var enabled, useTLS int
	err := s.queryRowContext(ctx, `
SELECT enabled, host, port, username, password, from_addr, use_tls FROM smtp_settings WHERE id=1`).
		Scan(&enabled, &c.Host, &c.Port, &c.Username, &c.Password, &c.FromAddr, &useTLS)
	if err != nil {
		return c, err
	}
	c.Enabled = enabled != 0
	c.UseTLS = useTLS != 0
	return c, nil
}

// SaveSMTPSettings updates the singleton outbound mail configuration.
func (s *Store) SaveSMTPSettings(ctx context.Context, c models.SMTPSettings) error {
	_, err := s.execContext(ctx, `
UPDATE smtp_settings SET enabled=?, host=?, port=?, username=?, password=?, from_addr=?, use_tls=?, updated_at=?
WHERE id=1`,
		boolToInt(c.Enabled), c.Host, c.Port, c.Username, c.Password, c.FromAddr, boolToInt(c.UseTLS), time.Now().Unix())
	return err
}

// GetKindleEmail returns the Kindle "send to" address for a user, or "" if unset.
func (s *Store) GetKindleEmail(ctx context.Context, userID int64) (string, error) {
	var email string
	err := s.queryRowContext(ctx, `SELECT email FROM user_kindle_email WHERE user_id=?`, userID).Scan(&email)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return email, err
}

// SaveKindleEmail sets or clears a user's Kindle "send to" address.
func (s *Store) SaveKindleEmail(ctx context.Context, userID int64, email string) error {
	_, err := s.execContext(ctx, `
INSERT INTO user_kindle_email (user_id, email, updated_at) VALUES (?,?,?)
ON CONFLICT(user_id) DO UPDATE SET email=excluded.email, updated_at=excluded.updated_at`,
		userID, email, time.Now().Unix())
	return err
}
