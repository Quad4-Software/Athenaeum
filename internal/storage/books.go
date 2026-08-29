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

// UpsertBook inserts or updates a book keyed by its relative path and
// returns the row id.
func (s *Store) UpsertBook(ctx context.Context, b *models.Book, mtime int64) (int64, error) {
	if b.LibraryID == 0 {
		b.LibraryID = 1
	}
	now := time.Now().Unix()
	return s.insertID(ctx, `
INSERT INTO books (library_id, title, author, series, series_index, format, rel_path, abs_path,
	file_size, has_cover, language, description, mtime, added_at, modified_at, meta_edited, cover_edited)
VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
ON CONFLICT(library_id, rel_path) DO UPDATE SET
	title=CASE WHEN books.meta_edited=1 THEN books.title ELSE excluded.title END,
	author=CASE WHEN books.meta_edited=1 THEN books.author ELSE excluded.author END,
	series=CASE WHEN books.meta_edited=1 THEN books.series ELSE excluded.series END,
	series_index=CASE WHEN books.meta_edited=1 THEN books.series_index ELSE excluded.series_index END,
	format=excluded.format,
	abs_path=excluded.abs_path,
	file_size=excluded.file_size,
	has_cover=CASE WHEN books.cover_edited=1 THEN books.has_cover ELSE excluded.has_cover END,
	language=CASE WHEN books.meta_edited=1 THEN books.language ELSE excluded.language END,
	description=CASE WHEN books.meta_edited=1 THEN books.description ELSE excluded.description END,
	mtime=excluded.mtime,
	modified_at=excluded.modified_at
RETURNING id`,
		b.LibraryID, b.Title, b.Author, b.Series, b.SeriesIndex, b.Format, b.RelPath, b.AbsPath,
		b.FileSize, boolToInt(b.HasCover), b.Language, b.Description, mtime, now, now,
		boolToInt(b.MetaEdited), boolToInt(b.CoverEdited))
}

// FileState returns the stored mtime and size for a path within a library.
func (s *Store) FileState(ctx context.Context, libraryID int64, relPath string) (mtime, size int64, ok bool, err error) {
	err = s.queryRowContext(ctx,
		`SELECT mtime, file_size FROM books WHERE library_id=? AND rel_path=?`, libraryID, relPath).
		Scan(&mtime, &size)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, 0, false, nil
	}
	if err != nil {
		return 0, 0, false, err
	}
	return mtime, size, true, nil
}

// PrunePaths removes books in libraryID whose relative paths are not in keep.
func (s *Store) PrunePaths(ctx context.Context, libraryID int64, keep map[string]struct{}) (int, error) {
	rows, err := s.queryContext(ctx, `SELECT rel_path FROM books WHERE library_id=?`, libraryID)
	if err != nil {
		return 0, err
	}
	var stale []string
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			_ = rows.Close()
			return 0, err
		}
		if _, ok := keep[p]; !ok {
			stale = append(stale, p)
		}
	}
	if err := rows.Close(); err != nil {
		return 0, err
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}
	for _, p := range stale {
		if _, err := s.execContext(ctx, `DELETE FROM books WHERE library_id=? AND rel_path=?`, libraryID, p); err != nil {
			return 0, err
		}
	}
	return len(stale), nil
}

// GetBook returns a single book by id.
func (s *Store) GetBook(ctx context.Context, id int64) (models.Book, error) {
	row := s.queryRowContext(ctx, selectColumns+` WHERE id=?`, id)
	b, err := scanBook(row)
	if errors.Is(err, sql.ErrNoRows) {
		return models.Book{}, ErrNotFound
	}
	return b, err
}

// DeleteBook removes a book row from the library index.
func (s *Store) DeleteBook(ctx context.Context, id int64) error {
	res, err := s.execContext(ctx, `DELETE FROM books WHERE id=?`, id)
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

const selectColumns = `SELECT id, library_id, title, author, series, series_index, format, rel_path,
	abs_path, file_size, has_cover, language, description, added_at, modified_at, meta_edited, cover_edited,
	content_hash, duplicate_of, hidden, audiobook_set_id FROM books`

const listSelectColumns = `SELECT books.id, books.library_id, books.title, books.author, books.series, books.series_index, books.format, books.rel_path,
	books.abs_path, books.file_size, books.has_cover, books.language, '' AS description, books.added_at, books.modified_at, books.meta_edited, books.cover_edited,
	books.content_hash, books.duplicate_of, COALESCE(progress.percent, 0) FROM books`

// GetBookByPath returns a book by library and relative path.
func (s *Store) GetBookByPath(ctx context.Context, libraryID int64, relPath string) (models.Book, error) {
	row := s.queryRowContext(ctx, selectColumns+` WHERE library_id=? AND rel_path=?`, libraryID, relPath)
	b, err := scanBook(row)
	if errors.Is(err, sql.ErrNoRows) {
		return models.Book{}, ErrNotFound
	}
	return b, err
}

// UpdateBookMetadata applies user-edited metadata and marks the row as edited.
func (s *Store) UpdateBookMetadata(ctx context.Context, id int64, u models.BookUpdate) (models.Book, error) {
	now := time.Now().Unix()
	_, err := s.execContext(ctx, `
UPDATE books SET
	title=?,
	author=?,
	series=?,
	series_index=?,
	language=?,
	description=?,
	meta_edited=1,
	modified_at=?
WHERE id=?`,
		u.Title, u.Author, u.Series, u.SeriesIndex, u.Language, u.Description, now, id)
	if err != nil {
		return models.Book{}, err
	}
	return s.GetBook(ctx, id)
}

// SetBookCover marks a book as having a user-uploaded cover.
func (s *Store) SetBookCover(ctx context.Context, id int64, hasCover bool) error {
	now := time.Now().Unix()
	_, err := s.execContext(ctx, `
UPDATE books SET has_cover=?, cover_edited=1, modified_at=? WHERE id=?`,
		boolToInt(hasCover), now, id)
	return err
}

// SetBookHasCover updates has_cover for scanner or maintenance jobs without marking cover_edited.
func (s *Store) SetBookHasCover(ctx context.Context, id int64, hasCover bool) error {
	now := time.Now().Unix()
	_, err := s.execContext(ctx, `
UPDATE books SET has_cover=?, modified_at=? WHERE id=? AND cover_edited=0`,
		boolToInt(hasCover), now, id)
	return err
}

// ListBooks returns a page of books matching the query.
func (s *Store) ListBooks(ctx context.Context, q models.BookQuery) (models.BookPage, error) {
	limit := q.Limit
	if limit <= 0 || limit > 200 {
		limit = 60
	}

	resolved, err := s.ResolveCollectionFilters(ctx, q.UserID, q)
	if err != nil {
		return models.BookPage{}, err
	}
	q = resolved

	ftsIDs, ftsErr := s.searchFTS(ctx, q.Search)
	useFTS := q.Search != "" && ftsErr == nil && len(ftsIDs) > 0

	var contentIDs []int64
	if q.Search != "" {
		contentIDs, _ = s.SearchBookContentIDs(ctx, q.Search)
	}

	where, args := bookWhereClause(q, ftsIDs, useFTS, contentIDs)

	clause := ""
	if len(where) > 0 {
		clause = " WHERE " + strings.Join(where, " AND ")
	}

	var total int64
	if err := s.queryRowContext(ctx, "SELECT COUNT(*) FROM books"+clause, args...).Scan(&total); err != nil {
		return models.BookPage{}, err
	}

	order := orderBy(q.Sort, useFTS, q.InProgress)
	listSQL := composeListBooksSQL(clause, order)
	listArgs := append([]any{q.UserID}, args...)
	listArgs = append(listArgs, limit, q.Offset)
	rows, err := s.queryContext(ctx, listSQL, listArgs...)
	if err != nil {
		return models.BookPage{}, err
	}
	defer rows.Close()

	page := models.BookPage{Total: total, Limit: limit, Offset: q.Offset, Items: []models.Book{}}
	for rows.Next() {
		b, err := scanListBook(rows)
		if err != nil {
			return models.BookPage{}, err
		}
		page.Items = append(page.Items, b)
	}
	return page, rows.Err()
}

func bookWhereClause(q models.BookQuery, ftsIDs []int64, useFTS bool, contentIDs []int64) (where []string, args []any) {
	where = append(where, "books.hidden = 0")
	if q.LibraryID > 0 {
		where = append(where, "books.library_id = ?")
		args = append(args, q.LibraryID)
	}
	if len(q.LibraryIDs) > 0 {
		placeholders := strings.Repeat("?,", len(q.LibraryIDs))
		placeholders = placeholders[:len(placeholders)-1]
		where = append(where, "books.library_id IN ("+placeholders+")")
		for _, id := range q.LibraryIDs {
			args = append(args, id)
		}
	}
	if q.CollectionID > 0 {
		where = append(where, `books.id IN (
			SELECT book_id FROM collection_items WHERE collection_id = ?
		)`)
		args = append(args, q.CollectionID)
	}
	if q.Favorites && q.UserID >= 0 {
		where = append(where, `books.id IN (
			SELECT book_id FROM user_favorites WHERE user_id = ?
		)`)
		args = append(args, q.UserID)
	}
	if q.InProgress && q.UserID >= 0 {
		where = append(where, `books.id IN (
			SELECT book_id FROM progress WHERE user_id = ? AND percent > 0 AND percent < 0.95
		)`)
		args = append(args, q.UserID)
	}
	if q.Tag != "" {
		where = append(where, `books.id IN (
			SELECT book_tags.book_id FROM book_tags
			JOIN tags ON tags.id = book_tags.tag_id
			WHERE LOWER(tags.name) = LOWER(?)
		)`)
		args = append(args, q.Tag)
	}
	if q.Author != "" {
		where = append(where, "books.author = ?")
		args = append(args, q.Author)
	}
	if q.Series != "" {
		where = append(where, "books.series = ?")
		args = append(args, q.Series)
	}
	if q.Format != "" {
		if q.Format == models.FormatAudio {
			where = append(where, audioFormatClause())
		} else if q.Format == models.FormatComic {
			where = append(where, "books.format IN ('cbz','cbr')")
		} else if q.Format == models.FormatKindle {
			where = append(where, "books.format IN ('mobi','azw3','azw')")
		} else {
			where = append(where, "books.format = ?")
			args = append(args, q.Format)
		}
	}
	if q.AddedAfter > 0 {
		where = append(where, "books.added_at >= ?")
		args = append(args, q.AddedAfter)
	}

	if q.Search != "" {
		var conds []string
		var condArgs []any
		if useFTS {
			placeholders := strings.Repeat("?,", len(ftsIDs))
			placeholders = placeholders[:len(placeholders)-1]
			conds = append(conds, "books.id IN ("+placeholders+")")
			for _, id := range ftsIDs {
				condArgs = append(condArgs, id)
			}
		} else {
			conds = append(conds, "(books.title LIKE ? OR books.author LIKE ? OR books.series LIKE ? OR books.description LIKE ? OR books.rel_path LIKE ?)")
			like := "%" + q.Search + "%"
			condArgs = append(condArgs, like, like, like, like, like)
		}
		if len(contentIDs) > 0 {
			placeholders := strings.Repeat("?,", len(contentIDs))
			placeholders = placeholders[:len(placeholders)-1]
			conds = append(conds, "books.id IN ("+placeholders+")")
			for _, id := range contentIDs {
				condArgs = append(condArgs, id)
			}
		}
		where = append(where, "("+strings.Join(conds, " OR ")+")")
		args = append(args, condArgs...)
	}
	return where, args
}

func audioFormatClause() string {
	formats := append([]string{}, models.AudioFormats...)
	formats = append(formats, models.FormatAudiobook)
	parts := make([]string, len(formats))
	for i, f := range formats {
		parts[i] = fmt.Sprintf("'%s'", f)
	}
	return "books.format IN (" + strings.Join(parts, ",") + ")"
}

func composeListBooksSQL(whereClause, orderClause string) string {
	// whereClause contains only static predicates with ? placeholders.
	// orderClause is selected from a fixed whitelist in orderBy().
	// #nosec G202
	return listSelectColumns + `
LEFT JOIN progress ON progress.book_id = books.id AND progress.user_id = ?` + whereClause + orderClause + " LIMIT ? OFFSET ?"
}

func orderBy(sort string, fts bool, inProgress bool) string {
	if inProgress {
		return " ORDER BY progress.updated_at DESC"
	}
	if fts {
		return " ORDER BY books.added_at DESC"
	}
	switch sort {
	case "title":
		return " ORDER BY LOWER(books.title) ASC"
	case "author":
		return " ORDER BY LOWER(books.author) ASC, LOWER(books.series) ASC, books.series_index ASC"
	case "oldest":
		return " ORDER BY books.added_at ASC"
	case "progress":
		return " ORDER BY books.added_at DESC"
	case "recent", "":
		return " ORDER BY books.added_at DESC"
	default:
		return " ORDER BY books.added_at DESC"
	}
}

// ListSeries returns distinct series names with book counts, optionally scoped to libraries.
func (s *Store) ListSeries(ctx context.Context, libraryID int64, libraryIDs []int64) ([]models.SeriesInfo, error) {
	clause := ""
	args := []any{}
	if libraryID > 0 {
		clause = " AND library_id = ?"
		args = append(args, libraryID)
	} else if len(libraryIDs) > 0 {
		placeholders := strings.Repeat("?,", len(libraryIDs))
		placeholders = placeholders[:len(placeholders)-1]
		clause = fmt.Sprintf(" AND library_id IN (%s)", placeholders)
		for _, id := range libraryIDs {
			args = append(args, id)
		}
	}
	rows, err := s.queryContext(ctx, fmt.Sprintf(`
SELECT series, COUNT(*) FROM books
WHERE series != ''%s
GROUP BY series
ORDER BY LOWER(series) ASC`, clause), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.SeriesInfo
	for rows.Next() {
		var info models.SeriesInfo
		if err := rows.Scan(&info.Name, &info.Count); err != nil {
			return nil, err
		}
		out = append(out, info)
	}
	return out, rows.Err()
}

// ListAuthors returns distinct author names with book counts.
func (s *Store) ListAuthors(ctx context.Context, libraryID int64, libraryIDs []int64) ([]models.AuthorInfo, error) {
	clause := ""
	args := []any{}
	if libraryID > 0 {
		clause = " AND library_id = ?"
		args = append(args, libraryID)
	} else if len(libraryIDs) > 0 {
		placeholders := strings.Repeat("?,", len(libraryIDs))
		placeholders = placeholders[:len(placeholders)-1]
		clause = fmt.Sprintf(" AND library_id IN (%s)", placeholders)
		for _, id := range libraryIDs {
			args = append(args, id)
		}
	}
	rows, err := s.queryContext(ctx, fmt.Sprintf(`
SELECT author, COUNT(*) FROM books
WHERE author != ''%s
GROUP BY author
ORDER BY LOWER(author) ASC`, clause), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.AuthorInfo
	for rows.Next() {
		var info models.AuthorInfo
		if err := rows.Scan(&info.Name, &info.Count); err != nil {
			return nil, err
		}
		out = append(out, info)
	}
	return out, rows.Err()
}

// Stats returns aggregate counts for the library, optionally scoped to one mount.
func (s *Store) Stats(ctx context.Context, libraryID int64, userID int64) (models.LibraryStats, error) {
	var st models.LibraryStats
	audioClause := audioFormatClause()
	libClause := ""
	args := []any{}
	if libraryID > 0 {
		libClause = " WHERE library_id = ?"
		args = append(args, libraryID)
	}
	weekAgo := time.Now().Add(-7 * 24 * time.Hour).Unix()
	err := s.queryRowContext(ctx, fmt.Sprintf(`
SELECT
	COUNT(*),
	COALESCE(SUM(CASE WHEN format='epub' THEN 1 ELSE 0 END),0),
	COALESCE(SUM(CASE WHEN format='pdf'  THEN 1 ELSE 0 END),0),
	COALESCE(SUM(CASE WHEN %s THEN 1 ELSE 0 END),0),
	COALESCE(SUM(file_size),0),
	COALESCE(COUNT(DISTINCT CASE WHEN author<>'' THEN author END),0),
	COALESCE(COUNT(DISTINCT CASE WHEN series<>'' THEN series END),0),
	COALESCE(SUM(CASE WHEN added_at>=? THEN 1 ELSE 0 END),0)
FROM books%s`, audioClause, libClause), append([]any{weekAgo}, args...)...).
		Scan(&st.TotalBooks, &st.EPUBCount, &st.PDFCount, &st.AudioCount,
			&st.TotalSizeBytes, &st.AuthorCount, &st.SeriesCount, &st.AddedLast7Days)
	if err != nil {
		return st, err
	}
	if err := s.queryRowContext(ctx, `SELECT COUNT(*) FROM libraries`).Scan(&st.LibraryCount); err != nil {
		return st, err
	}
	if err := s.queryRowContext(ctx, `SELECT COUNT(*) FROM collections`).Scan(&st.CollectionCount); err != nil {
		return st, err
	}
	auth, err := s.AuthRequired(ctx)
	if err != nil {
		return st, err
	}
	st.AuthEnabled = auth
	if auth {
		if err := s.queryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&st.UserCount); err != nil {
			return st, err
		}
	}
	if userID > 0 {
		if err := s.queryRowContext(ctx, `
SELECT
	COALESCE(SUM(CASE WHEN percent>0 AND percent<0.95 THEN 1 ELSE 0 END),0),
	COALESCE(SUM(CASE WHEN percent>=0.95 THEN 1 ELSE 0 END),0)
FROM progress WHERE user_id=?`, userID).
			Scan(&st.ReadingInProgress, &st.ReadingCompleted); err != nil {
			return st, err
		}
		if err := s.queryRowContext(ctx,
			`SELECT COUNT(*) FROM user_favorites WHERE user_id=?`, userID).
			Scan(&st.FavoriteCount); err != nil {
			return st, err
		}
	}
	return st, nil
}

// SaveProgress stores reading progress for a user and book.
func (s *Store) SaveProgress(ctx context.Context, userID int64, p models.Progress) error {
	_, err := s.execContext(ctx, `
INSERT INTO progress (user_id, book_id, location, percent, read_seconds, updated_at)
VALUES (?,?,?,?,COALESCE(?,0),?)
ON CONFLICT(user_id, book_id) DO UPDATE SET
	location=excluded.location,
	percent=excluded.percent,
	read_seconds=CASE WHEN excluded.read_seconds > 0 THEN progress.read_seconds + excluded.read_seconds ELSE progress.read_seconds END,
	updated_at=excluded.updated_at`,
		userID, p.BookID, p.Location, p.Percent, p.ReadSeconds, time.Now().Unix())
	return err
}

// GetProgress returns stored progress for a user and book.
func (s *Store) GetProgress(ctx context.Context, userID, bookID int64) (models.Progress, error) {
	var p models.Progress
	var updated int64
	err := s.queryRowContext(ctx,
		`SELECT book_id, user_id, location, percent, COALESCE(read_seconds,0), updated_at FROM progress WHERE user_id=? AND book_id=?`,
		userID, bookID).
		Scan(&p.BookID, &p.UserID, &p.Location, &p.Percent, &p.ReadSeconds, &updated)
	if errors.Is(err, sql.ErrNoRows) {
		return models.Progress{BookID: bookID, UserID: userID}, nil
	}
	if err != nil {
		return models.Progress{}, err
	}
	p.UpdatedAt = time.Unix(updated, 0)
	return p, nil
}

type scanner interface{ Scan(dest ...any) error }

func scanBook(row scanner) (models.Book, error) {
	var b models.Book
	var hasCover, metaEdited, coverEdited, hidden int
	var added, modified int64
	var setID int64
	err := row.Scan(&b.ID, &b.LibraryID, &b.Title, &b.Author, &b.Series, &b.SeriesIndex, &b.Format,
		&b.RelPath, &b.AbsPath, &b.FileSize, &hasCover, &b.Language, &b.Description, &added, &modified,
		&metaEdited, &coverEdited, &b.ContentHash, &b.DuplicateOf, &hidden, &setID)
	if err != nil {
		return models.Book{}, err
	}
	b.HasCover = hasCover != 0
	b.MetaEdited = metaEdited != 0
	b.CoverEdited = coverEdited != 0
	b.Hidden = hidden != 0
	b.AddedAt = time.Unix(added, 0)
	b.ModifiedAt = time.Unix(modified, 0)
	return b, nil
}

func scanListBook(row scanner) (models.Book, error) {
	var b models.Book
	var hasCover, metaEdited, coverEdited int
	var added, modified int64
	var progress float64
	err := row.Scan(&b.ID, &b.LibraryID, &b.Title, &b.Author, &b.Series, &b.SeriesIndex, &b.Format,
		&b.RelPath, &b.AbsPath, &b.FileSize, &hasCover, &b.Language, &b.Description, &added, &modified,
		&metaEdited, &coverEdited, &b.ContentHash, &b.DuplicateOf, &progress)
	if err != nil {
		return models.Book{}, err
	}
	b.HasCover = hasCover != 0
	b.MetaEdited = metaEdited != 0
	b.CoverEdited = coverEdited != 0
	b.AddedAt = time.Unix(added, 0)
	b.ModifiedAt = time.Unix(modified, 0)
	if progress > 0 {
		b.ProgressPercent = progress
	}
	return b, nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// ListBooksForMetadata returns books targeted by a bulk metadata job.
func (s *Store) ListBooksForMetadata(ctx context.Context, libraryID int64, ids []int64) ([]models.Book, error) {
	if len(ids) > 0 {
		placeholders := strings.Repeat("?,", len(ids))
		placeholders = placeholders[:len(placeholders)-1]
		args := make([]any, len(ids))
		for i, id := range ids {
			args[i] = id
		}
		rows, err := s.queryContext(ctx, fmt.Sprintf(`
SELECT id, library_id, title, author, series, series_index, format, rel_path, abs_path,
	file_size, content_hash, has_cover, meta_edited, cover_edited, language, description,
	added_at, modified_at, duplicate_of
FROM books WHERE id IN (%s) ORDER BY id ASC`, placeholders), args...)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		return scanMetadataBooks(rows)
	}

	clause := ""
	args := []any{}
	if libraryID > 0 {
		clause = " WHERE library_id = ?"
		args = append(args, libraryID)
	}
	rows, err := s.queryContext(ctx, `
SELECT id, library_id, title, author, series, series_index, format, rel_path, abs_path,
	file_size, content_hash, has_cover, meta_edited, cover_edited, language, description,
	added_at, modified_at, duplicate_of
FROM books`+clause+` ORDER BY id ASC`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMetadataBooks(rows)
}

func scanMetadataBooks(rows *sql.Rows) ([]models.Book, error) {
	var out []models.Book
	for rows.Next() {
		var b models.Book
		var hasCover, metaEdited, coverEdited int
		var added, modified int64
		var dup sql.NullInt64
		if err := rows.Scan(
			&b.ID, &b.LibraryID, &b.Title, &b.Author, &b.Series, &b.SeriesIndex, &b.Format,
			&b.RelPath, &b.AbsPath, &b.FileSize, &b.ContentHash, &hasCover, &metaEdited,
			&coverEdited, &b.Language, &b.Description, &added, &modified, &dup,
		); err != nil {
			return nil, err
		}
		b.HasCover = hasCover != 0
		b.MetaEdited = metaEdited != 0
		b.CoverEdited = coverEdited != 0
		b.AddedAt = time.Unix(added, 0)
		b.ModifiedAt = time.Unix(modified, 0)
		if dup.Valid {
			b.DuplicateOf = dup.Int64
		}
		out = append(out, b)
	}
	if out == nil {
		out = []models.Book{}
	}
	return out, rows.Err()
}
