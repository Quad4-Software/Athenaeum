package auth

import "testing"

func TestParseAPIKey(t *testing.T) {
	key := "ath_" + "aabbccddeeff00112233445566778899aabbccddeeff00112233445566778899"
	tests := []struct {
		auth, xkey string
		want       string
		ok         bool
	}{
		{"Bearer " + key, "", key, true},
		{"bearer " + key, "", "", false},
		{"", key, key, true},
		{"Basic dXNlcjpwYXNz", "", "", false},
		{"Bearer wrong", "", "", false},
	}
	for _, tc := range tests {
		got, ok := ParseAPIKey(tc.auth, tc.xkey)
		if ok != tc.ok || got != tc.want {
			t.Errorf("ParseAPIKey(%q, %q) = %q, %v; want %q, %v", tc.auth, tc.xkey, got, ok, tc.want, tc.ok)
		}
	}
}

func TestAPIKeyHashRoundTrip(t *testing.T) {
	full, prefix, hash, err := NewAPIKey()
	if err != nil {
		t.Fatal(err)
	}
	if len(prefix) != APIKeyLookupLen {
		t.Fatalf("prefix len=%d", len(prefix))
	}
	if !CheckAPIKey(hash, full) {
		t.Fatal("expected match")
	}
	if CheckAPIKey(hash, full+"x") {
		t.Fatal("expected mismatch")
	}
}
