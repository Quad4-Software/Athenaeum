package storage

import (
	"context"
	"database/sql"
	"errors"

	"athenaeum/internal/models"
)

// SaveKosyncProgress upserts a KOReader sync position for a user's document.
func (s *Store) SaveKosyncProgress(ctx context.Context, userID int64, document, progress string, percentage float64, device, deviceID string, timestamp int64) error {
	_, err := s.execContext(ctx, `
INSERT INTO kosync_documents (user_id, document, progress, percentage, device, device_id, timestamp)
VALUES (?,?,?,?,?,?,?)
ON CONFLICT(user_id, document) DO UPDATE SET
	progress=excluded.progress,
	percentage=excluded.percentage,
	device=excluded.device,
	device_id=excluded.device_id,
	timestamp=excluded.timestamp`,
		userID, document, progress, percentage, device, deviceID, timestamp)
	return err
}

// GetKosyncProgress returns a user's stored sync position for one document.
func (s *Store) GetKosyncProgress(ctx context.Context, userID int64, document string) (models.KosyncDocument, error) {
	var d models.KosyncDocument
	err := s.queryRowContext(ctx, `
SELECT document, progress, percentage, device, device_id, timestamp
FROM kosync_documents WHERE user_id=? AND document=?`, userID, document).
		Scan(&d.Document, &d.Progress, &d.Percentage, &d.Device, &d.DeviceID, &d.Timestamp)
	if errors.Is(err, sql.ErrNoRows) {
		return models.KosyncDocument{}, ErrNotFound
	}
	return d, err
}
