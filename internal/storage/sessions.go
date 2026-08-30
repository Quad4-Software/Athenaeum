package storage

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"athenaeum/internal/models"
)

// CreateUserSession stores linked access and refresh tokens with client metadata.
func (s *Store) CreateUserSession(ctx context.Context, sess models.SessionCreate) error {
	now := time.Now().Unix()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `
INSERT INTO sessions (token, user_id, expires_at, session_id, ip, user_agent, device, auth_method, created_at, last_seen_at)
VALUES (?,?,?,?,?,?,?,?,?,?)`,
		sess.AccessToken, sess.UserID, sess.AccessExpires.Unix(), sess.SessionID,
		sess.IP, sess.UserAgent, sess.Device, sess.AuthMethod, now, now); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO refresh_tokens (token, user_id, expires_at, session_id)
VALUES (?,?,?,?)`,
		sess.RefreshToken, sess.UserID, sess.RefreshExpires.Unix(), sess.SessionID); err != nil {
		return err
	}
	return tx.Commit()
}

// ListUserSessions returns active sessions for a user, newest first.
func (s *Store) ListUserSessions(ctx context.Context, userID int64, currentAccessToken string) ([]models.UserSession, error) {
	rows, err := s.queryContext(ctx, `
SELECT s.session_id, s.user_id, s.ip, s.user_agent, s.device, s.auth_method,
       s.created_at, s.last_seen_at, r.expires_at, s.token
FROM sessions s
JOIN refresh_tokens r ON r.session_id = s.session_id AND r.user_id = s.user_id
WHERE s.user_id = ? AND s.session_id != '' AND r.expires_at > ?
ORDER BY s.last_seen_at DESC, s.created_at DESC`,
		userID, time.Now().Unix())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.UserSession
	for rows.Next() {
		var sess models.UserSession
		var created, lastSeen, refreshExp int64
		var accessToken string
		if err := rows.Scan(&sess.ID, &sess.UserID, &sess.IP, &sess.UserAgent, &sess.Device,
			&sess.AuthMethod, &created, &lastSeen, &refreshExp, &accessToken); err != nil {
			return nil, err
		}
		sess.CreatedAt = time.Unix(created, 0)
		sess.LastSeenAt = time.Unix(lastSeen, 0)
		sess.ExpiresAt = time.Unix(refreshExp, 0)
		sess.Current = accessToken == currentAccessToken
		out = append(out, sess)
	}
	if out == nil {
		out = []models.UserSession{}
	}
	return out, rows.Err()
}

// RevokeSessionByID deletes one session bundle by its public id.
func (s *Store) RevokeSessionByID(ctx context.Context, userID int64, sessionID string) error {
	if sessionID == "" {
		return ErrNotFound
	}
	if _, err := s.execContext(ctx, `DELETE FROM sessions WHERE session_id=? AND user_id=?`, sessionID, userID); err != nil {
		return err
	}
	res, err := s.execContext(ctx, `DELETE FROM refresh_tokens WHERE session_id=? AND user_id=?`, sessionID, userID)
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

// RevokeOtherSessions deletes every session except the one tied to currentAccessToken.
func (s *Store) RevokeOtherSessions(ctx context.Context, userID int64, currentAccessToken string) (int64, error) {
	var currentSessionID string
	err := s.queryRowContext(ctx,
		`SELECT session_id FROM sessions WHERE token=? AND user_id=?`, currentAccessToken, userID).
		Scan(&currentSessionID)
	if errors.Is(err, sql.ErrNoRows) {
		currentSessionID = ""
	} else if err != nil {
		return 0, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()

	var res sql.Result
	if currentSessionID != "" {
		res, err = tx.ExecContext(ctx, `DELETE FROM sessions WHERE user_id=? AND session_id<>?`, userID, currentSessionID)
	} else {
		res, err = tx.ExecContext(ctx, `DELETE FROM sessions WHERE user_id=?`, userID)
	}
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()

	if currentSessionID != "" {
		_, err = tx.ExecContext(ctx, `DELETE FROM refresh_tokens WHERE user_id=? AND session_id<>?`, userID, currentSessionID)
	} else {
		_, err = tx.ExecContext(ctx, `DELETE FROM refresh_tokens WHERE user_id=?`, userID)
	}
	if err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return n, nil
}

// TouchSession updates last_seen_at for the access token.
func (s *Store) TouchSession(ctx context.Context, accessToken string) error {
	_, err := s.execContext(ctx,
		`UPDATE sessions SET last_seen_at=? WHERE token=?`, time.Now().Unix(), accessToken)
	return err
}

// SessionIDForToken returns the public session id for an access token.
func (s *Store) SessionIDForToken(ctx context.Context, accessToken string) (string, error) {
	var id string
	err := s.queryRowContext(ctx, `SELECT session_id FROM sessions WHERE token=?`, accessToken).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	return id, err
}

// DeleteSessionByAccessToken removes access and refresh rows for one token.
func (s *Store) DeleteSessionByAccessToken(ctx context.Context, accessToken string) error {
	sessionID, err := s.SessionIDForToken(ctx, accessToken)
	if errors.Is(err, ErrNotFound) {
		_, err = s.execContext(ctx, `DELETE FROM sessions WHERE token=?`, accessToken)
		return err
	}
	if err != nil {
		return err
	}
	if sessionID != "" {
		if _, err := s.execContext(ctx, `DELETE FROM refresh_tokens WHERE session_id=?`, sessionID); err != nil {
			return err
		}
	}
	_, err = s.execContext(ctx, `DELETE FROM sessions WHERE token=?`, accessToken)
	return err
}

// DeleteRefreshByToken removes a refresh token and its linked access session.
func (s *Store) DeleteRefreshByToken(ctx context.Context, refreshToken string) error {
	var sessionID string
	err := s.queryRowContext(ctx,
		`SELECT session_id FROM refresh_tokens WHERE token=?`, refreshToken).Scan(&sessionID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	if _, err := s.execContext(ctx, `DELETE FROM refresh_tokens WHERE token=?`, refreshToken); err != nil {
		return err
	}
	if sessionID != "" {
		_, err = s.execContext(ctx, `DELETE FROM sessions WHERE session_id=?`, sessionID)
	}
	return err
}

// RefreshTokenSessionID returns the session id linked to a refresh token.
func (s *Store) RefreshTokenSessionID(ctx context.Context, refreshToken string) (string, error) {
	var sessionID string
	err := s.queryRowContext(ctx,
		`SELECT session_id FROM refresh_tokens WHERE token=?`, refreshToken).Scan(&sessionID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	return sessionID, err
}

// RotateSessionTokens replaces access and refresh tokens for one session id.
func (s *Store) RotateSessionTokens(ctx context.Context, sessionID, oldRefresh, newAccess, newRefresh string, accessExp, refreshExp time.Time) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var userID int64
	var ip, ua, device, authMethod string
	var created int64
	err = tx.QueryRowContext(ctx, `
SELECT user_id, ip, user_agent, device, auth_method, created_at
FROM sessions WHERE session_id=? LIMIT 1`, sessionID).Scan(&userID, &ip, &ua, &device, &authMethod, &created)
	if errors.Is(err, sql.ErrNoRows) {
		err = tx.QueryRowContext(ctx, `
SELECT user_id FROM refresh_tokens WHERE session_id=? LIMIT 1`, sessionID).Scan(&userID)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM sessions WHERE session_id=?`, sessionID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM refresh_tokens WHERE token=?`, oldRefresh); err != nil {
		return err
	}

	now := time.Now().Unix()
	if created == 0 {
		created = now
	}

	if _, err := tx.ExecContext(ctx, `
INSERT INTO sessions (token, user_id, expires_at, session_id, ip, user_agent, device, auth_method, created_at, last_seen_at)
VALUES (?,?,?,?,?,?,?,?,?,?)`,
		newAccess, userID, accessExp.Unix(), sessionID, ip, ua, device, authMethod, created, now); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO refresh_tokens (token, user_id, expires_at, session_id)
VALUES (?,?,?,?)`,
		newRefresh, userID, refreshExp.Unix(), sessionID); err != nil {
		return err
	}
	return tx.Commit()
}

// FindUserByOIDCSub loads a user by OIDC subject.
func (s *Store) FindUserByOIDCSub(ctx context.Context, sub string) (models.User, error) {
	u, _, err := scanUserRow(s.queryRowContext(ctx,
		`SELECT `+userSelectCols+` FROM users WHERE oidc_sub=?`, sub))
	return u, err
}

// FindUserByEmail loads a user by email address.
func (s *Store) FindUserByEmail(ctx context.Context, email string) (models.User, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" {
		return models.User{}, ErrNotFound
	}
	u, _, err := scanUserRow(s.queryRowContext(ctx,
		`SELECT `+userSelectCols+` FROM users WHERE lower(email)=?`, email))
	return u, err
}

// CreateOIDCUser inserts an account provisioned via OpenID Connect.
func (s *Store) CreateOIDCUser(ctx context.Context, username, email, oidcSub string, admin bool) (int64, error) {
	return s.insertID(ctx, `
INSERT INTO users (username, password_hash, is_admin, created_at, email, oidc_sub)
VALUES (?,?,?,?,?,?) RETURNING id`,
		username, "", boolToInt(admin), time.Now().Unix(), email, oidcSub)
}

// LinkOIDCSub associates an OIDC subject with an existing account.
// Refuses to overwrite a different existing oidc_sub.
func (s *Store) LinkOIDCSub(ctx context.Context, userID int64, oidcSub, email string) error {
	res, err := s.execContext(ctx, `
UPDATE users SET oidc_sub=?, email=CASE WHEN email='' THEN ? ELSE email END
WHERE id=? AND (oidc_sub='' OR oidc_sub=?)`,
		oidcSub, email, userID, oidcSub)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrConflict
	}
	return nil
}
