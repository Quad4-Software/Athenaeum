package library

import (
	"testing"
)

func TestValidateOutboundURL(t *testing.T) {
	allowed := []string{
		"https://example.com/cover.jpg",
		"http://books.google.com/thumb.jpg",
	}
	for _, raw := range allowed {
		if _, err := ValidateOutboundURL(raw); err != nil {
			t.Errorf("expected %q allowed, got %v", raw, err)
		}
	}

	blocked := []string{
		"http://127.0.0.1/secret",
		"http://localhost/admin",
		"http://[::1]/",
		"http://10.0.0.1/internal",
		"http://192.168.1.1/",
		"http://169.254.169.254/latest/meta-data",
		"ftp://example.com/x",
		"",
		"not-a-url",
	}
	for _, raw := range blocked {
		if _, err := ValidateOutboundURL(raw); err == nil {
			t.Errorf("expected %q blocked", raw)
		}
	}
}

func TestFetchCoverImageBlocksPrivateURL(t *testing.T) {
	_, err := FetchCoverImage(t.Context(), "http://127.0.0.1/health")
	if err == nil {
		t.Fatal("expected private url to be rejected")
	}
}
