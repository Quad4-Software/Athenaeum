package server

import (
	"crypto/subtle"
	"errors"
	"net/http"
	"strings"

	"athenaeum/internal/auth"
)

var errCSRF = errors.New("csrf token missing or invalid")

func (s *Server) withCSRF(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if needsCSRF(r) && !s.usesExternalAuth(r) {
			if err := validateCSRF(r); err != nil {
				writeError(w, http.StatusForbidden, errCSRF)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func needsCSRF(r *http.Request) bool {
	switch r.Method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
	default:
		return false
	}
	return strings.HasPrefix(r.URL.Path, "/api/")
}

func validateCSRF(r *http.Request) error {
	c, err := r.Cookie(auth.CSRFCookie)
	if err != nil || c.Value == "" {
		return errCSRF
	}
	header := r.Header.Get(auth.CSRFHeader)
	if header == "" {
		return errCSRF
	}
	if subtle.ConstantTimeCompare([]byte(header), []byte(c.Value)) != 1 {
		return errCSRF
	}
	return nil
}

func (s *Server) usesExternalAuth(r *http.Request) bool {
	if strings.HasPrefix(r.Header.Get("Authorization"), "Basic ") {
		return true
	}
	_, ok := auth.ParseAPIKey(r.Header.Get("Authorization"), r.Header.Get("X-API-Key"))
	return ok
}

func (s *Server) ensureCSRFCookie(w http.ResponseWriter, r *http.Request) (string, error) {
	if c, err := r.Cookie(auth.CSRFCookie); err == nil && c.Value != "" {
		return c.Value, nil
	}
	token, err := auth.NewToken()
	if err != nil {
		return "", err
	}
	http.SetCookie(w, s.csrfCookie(r, token))
	return token, nil
}

func (s *Server) rotateCSRFCookie(w http.ResponseWriter, r *http.Request) (string, error) {
	token, err := auth.NewToken()
	if err != nil {
		return "", err
	}
	http.SetCookie(w, s.csrfCookie(r, token))
	return token, nil
}

// issueCSRFForAuthPages plants a CSRF cookie on public auth GETs so login,
// setup, and registration can submit with a double-submit token immediately.
func (s *Server) issueCSRFForAuthPages(w http.ResponseWriter, r *http.Request) {
	if _, err := s.ensureCSRFCookie(w, r); err != nil {
		s.log.Debug("csrf cookie issue failed", "err", err)
	}
}
