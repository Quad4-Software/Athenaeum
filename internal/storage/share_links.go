package storage

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"time"

	"athenaeum/internal/models"
)

// NewShareToken returns a URL-safe random token suitable for a share link.
func NewShareToken() (string, error) {
	buf := make([]byte, 20)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

const shareLinkColumns = `SELECT id, token, book_id, created_by, expires_at, created_at, download_count, max_downloads FROM share_links`

// CreateShareLink inserts a new share link for bookID and returns the stored row.
func (s *Store) CreateShareLink(ctx context.Context, bookID, createdBy int64, expiresAt *time.Time, maxDownloads int64) (models.ShareLink, error) {
	token, err := NewShareToken()
	if err != nil {
		return models.ShareLink{}, err
	}
	now := time.Now()
	var expiresUnix sql.NullInt64
	if expiresAt != nil {
		expiresUnix = sql.NullInt64{Int64: expiresAt.Unix(), Valid: true}
	}
	id, err := s.insertID(ctx, `
INSERT INTO share_links (token, book_id, created_by, expires_at, created_at, download_count, max_downloads)
VALUES (?,?,?,?,?,0,?) RETURNING id`,
		token, bookID, createdBy, expiresUnix, now.Unix(), maxDownloads)
	if err != nil {
		return models.ShareLink{}, err
	}
	return models.ShareLink{
		ID:           id,
		Token:        token,
		BookID:       bookID,
		CreatedBy:    createdBy,
		ExpiresAt:    expiresAt,
		CreatedAt:    now,
		MaxDownloads: maxDownloads,
	}, nil
}

func scanShareLink(row scanner) (models.ShareLink, error) {
	var sl models.ShareLink
	var expires sql.NullInt64
	var created int64
	err := row.Scan(&sl.ID, &sl.Token, &sl.BookID, &sl.CreatedBy, &expires, &created,
		&sl.DownloadCount, &sl.MaxDownloads)
	if err != nil {
		return models.ShareLink{}, err
	}
	sl.CreatedAt = time.Unix(created, 0)
	if expires.Valid {
		t := time.Unix(expires.Int64, 0)
		sl.ExpiresAt = &t
	}
	return sl, nil
}

// GetShareLinkByToken looks up a share link by its public token.
func (s *Store) GetShareLinkByToken(ctx context.Context, token string) (models.ShareLink, error) {
	row := s.queryRowContext(ctx, shareLinkColumns+` WHERE token=?`, token)
	sl, err := scanShareLink(row)
	if errors.Is(err, sql.ErrNoRows) {
		return models.ShareLink{}, ErrNotFound
	}
	return sl, err
}

// ListSharesForBook returns share links for a book, newest first.
func (s *Store) ListSharesForBook(ctx context.Context, bookID int64) ([]models.ShareLink, error) {
	rows, err := s.queryContext(ctx, shareLinkColumns+` WHERE book_id=? ORDER BY created_at DESC`, bookID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.ShareLink
	for rows.Next() {
		sl, err := scanShareLink(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, sl)
	}
	return out, rows.Err()
}

// DeleteShareLink removes a share link scoped to its owning book.
func (s *Store) DeleteShareLink(ctx context.Context, bookID, id int64) error {
	res, err := s.execContext(ctx, `DELETE FROM share_links WHERE id=? AND book_id=?`, id, bookID)
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

// IncrementShareDownload records one download against a share link with no max check.
func (s *Store) IncrementShareDownload(ctx context.Context, id int64) error {
	_, err := s.execContext(ctx, `UPDATE share_links SET download_count = download_count + 1 WHERE id=?`, id)
	return err
}

// TryIncrementShareDownload atomically records a download when under maxDownloads.
// A maxDownloads of 0 or less means unlimited. Returns false when the limit is reached.
func (s *Store) TryIncrementShareDownload(ctx context.Context, id int64, maxDownloads int64) (bool, error) {
	if maxDownloads <= 0 {
		return true, s.IncrementShareDownload(ctx, id)
	}
	res, err := s.execContext(ctx, `
UPDATE share_links SET download_count = download_count + 1
WHERE id=? AND download_count < ?`, id, maxDownloads)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}
