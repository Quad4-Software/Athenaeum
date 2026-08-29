package storage

import (
	"context"
	"strings"
	"time"

	"athenaeum/internal/models"
)

func (s *Store) ListBookmarks(ctx context.Context, userID, bookID int64) ([]models.Bookmark, error) {
	rows, err := s.queryContext(ctx, `
SELECT id, user_id, book_id, location, label, created_at
FROM bookmarks WHERE user_id=? AND book_id=? ORDER BY created_at DESC`, userID, bookID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Bookmark
	for rows.Next() {
		var b models.Bookmark
		var created int64
		if err := rows.Scan(&b.ID, &b.UserID, &b.BookID, &b.Location, &b.Label, &created); err != nil {
			return nil, err
		}
		b.CreatedAt = time.Unix(created, 0)
		out = append(out, b)
	}
	if out == nil {
		out = []models.Bookmark{}
	}
	return out, rows.Err()
}

func (s *Store) CreateBookmark(ctx context.Context, userID int64, b models.Bookmark) (int64, error) {
	return s.insertID(ctx, `
INSERT INTO bookmarks (user_id, book_id, location, label, created_at)
VALUES (?,?,?,?,?) RETURNING id`,
		userID, b.BookID, b.Location, b.Label, time.Now().Unix())
}

func (s *Store) DeleteBookmark(ctx context.Context, userID, id int64) error {
	res, err := s.execContext(ctx, `DELETE FROM bookmarks WHERE id=? AND user_id=?`, id, userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) ListHighlights(ctx context.Context, userID, bookID int64) ([]models.Highlight, error) {
	rows, err := s.queryContext(ctx, `
SELECT id, user_id, book_id, location, excerpt, note, color, created_at
FROM highlights WHERE user_id=? AND book_id=? ORDER BY created_at ASC`, userID, bookID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Highlight
	for rows.Next() {
		var h models.Highlight
		var created int64
		if err := rows.Scan(&h.ID, &h.UserID, &h.BookID, &h.Location, &h.Excerpt, &h.Note, &h.Color, &created); err != nil {
			return nil, err
		}
		h.CreatedAt = time.Unix(created, 0)
		out = append(out, h)
	}
	if out == nil {
		out = []models.Highlight{}
	}
	return out, rows.Err()
}

func (s *Store) CreateHighlight(ctx context.Context, userID int64, h models.Highlight) (int64, error) {
	if h.Color == "" {
		h.Color = "yellow"
	}
	return s.insertID(ctx, `
INSERT INTO highlights (user_id, book_id, location, excerpt, note, color, created_at)
VALUES (?,?,?,?,?,?,?) RETURNING id`,
		userID, h.BookID, h.Location, h.Excerpt, h.Note, h.Color, time.Now().Unix())
}

func (s *Store) DeleteHighlight(ctx context.Context, userID, id int64) error {
	res, err := s.execContext(ctx, `DELETE FROM highlights WHERE id=? AND user_id=?`, id, userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) AddReadSeconds(ctx context.Context, userID, bookID int64, seconds int64) error {
	if seconds <= 0 {
		return nil
	}
	_, err := s.execContext(ctx, `
INSERT INTO progress (user_id, book_id, location, percent, read_seconds, updated_at)
VALUES (?,?,'',0,?,?)
ON CONFLICT(user_id, book_id) DO UPDATE SET
	read_seconds = progress.read_seconds + excluded.read_seconds,
	updated_at = excluded.updated_at`,
		userID, bookID, seconds, time.Now().Unix())
	return err
}

func (s *Store) ReadingStats(ctx context.Context, userID int64) (models.ReadingStats, error) {
	var st models.ReadingStats
	err := s.queryRowContext(ctx, `
SELECT COALESCE(SUM(read_seconds),0) FROM progress WHERE user_id=?`, userID).Scan(&st.TotalReadSeconds)
	if err != nil {
		return st, err
	}
	_ = s.queryRowContext(ctx, `
SELECT COUNT(*) FROM progress WHERE user_id=? AND percent > 0 AND percent < 0.98`, userID).Scan(&st.BooksInProgress)
	_ = s.queryRowContext(ctx, `
SELECT COUNT(*) FROM progress WHERE user_id=? AND percent >= 0.98`, userID).Scan(&st.BooksCompleted)
	st.CurrentStreak = s.readingStreakDays(ctx, userID)
	return st, nil
}

func (s *Store) readingStreakDays(ctx context.Context, userID int64) int64 {
	rows, err := s.queryContext(ctx, `
SELECT DISTINCT `+s.unixDateExpr("updated_at")+` AS d
FROM progress WHERE user_id=? AND updated_at > 0
ORDER BY d DESC LIMIT 400`, userID)
	if err != nil {
		return 0
	}
	defer rows.Close()
	var days []string
	for rows.Next() {
		var d string
		if err := rows.Scan(&d); err != nil {
			return 0
		}
		days = append(days, d)
	}
	if len(days) == 0 {
		return 0
	}
	streak := int64(1)
	today := time.Now().UTC().Format("2006-01-02")
	if days[0] != today {
		yesterday := time.Now().UTC().Add(-24 * time.Hour).Format("2006-01-02")
		if days[0] != yesterday {
			return 0
		}
	}
	for i := 1; i < len(days); i++ {
		prev, err1 := time.Parse("2006-01-02", days[i-1])
		cur, err2 := time.Parse("2006-01-02", days[i])
		if err1 != nil || err2 != nil {
			break
		}
		if prev.Sub(cur) == 24*time.Hour {
			streak++
			continue
		}
		break
	}
	return streak
}

func (s *Store) ProgressMap(ctx context.Context, userID int64, bookIDs []int64) (map[int64]models.Progress, error) {
	out := map[int64]models.Progress{}
	if len(bookIDs) == 0 {
		return out, nil
	}
	var query strings.Builder
	query.WriteString(`SELECT book_id, user_id, location, percent, COALESCE(read_seconds,0), updated_at FROM progress WHERE user_id=? AND book_id IN (`)
	args := []any{userID}
	for i, id := range bookIDs {
		if i > 0 {
			query.WriteString(",")
		}
		query.WriteString("?")
		args = append(args, id)
	}
	query.WriteString(")")
	rows, err := s.queryContext(ctx, query.String(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var p models.Progress
		var updated int64
		if err := rows.Scan(&p.BookID, &p.UserID, &p.Location, &p.Percent, &p.ReadSeconds, &updated); err != nil {
			return nil, err
		}
		p.UpdatedAt = time.Unix(updated, 0)
		out[p.BookID] = p
	}
	return out, rows.Err()
}

func (s *Store) PingDB(ctx context.Context) error {
	return s.queryRowContext(ctx, `SELECT 1`).Scan(new(int))
}
