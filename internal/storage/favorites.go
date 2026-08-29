package storage

import (
	"context"
	"time"
)

// ListFavoriteIDs returns book ids favorited by userID.
func (s *Store) ListFavoriteIDs(ctx context.Context, userID int64) ([]int64, error) {
	rows, err := s.queryContext(ctx, `
SELECT book_id FROM user_favorites WHERE user_id=? ORDER BY created_at DESC`, userID)
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

// IsFavorite reports whether userID favorited bookID.
func (s *Store) IsFavorite(ctx context.Context, userID, bookID int64) (bool, error) {
	var n int
	err := s.queryRowContext(ctx,
		`SELECT COUNT(*) FROM user_favorites WHERE user_id=? AND book_id=?`, userID, bookID).
		Scan(&n)
	return n > 0, err
}

// SetFavorite adds or removes a favorite for userID.
func (s *Store) SetFavorite(ctx context.Context, userID, bookID int64, favorite bool) error {
	if _, err := s.GetBook(ctx, bookID); err != nil {
		return err
	}
	if favorite {
		_, err := s.execContext(ctx, `
INSERT INTO user_favorites (user_id, book_id, created_at) VALUES (?,?,?)
ON CONFLICT(user_id, book_id) DO NOTHING`,
			userID, bookID, time.Now().Unix())
		return err
	}
	_, err := s.execContext(ctx,
		`DELETE FROM user_favorites WHERE user_id=? AND book_id=?`, userID, bookID)
	return err
}

// CountFavorites returns how many favorites userID has.
func (s *Store) CountFavorites(ctx context.Context, userID int64) (int64, error) {
	var n int64
	err := s.queryRowContext(ctx,
		`SELECT COUNT(*) FROM user_favorites WHERE user_id=?`, userID).Scan(&n)
	return n, err
}
