package storage

import (
	"context"
	"database/sql"
	"errors"

	"athenaeum/internal/models"
)

// GetOIDCConfig loads OpenID Connect settings. ClientSecret is returned when includeSecret is true.
func (s *Store) GetOIDCConfig(ctx context.Context, includeSecret bool) (models.OIDCConfig, error) {
	var cfg models.OIDCConfig
	var enabled, loginLocal, autoReg, autoLaunch int
	var secret string
	err := s.queryRowContext(ctx, `
SELECT enabled, login_local, issuer_url, authorize_url, token_url, userinfo_url, jwks_url, logout_url,
       client_id, client_secret, signing_algorithm, button_text, match_by, auto_register, auto_launch,
       group_claim, admin_groups
FROM oidc_config WHERE id=1`).
		Scan(&enabled, &loginLocal, &cfg.IssuerURL, &cfg.AuthorizeURL, &cfg.TokenURL, &cfg.UserinfoURL,
			&cfg.JWKSURL, &cfg.LogoutURL, &cfg.ClientID, &secret, &cfg.SigningAlgorithm,
			&cfg.ButtonText, &cfg.MatchBy, &autoReg, &autoLaunch, &cfg.GroupClaim, &cfg.AdminGroups)
	if errors.Is(err, sql.ErrNoRows) {
		return models.OIDCConfig{LoginLocal: true, ButtonText: "Sign in with SSO", MatchBy: models.OIDCMatchUsername, SigningAlgorithm: "RS256", GroupClaim: "groups"}, nil
	}
	if err != nil {
		return cfg, err
	}
	cfg.Enabled = enabled != 0
	cfg.LoginLocal = loginLocal != 0
	cfg.AutoRegister = autoReg != 0
	cfg.AutoLaunch = autoLaunch != 0
	cfg.ClientSecretSet = secret != ""
	if includeSecret {
		cfg.ClientSecret = secret
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
	if cfg.GroupClaim == "" {
		cfg.GroupClaim = "groups"
	}
	return cfg, nil
}

// SaveOIDCConfig persists OpenID Connect settings. An empty clientSecret keeps the existing secret.
func (s *Store) SaveOIDCConfig(ctx context.Context, cfg models.OIDCConfig) error {
	existing, err := s.GetOIDCConfig(ctx, true)
	if err != nil {
		return err
	}
	secret := cfg.ClientSecret
	if secret == "" {
		secret = existing.ClientSecret
	}
	_, err = s.execContext(ctx, `
UPDATE oidc_config SET
	enabled=?, login_local=?, issuer_url=?, authorize_url=?, token_url=?, userinfo_url=?, jwks_url=?, logout_url=?,
	client_id=?, client_secret=?, signing_algorithm=?, button_text=?, match_by=?, auto_register=?, auto_launch=?,
	group_claim=?, admin_groups=?
WHERE id=1`,
		boolToInt(cfg.Enabled), boolToInt(cfg.LoginLocal),
		cfg.IssuerURL, cfg.AuthorizeURL, cfg.TokenURL, cfg.UserinfoURL, cfg.JWKSURL, cfg.LogoutURL,
		cfg.ClientID, secret, cfg.SigningAlgorithm, cfg.ButtonText, string(cfg.MatchBy),
		boolToInt(cfg.AutoRegister), boolToInt(cfg.AutoLaunch), cfg.GroupClaim, cfg.AdminGroups)
	return err
}

// AuthMethods reports enabled sign-in options.
func (s *Store) AuthMethods(ctx context.Context) (models.AuthMethods, error) {
	required, err := s.AuthRequired(ctx)
	if err != nil {
		return models.AuthMethods{}, err
	}
	out := models.AuthMethods{AuthEnabled: required, LoginLocal: true}
	authSettings, err := s.GetAuthSettings(ctx)
	if err != nil {
		return out, err
	}
	out.AllowRegistration = authSettings.AllowRegistration
	if !required {
		return out, nil
	}
	cfg, err := s.GetOIDCConfig(ctx, false)
	if err != nil {
		return out, err
	}
	out.LoginLocal = cfg.LoginLocal
	out.LoginOIDC = cfg.Enabled && cfg.IssuerURL != "" && cfg.ClientID != "" && cfg.ClientSecretSet
	out.OIDCButtonText = cfg.ButtonText
	out.OIDCAutoLaunch = cfg.AutoLaunch
	return out, nil
}
