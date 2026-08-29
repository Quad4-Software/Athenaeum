package models

import "time"

// APIKey is a stored API key without the secret value.
type APIKey struct {
	ID         int64      `json:"id"`
	UserID     int64      `json:"userId"`
	Name       string     `json:"name"`
	Prefix     string     `json:"prefix"`
	CreatedAt  time.Time  `json:"createdAt"`
	LastUsedAt *time.Time `json:"lastUsedAt,omitempty"`
}

// APIKeyCreated is returned once when a new API key is generated.
type APIKeyCreated struct {
	APIKey
	Key string `json:"key"`
}
