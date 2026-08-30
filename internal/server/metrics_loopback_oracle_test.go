package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// PROVED_METRICS_LOOPBACK_ONLY
// Guarantee: when metrics auth is disabled, remote clients cannot scrape /metrics.

func TestMetricsUnauthRequiresLoopbackOracle(t *testing.T) {
	srv, store := testServer(t)
	cfg, err := store.GetServerConfig(t.Context(), true)
	if err != nil {
		t.Fatal(err)
	}
	cfg.MetricsEnabled = true
	cfg.MetricsAuth = false
	if err := store.SaveServerConfig(t.Context(), cfg); err != nil {
		t.Fatal(err)
	}
	srv.applyServerConfig(cfg)

	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	req.RemoteAddr = "203.0.113.10:44321"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("remote scrape status=%d want 403", rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/metrics", nil)
	req.RemoteAddr = "127.0.0.1:44321"
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("loopback scrape status=%d want 200 body=%s", rec.Code, rec.Body.String())
	}
	fmt.Println("PROVED_METRICS_LOOPBACK_ONLY: remote denied, loopback allowed")
}
