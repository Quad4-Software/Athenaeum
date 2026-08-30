package server

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"athenaeum/internal/auth"
	"athenaeum/internal/models"
	"athenaeum/internal/storage"
)

func loginAdmin(t *testing.T, handler http.Handler, store *storage.Store) (*http.Cookie, *http.Cookie) {
	t.Helper()
	ctx := context.Background()
	hash, err := auth.HashPassword("longpassword")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.CreateUser(ctx, "admin", hash, true); err != nil {
		t.Fatal(err)
	}
	csrf := fetchCSRF(t, handler)
	login := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(`{"username":"admin","password":"longpassword"}`))
	login.Header.Set("Content-Type", "application/json")
	withCSRF(login, csrf)
	loginRec := httptest.NewRecorder()
	handler.ServeHTTP(loginRec, login)
	if loginRec.Code != http.StatusOK {
		t.Fatalf("login status=%d body=%s", loginRec.Code, loginRec.Body.String())
	}
	var sessionCookie *http.Cookie
	var loginCSRF *http.Cookie
	for _, c := range loginRec.Result().Cookies() {
		switch c.Name {
		case auth.SessionCookie:
			sessionCookie = c
		case auth.CSRFCookie:
			loginCSRF = c
		}
	}
	if sessionCookie == nil || loginCSRF == nil {
		t.Fatal("missing session or csrf cookie after login")
	}
	return sessionCookie, loginCSRF
}

func TestTTSAdminAndStatus(t *testing.T) {
	srv, store := testServer(t)
	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}
	session, csrf := loginAdmin(t, handler, store)

	req := httptest.NewRequest(http.MethodGet, "/api/tts/status", nil)
	req.AddCookie(session)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status code=%d body=%s", rec.Code, rec.Body.String())
	}
	var status models.TTSStatus
	if err := json.NewDecoder(rec.Body).Decode(&status); err != nil {
		t.Fatal(err)
	}
	if status.Enabled {
		t.Fatal("expected TTS disabled by default")
	}

	body, _ := json.Marshal(map[string]any{
		"enabled":      true,
		"baseUrl":      "http://127.0.0.1:9",
		"defaultVoice": "af_heart",
		"timeoutSec":   10,
	})
	req = httptest.NewRequest(http.MethodPut, "/api/admin/tts", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(session)
	withCSRF(req, csrf)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("put tts status=%d body=%s", rec.Code, rec.Body.String())
	}

	cfg, err := store.GetTTSSettings(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Enabled || cfg.BaseURL != "http://127.0.0.1:9" {
		t.Fatalf("unexpected saved cfg: %+v", cfg)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/tts/status", nil)
	req.AddCookie(session)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if err := json.NewDecoder(rec.Body).Decode(&status); err != nil {
		t.Fatal(err)
	}
	if !status.Enabled {
		t.Fatal("expected TTS enabled after save")
	}
}

func TestTTSSynthesizeRequiresConfig(t *testing.T) {
	srv, _ := testServer(t)
	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}
	csrf := fetchCSRF(t, handler)

	body, _ := json.Marshal(map[string]any{"text": "Hello"})
	req := httptest.NewRequest(http.MethodPost, "/api/tts/synthesize", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	withCSRF(req, csrf)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestTTSSynthesizeProxy(t *testing.T) {
	sidecar := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/health":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"ok":true}`))
		case r.URL.Path == "/voices":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"voices":[{"id":"af_heart","label":"Heart","lang":"en-us"}]}`))
		case r.URL.Path == "/v1/audio/speech":
			raw, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(raw), "Hello") {
				http.Error(w, "bad body", http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "audio/wav")
			_, _ = w.Write([]byte("RIFF....WAVEfmt "))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(sidecar.Close)

	srv, store := testServer(t)
	if err := store.SaveTTSSettings(context.Background(), models.TTSSettings{
		Enabled:      true,
		BaseURL:      sidecar.URL,
		DefaultVoice: "af_heart",
		TimeoutSec:   10,
	}); err != nil {
		t.Fatal(err)
	}

	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}
	session, csrf := loginAdmin(t, handler, store)

	req := httptest.NewRequest(http.MethodGet, "/api/tts/voices", nil)
	req.AddCookie(session)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("voices status=%d body=%s", rec.Code, rec.Body.String())
	}

	body, _ := json.Marshal(map[string]any{"text": "Hello", "voice": "af_heart", "speed": 1})
	req = httptest.NewRequest(http.MethodPost, "/api/tts/synthesize", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(session)
	withCSRF(req, csrf)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("synthesize status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Header().Get("Content-Type"), "audio") {
		t.Fatalf("unexpected content-type %q", rec.Header().Get("Content-Type"))
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("WAVE")) {
		t.Fatalf("unexpected audio body %q", rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, "/api/admin/tts/test", nil)
	req.AddCookie(session)
	withCSRF(req, csrf)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("test status=%d body=%s", rec.Code, rec.Body.String())
	}
	var testRes struct {
		OK bool `json:"ok"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&testRes); err != nil {
		t.Fatal(err)
	}
	if !testRes.OK {
		t.Fatalf("expected ok test: %s", rec.Body.String())
	}
}

func TestTTSPutValidation(t *testing.T) {
	srv, store := testServer(t)
	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}
	session, csrf := loginAdmin(t, handler, store)

	body, _ := json.Marshal(map[string]any{
		"enabled": true,
		"baseUrl": "",
	})
	req := httptest.NewRequest(http.MethodPut, "/api/admin/tts", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(session)
	withCSRF(req, csrf)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code == http.StatusOK {
		t.Fatal("expected validation failure for empty baseUrl")
	}
}
