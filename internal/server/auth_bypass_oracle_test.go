package server

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"athenaeum/internal/auth"
	"athenaeum/internal/models"
)

// PROVED_BASIC_AUTH_TOTP_BYPASS
// Guarantee: HTTP Basic must not authenticate users with TOTP enabled
// (password-only login requires a TOTP challenge).

func TestBasicAuthTOTPBypassOracle(t *testing.T) {
	srv, store := testServer(t)
	ctx := context.Background()
	hash, err := auth.HashPassword("longpassword")
	if err != nil {
		t.Fatal(err)
	}
	id, err := store.CreateUser(ctx, "totpuser", hash, false)
	if err != nil {
		t.Fatal(err)
	}
	secret, _, err := auth.GenerateTOTPSecret("totpuser")
	if err != nil {
		t.Fatal(err)
	}
	if err := store.SetUserTOTPSecret(ctx, id, secret); err != nil {
		t.Fatal(err)
	}
	if err := store.EnableUserTOTP(ctx, id); err != nil {
		t.Fatal(err)
	}

	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/api/books", nil)
	cred := base64.StdEncoding.EncodeToString([]byte("totpuser:longpassword"))
	req.Header.Set("Authorization", "Basic "+cred)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code == http.StatusOK {
		t.Fatalf("Basic auth bypassed TOTP status=%d", rec.Code)
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 got %d body=%s", rec.Code, rec.Body.String())
	}
	fmt.Println("PROVED_BASIC_AUTH_TOTP_BYPASS: Basic denied when TOTP enabled")
}

// PROVED_KOSYNC_TOTP_BYPASS
// Guarantee: kosync Basic auth must refuse TOTP-enabled accounts.

func TestKosyncTOTPBypassOracle(t *testing.T) {
	srv, store := testServer(t)
	ctx := context.Background()
	hash, err := auth.HashPassword("longpassword")
	if err != nil {
		t.Fatal(err)
	}
	id, err := store.CreateUser(ctx, "kosyncuser", hash, false)
	if err != nil {
		t.Fatal(err)
	}
	secret, _, err := auth.GenerateTOTPSecret("kosyncuser")
	if err != nil {
		t.Fatal(err)
	}
	if err := store.SetUserTOTPSecret(ctx, id, secret); err != nil {
		t.Fatal(err)
	}
	if err := store.EnableUserTOTP(ctx, id); err != nil {
		t.Fatal(err)
	}

	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/kosync/users/auth", nil)
	cred := base64.StdEncoding.EncodeToString([]byte("kosyncuser:longpassword"))
	req.Header.Set("Authorization", "Basic "+cred)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code == http.StatusOK {
		t.Fatalf("kosync Basic bypassed TOTP")
	}
	fmt.Println("PROVED_KOSYNC_TOTP_BYPASS: kosync denied when TOTP enabled status=", rec.Code)
}

// PROVED_TTS_REQUIRES_AUTH
// Guarantee: TTS synthesize is not reachable anonymously when accounts exist.

func TestTTSRequiresAuthOracle(t *testing.T) {
	srv, store := testServer(t)
	ctx := context.Background()
	hash, _ := auth.HashPassword("longpassword")
	if _, err := store.CreateUser(ctx, "admin", hash, true); err != nil {
		t.Fatal(err)
	}
	if err := store.SaveTTSSettings(ctx, models.TTSSettings{
		Enabled: true, BaseURL: "http://127.0.0.1:9", DefaultVoice: "af_heart", TimeoutSec: 5,
	}); err != nil {
		t.Fatal(err)
	}
	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/tts/synthesize", bytes.NewReader([]byte(`{"text":"hi"}`)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code == http.StatusOK {
		t.Fatalf("anonymous TTS synthesize succeeded")
	}
	if rec.Code != http.StatusUnauthorized && rec.Code != http.StatusForbidden {
		t.Fatalf("expected deny for anonymous TTS got %d", rec.Code)
	}
	fmt.Println("PROVED_TTS_REQUIRES_AUTH: anonymous synthesize denied status=", rec.Code)
}
