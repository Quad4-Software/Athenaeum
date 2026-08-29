package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

const (
	// SessionCookie stores the short-lived access token (HttpOnly).
	SessionCookie = "athenaeum_session"
	// RefreshCookie stores the long-lived refresh token (HttpOnly).
	RefreshCookie = "athenaeum_refresh"
	// CSRFCookie holds the CSRF token readable by browser JS (not HttpOnly).
	CSRFCookie = "athenaeum_csrf"
	// CSRFHeader is sent by clients alongside the CSRF cookie value.
	CSRFHeader = "X-CSRF-Token"

	// AccessTTL is how long an access session remains valid before refresh.
	AccessTTL = 15 * time.Minute
	// RefreshTTL is how long a refresh token remains valid.
	RefreshTTL = 30 * 24 * time.Hour
)

// NewToken returns a cryptographically random hex token.
func NewToken() (string, error) {
	var b [32]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("token: %w", err)
	}
	return hex.EncodeToString(b[:]), nil
}

// NewSessionToken returns a cryptographically random session token.
func NewSessionToken() (string, error) {
	return NewToken()
}
