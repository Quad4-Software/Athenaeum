// Package oidc resolves OIDC identities to local users and extracts group
// claims from ID tokens and userinfo responses.
package oidc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	gooidc "github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"

	"athenaeum/internal/models"
	"athenaeum/internal/storage"
)

// userinfoTimeout bounds the userinfo fallback request for group claims.
const userinfoTimeout = 10 * time.Second

// Resolver resolves OIDC identities to local users.
type Resolver struct {
	Store *storage.Store
	Emit  func(event string, data map[string]any)
}

func (r *Resolver) emit(event string, data map[string]any) {
	if r.Emit != nil {
		r.Emit(event, data)
	}
}

// ResolveUser finds or creates the local user for an OIDC identity: first by
// linked subject, then by the configured match strategy, then by
// auto-registration.
func (r *Resolver) ResolveUser(ctx context.Context, cfg models.OIDCConfig, sub, email, preferredUsername, name string, isAdminGroup, emailVerified bool) (models.User, error) {
	if u, err := r.Store.FindUserByOIDCSub(ctx, sub); err == nil {
		if isAdminGroup && !u.IsAdmin {
			if err := r.Store.SetUserAdmin(ctx, u.ID, true); err != nil {
				return models.User{}, err
			}
			return r.Store.GetUser(ctx, u.ID)
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
			matched, matchErr = r.Store.FindUserByEmail(ctx, email)
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
		username = SanitizeUsername(username)
		if username == "" {
			return models.User{}, errors.New("could not determine username from OIDC claims")
		}
		matched, _, matchErr = r.Store.GetUserByUsername(ctx, username)
	}
	if matchErr == nil {
		if err := r.Store.LinkOIDCSub(ctx, matched.ID, sub, email); err != nil {
			if errors.Is(err, storage.ErrConflict) {
				return models.User{}, errors.New("account is already linked to another identity")
			}
			return models.User{}, err
		}
		if isAdminGroup && !matched.IsAdmin {
			if err := r.Store.SetUserAdmin(ctx, matched.ID, true); err != nil {
				return models.User{}, err
			}
		}
		return r.Store.GetUser(ctx, matched.ID)
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
	username = SanitizeUsername(username)
	if username == "" {
		if len(sub) > 8 {
			username = "user-" + sub[:8]
		} else {
			username = "user-" + sub
		}
	}
	username, err := r.UniqueUsername(ctx, username)
	if err != nil {
		return models.User{}, err
	}
	id, err := r.Store.CreateOIDCUser(ctx, username, email, sub, isAdminGroup)
	if err != nil {
		return models.User{}, err
	}
	r.emit(models.WebhookEventUserCreate, map[string]any{
		"userId":   id,
		"username": username,
		"via":      "oidc",
	})
	return r.Store.GetUser(ctx, id)
}

// UniqueUsername appends a numeric suffix until the name is free.
func (r *Resolver) UniqueUsername(ctx context.Context, base string) (string, error) {
	candidate := base
	for i := range 20 {
		taken, err := r.Store.UsernameTaken(ctx, candidate, 0)
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

// Groups extracts the group membership claim from the ID token, falling back
// to the userinfo endpoint when the claim is absent there.
func Groups(ctx context.Context, idToken *gooidc.IDToken, token *oauth2.Token, oauthCfg *oauth2.Config, cfg models.OIDCConfig) []string {
	claimKey := strings.TrimSpace(cfg.GroupClaim)
	if claimKey == "" {
		claimKey = "groups"
	}
	var raw map[string]any
	if err := idToken.Claims(&raw); err == nil {
		if groups := GroupsFromClaim(raw, claimKey); len(groups) > 0 {
			return groups
		}
	}
	if cfg.UserinfoURL == "" {
		return nil
	}
	client := oauthCfg.Client(ctx, token)
	reqCtx, cancel := context.WithTimeout(ctx, userinfoTimeout)
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
	return GroupsFromClaim(userinfo, claimKey)
}

// GroupsFromClaim normalizes a claim value that may be a string array, a
// single string, or a comma-separated string into a list of group names.
func GroupsFromClaim(claims map[string]any, key string) []string {
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

// GroupMatchesAdmin reports whether any of groups matches a comma separated
// admin group list, case-insensitively.
func GroupMatchesAdmin(groups []string, adminGroups string) bool {
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

// SanitizeUsername reduces an OIDC claim value to a valid local username.
func SanitizeUsername(s string) string {
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
