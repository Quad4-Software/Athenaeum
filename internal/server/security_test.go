package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSecurityHeadersEmbeddableFileRoute(t *testing.T) {
	srv, _ := testServer(t)
	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/books/1/file", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get("X-Frame-Options"); got != "SAMEORIGIN" {
		t.Fatalf("file route X-Frame-Options = %q, want SAMEORIGIN", got)
	}
	if got := rec.Header().Get("Content-Security-Policy"); got != "frame-ancestors 'self'" {
		t.Fatalf("file route CSP = %q, want frame-ancestors 'self'", got)
	}
}

func TestSecurityHeadersSPABlocksFraming(t *testing.T) {
	srv, _ := testServer(t)
	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get("X-Frame-Options"); got != "DENY" {
		t.Fatalf("spa X-Frame-Options = %q, want DENY", got)
	}
	csp := rec.Header().Get("Content-Security-Policy")
	if csp == "" || !strings.Contains(csp, "frame-ancestors 'none'") {
		t.Fatalf("spa CSP = %q, want frame-ancestors 'none'", csp)
	}
	if !strings.Contains(csp, "frame-src 'self' blob:") {
		t.Fatalf("spa CSP = %q, want frame-src 'self' blob:", csp)
	}
	if !strings.Contains(csp, "style-src 'self' 'unsafe-inline' blob:") {
		t.Fatalf("spa CSP = %q, want style-src blob:", csp)
	}
	if !strings.Contains(csp, "script-src 'self' 'wasm-unsafe-eval'") {
		t.Fatalf("spa CSP = %q, want script-src wasm-unsafe-eval for PDF.js ICC", csp)
	}
	if !strings.Contains(csp, "https://huggingface.co") || !strings.Contains(csp, "https://*.huggingface.co") {
		t.Fatalf("spa CSP = %q, want Hugging Face hosts for Kokoro WASM models", csp)
	}
}
