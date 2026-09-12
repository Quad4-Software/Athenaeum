package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"

	"athenaeum/internal/auth"
	"athenaeum/internal/brand"
	"athenaeum/internal/models"
	athoidc "athenaeum/internal/oidc"
)

const oidcStateCookie = brand.OIDCStateCookie

func (s *Server) registerOIDCRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/auth/methods", s.handleAuthMethods)
	mux.HandleFunc("GET /api/auth/oidc/config", s.handleGetOIDCConfig)
	mux.HandleFunc("PUT /api/auth/oidc/config", s.handlePutOIDCConfig)
	mux.HandleFunc("POST /api/auth/oidc/discover", s.handleOIDCDiscover)
	mux.HandleFunc("GET /auth/oidc/login", s.handleOIDCLogin)
	mux.HandleFunc("GET /auth/oidc/callback", s.handleOIDCCallback)
}

func (s *Server) handleAuthMethods(w http.ResponseWriter, r *http.Request) {
	methods, err := s.store.AuthMethods(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	methods.PasswordPolicy = s.passwordPolicyPublic()
	if pub := s.altchaPublic(); pub.Enabled {
		methods.Altcha = &models.AltchaPublic{
			Enabled:         pub.Enabled,
			ChallengeURL:    pub.ChallengeURL,
			ProtectLogin:    pub.ProtectLogin,
			ProtectSetup:    pub.ProtectSetup,
			ProtectRegister: pub.ProtectRegister,
			Widget: models.AltchaWidgetPublic{
				Auto:       pub.Widget.Auto,
				Display:    pub.Widget.Display,
				HideFooter: pub.Widget.HideFooter,
				HideLogo:   pub.Widget.HideLogo,
				Language:   pub.Widget.Language,
				Name:       pub.Widget.Name,
				Theme:      pub.Widget.Theme,
				Type:       pub.Widget.Type,
				Workers:    pub.Widget.Workers,
			},
		}
	}
	s.issueCSRFForAuthPages(w, r)
	writeJSON(w, http.StatusOK, methods)
}

func (s *Server) handleGetOIDCConfig(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	cfg, err := s.store.GetOIDCConfig(r.Context(), false)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, cfg)
}

func (s *Server) handlePutOIDCConfig(w http.ResponseWriter, r *http.Request) {
	actor, ok := requireAdmin(w, r)
	if !ok {
		return
	}
	var cfg models.OIDCConfig
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&cfg); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if cfg.Enabled {
		if strings.TrimSpace(cfg.IssuerURL) == "" || strings.TrimSpace(cfg.ClientID) == "" {
			writeError(w, http.StatusBadRequest, errors.New("issuer URL and client ID are required when OIDC is enabled"))
			return
		}
		existing, err := s.store.GetOIDCConfig(r.Context(), true)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		if cfg.ClientSecret == "" && !existing.ClientSecretSet {
			writeError(w, http.StatusBadRequest, errors.New("client secret is required"))
			return
		}
	}
	if cfg.ButtonText == "" {
		cfg.ButtonText = "Sign in with SSO"
	}
	if cfg.MatchBy == "" {
		cfg.MatchBy = models.OIDCMatchUsername
	}
	if cfg.SigningAlgorithm == "" {
		cfg.SigningAlgorithm = "RS256"
	}
	if err := s.store.SaveOIDCConfig(r.Context(), cfg); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	s.logAudit(r, actor.ID, actor.Username, 0, "", "oidc.config", mapEnabled(cfg.Enabled))
	out, err := s.store.GetOIDCConfig(r.Context(), false)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func mapEnabled(enabled bool) string {
	if enabled {
		return "enabled"
	}
	return "disabled"
}

func (s *Server) handleOIDCDiscover(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	var req struct {
		IssuerURL string `json:"issuerUrl"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	endpoints, err := auth.DiscoverOIDC(r.Context(), req.IssuerURL)
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, models.OIDCDiscovery{
		IssuerURL:    endpoints.Issuer,
		AuthorizeURL: endpoints.AuthURL,
		TokenURL:     endpoints.TokenURL,
		UserinfoURL:  endpoints.UserinfoURL,
		JWKSURL:      endpoints.JWKSURL,
		LogoutURL:    endpoints.LogoutURL,
	})
}

func (s *Server) handleOIDCLogin(w http.ResponseWriter, r *http.Request) {
	cfg, secret, err := s.oidcRuntimeConfig(r.Context())
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	provider, oauthCfg, err := auth.OIDCProvider(r.Context(), cfg.IssuerURL, cfg.ClientID, secret, cfg.AuthorizeURL, cfg.TokenURL)
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	state, err := auth.NewToken()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	nonce, err := auth.NewToken()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	redirectURL := s.requestBaseURL(r) + "/auth/oidc/callback"
	oauthCfg.RedirectURL = redirectURL
	authURL := oauthCfg.AuthCodeURL(state, oidc.Nonce(nonce))
	http.SetCookie(w, s.oidcStateCookieValue(r, state+":"+nonce))
	_ = provider
	http.Redirect(w, r, authURL, http.StatusFound)
}

func (s *Server) handleOIDCCallback(w http.ResponseWriter, r *http.Request) {
	if errMsg := r.URL.Query().Get("error"); errMsg != "" {
		s.oidcErrorRedirect(w, r, errMsg)
		return
	}
	stateCookie, err := r.Cookie(oidcStateCookie)
	if err != nil || stateCookie.Value == "" {
		s.oidcErrorRedirect(w, r, "invalid_state")
		return
	}
	http.SetCookie(w, s.clearOIDCStateCookie(r))
	parts := strings.SplitN(stateCookie.Value, ":", 2)
	if len(parts) != 2 || parts[0] != r.URL.Query().Get("state") {
		s.oidcErrorRedirect(w, r, "state_mismatch")
		return
	}
	nonce := parts[1]

	cfg, secret, err := s.oidcRuntimeConfig(r.Context())
	if err != nil {
		s.oidcErrorRedirect(w, r, "config_error")
		return
	}
	provider, oauthCfg, err := auth.OIDCProvider(r.Context(), cfg.IssuerURL, cfg.ClientID, secret, cfg.AuthorizeURL, cfg.TokenURL)
	if err != nil {
		s.oidcErrorRedirect(w, r, "provider_error")
		return
	}
	oauthCfg.RedirectURL = s.requestBaseURL(r) + "/auth/oidc/callback"

	code := r.URL.Query().Get("code")
	if code == "" {
		s.oidcErrorRedirect(w, r, "missing_code")
		return
	}
	oauth2Token, err := oauthCfg.Exchange(r.Context(), code)
	if err != nil {
		s.oidcErrorRedirect(w, r, "token_exchange_failed")
		return
	}
	rawID, ok := oauth2Token.Extra("id_token").(string)
	if !ok || rawID == "" {
		s.oidcErrorRedirect(w, r, "missing_id_token")
		return
	}
	verifier := provider.Verifier(&oidc.Config{ClientID: cfg.ClientID})
	idToken, err := verifier.Verify(r.Context(), rawID)
	if err != nil {
		s.oidcErrorRedirect(w, r, "invalid_id_token")
		return
	}
	if idToken.Nonce != nonce {
		s.oidcErrorRedirect(w, r, "nonce_mismatch")
		return
	}

	var claims struct {
		Sub               string `json:"sub"`
		Email             string `json:"email"`
		EmailVerified     *bool  `json:"email_verified"`
		PreferredUsername string `json:"preferred_username"`
		Name              string `json:"name"`
	}
	if err := idToken.Claims(&claims); err != nil {
		s.oidcErrorRedirect(w, r, "invalid_claims")
		return
	}
	if claims.Sub == "" {
		s.oidcErrorRedirect(w, r, "missing_subject")
		return
	}

	groups := s.oidcGroups(r.Context(), idToken, oauth2Token, oauthCfg, cfg)
	isAdminGroup := oidcGroupMatchesAdmin(groups, cfg.AdminGroups)

	emailVerified := claims.EmailVerified == nil || *claims.EmailVerified
	u, err := s.resolveOIDCUser(r.Context(), cfg, claims.Sub, claims.Email, claims.PreferredUsername, claims.Name, isAdminGroup, emailVerified)
	if err != nil {
		s.oidcErrorRedirect(w, r, "account_error")
		return
	}
	s.completePendingInvite(w, r, u, claims.Email, emailVerified)
	if err := s.issueAuthTokens(w, r, u.ID, "oidc"); err != nil {
		s.oidcErrorRedirect(w, r, "session_error")
		return
	}
	s.logAudit(r, u.ID, u.Username, 0, "", "auth.login", "oidc")
	http.Redirect(w, r, "/?oidc=1", http.StatusFound)
}

// oidcResolver builds the OIDC user resolver with this server's store and
// webhook emitter.
func (s *Server) oidcResolver() *athoidc.Resolver {
	return &athoidc.Resolver{Store: s.store, Emit: s.emitWebhook}
}

func (s *Server) resolveOIDCUser(ctx context.Context, cfg models.OIDCConfig, sub, email, preferredUsername, name string, isAdminGroup, emailVerified bool) (models.User, error) {
	return s.oidcResolver().ResolveUser(ctx, cfg, sub, email, preferredUsername, name, isAdminGroup, emailVerified)
}

// oidcGroups extracts the group membership claim from the ID token, falling
// back to the userinfo endpoint when the claim is absent there.
func (s *Server) oidcGroups(ctx context.Context, idToken *oidc.IDToken, token *oauth2.Token, oauthCfg *oauth2.Config, cfg models.OIDCConfig) []string {
	return athoidc.Groups(ctx, idToken, token, oauthCfg, cfg)
}

// groupsFromClaim normalizes a claim value that may be a string array, a
// single string, or a comma-separated string into a list of group names.
func groupsFromClaim(claims map[string]any, key string) []string {
	return athoidc.GroupsFromClaim(claims, key)
}

// oidcGroupMatchesAdmin reports whether any of groups matches a comma
// separated admin group list, case-insensitively.
func oidcGroupMatchesAdmin(groups []string, adminGroups string) bool {
	return athoidc.GroupMatchesAdmin(groups, adminGroups)
}

func sanitizeUsername(s string) string {
	return athoidc.SanitizeUsername(s)
}

func (s *Server) uniqueUsername(ctx context.Context, base string) (string, error) {
	return s.oidcResolver().UniqueUsername(ctx, base)
}

func (s *Server) oidcRuntimeConfig(ctx context.Context) (models.OIDCConfig, string, error) {
	cfg, err := s.store.GetOIDCConfig(ctx, true)
	if err != nil {
		return cfg, "", err
	}
	if !cfg.Enabled || cfg.IssuerURL == "" || cfg.ClientID == "" || cfg.ClientSecret == "" {
		return cfg, "", errors.New("openid connect is not configured")
	}
	return cfg, cfg.ClientSecret, nil
}

func (s *Server) oidcErrorRedirect(w http.ResponseWriter, r *http.Request, code string) {
	q := url.Values{}
	q.Set("oidc_error", sanitizeOIDCErrorCode(code))
	http.Redirect(w, r, "/login?"+q.Encode(), http.StatusFound)
}

var oidcErrorCodes = map[string]struct{}{
	"access_denied":           {},
	"invalid_state":           {},
	"state_mismatch":          {},
	"config_error":            {},
	"provider_error":          {},
	"missing_code":            {},
	"token_exchange_failed":   {},
	"missing_id_token":        {},
	"invalid_id_token":        {},
	"nonce_mismatch":          {},
	"invalid_claims":          {},
	"missing_subject":         {},
	"account_error":           {},
	"session_error":           {},
	"login_required":          {},
	"interaction_required":    {},
	"consent_required":        {},
	"temporarily_unavailable": {},
}

func sanitizeOIDCErrorCode(code string) string {
	code = strings.TrimSpace(strings.ToLower(code))
	code = strings.ReplaceAll(code, " ", "_")
	if _, ok := oidcErrorCodes[code]; ok {
		return code
	}
	// Legacy / human messages from older call sites.
	switch code {
	case "invalid state":
		return "invalid_state"
	case "state mismatch":
		return "state_mismatch"
	case "missing code":
		return "missing_code"
	case "token exchange failed":
		return "token_exchange_failed"
	case "missing id token":
		return "missing_id_token"
	case "invalid id token":
		return "invalid_id_token"
	case "nonce mismatch":
		return "nonce_mismatch"
	case "invalid claims":
		return "invalid_claims"
	case "missing subject":
		return "missing_subject"
	case "session error":
		return "session_error"
	}
	return "provider_error"
}

func (s *Server) oidcStateCookieValue(r *http.Request, value string) *http.Cookie {
	// #nosec G124 -- Secure follows requestSecure() so local HTTP works and HTTPS sets Secure.
	return &http.Cookie{
		Name:     oidcStateCookie,
		Value:    value,
		Path:     "/auth/oidc",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   s.requestSecure(r),
		MaxAge:   600,
	}
}

func (s *Server) clearOIDCStateCookie(r *http.Request) *http.Cookie {
	// #nosec G124 -- Secure follows requestSecure() so local HTTP works and HTTPS sets Secure.
	return &http.Cookie{
		Name:     oidcStateCookie,
		Value:    "",
		Path:     "/auth/oidc",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   s.requestSecure(r),
		MaxAge:   -1,
	}
}

func (s *Server) invitePendingCookie(r *http.Request, token string) *http.Cookie {
	// #nosec G124 -- Secure follows requestSecure() so local HTTP works and HTTPS sets Secure.
	return &http.Cookie{
		Name:     brand.InvitePendingCookie,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   s.requestSecure(r),
		MaxAge:   600,
	}
}

func (s *Server) clearInvitePendingCookie(r *http.Request) *http.Cookie {
	// #nosec G124 -- Secure follows requestSecure() so local HTTP works and HTTPS sets Secure.
	return &http.Cookie{
		Name:     brand.InvitePendingCookie,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   s.requestSecure(r),
		MaxAge:   -1,
	}
}

func (s *Server) completePendingInvite(w http.ResponseWriter, r *http.Request, u models.User, email string, emailVerified bool) {
	c, err := r.Cookie(brand.InvitePendingCookie)
	http.SetCookie(w, s.clearInvitePendingCookie(r))
	if err != nil || c.Value == "" {
		return
	}
	inv, err := s.store.GetInviteByToken(r.Context(), c.Value)
	if err != nil {
		return
	}
	if inv.AcceptedAt != nil || inv.RevokedAt != nil {
		return
	}
	if inv.ExpiresAt != nil && time.Now().After(*inv.ExpiresAt) {
		return
	}
	if inv.Email != "" {
		if !emailVerified {
			return
		}
		if !strings.EqualFold(strings.TrimSpace(inv.Email), strings.TrimSpace(email)) {
			return
		}
	}
	if err := s.store.AcceptInvite(r.Context(), inv.ID, u.ID); err != nil {
		return
	}
	s.logAudit(r, u.ID, u.Username, 0, "", "invite.accepted", "oidc")
	s.emitWebhook(models.WebhookEventInviteAccepted, map[string]any{
		"inviteId": inv.ID,
		"kind":     inv.Kind,
		"via":      "oidc",
		"userId":   u.ID,
	})
}
