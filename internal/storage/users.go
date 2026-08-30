package storage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"athenaeum/internal/models"
)

// AuthRequired reports whether at least one user account exists.
func (s *Store) AuthRequired(ctx context.Context) (bool, error) {
	gen := s.authRequiredGen.Load()
	if cached := s.authRequired.Load(); cached != nil && cached.gen == gen {
		return cached.required, nil
	}
	var n int
	err := s.queryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&n)
	if err != nil {
		return false, err
	}
	required := n > 0
	s.authRequired.Store(&authRequiredState{required: required, gen: gen})
	if s.authRequiredGen.Load() != gen {
		s.authRequired.Store(nil)
	}
	return required, nil
}

func (s *Store) invalidateAuthRequired() {
	s.authRequiredGen.Add(1)
	s.authRequired.Store(nil)
}

// CreateUser inserts a new account and returns its id.
func (s *Store) CreateUser(ctx context.Context, username, passwordHash string, admin bool) (int64, error) {
	perms := models.DefaultUserPermissions
	if admin {
		perms = models.AllPermissions
	}
	id, err := s.insertID(ctx, `
INSERT INTO users (username, password_hash, is_admin, permissions, created_at)
VALUES (?,?,?,?,?) RETURNING id`,
		username, passwordHash, boolToInt(admin), perms, time.Now().Unix())
	if err != nil {
		return 0, err
	}
	s.invalidateAuthRequired()
	return id, nil
}

// GetUserByUsername loads a user and password hash by username.
func (s *Store) GetUserByUsername(ctx context.Context, username string) (models.User, string, error) {
	row := s.queryRowContext(ctx,
		`SELECT `+userSelectCols+` FROM users WHERE LOWER(username)=LOWER(?)`, username)
	return scanUserRow(row)
}

// GetUser returns a user by id without the password hash.
func (s *Store) GetUser(ctx context.Context, id int64) (models.User, error) {
	u, _, err := scanUserRow(s.queryRowContext(ctx,
		`SELECT `+userSelectCols+` FROM users WHERE id=?`, id))
	return u, err
}

const userSelectCols = `id, username, email, password_hash, is_admin, permissions, is_guest, expires_at, created_at, totp_enabled`

func scanUserRow(row *sql.Row) (models.User, string, error) {
	var u models.User
	var hash string
	var admin, guest, totpEnabled int
	var created, expires int64
	err := row.Scan(&u.ID, &u.Username, &u.Email, &hash, &admin, &u.Permissions, &guest, &expires, &created, &totpEnabled)
	if errors.Is(err, sql.ErrNoRows) {
		return models.User{}, "", ErrNotFound
	}
	if err != nil {
		return models.User{}, "", err
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
	if u.IsGuest && expires > 0 && time.Now().Unix() > expires {
		return models.User{}, "", ErrNotFound
	}
	return u, hash, nil
}

// CreateSession stores a new session token without refresh metadata (legacy helper).
func (s *Store) CreateSession(ctx context.Context, token string, userID int64, expires time.Time) error {
	now := time.Now().Unix()
	_, err := s.execContext(ctx, `
INSERT INTO sessions (token, user_id, expires_at, created_at, last_seen_at, auth_method)
VALUES (?,?,?,?,?,?)`,
		token, userID, expires.Unix(), now, now, "local")
	return err
}

// DeleteSession removes a session token.
func (s *Store) DeleteSession(ctx context.Context, token string) error {
	_, err := s.execContext(ctx, `DELETE FROM sessions WHERE token=?`, token)
	return err
}

// SessionUser resolves a session token to a user, deleting expired sessions.
func (s *Store) SessionUser(ctx context.Context, token string) (models.User, error) {
	var userID int64
	var expires int64
	err := s.queryRowContext(ctx,
		`SELECT user_id, expires_at FROM sessions WHERE token=?`, token).
		Scan(&userID, &expires)
	if errors.Is(err, sql.ErrNoRows) {
		return models.User{}, ErrNotFound
	}
	if err != nil {
		return models.User{}, err
	}
	if time.Now().Unix() > expires {
		_ = s.DeleteSession(ctx, token)
		return models.User{}, ErrNotFound
	}
	return s.GetUser(ctx, userID)
}

// PurgeExpiredSessions deletes sessions past their expiry.
func (s *Store) PurgeExpiredSessions(ctx context.Context) error {
	now := time.Now().Unix()
	if _, err := s.execContext(ctx, `DELETE FROM sessions WHERE expires_at < ?`, now); err != nil {
		return err
	}
	_, err := s.execContext(ctx, `DELETE FROM refresh_tokens WHERE expires_at < ?`, now)
	return err
}

// CreateRefreshToken stores a new refresh token.
func (s *Store) CreateRefreshToken(ctx context.Context, token string, userID int64, expires time.Time) error {
	_, err := s.execContext(ctx,
		`INSERT INTO refresh_tokens (token, user_id, expires_at) VALUES (?,?,?)`,
		token, userID, expires.Unix())
	return err
}

// DeleteRefreshToken removes a refresh token.
func (s *Store) DeleteRefreshToken(ctx context.Context, token string) error {
	_, err := s.execContext(ctx, `DELETE FROM refresh_tokens WHERE token=?`, token)
	return err
}

// RefreshTokenUser resolves a refresh token to a user, deleting expired tokens.
func (s *Store) RefreshTokenUser(ctx context.Context, token string) (models.User, error) {
	var userID int64
	var expires int64
	err := s.queryRowContext(ctx,
		`SELECT user_id, expires_at FROM refresh_tokens WHERE token=?`, token).
		Scan(&userID, &expires)
	if errors.Is(err, sql.ErrNoRows) {
		return models.User{}, ErrNotFound
	}
	if err != nil {
		return models.User{}, err
	}
	if time.Now().Unix() > expires {
		_ = s.DeleteRefreshToken(ctx, token)
		return models.User{}, ErrNotFound
	}
	return s.GetUser(ctx, userID)
}

// ListUsers returns all accounts ordered by creation time.
func (s *Store) ListUsers(ctx context.Context) ([]models.User, error) {
	rows, err := s.queryContext(ctx,
		`SELECT `+userSelectCols+` FROM users ORDER BY created_at ASC, id ASC`)
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

// UpdateUsername changes a user's login name.
func (s *Store) UpdateUsername(ctx context.Context, userID int64, username string) error {
	res, err := s.execContext(ctx,
		`UPDATE users SET username=? WHERE id=?`, username, userID)
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

// UpdateUserPassword replaces the stored password hash.
func (s *Store) UpdateUserPassword(ctx context.Context, userID int64, passwordHash string) error {
	res, err := s.execContext(ctx,
		`UPDATE users SET password_hash=? WHERE id=?`, passwordHash, userID)
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

// RevokeUserSessions deletes all sessions and refresh tokens for a user.
func (s *Store) RevokeUserSessions(ctx context.Context, userID int64) error {
	if _, err := s.execContext(ctx, `DELETE FROM sessions WHERE user_id=?`, userID); err != nil {
		return err
	}
	_, err := s.execContext(ctx, `DELETE FROM refresh_tokens WHERE user_id=?`, userID)
	return err
}

// UsernameTaken reports whether another account already uses the name.
func (s *Store) UsernameTaken(ctx context.Context, username string, excludeUserID int64) (bool, error) {
	var id int64
	err := s.queryRowContext(ctx,
		`SELECT id FROM users WHERE LOWER(username)=LOWER(?) AND id<>?`, username, excludeUserID).
		Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// DeleteUser removes an account and related sessions.
func (s *Store) DeleteUser(ctx context.Context, userID int64) error {
	if err := s.RevokeUserSessions(ctx, userID); err != nil {
		return err
	}
	res, err := s.execContext(ctx, `DELETE FROM users WHERE id=?`, userID)
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
	s.invalidateAuthRequired()
	return nil
}

// SetUserAdmin toggles the admin flag for a user.
func (s *Store) SetUserAdmin(ctx context.Context, userID int64, admin bool) error {
	perms := models.DefaultUserPermissions
	if admin {
		perms = models.AllPermissions
	}
	res, err := s.execContext(ctx, `UPDATE users SET is_admin=?, permissions=? WHERE id=?`, boolToInt(admin), perms, userID)
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

// SetUserPermissions replaces the permission mask for a non-admin user.
func (s *Store) SetUserPermissions(ctx context.Context, userID int64, perms int64) error {
	res, err := s.execContext(ctx, `UPDATE users SET permissions=? WHERE id=? AND is_admin=0`, perms, userID)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		var admin int
		err := s.queryRowContext(ctx, `SELECT is_admin FROM users WHERE id=?`, userID).Scan(&admin)
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		if err != nil {
			return err
		}
		if admin != 0 {
			return nil
		}
		return ErrNotFound
	}
	return nil
}

// SetUserTOTPSecret stores a pending or active TOTP secret for a user without
// changing whether TOTP is enforced at login.
func (s *Store) SetUserTOTPSecret(ctx context.Context, userID int64, secret string) error {
	res, err := s.execContext(ctx, `UPDATE users SET totp_secret=? WHERE id=?`, secret, userID)
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

// GetUserTOTPSecret returns the stored TOTP secret for a user, which may be
// empty if TOTP has never been set up.
func (s *Store) GetUserTOTPSecret(ctx context.Context, userID int64) (string, error) {
	var secret sql.NullString
	err := s.queryRowContext(ctx, `SELECT totp_secret FROM users WHERE id=?`, userID).Scan(&secret)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", err
	}
	return secret.String, nil
}

// EnableUserTOTP marks two-factor authentication as active for a user.
func (s *Store) EnableUserTOTP(ctx context.Context, userID int64) error {
	res, err := s.execContext(ctx, `UPDATE users SET totp_enabled=1 WHERE id=?`, userID)
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

// DisableUserTOTP clears the stored secret and turns off two-factor authentication.
func (s *Store) DisableUserTOTP(ctx context.Context, userID int64) error {
	res, err := s.execContext(ctx, `UPDATE users SET totp_enabled=0, totp_secret='' WHERE id=?`, userID)
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

// AdminCount returns how many admin accounts exist.
func (s *Store) AdminCount(ctx context.Context) (int64, error) {
	var n int64
	err := s.queryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE is_admin=1`).Scan(&n)
	return n, err
}

// UserCount returns how many user accounts exist.
func (s *Store) UserCount(ctx context.Context) (int64, error) {
	var n int64
	err := s.queryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&n)
	return n, err
}
