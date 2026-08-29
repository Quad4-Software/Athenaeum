package storage

import (
	"context"
	"time"
)

// ListOfflineGrants returns the book ids a user has approved for offline access.
func (s *Store) ListOfflineGrants(ctx context.Context, userID int64) ([]int64, error) {
	rows, err := s.queryContext(ctx,
		`SELECT book_id FROM offline_grants WHERE user_id=? ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// AddOfflineGrants approves the given books for offline access by a user.
func (s *Store) AddOfflineGrants(ctx context.Context, userID int64, bookIDs []int64) error {
	now := time.Now().Unix()
	for _, id := range bookIDs {
		if _, err := s.execContext(ctx, `
INSERT INTO offline_grants (user_id, book_id, created_at) VALUES (?,?,?)
ON CONFLICT(user_id, book_id) DO NOTHING`, userID, id, now); err != nil {
			return err
		}
	}
	return nil
}

// RemoveOfflineGrants revokes offline access for the given books.
func (s *Store) RemoveOfflineGrants(ctx context.Context, userID int64, bookIDs []int64) error {
	for _, id := range bookIDs {
		if _, err := s.execContext(ctx,
			`DELETE FROM offline_grants WHERE user_id=? AND book_id=?`, userID, id); err != nil {
			return err
		}
	}
	return nil
}
