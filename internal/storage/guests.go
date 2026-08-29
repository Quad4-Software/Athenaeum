package storage

import (
	"context"
	"time"

	"athenaeum/internal/models"
)

// ListGuestUsers returns guest accounts, optionally filtered to those expiring soon.
func (s *Store) ListGuestUsers(ctx context.Context, expiringWithinHours int) ([]models.User, error) {
	q := `SELECT ` + userSelectCols + ` FROM users WHERE is_guest=1`
	var args []any
	if expiringWithinHours > 0 {
		deadline := time.Now().Add(time.Duration(expiringWithinHours) * time.Hour).Unix()
		now := time.Now().Unix()
		q += ` AND expires_at > ? AND expires_at <= ?`
		args = append(args, now, deadline)
	}
	q += ` ORDER BY expires_at ASC, LOWER(username) ASC`
	rows, err := s.queryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.User
	for rows.Next() {
		var u models.User
		var hash string
		var admin, guest, totpEnabled int
		var created, expires int64
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &hash, &admin, &u.Permissions, &guest, &expires, &created, &totpEnabled); err != nil {
			return nil, err
		}
		u.IsAdmin = admin != 0
		u.IsGuest = guest != 0
		u.LocalAuth = hash != ""
		u.TOTPEnabled = totpEnabled != 0
		u.CreatedAt = time.Unix(created, 0)
		if expires > 0 {
			t := time.Unix(expires, 0)
			u.ExpiresAt = &t
		}
		out = append(out, u)
	}
	if out == nil {
		out = []models.User{}
	}
	return out, rows.Err()
}

// ExtendGuestExpiry updates a guest account expiry.
func (s *Store) ExtendGuestExpiry(ctx context.Context, userID int64, expiresAt time.Time) error {
	res, err := s.execContext(ctx, `
UPDATE users SET expires_at=? WHERE id=? AND is_guest=1`, expiresAt.Unix(), userID)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// DeleteGuestUsers removes multiple guest accounts.
func (s *Store) DeleteGuestUsers(ctx context.Context, ids []int64) (int64, error) {
	var n int64
	for _, id := range ids {
		u, err := s.GetUser(ctx, id)
		if err != nil {
			continue
		}
		if !u.IsGuest {
			continue
		}
		if err := s.DeleteUser(ctx, id); err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}
