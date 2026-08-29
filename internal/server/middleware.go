package server

import (
	"encoding/base64"
	"errors"
	"net/http"
	"strings"

	"athenaeum/internal/auth"
	"athenaeum/internal/models"
)

var errUnauthorized = errors.New("authentication required")

func (s *Server) withAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if user, ok := s.userFromRequest(r); ok {
			ctx = WithUser(ctx, user)
			r = r.WithContext(ctx)
		} else if r.URL.Path != "/api/auth/refresh" {
			if user, ok := s.trySilentRefresh(w, r); ok {
				ctx = WithUser(ctx, user)
				r = r.WithContext(ctx)
			}
		}

		required, err := s.store.AuthRequired(ctx)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		if required && !isPublicRoute(r.Method, r.URL.Path) {
			if _, ok := UserFromContext(ctx); !ok {
				writeErrorReq(w, r, http.StatusUnauthorized, errUnauthorized)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) userFromRequest(r *http.Request) (models.User, bool) {
	if c, err := r.Cookie(auth.SessionCookie); err == nil && c.Value != "" {
		u, err := s.store.SessionUser(r.Context(), c.Value)
		if err == nil {
			return u, true
		}
	}
	if key, ok := auth.ParseAPIKey(r.Header.Get("Authorization"), r.Header.Get("X-API-Key")); ok {
		u, _, err := s.store.UserFromAPIKey(r.Context(), key)
		if err == nil {
			return u, true
		}
	}
	if u, ok := s.userFromBasicAuth(r); ok {
		return u, true
	}
	return models.User{}, false
}

func (s *Server) userFromBasicAuth(r *http.Request) (models.User, bool) {
	h := r.Header.Get("Authorization")
	if !strings.HasPrefix(h, "Basic ") {
		return models.User{}, false
	}
	raw, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(h, "Basic "))
	if err != nil {
		return models.User{}, false
	}
	parts := strings.SplitN(string(raw), ":", 2)
	if len(parts) != 2 {
		return models.User{}, false
	}
	u, hash, err := s.store.GetUserByUsername(r.Context(), parts[0])
	if err != nil {
		return models.User{}, false
	}
	if !auth.CheckPassword(hash, parts[1]) || hash == "" {
		return models.User{}, false
	}
	return u, true
}

// isSPARoute reports paths served by the frontend SPA. These stay
// reachable without a session so the client can run setup/login flows.
func isSPARoute(path string) bool {
	if strings.HasPrefix(path, "/api/") {
		return false
	}
	if path == "/opds" || strings.HasPrefix(path, "/opds/") {
		return false
	}
	if path == "/metrics" {
		return false
	}
	if path == "/docs" || strings.HasPrefix(path, "/docs/") {
		return false
	}
	return true
}

func requireAdmin(w http.ResponseWriter, r *http.Request) (models.User, bool) {
	u, ok := UserFromContext(r.Context())
	if !ok {
		writeErrorReq(w, r, http.StatusUnauthorized, errUnauthorized)
		return models.User{}, false
	}
	if !u.IsAdmin {
		writeErrorReq(w, r, http.StatusForbidden, errForbidden)
		return models.User{}, false
	}
	return u, true
}

func (s *Server) requirePermission(w http.ResponseWriter, r *http.Request, perm int64) (models.User, bool) {
	required, err := s.store.AuthRequired(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return models.User{}, false
	}
	if !required {
		return models.User{}, true
	}
	u, ok := UserFromContext(r.Context())
	if !ok {
		writeErrorReq(w, r, http.StatusUnauthorized, errUnauthorized)
		return models.User{}, false
	}
	if models.HasPermission(models.EffectivePermissions(u), perm) {
		return u, true
	}
	writeErrorReq(w, r, http.StatusForbidden, errForbidden)
	return models.User{}, false
}
