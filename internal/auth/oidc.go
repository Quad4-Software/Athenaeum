package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

const oidcHTTPTimeout = 15 * time.Second

// DiscoverOIDC fetches provider endpoints from an issuer or discovery URL.
func DiscoverOIDC(ctx context.Context, issuerURL string) (endpoints struct {
	Issuer      string
	AuthURL     string
	TokenURL    string
	UserinfoURL string
	JWKSURL     string
	LogoutURL   string
}, err error) {
	issuerURL = strings.TrimSpace(issuerURL)
	if issuerURL == "" {
		return endpoints, fmt.Errorf("issuer url is required")
	}
	discoveryURL := issuerURL
	if !strings.HasSuffix(discoveryURL, "/.well-known/openid-configuration") {
		discoveryURL = strings.TrimRight(discoveryURL, "/") + "/.well-known/openid-configuration"
	}

	reqCtx, cancel := context.WithTimeout(ctx, oidcHTTPTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, discoveryURL, nil)
	if err != nil {
		return endpoints, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return endpoints, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(res.Body, 512))
		return endpoints, fmt.Errorf("discovery failed: %s", strings.TrimSpace(string(body)))
	}
	var doc struct {
		Issuer                string `json:"issuer"`
		AuthorizationEndpoint string `json:"authorization_endpoint"`
		TokenEndpoint         string `json:"token_endpoint"`
		UserinfoEndpoint      string `json:"userinfo_endpoint"`
		JWKSURI               string `json:"jwks_uri"`
		EndSessionEndpoint    string `json:"end_session_endpoint"`
	}
	if err := json.NewDecoder(io.LimitReader(res.Body, 1<<20)).Decode(&doc); err != nil {
		return endpoints, err
	}
	if doc.Issuer == "" || doc.AuthorizationEndpoint == "" || doc.TokenEndpoint == "" {
		return endpoints, fmt.Errorf("incomplete discovery document")
	}
	endpoints.Issuer = doc.Issuer
	endpoints.AuthURL = doc.AuthorizationEndpoint
	endpoints.TokenURL = doc.TokenEndpoint
	endpoints.UserinfoURL = doc.UserinfoEndpoint
	endpoints.JWKSURL = doc.JWKSURI
	endpoints.LogoutURL = doc.EndSessionEndpoint
	return endpoints, nil
}

// OIDCProvider builds a verifier and oauth2 config from stored settings.
func OIDCProvider(ctx context.Context, issuerURL, clientID, clientSecret, authURL, tokenURL string) (*oidc.Provider, *oauth2.Config, error) {
	issuerURL = strings.TrimSpace(issuerURL)
	if issuerURL == "" {
		return nil, nil, fmt.Errorf("issuer url is required")
	}
	provider, err := oidc.NewProvider(ctx, issuerURL)
	if err != nil {
		return nil, nil, err
	}
	cfg := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email", "groups"},
	}
	if authURL != "" || tokenURL != "" {
		cfg.Endpoint = oauth2.Endpoint{
			AuthURL:  authURL,
			TokenURL: tokenURL,
		}
	}
	return provider, cfg, nil
}
