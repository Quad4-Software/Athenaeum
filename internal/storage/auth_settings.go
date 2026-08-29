package storage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"athenaeum/internal/models"
)

// GetAuthSettings loads server-wide authentication policy toggles.
func (s *Store) GetAuthSettings(ctx context.Context) (models.AuthSettings, error) {
	var allowReg, requireTOTP int
	err := s.queryRowContext(ctx,
		`SELECT allow_registration, require_totp FROM auth_settings WHERE id=1`).
		Scan(&allowReg, &requireTOTP)
	if errors.Is(err, sql.ErrNoRows) {
		return models.AuthSettings{}, nil
	}
	if err != nil {
		return models.AuthSettings{}, err
	}
	return models.AuthSettings{
		AllowRegistration: allowReg != 0,
		RequireTOTP:       requireTOTP != 0,
	}, nil
}

// SaveAuthSettings persists server-wide authentication policy toggles.
func (s *Store) SaveAuthSettings(ctx context.Context, settings models.AuthSettings) error {
	_, err := s.execContext(ctx, `
INSERT INTO auth_settings (id, allow_registration, require_totp, updated_at)
VALUES (1, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET allow_registration=excluded.allow_registration,
	require_totp=excluded.require_totp, updated_at=excluded.updated_at`,
		boolToInt(settings.AllowRegistration), boolToInt(settings.RequireTOTP), time.Now().Unix())
	return err
}
