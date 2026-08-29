package auth

import (
	"strings"
	"testing"
)

func FuzzHashAPIKey(f *testing.F) {
	f.Add("")
	f.Add("ath_testkey")
	f.Add(APIKeyPrefix + "abcdef")
	f.Add(strings.Repeat("k", 256))

	f.Fuzz(func(t *testing.T, key string) {
		h1 := HashAPIKey(key)
		h2 := HashAPIKey(key)
		if h1 != h2 {
			t.Fatal("hash not deterministic")
		}
		if len(h1) != 64 {
			t.Fatalf("sha256 hex len=%d", len(h1))
		}
		if !CheckAPIKey(h1, key) {
			t.Fatal("CheckAPIKey failed for matching key")
		}
		other := key + "\x00"
		if other != key && CheckAPIKey(h1, other) {
			t.Fatal("CheckAPIKey matched different key")
		}
	})
}

func FuzzParseAPIKey(f *testing.F) {
	f.Add("", "")
	f.Add("Bearer "+APIKeyPrefix+"abc", "")
	f.Add("", APIKeyPrefix+"xyz")
	f.Add("Bearer tok", "other")

	f.Fuzz(func(t *testing.T, authHeader, xHeader string) {
		key, ok := ParseAPIKey(authHeader, xHeader)
		if !ok {
			return
		}
		if !isAPIKey(key) {
			t.Fatalf("ParseAPIKey returned non-key %q", key)
		}
	})
}
