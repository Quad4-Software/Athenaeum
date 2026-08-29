package storage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"athenaeum/internal/models"
)

// CreateUploadSession inserts a new resumable upload.
func (s *Store) CreateUploadSession(ctx context.Context, sess models.UploadSession) error {
	now := time.Now().Unix()
	_, err := s.execContext(ctx, `
INSERT INTO upload_sessions (id, library_id, user_id, rel_path, total_size, "offset", done, book_id, created_at, updated_at)
VALUES (?,?,?,?,?,?,0,0,?,?)`,
		sess.ID, sess.LibraryID, sess.UserID, sess.RelPath, sess.TotalSize, 0, now, now)
	return err
}

// GetUploadSession loads one upload session.
func (s *Store) GetUploadSession(ctx context.Context, id string) (models.UploadSession, error) {
	var sess models.UploadSession
	var done int
	var created, updated int64
	err := s.queryRowContext(ctx, `
SELECT id, library_id, user_id, rel_path, total_size, "offset", done, book_id, created_at, updated_at
FROM upload_sessions WHERE id=?`, id).
		Scan(&sess.ID, &sess.LibraryID, &sess.UserID, &sess.RelPath, &sess.TotalSize, &sess.Offset,
			&done, &sess.BookID, &created, &updated)
	if errors.Is(err, sql.ErrNoRows) {
		return models.UploadSession{}, ErrNotFound
	}
	if err != nil {
		return models.UploadSession{}, err
	}
	sess.Done = done != 0
	sess.CreatedAt = time.Unix(created, 0)
	sess.UpdatedAt = time.Unix(updated, 0)
	return sess, nil
}

// UpdateUploadOffset advances the uploaded byte count.
func (s *Store) UpdateUploadOffset(ctx context.Context, id string, offset int64) error {
	res, err := s.execContext(ctx,
		`UPDATE upload_sessions SET "offset"=?, updated_at=? WHERE id=? AND done=0`,
		offset, time.Now().Unix(), id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// CompleteUploadSession marks an upload finished and links the indexed book.
func (s *Store) CompleteUploadSession(ctx context.Context, id string, bookID int64) error {
	res, err := s.execContext(ctx,
		`UPDATE upload_sessions SET done=1, book_id=?, updated_at=? WHERE id=?`,
		bookID, time.Now().Unix(), id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// DeleteUploadSession removes session metadata.
func (s *Store) DeleteUploadSession(ctx context.Context, id string) error {
	res, err := s.execContext(ctx, `DELETE FROM upload_sessions WHERE id=?`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// FindBookByContentHash returns another book id with the same hash, if any.
func (s *Store) FindBookByContentHash(ctx context.Context, hash string, excludeID int64) (int64, error) {
	if hash == "" {
		return 0, nil
	}
	var id int64
	err := s.queryRowContext(ctx,
		`SELECT id FROM books WHERE content_hash=? AND id<>? ORDER BY id ASC LIMIT 1`, hash, excludeID).
		Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	return id, err
}

// SetBookDuplicateOf links a book to an earlier duplicate.
func (s *Store) SetBookDuplicateOf(ctx context.Context, bookID, duplicateOf int64) error {
	_, err := s.execContext(ctx,
		`UPDATE books SET duplicate_of=? WHERE id=?`, duplicateOf, bookID)
	return err
}

// ApplyContentHash stores hash and duplicate link after indexing.
func (s *Store) ApplyContentHash(ctx context.Context, bookID int64, hash string) error {
	dupID, err := s.FindBookByContentHash(ctx, hash, bookID)
	if err != nil {
		return err
	}
	_, err = s.execContext(ctx,
		`UPDATE books SET content_hash=?, duplicate_of=? WHERE id=?`, hash, dupID, bookID)
	return err
}
