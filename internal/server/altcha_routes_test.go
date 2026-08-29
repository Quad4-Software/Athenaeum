package server

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	altchalib "github.com/altcha-org/altcha-lib-go/v2"

	"athenaeum/internal/auth"
	"athenaeum/internal/config"
	"athenaeum/internal/library"
	"athenaeum/internal/storage"
)

func TestAltchaProtectsLogin(t *testing.T) {
	dir := t.TempDir()
	store, err := storage.Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { store.Close() })

	hash, err := auth.HashPassword("longpassword")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.CreateUser(context.Background(), "admin", hash, true); err != nil {
		t.Fatal(err)
	}

	cfg := config.Config{
		DataDir:             dir,
		LibraryDir:          filepath.Join(dir, "lib"),
		AltchaEnabled:       true,
		AltchaMode:          "builtin",
		AltchaHMACSecret:    "test-hmac",
		AltchaHMACKeySecret: "test-hmac-key",
		AltchaCost:          100,
		AltchaExpiresSecs:   120,
		AltchaProtect:       "login,setup",
	}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	scanner := library.New(store, cfg.CoverDir(), cfg.TempDir(), log, 2)
	srv, err := New(context.Background(), cfg, store, scanner, log)
	if err != nil {
		t.Fatal(err)
	}
	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/altcha/challenge", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("challenge status=%d body=%s", rec.Code, rec.Body.String())
	}
	var challenge altchalib.Challenge
	if err := json.NewDecoder(rec.Body).Decode(&challenge); err != nil {
		t.Fatal(err)
	}

	req = httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(`{"username":"admin","password":"longpassword"}`))
	req.Header.Set("Content-Type", "application/json")
	csrf := fetchCSRF(t, handler)
	withCSRF(req, csrf)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("login without altcha status=%d body=%s", rec.Code, rec.Body.String())
	}

	solution, err := altchalib.SolveChallenge(altchalib.SolveChallengeOptions{
		Challenge: challenge,
		DeriveKey: altchalib.DeriveKeyPBKDF2(),
	})
	if err != nil || solution == nil {
		t.Fatalf("solve: %v", err)
	}
	raw, err := json.Marshal(altchalib.Payload{Challenge: challenge, Solution: *solution})
	if err != nil {
		t.Fatal(err)
	}
	payload := base64.StdEncoding.EncodeToString(raw)
	body, _ := json.Marshal(map[string]string{
		"username": "admin",
		"password": "longpassword",
		"altcha":   payload,
	})
	req = httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	csrf = fetchCSRF(t, handler)
	withCSRF(req, csrf)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("login with altcha status=%d body=%s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/auth/methods", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("methods status=%d", rec.Code)
	}
	var methods map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&methods); err != nil {
		t.Fatal(err)
	}
	altchaCfg, ok := methods["altcha"].(map[string]any)
	if !ok || altchaCfg["enabled"] != true {
		t.Fatalf("methods altcha=%v", methods["altcha"])
	}
}
