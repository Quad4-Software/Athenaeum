package models

import "time"

// UserSession is an active login session with client metadata.
type UserSession struct {
	ID         string    `json:"id"`
	UserID     int64     `json:"userId,omitempty"`
	IP         string    `json:"ip,omitempty"`
	UserAgent  string    `json:"userAgent,omitempty"`
	Device     string    `json:"device,omitempty"`
	AuthMethod string    `json:"authMethod"`
	CreatedAt  time.Time `json:"createdAt"`
	LastSeenAt time.Time `json:"lastSeenAt"`
	ExpiresAt  time.Time `json:"expiresAt"`
	Current    bool      `json:"current"`
}

// SessionCreate bundles access and refresh tokens for one browser login.
type SessionCreate struct {
	SessionID      string
	AccessToken    string
	RefreshToken   string
	UserID         int64
	IP             string
	UserAgent      string
	Device         string
	AuthMethod     string
	AccessExpires  time.Time
	RefreshExpires time.Time
}
