package storage

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"athenaeum/internal/models"
)

const feedTokenColumns = `SELECT id, token, user_id, name, library_id, collection_id, created_at FROM feed_tokens`

// CreateFeedToken inserts a token-scoped audio feed for userID and returns
// the stored row. A libraryID or collectionID of 0 leaves the feed unscoped
// on that axis.
func (s *Store) CreateFeedToken(ctx context.Context, userID int64, name string, libraryID, collectionID int64) (models.FeedToken, error) {
	token, err := NewShareToken()
	if err != nil {
		return models.FeedToken{}, err
	}
	now := time.Now()
	id, err := s.insertID(ctx, `
INSERT INTO feed_tokens (token, user_id, name, library_id, collection_id, created_at)
VALUES (?,?,?,?,?,?) RETURNING id`,
		token, userID, name, libraryID, collectionID, now.Unix())
	if err != nil {
		return models.FeedToken{}, err
	}
	return models.FeedToken{
		ID:           id,
		Token:        token,
		UserID:       userID,
		Name:         name,
		LibraryID:    libraryID,
		CollectionID: collectionID,
		CreatedAt:    now,
	}, nil
}

func scanFeedToken(row scanner) (models.FeedToken, error) {
	var ft models.FeedToken
	var created int64
	err := row.Scan(&ft.ID, &ft.Token, &ft.UserID, &ft.Name, &ft.LibraryID, &ft.CollectionID, &created)
	if err != nil {
		return models.FeedToken{}, err
	}
	ft.CreatedAt = time.Unix(created, 0)
	return ft, nil
}

// GetFeedTokenByToken looks up a feed by its public token.
func (s *Store) GetFeedTokenByToken(ctx context.Context, token string) (models.FeedToken, error) {
	row := s.queryRowContext(ctx, feedTokenColumns+` WHERE token=?`, token)
	ft, err := scanFeedToken(row)
	if errors.Is(err, sql.ErrNoRows) {
		return models.FeedToken{}, ErrNotFound
	}
	return ft, err
}

// ListFeedTokens returns the feeds owned by userID, newest first.
func (s *Store) ListFeedTokens(ctx context.Context, userID int64) ([]models.FeedToken, error) {
	rows, err := s.queryContext(ctx, feedTokenColumns+` WHERE user_id=? ORDER BY created_at DESC, id DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.FeedToken
	for rows.Next() {
		ft, err := scanFeedToken(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, ft)
	}
	return out, rows.Err()
}

// DeleteFeedToken removes a feed scoped to its owner.
func (s *Store) DeleteFeedToken(ctx context.Context, userID, id int64) error {
	res, err := s.execContext(ctx, `DELETE FROM feed_tokens WHERE id=? AND user_id=?`, id, userID)
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

// CollectionContainsBook reports whether bookID is inside the collection
// owned by userID. Manual and reading shelves check collection_items; smart
// and auto shelves evaluate their stored query against the book row.
func (s *Store) CollectionContainsBook(ctx context.Context, userID, collectionID, bookID int64) (bool, error) {
	c, err := s.GetCollection(ctx, userID, collectionID)
	if errors.Is(err, ErrNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	smart := c.Kind == models.CollectionSmart || c.Kind == models.CollectionAuto
	if smart && c.Query != nil {
		q := ApplySmartQuery(models.BookQuery{}, *c.Query)
		ftsIDs, ftsErr := s.searchFTS(ctx, q.Search)
		useFTS := q.Search != "" && ftsErr == nil && len(ftsIDs) > 0
		where, args := bookWhereClause(q, ftsIDs, useFTS, nil)
		where = append(where, "books.id = ?")
		args = append(args, bookID)
		var n int
		if err := s.queryRowContext(ctx,
			"SELECT COUNT(*) FROM books WHERE "+strings.Join(where, " AND "), args...).Scan(&n); err != nil {
			return false, err
		}
		return n > 0, nil
	}
	var one int
	err = s.queryRowContext(ctx,
		`SELECT 1 FROM collection_items WHERE collection_id=? AND book_id=?`, collectionID, bookID).Scan(&one)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}
