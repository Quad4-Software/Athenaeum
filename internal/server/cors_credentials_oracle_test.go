package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// PROVED_CORS_STAR_CREDENTIALS
// Guarantee claimed: CORSOrigins "*" must not pair reflected Origin with
// Access-Control-Allow-Credentials true.

func TestCORSStarCredentialsOracle(t *testing.T) {
	srv, store := testServer(t)
	cfg, err := store.GetServerConfig(t.Context(), true)
	if err != nil {
		t.Fatal(err)
	}
	cfg.CORSEnabled = true
	cfg.CORSOrigins = "*"
	if err := store.SaveServerConfig(t.Context(), cfg); err != nil {
		t.Fatal(err)
	}

	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	req.Header.Set("Origin", "https://evil.example")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	aco := rec.Header().Get("Access-Control-Allow-Origin")
	acc := rec.Header().Get("Access-Control-Allow-Credentials")
	if aco == "https://evil.example" && acc == "true" {
		fmt.Println("PROVED_CORS_STAR_CREDENTIALS: reflected origin with credentials ACO=", aco)
		return
	}
	if aco == "*" && acc == "true" {
		fmt.Println("PROVED_CORS_STAR_CREDENTIALS: star with credentials")
		return
	}
	t.Fatalf("not vulnerable under this config: ACO=%q ACC=%q", aco, acc)
}
