package server

import (
	"fmt"
	"testing"
)

// PROVED_WEBHOOK_URL_SCHEME
// Guarantee: webhook URLs must parse as http(s) with a host.
// Prefix checks like HasPrefix(url, "http") accept junk such as "httpbogus".

func TestWebhookURLValidationOracle(t *testing.T) {
	cases := []struct {
		url    string
		reject bool
	}{
		{"https://example.com/hook", false},
		{"http://example.com/hook", false},
		{"", true},
		{"ftp://example.com/hook", true},
		{"javascript:alert(1)", true},
		{"http", true},
		{"httpbogus://example.com", true},
		{"https://", true},
	}
	for _, tc := range cases {
		err := validateWebhookURL(tc.url)
		if tc.reject && err == nil {
			t.Fatalf("accepted %q", tc.url)
		}
		if !tc.reject && err != nil {
			t.Fatalf("rejected %q: %v", tc.url, err)
		}
	}
	fmt.Println("PROVED_WEBHOOK_URL_SCHEME: scheme/host validation enforced")
}
