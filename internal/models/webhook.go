package models

import "time"

// Webhook event names for v1.
const (
	WebhookEventUserCreate          = "user.create"
	WebhookEventUserDelete          = "user.delete"
	WebhookEventInviteCreated       = "invite.created"
	WebhookEventInviteAccepted      = "invite.accepted"
	WebhookEventBookUpload          = "book.upload"
	WebhookEventLibraryScanComplete = "library.scan.complete"
	WebhookEventPing                = "ping"
)

// WebhookEventsV1 is the set of events admins can subscribe to.
var WebhookEventsV1 = []string{
	WebhookEventUserCreate,
	WebhookEventUserDelete,
	WebhookEventInviteCreated,
	WebhookEventInviteAccepted,
	WebhookEventBookUpload,
	WebhookEventLibraryScanComplete,
}

// Webhook is an outbound event subscription.
type Webhook struct {
	ID        int64     `json:"id"`
	URL       string    `json:"url"`
	Secret    string    `json:"secret,omitempty"`
	Events    []string  `json:"events"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// WebhookPublic masks the signing secret.
type WebhookPublic struct {
	ID        int64     `json:"id"`
	URL       string    `json:"url"`
	SecretSet bool      `json:"secretSet"`
	Events    []string  `json:"events"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Public strips the secret for API responses.
func (w Webhook) Public() WebhookPublic {
	events := w.Events
	if events == nil {
		events = []string{}
	}
	return WebhookPublic{
		ID:        w.ID,
		URL:       w.URL,
		SecretSet: w.Secret != "",
		Events:    events,
		Enabled:   w.Enabled,
		CreatedAt: w.CreatedAt,
		UpdatedAt: w.UpdatedAt,
	}
}

// WebhookDelivery is one delivery attempt record.
type WebhookDelivery struct {
	ID          int64      `json:"id"`
	WebhookID   int64      `json:"webhookId"`
	Event       string     `json:"event"`
	Payload     string     `json:"payload"`
	StatusCode  int        `json:"statusCode"`
	Success     bool       `json:"success"`
	Attempts    int        `json:"attempts"`
	LastError   string     `json:"lastError,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
	DeliveredAt *time.Time `json:"deliveredAt,omitempty"`
}
