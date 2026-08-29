package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"athenaeum/internal/models"
	"athenaeum/internal/pocketid"
)

func mockPocketIDAdmin(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-API-KEY") == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": []any{}})
	})
	mux.HandleFunc("/api/signup-tokens", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []pocketid.SignupToken{{ID: "st1", Token: "tok1", UsageLimit: 1}},
			})
		case http.MethodPost:
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(pocketid.SignupToken{
				ID: "st-new", Token: "fresh", UsageLimit: 1, ExpiresAt: "tomorrow",
			})
		default:
			http.Error(w, "method", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/signup-tokens/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "method", http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
		base := "http://" + r.Host
		_ = json.NewEncoder(w).Encode(map[string]any{
			"issuer":                                base,
			"authorization_endpoint":                base + "/authorize",
			"token_endpoint":                        base + "/token",
			"userinfo_endpoint":                     base + "/userinfo",
			"jwks_uri":                              base + "/jwks",
			"end_session_endpoint":                  base + "/logout",
			"id_token_signing_alg_values_supported": []string{"RS256"},
		})
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func TestPocketIDSettingsAndTest(t *testing.T) {
	srv, store := testServer(t)
	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}
	session, csrf := loginAdmin(t, handler, store)
	mock := mockPocketIDAdmin(t)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/pocketid", nil)
	req.AddCookie(session)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("get status=%d body=%s", rec.Code, rec.Body.String())
	}

	body := bytes.NewBufferString(`{
		"enabled": true,
		"baseUrl": "` + mock.URL + `",
		"apiKey": "test-key",
		"defaultGroupIds": ["g1"]
	}`)
	req = httptest.NewRequest(http.MethodPut, "/api/admin/pocketid", body)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(session)
	withCSRF(req, csrf)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("put status=%d body=%s", rec.Code, rec.Body.String())
	}
	var pub models.PocketIDSettingsPublic
	if err := json.NewDecoder(rec.Body).Decode(&pub); err != nil {
		t.Fatal(err)
	}
	if !pub.Enabled || !pub.APIKeySet || pub.BaseURL != mock.URL {
		t.Fatalf("public=%+v", pub)
	}

	req = httptest.NewRequest(http.MethodPut, "/api/admin/pocketid", bytes.NewBufferString(`{
		"enabled": true,
		"baseUrl": "`+mock.URL+`",
		"defaultGroupIds": ["g1","g2"]
	}`))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(session)
	withCSRF(req, csrf)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("put keep key status=%d body=%s", rec.Code, rec.Body.String())
	}
	saved, err := store.GetPocketIDSettings(context.Background())
	if err != nil || saved.APIKey != "test-key" {
		t.Fatalf("saved key=%q err=%v", saved.APIKey, err)
	}

	req = httptest.NewRequest(http.MethodPost, "/api/admin/pocketid/test", nil)
	req.AddCookie(session)
	withCSRF(req, csrf)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("test status=%d body=%s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/admin/pocketid", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized && rec.Code != http.StatusForbidden {
		t.Fatalf("anon get status=%d", rec.Code)
	}
}

func TestPocketIDApplyOIDCAndSignupTokens(t *testing.T) {
	srv, store := testServer(t)
	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}
	session, csrf := loginAdmin(t, handler, store)
	mock := mockPocketIDAdmin(t)

	if err := store.SavePocketIDSettings(context.Background(), models.PocketIDSettings{
		Enabled: true, BaseURL: mock.URL, APIKey: "test-key", DefaultGroupIDs: []string{},
	}); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/admin/pocketid/apply-oidc", nil)
	req.AddCookie(session)
	withCSRF(req, csrf)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("apply-oidc status=%d body=%s", rec.Code, rec.Body.String())
	}
	var oidcCfg models.OIDCConfig
	if err := json.NewDecoder(rec.Body).Decode(&oidcCfg); err != nil {
		t.Fatal(err)
	}
	if !oidcCfg.Enabled || oidcCfg.IssuerURL != mock.URL {
		t.Fatalf("oidc cfg=%+v", oidcCfg)
	}
	if oidcCfg.MatchBy != models.OIDCMatchEmail || !oidcCfg.AutoRegister {
		t.Fatalf("oidc match=%+v", oidcCfg)
	}
	if oidcCfg.GroupClaim != "groups" || oidcCfg.ButtonText == "" {
		t.Fatalf("oidc text/claim=%+v", oidcCfg)
	}

	req = httptest.NewRequest(http.MethodPost, "/api/admin/pocketid/signup-tokens", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(session)
	withCSRF(req, csrf)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create token status=%d body=%s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/admin/pocketid/signup-tokens", nil)
	req.AddCookie(session)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list tokens status=%d body=%s", rec.Code, rec.Body.String())
	}
	var list []pocketid.SignupToken
	if err := json.NewDecoder(rec.Body).Decode(&list); err != nil {
		t.Fatal(err)
	}
	if len(list) < 1 {
		t.Fatalf("list=%v", list)
	}

	req = httptest.NewRequest(http.MethodDelete, "/api/admin/pocketid/signup-tokens/st1", nil)
	req.AddCookie(session)
	withCSRF(req, csrf)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("delete token status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestPocketIDNotConfigured(t *testing.T) {
	srv, store := testServer(t)
	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}
	session, csrf := loginAdmin(t, handler, store)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/pocketid/test", nil)
	req.AddCookie(session)
	withCSRF(req, csrf)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("unconfigured test status=%d body=%s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, "/api/admin/pocketid/apply-oidc", nil)
	req.AddCookie(session)
	withCSRF(req, csrf)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "base url") {
		t.Fatalf("apply without url status=%d body=%s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, "/api/admin/pocketid/signup-tokens", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(session)
	withCSRF(req, csrf)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("create without config status=%d", rec.Code)
	}
}
