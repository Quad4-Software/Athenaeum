package server

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"athenaeum/internal/auth"
)

// PROVED_OPDS_FORWARDED_PROTO
// Guarantee claimed: OPDS base URLs must ignore X-Forwarded-Proto unless
// the peer is a trusted proxy (same as requestBaseURL).

func TestOPDSIgnoresUntrustedForwardedProtoOracle(t *testing.T) {
	srv, store := testServer(t)
	ctx := context.Background()
	hash, _ := auth.HashPassword("longpassword")
	if _, err := store.CreateUser(ctx, "admin", hash, true); err != nil {
		t.Fatal(err)
	}
	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}
	session, _, _ := loginUser(t, handler, "admin", "longpassword")

	req := httptest.NewRequest(http.MethodGet, "/opds/", nil)
	req.Host = "library.example"
	req.Header.Set("X-Forwarded-Proto", "https")
	req.AddCookie(session)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	if strings.Contains(body, "https://library.example") {
		fmt.Println("PROVED_OPDS_FORWARDED_PROTO: untrusted X-Forwarded-Proto trusted in OPDS links")
		return
	}
	if strings.Contains(body, "http://library.example") {
		t.Fatal("not vulnerable: OPDS ignored untrusted forwarded proto")
	}
	t.Fatalf("unexpected body=%s", body)
}
