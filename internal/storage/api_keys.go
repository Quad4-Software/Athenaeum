package storage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"athenaeum/internal/auth"
	"athenaeum/internal/models"
)

// CreateAPIKey stores a new API key for userID and returns the record with the secret.
func (s *Store) CreateAPIKey(ctx context.Context, userID int64, name string) (models.APIKeyCreated, error) {
	full, prefix, hash, err := auth.NewAPIKey()
	if err != nil {
		return models.APIKeyCreated{}, err
	}
	now := time.Now().Unix()
	id, err := s.insertID(ctx, `
INSERT INTO api_keys (user_id, name, prefix, key_hash, created_at, last_used_at)
VALUES (?,?,?,?,?,0) RETURNING id`,
		userID, name, prefix, hash, now)
	if err != nil {
		return models.APIKeyCreated{}, err
	}
	return models.APIKeyCreated{
		APIKey: models.APIKey{
			ID:        id,
			UserID:    userID,
			Name:      name,
			Prefix:    prefix,
			CreatedAt: time.Unix(now, 0),
		},
		Key: full,
	}, nil
}

// ListAPIKeys returns API keys owned by userID.
func (s *Store) ListAPIKeys(ctx context.Context, userID int64) ([]models.APIKey, error) {
	rows, err := s.queryContext(ctx, `
SELECT id, user_id, name, prefix, created_at, last_used_at
FROM api_keys
WHERE user_id=?
ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAPIKeys(rows)
}

// DeleteAPIKey removes an API key owned by userID.
func (s *Store) DeleteAPIKey(ctx context.Context, userID, keyID int64) error {
	res, err := s.execContext(ctx, `DELETE FROM api_keys WHERE id=? AND user_id=?`, keyID, userID)
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

// DeleteAPIKeyAdmin removes any API key by id.
func (s *Store) DeleteAPIKeyAdmin(ctx context.Context, keyID int64) error {
	res, err := s.execContext(ctx, `DELETE FROM api_keys WHERE id=?`, keyID)
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

// UserFromAPIKey resolves a user from a full API key value.
func (s *Store) UserFromAPIKey(ctx context.Context, key string) (models.User, int64, error) {
	if len(key) < auth.APIKeyLookupLen {
		return models.User{}, 0, ErrNotFound
	}
	prefix := key[:auth.APIKeyLookupLen]
	var (
		id, userID, created, lastUsed int64
		name, storedHash              string
	)
	err := s.queryRowContext(ctx, `
SELECT id, user_id, name, key_hash, created_at, last_used_at
FROM api_keys
WHERE prefix=?`, prefix).
		Scan(&id, &userID, &name, &storedHash, &created, &lastUsed)
	if errors.Is(err, sql.ErrNoRows) {
		return models.User{}, 0, ErrNotFound
	}
	if err != nil {
		return models.User{}, 0, err
	}
	if !auth.CheckAPIKey(storedHash, key) {
		return models.User{}, 0, ErrNotFound
	}
	u, err := s.GetUser(ctx, userID)
	if err != nil {
		return models.User{}, 0, err
	}
	_, _ = s.execContext(ctx, `UPDATE api_keys SET last_used_at=? WHERE id=?`, time.Now().Unix(), id)
	return u, id, nil
}

func scanAPIKeys(rows *sql.Rows) ([]models.APIKey, error) {
	var out []models.APIKey
	for rows.Next() {
		var k models.APIKey
		var created, lastUsed int64
		if err := rows.Scan(&k.ID, &k.UserID, &k.Name, &k.Prefix, &created, &lastUsed); err != nil {
			return nil, err
		}
		k.CreatedAt = time.Unix(created, 0)
		if lastUsed > 0 {
			t := time.Unix(lastUsed, 0)
			k.LastUsedAt = &t
		}
		out = append(out, k)
	}
	return out, rows.Err()
}
