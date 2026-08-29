package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"athenaeum/internal/auth"
)

func fetchCSRF(t *testing.T, handler http.Handler) *http.Cookie {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/auth/csrf", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("csrf status=%d body=%s", rec.Code, rec.Body.String())
	}
	for _, c := range rec.Result().Cookies() {
		if c.Name == auth.CSRFCookie {
			return c
		}
	}
	t.Fatal("csrf cookie not set")
	return nil
}

func withCSRF(req *http.Request, csrf *http.Cookie) {
	req.AddCookie(csrf)
	req.Header.Set("X-CSRF-Token", csrf.Value)
}
