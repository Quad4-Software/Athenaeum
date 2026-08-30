package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"
)

// webhookAllowLocal permits loopback/private webhook targets (tests only).
var webhookAllowLocal bool

func validateWebhookURL(raw string) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fmt.Errorf("url is required")
	}
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("invalid url")
	}
	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return fmt.Errorf("url must be an http(s) endpoint")
	}
	if u.Host == "" {
		return fmt.Errorf("url host is required")
	}
	host := u.Hostname()
	if host == "" {
		return fmt.Errorf("url host is required")
	}
	if looksLikeMetadataHost(host) {
		return fmt.Errorf("url host is not allowed")
	}
	ips, err := net.LookupIP(host)
	if err != nil {
		// Literal IPs still parse via LookupIP in Go.
		if ip := net.ParseIP(host); ip != nil {
			if isBlockedWebhookIP(ip) {
				return fmt.Errorf("url host is not allowed")
			}
			return nil
		}
		return fmt.Errorf("url host could not be resolved")
	}
	if len(ips) == 0 {
		return fmt.Errorf("url host could not be resolved")
	}
	if slices.ContainsFunc(ips, isBlockedWebhookIP) {
		return fmt.Errorf("url host is not allowed")
	}
	return nil
}

func looksLikeMetadataHost(host string) bool {
	h := strings.ToLower(host)
	switch h {
	case "metadata.google.internal", "metadata.google", "instance-data":
		return true
	}
	return strings.HasSuffix(h, ".internal") && strings.Contains(h, "metadata")
}

func isBlockedWebhookIP(ip net.IP) bool {
	if webhookAllowLocal {
		return false
	}
	if ip == nil {
		return true
	}
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return true
	}
	if ip.IsUnspecified() || ip.IsMulticast() {
		return true
	}
	// Carrier-grade NAT / documentation / benchmarking ranges often used in SSRF labs.
	if ip4 := ip.To4(); ip4 != nil {
		if ip4[0] == 169 && ip4[1] == 254 {
			return true
		}
		if ip4[0] == 100 && ip4[1] >= 64 && ip4[1] <= 127 {
			return true
		}
	}
	return false
}

func webhookHTTPClient() *http.Client {
	dialer := &net.Dialer{Timeout: 5 * time.Second}
	transport := &http.Transport{
		Proxy: nil,
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			host, port, err := net.SplitHostPort(addr)
			if err != nil {
				return nil, err
			}
			ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
			if err != nil {
				return nil, err
			}
			var lastErr error
			for _, ipa := range ips {
				if isBlockedWebhookIP(ipa.IP) {
					lastErr = fmt.Errorf("url host is not allowed")
					continue
				}
				conn, err := dialer.DialContext(ctx, network, net.JoinHostPort(ipa.IP.String(), port))
				if err == nil {
					return conn, nil
				}
				lastErr = err
			}
			if lastErr == nil {
				lastErr = fmt.Errorf("url host is not allowed")
			}
			return nil, lastErr
		},
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          4,
		IdleConnTimeout:       30 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	return &http.Client{
		Timeout:   webhookHTTPTimeout,
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 3 {
				return fmt.Errorf("too many redirects")
			}
			return validateWebhookURL(req.URL.String())
		},
	}
}
