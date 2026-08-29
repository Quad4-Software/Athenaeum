package auth

import "testing"

func TestParseDevice(t *testing.T) {
	tests := []struct {
		ua   string
		want string
	}{
		{"", "Unknown"},
		{
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
			"Chrome on Windows",
		},
		{
			"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Safari/605.1.15",
			"Safari on macOS",
		},
	}
	for _, tc := range tests {
		if got := ParseDevice(tc.ua); got != tc.want {
			t.Errorf("ParseDevice(%q) = %q, want %q", tc.ua, got, tc.want)
		}
	}
}
