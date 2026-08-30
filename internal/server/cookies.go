package server

import (
	"net/http"
	"strings"
	"time"

	"athenaeum/internal/auth"
)

func (s *Server) requestSecure(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}
	if s.proxyTrusted(r) && strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https") {
		return true
	}
	return false
}

func (s *Server) sessionCookie(r *http.Request, token string, expires time.Time) *http.Cookie {
	// #nosec G124 -- Secure follows requestSecure() so local HTTP works and HTTPS sets Secure.
	return &http.Cookie{
		Name:     auth.SessionCookie,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   s.requestSecure(r),
		Expires:  expires,
	}
}

func (s *Server) clearSessionCookie(r *http.Request) *http.Cookie {
	// #nosec G124 -- Secure follows requestSecure() so local HTTP works and HTTPS sets Secure.
	return &http.Cookie{
		Name:     auth.SessionCookie,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   s.requestSecure(r),
		MaxAge:   -1,
	}
}

func (s *Server) refreshCookie(r *http.Request, token string, expires time.Time) *http.Cookie {
	// #nosec G124 -- Secure follows requestSecure() so local HTTP works and HTTPS sets Secure.
	return &http.Cookie{
		Name:     auth.RefreshCookie,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   s.requestSecure(r),
		Expires:  expires,
	}
}

func (s *Server) clearRefreshCookie(r *http.Request) *http.Cookie {
	// #nosec G124 -- Secure follows requestSecure() so local HTTP works and HTTPS sets Secure.
	return &http.Cookie{
		Name:     auth.RefreshCookie,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   s.requestSecure(r),
		MaxAge:   -1,
	}
}

func (s *Server) csrfCookie(r *http.Request, token string) *http.Cookie {
	// #nosec G124 -- Secure follows requestSecure() so local HTTP works and HTTPS sets Secure.
	return &http.Cookie{
		Name:     auth.CSRFCookie,
		Value:    token,
		Path:     "/",
		HttpOnly: false,
		SameSite: http.SameSiteStrictMode,
		Secure:   s.requestSecure(r),
		MaxAge:   int((24 * time.Hour).Seconds()),
	}
}

func (s *Server) clearCSRFCookie(r *http.Request) *http.Cookie {
	// #nosec G124 -- Secure follows requestSecure() so local HTTP works and HTTPS sets Secure.
	return &http.Cookie{
		Name:     auth.CSRFCookie,
		Value:    "",
		Path:     "/",
		HttpOnly: false,
		SameSite: http.SameSiteStrictMode,
		Secure:   s.requestSecure(r),
		MaxAge:   -1,
	}
}

func (s *Server) clearSessionCookies(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, s.clearSessionCookie(r))
	http.SetCookie(w, s.clearRefreshCookie(r))
}

func (s *Server) clearAuthCookies(w http.ResponseWriter, r *http.Request) {
	s.clearSessionCookies(w, r)
	http.SetCookie(w, s.clearCSRFCookie(r))
}

// clearSessionCookiesUnlessSessionValid drops session/refresh cookies after a
// failed refresh without touching CSRF. Clearing CSRF here breaks login: the
// SPA plants a CSRF token then calls /api/auth/me, which triggers refresh and
// would otherwise wipe the token the login POST still has cached.
func (s *Server) clearSessionCookiesUnlessSessionValid(w http.ResponseWriter, r *http.Request) {
	if s.hasValidSessionCookie(r) {
		return
	}
	s.clearSessionCookies(w, r)
}

func (s *Server) hasValidSessionCookie(r *http.Request) bool {
	c, err := r.Cookie(auth.SessionCookie)
	if err != nil || c.Value == "" {
		return false
	}
	_, err = s.store.SessionUser(r.Context(), c.Value)
	return err == nil
}

func (s *Server) setAuthCookies(w http.ResponseWriter, r *http.Request, accessToken, refreshToken string, accessExp, refreshExp time.Time) {
	http.SetCookie(w, s.sessionCookie(r, accessToken, accessExp))
	http.SetCookie(w, s.refreshCookie(r, refreshToken, refreshExp))
}
