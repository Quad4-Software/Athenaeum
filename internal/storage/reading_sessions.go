package storage

import (
	"context"
	"time"

	"athenaeum/internal/models"
)

// sessionMergeWindow is how far back the latest session's end may lie for a
// new heartbeat to extend it instead of opening a new session.
const sessionMergeWindow = 30 * time.Minute

const (
	defaultSessionListLimit = 20
	maxSessionListLimit     = 100
)

// recordSession folds seconds into the latest session for the user+book pair
// when it ended within sessionMergeWindow of now, otherwise inserts a new
// session row spanning only this heartbeat.
func (s *Store) recordSession(ctx context.Context, userID, bookID, seconds, now int64) error {
	if seconds <= 0 {
		return nil
	}
	res, err := s.execContext(ctx, `
UPDATE reading_sessions SET ended_at=?, seconds=seconds+?
WHERE id = (
	SELECT id FROM reading_sessions
	WHERE user_id=? AND book_id=?
	ORDER BY ended_at DESC, id DESC LIMIT 1
) AND ended_at >= ?`,
		now, seconds, userID, bookID, now-int64(sessionMergeWindow/time.Second))
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	_, err = s.execContext(ctx, `
INSERT INTO reading_sessions (user_id, book_id, started_at, ended_at, seconds)
VALUES (?,?,?,?,?)`, userID, bookID, now, now, seconds)
	return err
}

// ListReadingSessions returns a user's sessions newest first, joined with the
// book title. limit is clamped to [1, maxSessionListLimit].
func (s *Store) ListReadingSessions(ctx context.Context, userID int64, limit int) ([]models.ReadingSession, error) {
	if limit <= 0 {
		limit = defaultSessionListLimit
	}
	if limit > maxSessionListLimit {
		limit = maxSessionListLimit
	}
	rows, err := s.queryContext(ctx, `
SELECT rs.id, rs.book_id, b.title, rs.started_at, rs.ended_at, rs.seconds
FROM reading_sessions rs JOIN books b ON b.id = rs.book_id
WHERE rs.user_id=?
ORDER BY rs.ended_at DESC, rs.id DESC
LIMIT ?`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.ReadingSession
	for rows.Next() {
		var sess models.ReadingSession
		var started, ended int64
		if err := rows.Scan(&sess.ID, &sess.BookID, &sess.Title, &started, &ended, &sess.Seconds); err != nil {
			return nil, err
		}
		sess.StartedAt = time.Unix(started, 0)
		sess.EndedAt = time.Unix(ended, 0)
		out = append(out, sess)
	}
	if out == nil {
		out = []models.ReadingSession{}
	}
	return out, rows.Err()
}

// recentSessionStats returns session count and seconds for the trailing 7
// days plus seconds for the trailing 30 days.
func (s *Store) recentSessionStats(ctx context.Context, userID int64) (sessions7d, seconds7d, seconds30d int64) {
	now := time.Now()
	weekAgo := now.Add(-7 * 24 * time.Hour).Unix()
	monthAgo := now.Add(-30 * 24 * time.Hour).Unix()
	_ = s.queryRowContext(ctx, `
SELECT
	COALESCE(SUM(CASE WHEN ended_at >= ? THEN 1 ELSE 0 END),0),
	COALESCE(SUM(CASE WHEN ended_at >= ? THEN seconds ELSE 0 END),0),
	COALESCE(SUM(CASE WHEN ended_at >= ? THEN seconds ELSE 0 END),0)
FROM reading_sessions WHERE user_id=?`,
		weekAgo, weekAgo, monthAgo, userID).
		Scan(&sessions7d, &seconds7d, &seconds30d)
	return sessions7d, seconds7d, seconds30d
}
