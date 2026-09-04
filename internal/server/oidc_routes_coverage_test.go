package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"athenaeum/internal/auth"
	"athenaeum/internal/brand"
	"athenaeum/internal/models"
	"athenaeum/internal/storage"
)

func mockOIDCIssuer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
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
	mux.HandleFunc("/jwks", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"keys": []any{}})
	})
	mux.HandleFunc("/userinfo", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"groups": []string{"admins"}})
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func saveEnabledOIDC(t *testing.T, store *storage.Store, issuerURL string) {
	t.Helper()
	cfg := models.OIDCConfig{
		Enabled:      true,
		LoginLocal:   true,
		IssuerURL:    issuerURL,
		AuthorizeURL: issuerURL + "/authorize",
		TokenURL:     issuerURL + "/token",
		UserinfoURL:  issuerURL + "/userinfo",
		JWKSURL:      issuerURL + "/jwks",
		LogoutURL:    issuerURL + "/logout",
		ClientID:     "reader-client",
		ClientSecret: "reader-secret",
		ButtonText:   "SSO",
		MatchBy:      models.OIDCMatchEmail,
		AutoRegister: true,
		GroupClaim:   "groups",
		AdminGroups:  "admins",
	}
	if err := store.SaveOIDCConfig(context.Background(), cfg); err != nil {
		t.Fatal(err)
	}
}

func TestOIDCDiscoverAndLogin(t *testing.T) {
	srv, store := testServer(t)
	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}
	session, csrf := loginAdmin(t, handler, store)
	issuer := mockOIDCIssuer(t)

	body := bytes.NewBufferString(`{"issuerUrl":"` + issuer.URL + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/oidc/discover", body)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(session)
	withCSRF(req, csrf)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("discover status=%d body=%s", rec.Code, rec.Body.String())
	}
	var disc models.OIDCDiscovery
	if err := json.NewDecoder(rec.Body).Decode(&disc); err != nil {
		t.Fatal(err)
	}
	if disc.IssuerURL != issuer.URL || disc.AuthorizeURL == "" || disc.TokenURL == "" {
		t.Fatalf("discovery=%+v", disc)
	}

	req = httptest.NewRequest(http.MethodPost, "/api/auth/oidc/discover", bytes.NewBufferString(`{`))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(session)
	withCSRF(req, csrf)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("bad json status=%d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/api/auth/oidc/discover", bytes.NewBufferString(`{"issuerUrl":"http://127.0.0.1:9"}`))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(session)
	withCSRF(req, csrf)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadGateway {
		t.Fatalf("unreachable discover status=%d", rec.Code)
	}

	saveEnabledOIDC(t, store, issuer.URL)

	req = httptest.NewRequest(http.MethodGet, "/auth/oidc/login", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusFound {
		t.Fatalf("login status=%d body=%s", rec.Code, rec.Body.String())
	}
	loc := rec.Header().Get("Location")
	if !strings.Contains(loc, issuer.URL+"/authorize") || !strings.Contains(loc, "state=") {
		t.Fatalf("auth location=%q", loc)
	}
	var stateCookie *http.Cookie
	for _, c := range rec.Result().Cookies() {
		if c.Name == brand.OIDCStateCookie {
			stateCookie = c
			break
		}
	}
	if stateCookie == nil || !strings.Contains(stateCookie.Value, ":") {
		t.Fatalf("state cookie=%v", stateCookie)
	}
}

func TestOIDCLoginNotConfigured(t *testing.T) {
	srv, _ := testServer(t)
	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/auth/oidc/login", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestOIDCCallbackErrorPaths(t *testing.T) {
	srv, store := testServer(t)
	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}
	issuer := mockOIDCIssuer(t)
	saveEnabledOIDC(t, store, issuer.URL)

	expectOIDCErr := func(path string, cookies ...*http.Cookie) {
		t.Helper()
		req := httptest.NewRequest(http.MethodGet, path, nil)
		for _, c := range cookies {
			if c != nil {
				req.AddCookie(c)
			}
		}
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusFound {
			t.Fatalf("status=%d body=%s path=%s", rec.Code, rec.Body.String(), path)
		}
		loc := rec.Header().Get("Location")
		if !strings.HasPrefix(loc, "/login?") || !strings.Contains(loc, "oidc_error=") {
			t.Fatalf("location=%q path=%s", loc, path)
		}
	}

	expectOIDCErr("/auth/oidc/callback?error=access_denied")
	expectOIDCErr("/auth/oidc/callback?code=x&state=y")
	expectOIDCErr("/auth/oidc/callback?code=x&state=y", &http.Cookie{Name: brand.OIDCStateCookie, Value: ""})
	expectOIDCErr("/auth/oidc/callback?code=x&state=wrong", &http.Cookie{Name: brand.OIDCStateCookie, Value: "right:nonce"})
	expectOIDCErr("/auth/oidc/callback?state=st", &http.Cookie{Name: brand.OIDCStateCookie, Value: "st:nonce"})
	expectOIDCErr("/auth/oidc/callback?code=bad&state=st", &http.Cookie{Name: brand.OIDCStateCookie, Value: "st:nonce"})

	cfg, secret, err := srv.oidcRuntimeConfig(context.Background())
	if err != nil || secret == "" || !cfg.Enabled {
		t.Fatalf("runtime cfg=%+v secret set=%v err=%v", cfg, secret != "", err)
	}
}

func TestOIDCCallbackWhenDisabled(t *testing.T) {
	srv, _ := testServer(t)
	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/auth/oidc/callback?code=x&state=st", nil)
	req.AddCookie(&http.Cookie{Name: brand.OIDCStateCookie, Value: "st:nonce"})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusFound || !strings.Contains(rec.Header().Get("Location"), "oidc_error=") {
		t.Fatalf("status=%d loc=%s", rec.Code, rec.Header().Get("Location"))
	}
}

func TestOIDCResolveUserPaths(t *testing.T) {
	srv, store := testServer(t)
	ctx := context.Background()

	cfg := models.OIDCConfig{MatchBy: models.OIDCMatchEmail, AutoRegister: true, AdminGroups: "admins"}

	u, err := srv.resolveOIDCUser(ctx, cfg, "sub-new-1", "new1@example.com", "newuser1", "New User", false, true)
	if err != nil {
		t.Fatal(err)
	}
	if u.Username == "" || u.ID == 0 {
		t.Fatalf("created user=%+v", u)
	}
	again, err := srv.resolveOIDCUser(ctx, cfg, "sub-new-1", "new1@example.com", "newuser1", "New User", true, true)
	if err != nil {
		t.Fatal(err)
	}
	if !again.IsAdmin {
		t.Fatal("expected admin promotion on existing oidc sub")
	}

	invited, err := store.CreateInvitedUser(ctx, "linkedmail", "hash", "link@example.com", models.DefaultUserPermissions)
	if err != nil {
		t.Fatal(err)
	}
	linked, err := srv.resolveOIDCUser(ctx, cfg, "sub-link-email", "link@example.com", "", "", true, true)
	if err != nil {
		t.Fatal(err)
	}
	if linked.ID != invited || !linked.IsAdmin {
		t.Fatalf("linked=%+v want id=%d", linked, invited)
	}

	hash, err := auth.HashPassword("longpassword")
	if err != nil {
		t.Fatal(err)
	}
	uid, err := store.CreateUser(ctx, "matchname", hash, false)
	if err != nil {
		t.Fatal(err)
	}
	byName, err := srv.resolveOIDCUser(ctx, models.OIDCConfig{MatchBy: models.OIDCMatchUsername, AutoRegister: false}, "sub-name", "", "matchname", "", false, true)
	if err != nil {
		t.Fatal(err)
	}
	if byName.ID != uid {
		t.Fatalf("username match=%+v want %d", byName, uid)
	}

	_, err = srv.resolveOIDCUser(ctx, models.OIDCConfig{MatchBy: models.OIDCMatchSub, AutoRegister: false}, "sub-nomatch", "x@y.z", "x", "x", false, true)
	if err == nil || !strings.Contains(err.Error(), "no matching account") {
		t.Fatalf("auto register off err=%v", err)
	}

	created, err := srv.resolveOIDCUser(ctx, models.OIDCConfig{MatchBy: models.OIDCMatchSub, AutoRegister: true}, "sub-auto-2", "", "", "", false, true)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(created.Username, "user-") {
		t.Fatalf("fallback username=%q", created.Username)
	}

	fromEmailLocal, err := srv.resolveOIDCUser(ctx, models.OIDCConfig{MatchBy: models.OIDCMatchUsername, AutoRegister: true}, "sub-auto-3", "prefixuser@example.com", "", "", false, true)
	if err != nil {
		t.Fatal(err)
	}
	if fromEmailLocal.Username != "prefixuser" {
		t.Fatalf("username from email=%q", fromEmailLocal.Username)
	}
}

func TestOIDCCookieHelpers(t *testing.T) {
	srv, _ := testServer(t)
	req := httptest.NewRequest(http.MethodGet, "http://localhost/auth/oidc/login", nil)
	c := srv.oidcStateCookieValue(req, "state:nonce")
	if c.Name != brand.OIDCStateCookie || c.Value != "state:nonce" || c.MaxAge != 600 || !c.HttpOnly {
		t.Fatalf("state cookie=%+v", c)
	}
	clear := srv.clearOIDCStateCookie(req)
	if clear.MaxAge != -1 || clear.Value != "" {
		t.Fatalf("clear cookie=%+v", clear)
	}

	rec := httptest.NewRecorder()
	srv.oidcErrorRedirect(rec, req, "<script>alert(1)</script>")
	loc := rec.Header().Get("Location")
	if rec.Code != http.StatusFound || !strings.Contains(loc, "oidc_error=provider_error") || strings.Contains(loc, "<script") {
		t.Fatalf("redirect status=%d loc=%s", rec.Code, loc)
	}

	_, _, err := srv.oidcRuntimeConfig(context.Background())
	if err == nil {
		t.Fatal("expected not configured")
	}
}

func TestOIDCPutConfigValidation(t *testing.T) {
	srv, store := testServer(t)
	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}
	session, csrf := loginAdmin(t, handler, store)

	req := httptest.NewRequest(http.MethodPut, "/api/auth/oidc/config", bytes.NewBufferString(`{"enabled":true,"issuerUrl":"","clientId":""}`))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(session)
	withCSRF(req, csrf)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("missing issuer status=%d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodPut, "/api/auth/oidc/config", bytes.NewBufferString(`{"enabled":true,"issuerUrl":"https://x","clientId":"c"}`))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(session)
	withCSRF(req, csrf)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("missing secret status=%d body=%s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/auth/oidc/config", nil)
	req.AddCookie(session)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("get config status=%d", rec.Code)
	}
}
