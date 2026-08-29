package library

import (
	"net"
	"net/url"
	"strings"
	"testing"
)

func FuzzBlockedIP(f *testing.F) {
	f.Add(byte(127), byte(0), byte(0), byte(1))
	f.Add(byte(8), byte(8), byte(8), byte(8))
	f.Add(byte(10), byte(0), byte(0), byte(1))
	f.Add(byte(192), byte(168), byte(1), byte(1))
	f.Add(byte(169), byte(254), byte(1), byte(1))
	f.Add(byte(0), byte(0), byte(0), byte(0))

	f.Fuzz(func(t *testing.T, a, b, c, d byte) {
		ip := net.IPv4(a, b, c, d)
		got := blockedIP(ip)
		category := ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() ||
			ip.IsLinkLocalMulticast() || ip.IsUnspecified()
		if got != category {
			t.Fatalf("blockedIP(%v) = %v, category=%v", ip, got, category)
		}
	})
}

func FuzzValidateOutboundURLLiteral(f *testing.F) {
	seeds := []string{
		"",
		"not-a-url",
		"ftp://example.com/",
		"http://127.0.0.1/",
		"https://8.8.8.8/",
		"http://10.0.0.1/",
		"http://192.168.0.1/x",
		"http://[::1]/",
		"https://169.254.169.254/latest",
		"http://localhost/",
		"https://example.localhost/",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, raw string) {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			return
		}
		parsed, err := url.Parse(raw)
		if err != nil {
			return
		}
		host := strings.TrimSuffix(strings.ToLower(parsed.Hostname()), ".")
		if host == "" {
			return
		}
		// Only exercise the literal-IP branch. Hostname inputs would call the
		// system resolver and hang fuzz workers on DNS timeouts.
		ip := net.ParseIP(host)
		if ip == nil {
			if host == "localhost" || strings.HasSuffix(host, ".localhost") {
				if _, err := ValidateOutboundURL(raw); err == nil {
					t.Fatalf("accepted localhost hostname %q", raw)
				}
			}
			return
		}

		u, err := ValidateOutboundURL(raw)
		if err != nil {
			return
		}
		gotHost := strings.TrimSuffix(strings.ToLower(u.Hostname()), ".")
		gotIP := net.ParseIP(gotHost)
		if gotIP == nil {
			t.Fatalf("accepted non-literal host %q from %q", gotHost, raw)
		}
		if blockedIP(gotIP) {
			t.Fatalf("accepted blocked literal %v from %q", gotIP, raw)
		}
	})
}
