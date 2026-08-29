package server

import (
	"net/http"
	"strings"
)

// publicAnyMethod lists exact paths that skip auth for every HTTP method.
var publicAnyMethod = map[string]struct{}{
	"/api/health": {},
}

// publicExact lists method+path pairs that skip auth when auth is required.
var publicExact = map[string]struct{}{
	http.MethodGet + " /api/auth/csrf":             {},
	http.MethodPost + " /api/auth/logout":          {},
	http.MethodPost + " /api/auth/login":           {},
	http.MethodPost + " /api/auth/refresh":         {},
	http.MethodPost + " /api/auth/password/check":  {},
	http.MethodGet + " /api/auth/setup":            {},
	http.MethodPost + " /api/auth/setup":           {},
	http.MethodGet + " /api/auth/methods":          {},
	http.MethodPost + " /api/auth/register-public": {},
	http.MethodPost + " /api/auth/totp/verify":     {},
	http.MethodGet + " /api/altcha/challenge":      {},
	http.MethodGet + " /api/docs":                  {},
	http.MethodGet + " /api/openapi.json":          {},
	http.MethodGet + " /docs":                      {},
	http.MethodGet + " /docs/app.js":               {},
	http.MethodGet + " /auth/oidc/login":           {},
	http.MethodGet + " /auth/oidc/callback":        {},
	http.MethodGet + " /api/tts/status":            {},
	http.MethodGet + " /api/tts/voices":            {},
	http.MethodPost + " /api/tts/synthesize":       {},
	http.MethodGet + " /metrics":                   {},
}

// publicPrefix is a path prefix that is public for a specific method,
// or for any method when Method is empty.
type publicPrefix struct {
	Prefix string
	Method string
}

var publicPrefixes = []publicPrefix{
	{Prefix: "/api/i18n/", Method: http.MethodGet},
	{Prefix: "/api/share/", Method: http.MethodGet},
	{Prefix: "/share/", Method: ""},
	{Prefix: "/kosync/", Method: ""},
	{Prefix: "/api/invite/", Method: ""},
}

func isPublicRoute(method, path string) bool {
	if isSPARoute(path) {
		return true
	}
	if _, ok := publicAnyMethod[path]; ok {
		return true
	}
	if _, ok := publicExact[method+" "+path]; ok {
		return true
	}
	for _, p := range publicPrefixes {
		if !strings.HasPrefix(path, p.Prefix) {
			continue
		}
		if p.Method == "" || p.Method == method {
			return true
		}
	}
	return false
}
