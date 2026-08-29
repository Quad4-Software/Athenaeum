package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"athenaeum/internal/models"
)

// CreateWebhook inserts a webhook subscription.
func (s *Store) CreateWebhook(ctx context.Context, wh models.Webhook) (models.Webhook, error) {
	now := time.Now()
	eventsJSON, err := json.Marshal(wh.Events)
	if err != nil {
		return models.Webhook{}, err
	}
	id, err := s.insertID(ctx, `
INSERT INTO webhooks (url, secret, events, enabled, created_at, updated_at)
VALUES (?,?,?,?,?,?) RETURNING id`,
		wh.URL, wh.Secret, string(eventsJSON), boolToInt(wh.Enabled), now.Unix(), now.Unix())
	if err != nil {
		return models.Webhook{}, err
	}
	wh.ID = id
	wh.CreatedAt = now
	wh.UpdatedAt = now
	return wh, nil
}

// GetWebhook loads a webhook by id.
func (s *Store) GetWebhook(ctx context.Context, id int64) (models.Webhook, error) {
	row := s.queryRowContext(ctx, `
SELECT id, url, secret, events, enabled, created_at, updated_at FROM webhooks WHERE id=?`, id)
	return scanWebhook(row)
}

// ListWebhooks returns all webhook subscriptions.
func (s *Store) ListWebhooks(ctx context.Context) ([]models.Webhook, error) {
	rows, err := s.queryContext(ctx, `
SELECT id, url, secret, events, enabled, created_at, updated_at FROM webhooks ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Webhook
	for rows.Next() {
		wh, err := scanWebhook(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, wh)
	}
	return out, rows.Err()
}

// ListEnabledWebhooksForEvent returns enabled hooks subscribed to event or "*".
func (s *Store) ListEnabledWebhooksForEvent(ctx context.Context, event string) ([]models.Webhook, error) {
	all, err := s.ListWebhooks(ctx)
	if err != nil {
		return nil, err
	}
	var out []models.Webhook
	for _, wh := range all {
		if !wh.Enabled {
			continue
		}
		for _, e := range wh.Events {
			if e == event || e == "*" {
				out = append(out, wh)
				break
			}
		}
	}
	return out, nil
}

// UpdateWebhook updates a webhook subscription.
func (s *Store) UpdateWebhook(ctx context.Context, wh models.Webhook) error {
	eventsJSON, err := json.Marshal(wh.Events)
	if err != nil {
		return err
	}
	res, err := s.execContext(ctx, `
UPDATE webhooks SET url=?, secret=?, events=?, enabled=?, updated_at=? WHERE id=?`,
		wh.URL, wh.Secret, string(eventsJSON), boolToInt(wh.Enabled), time.Now().Unix(), wh.ID)
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

// DeleteWebhook removes a webhook and its deliveries.
func (s *Store) DeleteWebhook(ctx context.Context, id int64) error {
	_, _ = s.execContext(ctx, `DELETE FROM webhook_deliveries WHERE webhook_id=?`, id)
	res, err := s.execContext(ctx, `DELETE FROM webhooks WHERE id=?`, id)
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

// InsertWebhookDelivery records a delivery attempt.
func (s *Store) InsertWebhookDelivery(ctx context.Context, d models.WebhookDelivery) (int64, error) {
	var delivered sql.NullInt64
	if d.DeliveredAt != nil {
		delivered = sql.NullInt64{Int64: d.DeliveredAt.Unix(), Valid: true}
	}
	return s.insertID(ctx, `
INSERT INTO webhook_deliveries (webhook_id, event, payload, status_code, success, attempts, last_error, created_at, delivered_at)
VALUES (?,?,?,?,?,?,?,?,?) RETURNING id`,
		d.WebhookID, d.Event, d.Payload, d.StatusCode, boolToInt(d.Success), d.Attempts, d.LastError,
		d.CreatedAt.Unix(), delivered)
}

// ListWebhookDeliveries returns recent deliveries for a webhook.
func (s *Store) ListWebhookDeliveries(ctx context.Context, webhookID int64, limit, offset int) ([]models.WebhookDelivery, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	rows, err := s.queryContext(ctx, `
SELECT id, webhook_id, event, payload, status_code, success, attempts, last_error, created_at, delivered_at
FROM webhook_deliveries WHERE webhook_id=? ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		webhookID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.WebhookDelivery
	for rows.Next() {
		d, err := scanWebhookDelivery(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func scanWebhook(row scanner) (models.Webhook, error) {
	var wh models.Webhook
	var eventsJSON string
	var enabled int
	var created, updated int64
	err := row.Scan(&wh.ID, &wh.URL, &wh.Secret, &eventsJSON, &enabled, &created, &updated)
	if errors.Is(err, sql.ErrNoRows) {
		return models.Webhook{}, ErrNotFound
	}
	if err != nil {
		return models.Webhook{}, err
	}
	wh.Enabled = enabled != 0
	wh.CreatedAt = time.Unix(created, 0)
	wh.UpdatedAt = time.Unix(updated, 0)
	if eventsJSON != "" {
		_ = json.Unmarshal([]byte(eventsJSON), &wh.Events)
	}
	if wh.Events == nil {
		wh.Events = []string{}
	}
	return wh, nil
}

func scanWebhookDelivery(row scanner) (models.WebhookDelivery, error) {
	var d models.WebhookDelivery
	var success int
	var created int64
	var delivered sql.NullInt64
	err := row.Scan(&d.ID, &d.WebhookID, &d.Event, &d.Payload, &d.StatusCode, &success,
		&d.Attempts, &d.LastError, &created, &delivered)
	if err != nil {
		return models.WebhookDelivery{}, err
	}
	d.Success = success != 0
	d.CreatedAt = time.Unix(created, 0)
	if delivered.Valid {
		t := time.Unix(delivered.Int64, 0)
		d.DeliveredAt = &t
	}
	return d, nil
}
