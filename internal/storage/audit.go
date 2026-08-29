package storage

import (
	"context"
	"strings"
	"time"

	"athenaeum/internal/models"
)

// InsertAudit appends one audit log row.
func (s *Store) InsertAudit(ctx context.Context, e models.AuditEntry) error {
	_, err := s.execContext(ctx, `
INSERT INTO audit_log (actor_id, actor_name, target_user_id, target_name, action, details, ip, created_at)
VALUES (?,?,?,?,?,?,?,?)`,
		e.ActorID, e.ActorName, e.TargetUserID, e.TargetName, e.Action, e.Details, e.IP, time.Now().Unix())
	return err
}

// ListAudit returns paginated audit entries, newest first.
// action filters exact action when non-empty. q matches actor, target, action, details, or ip.
func (s *Store) ListAudit(ctx context.Context, limit, offset int, action, q string) (models.AuditPage, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	var page models.AuditPage
	page.Limit = limit
	page.Offset = offset

	where, args := auditWhere(action, q)
	countSQL := `SELECT COUNT(*) FROM audit_log` + where
	listSQL := `
SELECT id, actor_id, actor_name, target_user_id, target_name, action, details, ip, created_at
FROM audit_log` + where + `
ORDER BY created_at DESC, id DESC
LIMIT ? OFFSET ?`

	if err := s.queryRowContext(ctx, countSQL, args...).Scan(&page.Total); err != nil {
		return page, err
	}
	listArgs := append(append([]any{}, args...), limit, offset)
	rows, err := s.queryContext(ctx, listSQL, listArgs...)
	if err != nil {
		return page, err
	}
	defer rows.Close()
	for rows.Next() {
		var e models.AuditEntry
		var created int64
		if err := rows.Scan(&e.ID, &e.ActorID, &e.ActorName, &e.TargetUserID, &e.TargetName,
			&e.Action, &e.Details, &e.IP, &created); err != nil {
			return page, err
		}
		e.CreatedAt = time.Unix(created, 0)
		page.Items = append(page.Items, e)
	}
	if page.Items == nil {
		page.Items = []models.AuditEntry{}
	}
	return page, rows.Err()
}

func auditWhere(action, q string) (string, []any) {
	var parts []string
	var args []any
	action = strings.TrimSpace(action)
	q = strings.TrimSpace(q)
	if action != "" {
		parts = append(parts, "action = ?")
		args = append(args, action)
	}
	if q != "" {
		like := "%" + q + "%"
		parts = append(parts, `(LOWER(actor_name) LIKE LOWER(?)
OR LOWER(COALESCE(target_name, '')) LIKE LOWER(?)
OR LOWER(action) LIKE LOWER(?)
OR LOWER(COALESCE(details, '')) LIKE LOWER(?)
OR LOWER(COALESCE(ip, '')) LIKE LOWER(?))`)
		args = append(args, like, like, like, like, like)
	}
	if len(parts) == 0 {
		return "", nil
	}
	return " WHERE " + strings.Join(parts, " AND "), args
}
