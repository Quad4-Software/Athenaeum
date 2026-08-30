package server

import (
	"net/http"
	"strings"
	"testing"
)

func TestPublicRouteExactAllowlist(t *testing.T) {
	cases := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/health"},
		{http.MethodPost, "/api/health"},
		{http.MethodGet, "/api/auth/csrf"},
		{http.MethodPost, "/api/auth/logout"},
		{http.MethodPost, "/api/auth/login"},
		{http.MethodPost, "/api/auth/refresh"},
		{http.MethodPost, "/api/auth/password/check"},
		{http.MethodGet, "/api/auth/setup"},
		{http.MethodPost, "/api/auth/setup"},
		{http.MethodGet, "/api/auth/methods"},
		{http.MethodPost, "/api/auth/register-public"},
		{http.MethodPost, "/api/auth/totp/verify"},
		{http.MethodGet, "/api/altcha/challenge"},
		{http.MethodGet, "/api/docs"},
		{http.MethodGet, "/api/openapi.json"},
		{http.MethodGet, "/docs"},
		{http.MethodGet, "/docs/app.js"},
		{http.MethodGet, "/auth/oidc/login"},
		{http.MethodGet, "/auth/oidc/callback"},
		{http.MethodGet, "/metrics"},
		{http.MethodGet, "/api/i18n/en"},
		{http.MethodGet, "/api/share/abc"},
		{http.MethodGet, "/share/abc"},
		{http.MethodPost, "/share/abc"},
		{http.MethodGet, "/kosync/syncs/progress"},
		{http.MethodPost, "/api/invite/tok"},
		{http.MethodGet, "/api/invite/tok"},
	}
	for _, tc := range cases {
		if !isPublicRoute(tc.method, tc.path) {
			t.Errorf("expected public: %s %s", tc.method, tc.path)
		}
	}
}

func TestPublicRouteProtectedSample(t *testing.T) {
	cases := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/books"},
		{http.MethodPut, "/api/auth/profile"},
		{http.MethodPost, "/api/auth/register"},
		{http.MethodGet, "/api/auth/me"},
		{http.MethodGet, "/api/auth/users"},
		{http.MethodPost, "/api/i18n/en"},
		{http.MethodPost, "/api/share/abc"},
		{http.MethodGet, "/api/tts/status"},
		{http.MethodGet, "/api/tts/voices"},
		{http.MethodPost, "/api/tts/synthesize"},
		{http.MethodDelete, "/api/tts/status"},
	}
	for _, tc := range cases {
		if isPublicRoute(tc.method, tc.path) {
			t.Errorf("expected protected: %s %s", tc.method, tc.path)
		}
	}
}

func TestPublicRouteDocumentedAuthPublic(t *testing.T) {
	for _, ep := range documentedEndpoints() {
		if !strings.EqualFold(ep.Auth, "public") {
			continue
		}
		path := concretePublicPath(ep.Path)
		if !isPublicRoute(ep.Method, path) {
			t.Errorf("documented Auth=public not allowed: %s %s (concrete %s)", ep.Method, ep.Path, path)
		}
	}
}

func concretePublicPath(path string) string {
	out := path
	out = strings.ReplaceAll(out, "{locale}", "en")
	out = strings.ReplaceAll(out, "{token}", "sample-token")
	out = strings.ReplaceAll(out, "{id}", "1")
	return out
}
