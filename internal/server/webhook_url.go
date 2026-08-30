package server

import (
	"fmt"
	"net/url"
	"strings"
)

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
	return nil
}
