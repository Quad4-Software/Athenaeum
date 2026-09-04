package server

import (
	"net/http"
	"strings"

	"athenaeum/internal/telemetry"
)

// defaultCSP covers the SPA plus browser Kokoro TTS (ONNX weights from Hugging Face).
const defaultCSP = "default-src 'self'; script-src 'self' 'wasm-unsafe-eval'; style-src 'self' 'unsafe-inline' blob:; img-src 'self' blob: data: https://covers.openlibrary.org https://books.google.com https://*.googleusercontent.com; font-src 'self' blob: data:; connect-src 'self' https://huggingface.co https://*.huggingface.co https://*.hf.co; media-src 'self' blob:; frame-src 'self' blob:; worker-src 'self' blob:; object-src 'none'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'"

const embeddableCSP = "frame-ancestors 'self'"

func sameOriginEmbeddablePath(path string) bool {
	return strings.HasPrefix(path, "/api/books/") && strings.HasSuffix(path, "/file")
}

func (s *Server) withSecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg := s.currentServerConfig()
		embeddable := sameOriginEmbeddablePath(r.URL.Path)

		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		if embeddable {
			w.Header().Set("X-Frame-Options", "SAMEORIGIN")
		} else {
			w.Header().Set("X-Frame-Options", "DENY")
		}
		if cfg.CSPEnabled {
			if embeddable {
				w.Header().Set("Content-Security-Policy", embeddableCSP)
			} else {
				policy := cfg.CSPPolicy
				if policy == "" {
					policy = defaultCSP
				}
				policy = withSentryConnectSrc(policy, s.cfg.SentryPublicDSN())
				policy = withAltchaConnectSrc(policy, s.altchaPublic().ChallengeURL)
				w.Header().Set("Content-Security-Policy", policy)
			}
		}
		if cfg.CORSEnabled {
			origin := r.Header.Get("Origin")
			if origin != "" && corsOriginAllowed(origin, cfg.CORSOrigins) {
				allowCredentials := strings.TrimSpace(cfg.CORSOrigins) != "*"
				if allowCredentials {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Vary", "Origin")
					w.Header().Set("Access-Control-Allow-Credentials", "true")
				} else {
					w.Header().Set("Access-Control-Allow-Origin", "*")
				}
				w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-API-Key, X-CSRF-Token, Content-Range")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
				if r.Method == http.MethodOptions {
					w.WriteHeader(http.StatusNoContent)
					return
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}

func corsOriginAllowed(origin, allowed string) bool {
	if allowed == "*" {
		return true
	}
	for part := range strings.SplitSeq(allowed, ",") {
		if strings.TrimSpace(part) == origin {
			return true
		}
	}
	return false
}

func withSentryConnectSrc(policy, dsn string) string {
	host := telemetry.ConnectHost(dsn)
	if host == "" {
		return policy
	}
	const needle = "connect-src 'self'"
	if !strings.Contains(policy, needle) {
		return policy
	}
	return strings.Replace(policy, needle, "connect-src 'self' "+host, 1)
}

func withAltchaConnectSrc(policy, challengeURL string) string {
	host := absoluteOrigin(challengeURL)
	if host == "" {
		return policy
	}
	const needle = "connect-src 'self'"
	if !strings.Contains(policy, needle) {
		return policy
	}
	if strings.Contains(policy, host) {
		return policy
	}
	return strings.Replace(policy, needle, "connect-src 'self' "+host, 1)
}

func absoluteOrigin(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" || strings.HasPrefix(raw, "/") {
		return ""
	}
	if !strings.Contains(raw, "://") {
		return ""
	}
	rest := raw
	schemeEnd := strings.Index(rest, "://")
	if schemeEnd < 0 {
		return ""
	}
	rest = rest[schemeEnd+3:]
	slash := strings.IndexByte(rest, '/')
	if slash >= 0 {
		rest = rest[:slash]
	}
	if rest == "" {
		return ""
	}
	return raw[:schemeEnd+3] + rest
}
