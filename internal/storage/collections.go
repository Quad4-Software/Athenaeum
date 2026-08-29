package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"athenaeum/internal/models"
)

// ListCollections returns shelves owned by userID.
func (s *Store) ListCollections(ctx context.Context, userID int64) ([]models.Collection, error) {
	rows, err := s.queryContext(ctx, `
SELECT id, user_id, name, description, kind, query_json, created_at
FROM collections
WHERE user_id = ?
ORDER BY
  CASE kind WHEN 'auto' THEN 0 WHEN 'smart' THEN 1 ELSE 2 END,
  LOWER(name) ASC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Collection
	for rows.Next() {
		c, err := scanCollectionFields(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := s.fillCollectionCounts(ctx, out); err != nil {
		return nil, err
	}
	return out, nil
}

// fillCollectionCounts batches manual/reading counts and fills smart/auto counts.
func (s *Store) fillCollectionCounts(ctx context.Context, cols []models.Collection) error {
	manualIDs := make([]int64, 0, len(cols))
	manualIdx := make(map[int64]int, len(cols))
	for i := range cols {
		kind := cols[i].Kind
		if kind == models.CollectionManual || kind == models.CollectionReading {
			manualIDs = append(manualIDs, cols[i].ID)
			manualIdx[cols[i].ID] = i
			continue
		}
		if err := s.fillCollectionCount(ctx, &cols[i]); err != nil {
			return err
		}
	}
	if len(manualIDs) == 0 {
		return nil
	}
	placeholders := strings.Repeat("?,", len(manualIDs))
	placeholders = placeholders[:len(placeholders)-1]
	args := make([]any, len(manualIDs))
	for i, id := range manualIDs {
		args[i] = id
	}
	qrows, err := s.queryContext(ctx,
		// placeholders is only "?" repeats sized to manualIDs.
		`SELECT collection_id, COUNT(*) FROM collection_items WHERE collection_id IN (`+placeholders+`) GROUP BY collection_id`, // #nosec G202
		args...)
	if err != nil {
		return err
	}
	defer qrows.Close()
	for qrows.Next() {
		var id, count int64
		if err := qrows.Scan(&id, &count); err != nil {
			return err
		}
		if i, ok := manualIdx[id]; ok {
			cols[i].BookCount = count
		}
	}
	return qrows.Err()
}

func scanCollectionFields(row scanner) (models.Collection, error) {
	var c models.Collection
	var created int64
	var kind, queryJSON string
	err := row.Scan(&c.ID, &c.UserID, &c.Name, &c.Description, &kind, &queryJSON, &created)
	if err != nil {
		return models.Collection{}, err
	}
	c.Kind = kind
	if c.Kind == "" {
		c.Kind = models.CollectionManual
	}
	c.CreatedAt = time.Unix(created, 0)
	if queryJSON != "" {
		var sq models.SmartQuery
		if err := json.Unmarshal([]byte(queryJSON), &sq); err == nil {
			c.Query = &sq
		}
	}
	return c, nil
}

func (s *Store) fillCollectionCount(ctx context.Context, c *models.Collection) error {
	if c.Kind == models.CollectionManual || c.Kind == models.CollectionReading {
		return s.queryRowContext(ctx,
			`SELECT COUNT(*) FROM collection_items WHERE collection_id=?`, c.ID).Scan(&c.BookCount)
	}
	if c.Query != nil {
		var err error
		c.BookCount, err = s.countSmartBooks(ctx, *c.Query)
		return err
	}
	return nil
}

// GetCollection returns one collection if it belongs to userID.
func (s *Store) GetCollection(ctx context.Context, userID, id int64) (models.Collection, error) {
	row := s.queryRowContext(ctx, `
SELECT id, user_id, name, description, kind, query_json, created_at
FROM collections WHERE id = ? AND user_id = ?`, id, userID)
	c, err := scanCollectionFields(row)
	if errors.Is(err, sql.ErrNoRows) {
		return models.Collection{}, ErrNotFound
	}
	if err != nil {
		return models.Collection{}, err
	}
	if err := s.fillCollectionCount(ctx, &c); err != nil {
		return models.Collection{}, err
	}
	return c, nil
}

// CreateReadingCollection inserts a reading list for userID.
func (s *Store) CreateReadingCollection(ctx context.Context, userID int64, name, description string) (models.Collection, error) {
	return s.createCollection(ctx, userID, name, description, models.CollectionReading, nil)
}

// CreateCollection inserts a manual shelf for userID.
func (s *Store) CreateCollection(ctx context.Context, userID int64, name, description string) (models.Collection, error) {
	return s.createCollection(ctx, userID, name, description, models.CollectionManual, nil)
}

// CreateSmartCollection inserts a query-driven shelf.
func (s *Store) CreateSmartCollection(ctx context.Context, userID int64, name, description string, q models.SmartQuery) (models.Collection, error) {
	return s.createCollection(ctx, userID, name, description, models.CollectionSmart, &q)
}

func (s *Store) createCollection(ctx context.Context, userID int64, name, description, kind string, q *models.SmartQuery) (models.Collection, error) {
	queryJSON := ""
	if q != nil {
		raw, err := json.Marshal(q)
		if err != nil {
			return models.Collection{}, err
		}
		queryJSON = string(raw)
	}
	now := time.Now().Unix()
	id, err := s.insertID(ctx, `
INSERT INTO collections (user_id, name, description, kind, query_json, created_at)
VALUES (?,?,?,?,?,?) RETURNING id`,
		userID, name, description, kind, queryJSON, now)
	if err != nil {
		return models.Collection{}, err
	}
	return s.GetCollection(ctx, userID, id)
}

// UpdateCollection renames or updates a manual/smart collection description.
func (s *Store) UpdateCollection(ctx context.Context, userID, id int64, name, description string) (models.Collection, error) {
	c, err := s.GetCollection(ctx, userID, id)
	if err != nil {
		return models.Collection{}, err
	}
	if c.Kind == models.CollectionAuto {
		return models.Collection{}, errors.New("auto collections cannot be edited")
	}
	res, err := s.execContext(ctx, `
UPDATE collections SET name=?, description=? WHERE id=? AND user_id=?`,
		name, description, id, userID)
	if err != nil {
		return models.Collection{}, err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return models.Collection{}, ErrNotFound
	}
	return s.GetCollection(ctx, userID, id)
}

// DeleteCollection removes a shelf. Auto collections are recreated on the next scan.
func (s *Store) DeleteCollection(ctx context.Context, userID, id int64) error {
	res, err := s.execContext(ctx, `DELETE FROM collections WHERE id=? AND user_id=?`, id, userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// AddToCollection links a book to a manual shelf.
func (s *Store) AddToCollection(ctx context.Context, userID, collectionID, bookID int64) error {
	c, err := s.GetCollection(ctx, userID, collectionID)
	if err != nil {
		return err
	}
	if c.Kind != models.CollectionManual && c.Kind != models.CollectionReading {
		return errors.New("only manual collections and reading lists accept book assignments")
	}
	if _, err := s.GetBook(ctx, bookID); err != nil {
		return err
	}
	var maxOrder int
	_ = s.queryRowContext(ctx,
		`SELECT COALESCE(MAX(sort_order),0) FROM collection_items WHERE collection_id=?`, collectionID).
		Scan(&maxOrder)
	_, err = s.execContext(ctx, `
INSERT INTO collection_items (collection_id, book_id, sort_order)
VALUES (?,?,?)
ON CONFLICT(collection_id, book_id) DO NOTHING`,
		collectionID, bookID, maxOrder+1)
	return err
}

// RemoveFromCollection unlinks a book from a manual shelf.
func (s *Store) RemoveFromCollection(ctx context.Context, userID, collectionID, bookID int64) error {
	c, err := s.GetCollection(ctx, userID, collectionID)
	if err != nil {
		return err
	}
	if c.Kind != models.CollectionManual && c.Kind != models.CollectionReading {
		return errors.New("only manual collections and reading lists accept book assignments")
	}
	res, err := s.execContext(ctx,
		`DELETE FROM collection_items WHERE collection_id=? AND book_id=?`, collectionID, bookID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// ResolveCollectionFilters loads a collection and merges its smart query into q.
func (s *Store) ResolveCollectionFilters(ctx context.Context, userID int64, q models.BookQuery) (models.BookQuery, error) {
	if q.CollectionID <= 0 {
		return q, nil
	}
	c, err := s.GetCollection(ctx, userID, q.CollectionID)
	if err != nil {
		return q, err
	}
	if c.Kind == models.CollectionManual || c.Kind == models.CollectionReading {
		return q, nil
	}
	if c.Query == nil {
		return q, nil
	}
	merged := ApplySmartQuery(q, *c.Query)
	merged.CollectionID = 0
	return merged, nil
}
