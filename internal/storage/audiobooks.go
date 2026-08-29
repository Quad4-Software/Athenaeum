package storage

import (
	"context"
	"time"

	"athenaeum/internal/models"
)

// ListAudioBooksByLibrary returns non-hidden audio format books in a library.
func (s *Store) ListAudioBooksByLibrary(ctx context.Context, libraryID int64) ([]models.Book, error) {
	rows, err := s.queryContext(ctx, selectColumns+`
 WHERE library_id=? AND hidden=0 AND format IN ('mp3','m4b','m4a','ogg','flac')`, libraryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Book
	for rows.Next() {
		b, err := scanBook(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

// ListAudiobookTracks returns ordered tracks for a set book.
func (s *Store) UpsertAudiobookSet(ctx context.Context, b *models.Book, tracks []models.AudiobookTrack) (int64, error) {
	now := time.Now().Unix()
	id, err := s.insertID(ctx, `
INSERT INTO books (library_id, title, author, series, series_index, format, rel_path, abs_path,
	file_size, has_cover, language, description, mtime, added_at, modified_at, meta_edited, cover_edited, hidden)
VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,0)
ON CONFLICT(library_id, rel_path) DO UPDATE SET
	title=excluded.title,
	author=excluded.author,
	series=excluded.series,
	format=excluded.format,
	abs_path=excluded.abs_path,
	file_size=excluded.file_size,
	modified_at=excluded.modified_at
RETURNING id`, b.LibraryID, b.Title, b.Author, b.Series, b.SeriesIndex, b.Format, b.RelPath, b.AbsPath,
		b.FileSize, boolToInt(b.HasCover), b.Language, b.Description, now, now, now,
		boolToInt(b.MetaEdited), boolToInt(b.CoverEdited))
	if err != nil {
		return 0, err
	}
	if _, err := s.execContext(ctx, `DELETE FROM audiobook_tracks WHERE set_book_id=?`, id); err != nil {
		return 0, err
	}
	for _, t := range tracks {
		_, err := s.execContext(ctx, `
INSERT INTO audiobook_tracks (set_book_id, track_index, rel_path, title, format, file_size)
VALUES (?,?,?,?,?,?)`, id, t.Index, t.RelPath, t.Title, t.Format, t.FileSize)
		if err != nil {
			return 0, err
		}
	}
	return id, nil
}

// HideBook marks a per-file track as hidden and links it to a set book.
func (s *Store) HideBook(ctx context.Context, bookID, setID int64) error {
	_, err := s.execContext(ctx, `UPDATE books SET hidden=1, audiobook_set_id=? WHERE id=?`, setID, bookID)
	return err
}

// PruneOrphanAudiobookSets removes set rows whose directories no longer qualify.
func (s *Store) PruneOrphanAudiobookSets(ctx context.Context, libraryID int64) error {
	_, err := s.execContext(ctx, `
DELETE FROM books WHERE library_id=? AND format=? AND id NOT IN (
	SELECT DISTINCT audiobook_set_id FROM books WHERE audiobook_set_id > 0 AND library_id=?
)`, libraryID, models.FormatAudiobook, libraryID)
	return err
}

// ListAudiobookSetRelPaths returns rel_paths for multi-file audiobook set rows.
func (s *Store) ListAudiobookSetRelPaths(ctx context.Context, libraryID int64) ([]string, error) {
	rows, err := s.queryContext(ctx,
		`SELECT rel_path FROM books WHERE library_id=? AND format=?`, libraryID, models.FormatAudiobook)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *Store) ListAudiobookTracks(ctx context.Context, setBookID int64) ([]models.AudiobookTrack, error) {
	rows, err := s.queryContext(ctx, `
SELECT track_index, rel_path, title, format, file_size
FROM audiobook_tracks WHERE set_book_id=? ORDER BY track_index`, setBookID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.AudiobookTrack
	for rows.Next() {
		var t models.AudiobookTrack
		if err := rows.Scan(&t.Index, &t.RelPath, &t.Title, &t.Format, &t.FileSize); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}
