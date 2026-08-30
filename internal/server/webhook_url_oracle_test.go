package server

import (
	"fmt"
	"testing"
)

// PROVED_WEBHOOK_URL_SCHEME
// Guarantee: webhook URLs must parse as http(s) with a host, and must not
// target loopback/private/link-local addresses.

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
		{"http://127.0.0.1/hook", true},
		{"http://localhost/hook", true},
		{"http://169.254.169.254/latest", true},
		{"http://10.0.0.1/hook", true},
		{"http://[::1]/hook", true},
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
	fmt.Println("PROVED_WEBHOOK_URL_SCHEME: scheme/host/private-IP validation enforced")
}
