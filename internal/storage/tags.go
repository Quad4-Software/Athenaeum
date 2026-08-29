package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"athenaeum/internal/models"
)

// ListTags returns all tags ordered by name.
func (s *Store) ListTags(ctx context.Context) ([]models.Tag, error) {
	rows, err := s.queryContext(ctx, `SELECT id, name FROM tags ORDER BY LOWER(name)`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Tag
	for rows.Next() {
		var t models.Tag
		if err := rows.Scan(&t.ID, &t.Name); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// CreateTag inserts a tag by name, or returns the existing one.
func (s *Store) CreateTag(ctx context.Context, name string) (models.Tag, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return models.Tag{}, errors.New("tag name required")
	}
	var t models.Tag
	err := s.queryRowContext(ctx, `SELECT id, name FROM tags WHERE LOWER(name) = LOWER(?)`, name).
		Scan(&t.ID, &t.Name)
	if err == nil {
		return t, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return models.Tag{}, err
	}
	id, err := s.insertID(ctx, `INSERT INTO tags (name) VALUES (?) RETURNING id`, name)
	if err != nil {
		return models.Tag{}, err
	}
	return models.Tag{ID: id, Name: name}, nil
}

// ListBookTags returns tag names attached to a book.
func (s *Store) ListBookTags(ctx context.Context, bookID int64) ([]string, error) {
	rows, err := s.queryContext(ctx, `
SELECT t.name FROM tags t
JOIN book_tags bt ON bt.tag_id = t.id
WHERE bt.book_id = ?
ORDER BY LOWER(t.name)`, bookID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		out = append(out, name)
	}
	return out, rows.Err()
}

// ListBookTagsBatch returns tag names keyed by book ID.
func (s *Store) ListBookTagsBatch(ctx context.Context, bookIDs []int64) (map[int64][]string, error) {
	out := make(map[int64][]string, len(bookIDs))
	if len(bookIDs) == 0 {
		return out, nil
	}
	placeholders := make([]string, len(bookIDs))
	args := make([]any, len(bookIDs))
	for i, id := range bookIDs {
		placeholders[i] = "?"
		args[i] = id
		out[id] = []string{}
	}
	q := fmt.Sprintf(`
SELECT bt.book_id, t.name FROM book_tags bt
JOIN tags t ON t.id = bt.tag_id
WHERE bt.book_id IN (%s)
ORDER BY LOWER(t.name)`, strings.Join(placeholders, ",")) // #nosec G201 -- placeholders are only "?" repeats
	rows, err := s.queryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var bookID int64
		var name string
		if err := rows.Scan(&bookID, &name); err != nil {
			return nil, err
		}
		out[bookID] = append(out[bookID], name)
	}
	return out, rows.Err()
}

// SetBookTags replaces all tags on a book with the given names.
func (s *Store) SetBookTags(ctx context.Context, bookID int64, names []string) ([]string, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, s.rebind(`DELETE FROM book_tags WHERE book_id = ?`), bookID); err != nil {
		return nil, err
	}
	seen := map[string]struct{}{}
	var out []string
	for _, raw := range names {
		name := strings.TrimSpace(raw)
		if name == "" {
			continue
		}
		key := strings.ToLower(name)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		var tagID int64
		var storedName string
		err := tx.QueryRowContext(ctx, s.rebind(`SELECT id, name FROM tags WHERE LOWER(name) = LOWER(?)`), name).
			Scan(&tagID, &storedName)
		if errors.Is(err, sql.ErrNoRows) {
			ierr := tx.QueryRowContext(ctx, s.rebind(`INSERT INTO tags (name) VALUES (?) RETURNING id`), name).Scan(&tagID)
			if ierr != nil {
				return nil, ierr
			}
			storedName = name
		} else if err != nil {
			return nil, err
		}
		if _, err := tx.ExecContext(ctx, s.rebind(`INSERT INTO book_tags (book_id, tag_id) VALUES (?,?)`), bookID, tagID); err != nil {
			return nil, err
		}
		out = append(out, storedName)
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return out, nil
}

// AddBookTag attaches a tag by name and returns the updated name list.
func (s *Store) AddBookTag(ctx context.Context, bookID int64, name string) ([]string, error) {
	tag, err := s.CreateTag(ctx, name)
	if err != nil {
		return nil, err
	}
	_, err = s.execContext(ctx, `
INSERT INTO book_tags (book_id, tag_id) VALUES (?,?)
ON CONFLICT (book_id, tag_id) DO NOTHING`, bookID, tag.ID)
	if err != nil {
		return nil, err
	}
	return s.ListBookTags(ctx, bookID)
}

// RemoveBookTag detaches a tag from a book by tag ID.
func (s *Store) RemoveBookTag(ctx context.Context, bookID, tagID int64) error {
	res, err := s.execContext(ctx, `DELETE FROM book_tags WHERE book_id=? AND tag_id=?`, bookID, tagID)
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

// GetRating returns the user's rating for a book.
func (s *Store) GetRating(ctx context.Context, userID, bookID int64) (models.BookRating, error) {
	var r models.BookRating
	err := s.queryRowContext(ctx, `
SELECT user_id, book_id, rating, updated_at FROM book_ratings WHERE user_id=? AND book_id=?`,
		userID, bookID).Scan(&r.UserID, &r.BookID, &r.Rating, &r.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return models.BookRating{UserID: userID, BookID: bookID}, nil
	}
	return r, err
}

// RatingsBatch returns ratings for many books for one user.
func (s *Store) RatingsBatch(ctx context.Context, userID int64, bookIDs []int64) (map[int64]int, error) {
	out := make(map[int64]int, len(bookIDs))
	if len(bookIDs) == 0 || userID <= 0 {
		return out, nil
	}
	placeholders := make([]string, len(bookIDs))
	args := make([]any, 0, len(bookIDs)+1)
	args = append(args, userID)
	for i, id := range bookIDs {
		placeholders[i] = "?"
		args = append(args, id)
	}
	q := fmt.Sprintf(`SELECT book_id, rating FROM book_ratings WHERE user_id=? AND book_id IN (%s)`,
		strings.Join(placeholders, ",")) // #nosec G201 -- placeholders are only "?" repeats
	rows, err := s.queryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var bookID int64
		var rating int
		if err := rows.Scan(&bookID, &rating); err != nil {
			return nil, err
		}
		out[bookID] = rating
	}
	return out, rows.Err()
}

// SetRating stores a 1-5 rating. Rating 0 clears it.
func (s *Store) SetRating(ctx context.Context, userID, bookID int64, rating int) (models.BookRating, error) {
	if rating < 0 || rating > 5 {
		return models.BookRating{}, errors.New("rating must be between 0 and 5")
	}
	if rating == 0 {
		_, err := s.execContext(ctx, `DELETE FROM book_ratings WHERE user_id=? AND book_id=?`, userID, bookID)
		return models.BookRating{UserID: userID, BookID: bookID}, err
	}
	now := time.Now().Unix()
	_, err := s.execContext(ctx, `
INSERT INTO book_ratings (user_id, book_id, rating, updated_at) VALUES (?,?,?,?)
ON CONFLICT(user_id, book_id) DO UPDATE SET rating=excluded.rating, updated_at=excluded.updated_at`,
		userID, bookID, rating, now)
	if err != nil {
		return models.BookRating{}, err
	}
	return models.BookRating{UserID: userID, BookID: bookID, Rating: rating, UpdatedAt: now}, nil
}
