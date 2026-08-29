package storage

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"athenaeum/internal/models"
)

const autoAuthorMinBooks = 2

// EnsureAutoCollections creates or refreshes system-managed smart shelves
// after each library scan (Recently Added, and one shelf per prolific author).
func (s *Store) EnsureAutoCollections(ctx context.Context) error {
	userIDs, err := s.allUserIDs(ctx)
	if err != nil {
		return err
	}
	for _, uid := range userIDs {
		if err := s.ensureRecentAuto(ctx, uid); err != nil {
			return err
		}
		if err := s.syncAuthorAuto(ctx, uid); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) allUserIDs(ctx context.Context) ([]int64, error) {
	ids := []int64{models.AnonymousUserID}
	rows, err := s.queryContext(ctx, `SELECT id FROM users`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (s *Store) ensureRecentAuto(ctx context.Context, userID int64) error {
	q := models.SmartQuery{AddedDays: 30}
	return s.upsertAutoCollection(ctx, userID, "Recently Added", "Books added in the last 30 days", q)
}

func (s *Store) syncAuthorAuto(ctx context.Context, userID int64) error {
	rows, err := s.queryContext(ctx, `
SELECT author, COUNT(*) AS n FROM books
WHERE author != ''
GROUP BY author
HAVING n >= ?`, autoAuthorMinBooks)
	if err != nil {
		return err
	}

	type authorShelf struct {
		author string
		name   string
	}
	var shelves []authorShelf
	for rows.Next() {
		var author string
		var n int64
		if err := rows.Scan(&author, &n); err != nil {
			_ = rows.Close()
			return err
		}
		shelves = append(shelves, authorShelf{author: author, name: "By " + author})
	}
	if err := rows.Close(); err != nil {
		return err
	}
	if err := rows.Err(); err != nil {
		return err
	}

	active := make(map[string]struct{}, len(shelves))
	for _, sh := range shelves {
		active[sh.name] = struct{}{}
		q := models.SmartQuery{Author: sh.author}
		if err := s.upsertAutoCollection(ctx, userID, sh.name, "All books by "+sh.author, q); err != nil {
			return err
		}
	}
	return s.pruneStaleAuthorAuto(ctx, userID, active)
}

func (s *Store) upsertAutoCollection(ctx context.Context, userID int64, name, desc string, q models.SmartQuery) error {
	raw, err := json.Marshal(q)
	if err != nil {
		return err
	}
	now := time.Now().Unix()
	_, err = s.execContext(ctx, `
INSERT INTO collections (user_id, name, description, kind, query_json, created_at)
VALUES (?,?,?,?,?,?)
ON CONFLICT(user_id, name) DO UPDATE SET
	description=excluded.description,
	kind=excluded.kind,
	query_json=excluded.query_json`,
		userID, name, desc, models.CollectionAuto, string(raw), now)
	return err
}

func (s *Store) pruneStaleAuthorAuto(ctx context.Context, userID int64, active map[string]struct{}) error {
	rows, err := s.queryContext(ctx, `
SELECT id, name FROM collections
WHERE user_id=? AND kind=? AND name LIKE 'By %'`, userID, models.CollectionAuto)
	if err != nil {
		return err
	}
	var stale []int64
	for rows.Next() {
		var id int64
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			_ = rows.Close()
			return err
		}
		if _, ok := active[name]; !ok {
			stale = append(stale, id)
		}
	}
	if err := rows.Close(); err != nil {
		return err
	}
	if err := rows.Err(); err != nil {
		return err
	}
	for _, id := range stale {
		if _, err := s.execContext(ctx, `DELETE FROM collections WHERE id=?`, id); err != nil {
			return err
		}
	}
	return nil
}

// ApplySmartQuery merges a smart collection query into a book list query.
func ApplySmartQuery(q models.BookQuery, sq models.SmartQuery) models.BookQuery {
	if sq.Format != "" {
		q.Format = sq.Format
	}
	if sq.Author != "" {
		q.Author = sq.Author
	}
	if sq.Series != "" {
		q.Series = sq.Series
	}
	if sq.Search != "" {
		q.Search = sq.Search
	}
	if sq.AddedDays > 0 {
		q.AddedAfter = time.Now().Add(-time.Duration(sq.AddedDays) * 24 * time.Hour).Unix()
	}
	return q
}

func (s *Store) countSmartBooks(ctx context.Context, sq models.SmartQuery) (int64, error) {
	q := ApplySmartQuery(models.BookQuery{}, sq)
	return s.countBooks(ctx, q)
}

func (s *Store) countBooks(ctx context.Context, q models.BookQuery) (int64, error) {
	ftsIDs, ftsErr := s.searchFTS(ctx, q.Search)
	useFTS := q.Search != "" && ftsErr == nil && len(ftsIDs) > 0
	where, args := bookWhereClause(q, ftsIDs, useFTS, nil)
	clause := ""
	if len(where) > 0 {
		clause = " WHERE " + strings.Join(where, " AND ")
	}
	var total int64
	err := s.queryRowContext(ctx, "SELECT COUNT(*) FROM books"+clause, args...).Scan(&total)
	return total, err
}
