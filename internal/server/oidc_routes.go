package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"

	"athenaeum/internal/auth"
	"athenaeum/internal/brand"
	"athenaeum/internal/models"
	"athenaeum/internal/storage"
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
		s.oidcErrorRedirect(w, r, "invalid state")
		return
	}
	http.SetCookie(w, s.clearOIDCStateCookie(r))
	parts := strings.SplitN(stateCookie.Value, ":", 2)
	if len(parts) != 2 || parts[0] != r.URL.Query().Get("state") {
		s.oidcErrorRedirect(w, r, "state mismatch")
		return
	}
	nonce := parts[1]

	cfg, secret, err := s.oidcRuntimeConfig(r.Context())
	if err != nil {
		s.oidcErrorRedirect(w, r, err.Error())
		return
	}
	provider, oauthCfg, err := auth.OIDCProvider(r.Context(), cfg.IssuerURL, cfg.ClientID, secret, cfg.AuthorizeURL, cfg.TokenURL)
	if err != nil {
		s.oidcErrorRedirect(w, r, err.Error())
		return
	}
	oauthCfg.RedirectURL = s.requestBaseURL(r) + "/auth/oidc/callback"

	code := r.URL.Query().Get("code")
	if code == "" {
		s.oidcErrorRedirect(w, r, "missing code")
		return
	}
	oauth2Token, err := oauthCfg.Exchange(r.Context(), code)
	if err != nil {
		s.oidcErrorRedirect(w, r, "token exchange failed")
		return
	}
	rawID, ok := oauth2Token.Extra("id_token").(string)
	if !ok || rawID == "" {
		s.oidcErrorRedirect(w, r, "missing id token")
		return
	}
	verifier := provider.Verifier(&oidc.Config{ClientID: cfg.ClientID})
	idToken, err := verifier.Verify(r.Context(), rawID)
	if err != nil {
		s.oidcErrorRedirect(w, r, "invalid id token")
		return
	}
	if idToken.Nonce != nonce {
		s.oidcErrorRedirect(w, r, "nonce mismatch")
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
		s.oidcErrorRedirect(w, r, "invalid claims")
		return
	}
	if claims.Sub == "" {
		s.oidcErrorRedirect(w, r, "missing subject")
		return
	}

	groups := s.oidcGroups(r.Context(), idToken, oauth2Token, oauthCfg, cfg)
	isAdminGroup := oidcGroupMatchesAdmin(groups, cfg.AdminGroups)

	emailVerified := claims.EmailVerified == nil || *claims.EmailVerified
	u, err := s.resolveOIDCUser(r.Context(), cfg, claims.Sub, claims.Email, claims.PreferredUsername, claims.Name, isAdminGroup, emailVerified)
	if err != nil {
		s.oidcErrorRedirect(w, r, err.Error())
		return
	}
	s.completePendingInvite(w, r, u, claims.Email, emailVerified)
	if err := s.issueAuthTokens(w, r, u.ID, "oidc"); err != nil {
		s.oidcErrorRedirect(w, r, "session error")
		return
	}
	s.logAudit(r, u.ID, u.Username, 0, "", "auth.login", "oidc")
	http.Redirect(w, r, "/?oidc=1", http.StatusFound)
}

func (s *Server) resolveOIDCUser(ctx context.Context, cfg models.OIDCConfig, sub, email, preferredUsername, name string, isAdminGroup, emailVerified bool) (models.User, error) {
	if u, err := s.store.FindUserByOIDCSub(ctx, sub); err == nil {
		if isAdminGroup && !u.IsAdmin {
			if err := s.store.SetUserAdmin(ctx, u.ID, true); err != nil {
				return models.User{}, err
			}
			return s.store.GetUser(ctx, u.ID)
		}
		return u, nil
	} else if !errors.Is(err, storage.ErrNotFound) {
		return models.User{}, err
	}

	var matched models.User
	var matchErr error
	switch cfg.MatchBy {
	case models.OIDCMatchEmail:
		if email == "" {
			matchErr = storage.ErrNotFound
		} else if !emailVerified {
			return models.User{}, errors.New("email claim is not verified")
		} else {
			matched, matchErr = s.store.FindUserByEmail(ctx, email)
		}
	case models.OIDCMatchSub:
		matchErr = storage.ErrNotFound
	default:
		username := preferredUsername
		if username == "" && email != "" {
			username = strings.Split(email, "@")[0]
		}
		if username == "" {
			username = name
		}
		username = sanitizeUsername(username)
		if username == "" {
			return models.User{}, errors.New("could not determine username from OIDC claims")
		}
		matched, _, matchErr = s.store.GetUserByUsername(ctx, username)
	}
	if matchErr == nil {
		if err := s.store.LinkOIDCSub(ctx, matched.ID, sub, email); err != nil {
			if errors.Is(err, storage.ErrConflict) {
				return models.User{}, errors.New("account is already linked to another identity")
			}
			return models.User{}, err
		}
		if isAdminGroup && !matched.IsAdmin {
			if err := s.store.SetUserAdmin(ctx, matched.ID, true); err != nil {
				return models.User{}, err
			}
		}
		return s.store.GetUser(ctx, matched.ID)
	}
	if !errors.Is(matchErr, storage.ErrNotFound) {
		return models.User{}, matchErr
	}
	if !cfg.AutoRegister {
		return models.User{}, errors.New("no matching account; contact an administrator")
	}

	username := preferredUsername
	if username == "" && email != "" {
		username = strings.Split(email, "@")[0]
	}
	if username == "" {
		username = name
	}
	username = sanitizeUsername(username)
	if username == "" {
		if len(sub) > 8 {
			username = "user-" + sub[:8]
		} else {
			username = "user-" + sub
		}
	}
	username, err := s.uniqueUsername(ctx, username)
	if err != nil {
		return models.User{}, err
	}
	id, err := s.store.CreateOIDCUser(ctx, username, email, sub, isAdminGroup)
	if err != nil {
		return models.User{}, err
	}
	s.emitWebhook(models.WebhookEventUserCreate, map[string]any{
		"userId":   id,
		"username": username,
		"via":      "oidc",
	})
	return s.store.GetUser(ctx, id)
}

// oidcGroups extracts the group membership claim from the ID token, falling
// back to the userinfo endpoint when the claim is absent there.
func (s *Server) oidcGroups(ctx context.Context, idToken *oidc.IDToken, token *oauth2.Token, oauthCfg *oauth2.Config, cfg models.OIDCConfig) []string {
	claimKey := strings.TrimSpace(cfg.GroupClaim)
	if claimKey == "" {
		claimKey = "groups"
	}
	var raw map[string]any
	if err := idToken.Claims(&raw); err == nil {
		if groups := groupsFromClaim(raw, claimKey); len(groups) > 0 {
			return groups
		}
	}
	if cfg.UserinfoURL == "" {
		return nil
	}
	client := oauthCfg.Client(ctx, token)
	reqCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, cfg.UserinfoURL, nil)
	if err != nil {
		return nil
	}
	res, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil
	}
	var userinfo map[string]any
	if err := json.NewDecoder(res.Body).Decode(&userinfo); err != nil {
		return nil
	}
	return groupsFromClaim(userinfo, claimKey)
}

// groupsFromClaim normalizes a claim value that may be a string array, a
// single string, or a comma-separated string into a list of group names.
func groupsFromClaim(claims map[string]any, key string) []string {
	value, ok := claims[key]
	if !ok {
		return nil
	}
	switch v := value.(type) {
	case []any:
		out := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok && s != "" {
				out = append(out, s)
			}
		}
		return out
	case []string:
		return v
	case string:
		var out []string
		for part := range strings.SplitSeq(v, ",") {
			part = strings.TrimSpace(part)
			if part != "" {
				out = append(out, part)
			}
		}
		return out
	default:
		return nil
	}
}

// oidcGroupMatchesAdmin reports whether any of groups matches a comma
// separated admin group list, case-insensitively.
func oidcGroupMatchesAdmin(groups []string, adminGroups string) bool {
	adminGroups = strings.TrimSpace(adminGroups)
	if adminGroups == "" || len(groups) == 0 {
		return false
	}
	var wanted []string
	for part := range strings.SplitSeq(adminGroups, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			wanted = append(wanted, part)
		}
	}
	for _, g := range groups {
		for _, w := range wanted {
			if strings.EqualFold(g, w) {
				return true
			}
		}
	}
	return false
}

func sanitizeUsername(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_', r == '.':
			b.WriteRune(r)
		case r == ' ' || r == '@':
			b.WriteRune('_')
		}
	}
	out := strings.Trim(b.String(), "._-")
	if len(out) < 2 {
		return ""
	}
	if len(out) > 64 {
		out = out[:64]
	}
	return out
}

func (s *Server) uniqueUsername(ctx context.Context, base string) (string, error) {
	candidate := base
	for i := range 20 {
		taken, err := s.store.UsernameTaken(ctx, candidate, 0)
		if err != nil {
			return "", err
		}
		if !taken {
			return candidate, nil
		}
		candidate = fmt.Sprintf("%s-%d", base, i+2)
	}
	return "", errors.New("could not allocate unique username")
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

func (s *Server) oidcErrorRedirect(w http.ResponseWriter, r *http.Request, msg string) {
	q := url.Values{}
	q.Set("oidc_error", msg)
	http.Redirect(w, r, "/login?"+q.Encode(), http.StatusFound)
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
