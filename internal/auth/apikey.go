package auth

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"strings"

	"athenaeum/internal/brand"
)

const (
	// APIKeyPrefix identifies API keys in Authorization headers.
	APIKeyPrefix = brand.APIKeyPrefix
	// APIKeyLookupLen is how many characters of the key are stored for lookup.
	APIKeyLookupLen    = 12
	legacyAPIKeyPrefix = "rdr_"
)

func isAPIKey(token string) bool {
	return strings.HasPrefix(token, APIKeyPrefix) || strings.HasPrefix(token, legacyAPIKeyPrefix)
}

// NewAPIKey returns a new API key, its lookup prefix, and SHA-256 hash.
func NewAPIKey() (full, prefix, hash string, err error) {
	token, err := NewToken()
	if err != nil {
		return "", "", "", err
	}
	full = APIKeyPrefix + token
	if len(full) < APIKeyLookupLen {
		return "", "", "", fmt.Errorf("api key too short")
	}
	prefix = full[:APIKeyLookupLen]
	hash = HashAPIKey(full)
	return full, prefix, hash, nil
}

// HashAPIKey returns the SHA-256 hex digest of an API key.
func HashAPIKey(key string) string {
	sum := sha256.Sum256([]byte(key))
	return hex.EncodeToString(sum[:])
}

// CheckAPIKey reports whether key matches the stored hash.
func CheckAPIKey(hash, key string) bool {
	return subtle.ConstantTimeCompare([]byte(hash), []byte(HashAPIKey(key))) == 1
}

// ParseAPIKey extracts an API key from Authorization Bearer or X-API-Key.
func ParseAPIKey(rHeader, xHeader string) (string, bool) {
	if x := strings.TrimSpace(xHeader); x != "" {
		if isAPIKey(x) {
			return x, true
		}
	}
	h := strings.TrimSpace(rHeader)
	if after, ok := strings.CutPrefix(h, "Bearer "); ok {
		token := strings.TrimSpace(after)
		if isAPIKey(token) {
			return token, true
		}
	}
	return "", false
}
