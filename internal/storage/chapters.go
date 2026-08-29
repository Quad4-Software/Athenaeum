package storage

import (
	"context"

	"athenaeum/internal/models"
)

// ReplaceChapters stores chapter markers for a book, replacing any prior rows.
func (s *Store) ReplaceChapters(ctx context.Context, bookID int64, chapters []models.Chapter) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `DELETE FROM book_chapters WHERE book_id=?`, bookID); err != nil {
		return err
	}
	for _, ch := range chapters {
		startMS := int64(ch.StartSec * 1000)
		if _, err := tx.ExecContext(ctx, `
INSERT INTO book_chapters (book_id, idx, title, start_ms) VALUES (?,?,?,?)`,
			bookID, ch.Index, ch.Title, startMS); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// ListChapters returns chapter markers for a book in display order.
func (s *Store) ListChapters(ctx context.Context, bookID int64) ([]models.Chapter, error) {
	rows, err := s.queryContext(ctx, `
SELECT idx, title, start_ms FROM book_chapters WHERE book_id=? ORDER BY idx ASC`, bookID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Chapter
	for rows.Next() {
		var ch models.Chapter
		var startMS int64
		if err := rows.Scan(&ch.Index, &ch.Title, &startMS); err != nil {
			return nil, err
		}
		ch.StartSec = float64(startMS) / 1000
		out = append(out, ch)
	}
	return out, rows.Err()
}
