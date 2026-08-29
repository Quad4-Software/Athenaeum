package models

import "time"

// AuditEntry records a security-relevant action.
type AuditEntry struct {
	ID           int64     `json:"id"`
	ActorID      int64     `json:"actorId"`
	ActorName    string    `json:"actorName"`
	TargetUserID int64     `json:"targetUserId,omitempty"`
	TargetName   string    `json:"targetName,omitempty"`
	Action       string    `json:"action"`
	Details      string    `json:"details,omitempty"`
	IP           string    `json:"ip,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
}

// AuditPage is a paginated audit log slice.
type AuditPage struct {
	Items  []AuditEntry `json:"items"`
	Total  int64        `json:"total"`
	Limit  int          `json:"limit"`
	Offset int          `json:"offset"`
}
