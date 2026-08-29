package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDiscoverOIDC(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{
			"issuer":                 "http://" + r.Host,
			"authorization_endpoint": "http://" + r.Host + "/authorize",
			"token_endpoint":         "http://" + r.Host + "/token",
			"userinfo_endpoint":      "http://" + r.Host + "/userinfo",
			"jwks_uri":               "http://" + r.Host + "/jwks",
			"end_session_endpoint":   "http://" + r.Host + "/logout",
		})
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	ep, err := DiscoverOIDC(context.Background(), srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	if ep.Issuer != srv.URL {
		t.Fatalf("issuer=%q", ep.Issuer)
	}
	if ep.AuthURL != srv.URL+"/authorize" || ep.TokenURL != srv.URL+"/token" {
		t.Fatalf("endpoints=%+v", ep)
	}
	if ep.UserinfoURL != srv.URL+"/userinfo" || ep.JWKSURL != srv.URL+"/jwks" {
		t.Fatalf("endpoints=%+v", ep)
	}
	if ep.LogoutURL != srv.URL+"/logout" {
		t.Fatalf("logout=%q", ep.LogoutURL)
	}

	ep2, err := DiscoverOIDC(context.Background(), srv.URL+"/.well-known/openid-configuration")
	if err != nil {
		t.Fatal(err)
	}
	if ep2.Issuer != srv.URL {
		t.Fatalf("discovery url issuer=%q", ep2.Issuer)
	}
}

func TestDiscoverOIDCErrors(t *testing.T) {
	if _, err := DiscoverOIDC(context.Background(), ""); err == nil || !strings.Contains(err.Error(), "required") {
		t.Fatalf("empty issuer err=%v", err)
	}

	bad := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusInternalServerError)
	}))
	t.Cleanup(bad.Close)
	if _, err := DiscoverOIDC(context.Background(), bad.URL); err == nil || !strings.Contains(err.Error(), "discovery failed") {
		t.Fatalf("http error err=%v", err)
	}

	incomplete := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"issuer": "http://x"})
	}))
	t.Cleanup(incomplete.Close)
	if _, err := DiscoverOIDC(context.Background(), incomplete.URL); err == nil || !strings.Contains(err.Error(), "incomplete") {
		t.Fatalf("incomplete err=%v", err)
	}

	invalidJSON := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("{"))
	}))
	t.Cleanup(invalidJSON.Close)
	if _, err := DiscoverOIDC(context.Background(), invalidJSON.URL); err == nil {
		t.Fatal("expected json decode error")
	}
}

func TestOIDCProvider(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
		base := "http://" + r.Host
		_ = json.NewEncoder(w).Encode(map[string]any{
			"issuer":                                base,
			"authorization_endpoint":                base + "/authorize",
			"token_endpoint":                        base + "/token",
			"jwks_uri":                              base + "/jwks",
			"id_token_signing_alg_values_supported": []string{"RS256"},
		})
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	provider, cfg, err := OIDCProvider(context.Background(), srv.URL, "client", "secret", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if provider == nil || cfg.ClientID != "client" || cfg.Endpoint.AuthURL == "" {
		t.Fatalf("provider=%v cfg=%+v", provider, cfg)
	}

	_, cfg2, err := OIDCProvider(context.Background(), srv.URL, "c2", "s2", srv.URL+"/custom-auth", srv.URL+"/custom-token")
	if err != nil {
		t.Fatal(err)
	}
	if cfg2.Endpoint.AuthURL != srv.URL+"/custom-auth" || cfg2.Endpoint.TokenURL != srv.URL+"/custom-token" {
		t.Fatalf("override endpoints=%+v", cfg2.Endpoint)
	}

	if _, _, err := OIDCProvider(context.Background(), "", "c", "s", "", ""); err == nil {
		t.Fatal("expected empty issuer error")
	}
}
