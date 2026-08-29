package storage

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"athenaeum/internal/models"
)

const inviteColumns = `id, token, kind, email, permissions, created_by, expires_at, guest_expires_at,
	pocket_id_user_id, accepted_at, accepted_user_id, revoked_at, created_at`

// CreateInvite inserts a new invite token.
func (s *Store) CreateInvite(ctx context.Context, inv models.Invite) (models.Invite, error) {
	token := inv.Token
	if token == "" {
		var err error
		token, err = NewShareToken()
		if err != nil {
			return models.Invite{}, err
		}
	}
	now := time.Now()
	var expires, guestExpires sql.NullInt64
	if inv.ExpiresAt != nil {
		expires = sql.NullInt64{Int64: inv.ExpiresAt.Unix(), Valid: true}
	}
	if inv.GuestExpiresAt != nil {
		guestExpires = sql.NullInt64{Int64: inv.GuestExpiresAt.Unix(), Valid: true}
	}
	id, err := s.insertID(ctx, `
INSERT INTO invites (token, kind, email, permissions, created_by, expires_at, guest_expires_at,
	pocket_id_user_id, created_at)
VALUES (?,?,?,?,?,?,?,?,?) RETURNING id`,
		token, inv.Kind, strings.TrimSpace(inv.Email), inv.Permissions, inv.CreatedBy,
		expires, guestExpires, inv.PocketIDUserID, now.Unix())
	if err != nil {
		return models.Invite{}, err
	}
	inv.ID = id
	inv.Token = token
	inv.CreatedAt = now
	return inv, nil
}

// GetInviteByToken loads an invite by its public token.
func (s *Store) GetInviteByToken(ctx context.Context, token string) (models.Invite, error) {
	row := s.queryRowContext(ctx, `SELECT `+inviteColumns+` FROM invites WHERE token=?`, token)
	return scanInvite(row)
}

// GetInvite loads an invite by id.
func (s *Store) GetInvite(ctx context.Context, id int64) (models.Invite, error) {
	row := s.queryRowContext(ctx, `SELECT `+inviteColumns+` FROM invites WHERE id=?`, id)
	return scanInvite(row)
}

// ListInvites returns invites newest first, optionally filtered by status.
func (s *Store) ListInvites(ctx context.Context, status string) ([]models.Invite, error) {
	rows, err := s.queryContext(ctx, `SELECT `+inviteColumns+` FROM invites ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Invite
	now := time.Now()
	for rows.Next() {
		inv, err := scanInvite(rows)
		if err != nil {
			return nil, err
		}
		if status != "" && inviteStatus(inv, now) != status {
			continue
		}
		out = append(out, inv)
	}
	return out, rows.Err()
}

// SetInvitePocketID stores the Pocket ID user id on an invite.
func (s *Store) SetInvitePocketID(ctx context.Context, inviteID int64, pocketUserID string) error {
	_, err := s.execContext(ctx, `UPDATE invites SET pocket_id_user_id=? WHERE id=?`, pocketUserID, inviteID)
	return err
}

// AcceptInvite marks an invite accepted for the given user.
func (s *Store) AcceptInvite(ctx context.Context, inviteID, userID int64) error {
	now := time.Now().Unix()
	res, err := s.execContext(ctx, `
UPDATE invites SET accepted_at=?, accepted_user_id=?
WHERE id=? AND accepted_at IS NULL AND revoked_at IS NULL`,
		now, userID, inviteID)
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

// AcceptInviteSSO marks a Pocket ID invite accepted without a local user id yet.
func (s *Store) AcceptInviteSSO(ctx context.Context, inviteID int64) error {
	now := time.Now().Unix()
	res, err := s.execContext(ctx, `
UPDATE invites SET accepted_at=?
WHERE id=? AND accepted_at IS NULL AND revoked_at IS NULL`,
		now, inviteID)
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

// RevokeInvite sets revoked_at on a pending invite.
func (s *Store) RevokeInvite(ctx context.Context, id int64) error {
	now := time.Now().Unix()
	res, err := s.execContext(ctx, `
UPDATE invites SET revoked_at=? WHERE id=? AND accepted_at IS NULL AND revoked_at IS NULL`,
		now, id)
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

// CreateInvitedUser creates a permanent local user with email and permissions.
func (s *Store) CreateInvitedUser(ctx context.Context, username, passwordHash, email string, perms int64) (int64, error) {
	if perms == 0 {
		perms = models.DefaultUserPermissions
	}
	return s.insertID(ctx, `
INSERT INTO users (username, password_hash, is_admin, permissions, email, created_at)
VALUES (?,?,0,?,?,?) RETURNING id`,
		username, passwordHash, perms, strings.TrimSpace(email), time.Now().Unix())
}

func inviteStatus(inv models.Invite, now time.Time) string {
	if inv.RevokedAt != nil {
		return "revoked"
	}
	if inv.AcceptedAt != nil {
		return "accepted"
	}
	if inv.ExpiresAt != nil && now.After(*inv.ExpiresAt) {
		return "expired"
	}
	return "pending"
}

func scanInvite(row scanner) (models.Invite, error) {
	var inv models.Invite
	var expires, guestExpires, accepted, acceptedUser, revoked sql.NullInt64
	var created int64
	err := row.Scan(
		&inv.ID, &inv.Token, &inv.Kind, &inv.Email, &inv.Permissions, &inv.CreatedBy,
		&expires, &guestExpires, &inv.PocketIDUserID, &accepted, &acceptedUser, &revoked, &created,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return models.Invite{}, ErrNotFound
	}
	if err != nil {
		return models.Invite{}, err
	}
	inv.CreatedAt = time.Unix(created, 0)
	if expires.Valid {
		t := time.Unix(expires.Int64, 0)
		inv.ExpiresAt = &t
	}
	if guestExpires.Valid {
		t := time.Unix(guestExpires.Int64, 0)
		inv.GuestExpiresAt = &t
	}
	if accepted.Valid {
		t := time.Unix(accepted.Int64, 0)
		inv.AcceptedAt = &t
	}
	if acceptedUser.Valid {
		id := acceptedUser.Int64
		inv.AcceptedUserID = &id
	}
	if revoked.Valid {
		t := time.Unix(revoked.Int64, 0)
		inv.RevokedAt = &t
	}
	return inv, nil
}
