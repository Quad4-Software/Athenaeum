package storage

import (
	"context"
	"testing"
	"time"

	"athenaeum/internal/auth"
	"athenaeum/internal/models"
)

func TestUserSessionsLifecycle(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	hash, _ := auth.HashPassword("secretpass")
	id, err := s.CreateUser(ctx, "alice", hash, true)
	if err != nil {
		t.Fatal(err)
	}

	sessID, _ := auth.NewToken()
	access, _ := auth.NewSessionToken()
	refresh, _ := auth.NewSessionToken()
	exp := time.Now().Add(24 * time.Hour)
	if err := s.CreateUserSession(ctx, models.SessionCreate{
		SessionID:      sessID,
		AccessToken:    access,
		RefreshToken:   refresh,
		UserID:         id,
		IP:             "203.0.113.1",
		UserAgent:      "TestAgent",
		Device:         "Chrome on Linux",
		AuthMethod:     "local",
		AccessExpires:  exp,
		RefreshExpires: exp,
	}); err != nil {
		t.Fatal(err)
	}

	list, err := s.ListUserSessions(ctx, id, access)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("sessions=%d", len(list))
	}
	if !list[0].Current || list[0].Device != "Chrome on Linux" || list[0].IP != "203.0.113.1" {
		t.Fatalf("session=%+v", list[0])
	}

	n, err := s.RevokeOtherSessions(ctx, id, access)
	if err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("revoked others=%d", n)
	}

	if err := s.RevokeSessionByID(ctx, id, sessID); err != nil {
		t.Fatal(err)
	}
	list, err = s.ListUserSessions(ctx, id, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 0 {
		t.Fatalf("expected 0 sessions, got %d", len(list))
	}
}

func TestOIDCConfigRoundTrip(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	hash, _ := auth.HashPassword("secretpass")
	if _, err := s.CreateUser(ctx, "admin", hash, true); err != nil {
		t.Fatal(err)
	}

	cfg := models.OIDCConfig{
		Enabled:          true,
		LoginLocal:       true,
		IssuerURL:        "https://auth.example.com",
		AuthorizeURL:     "https://auth.example.com/authorize",
		TokenURL:         "https://auth.example.com/token",
		UserinfoURL:      "https://auth.example.com/userinfo",
		JWKSURL:          "https://auth.example.com/jwks",
		ClientID:         "reader",
		ClientSecret:     "s3cret",
		ButtonText:       "Login with Example",
		MatchBy:          models.OIDCMatchEmail,
		AutoRegister:     true,
		SigningAlgorithm: "RS256",
	}
	if err := s.SaveOIDCConfig(ctx, cfg); err != nil {
		t.Fatal(err)
	}
	got, err := s.GetOIDCConfig(ctx, true)
	if err != nil {
		t.Fatal(err)
	}
	if got.ClientSecret != "s3cret" || !got.Enabled || got.MatchBy != models.OIDCMatchEmail {
		t.Fatalf("config=%+v", got)
	}

	methods, err := s.AuthMethods(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !methods.AuthEnabled || !methods.LoginOIDC || methods.OIDCButtonText != "Login with Example" {
		t.Fatalf("methods=%+v", methods)
	}
}

func TestCreateOIDCUser(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	id, err := s.CreateOIDCUser(ctx, "oidc-user", "user@example.com", "sub-123", false)
	if err != nil {
		t.Fatal(err)
	}
	u, err := s.FindUserByOIDCSub(ctx, "sub-123")
	if err != nil || u.ID != id || u.LocalAuth {
		t.Fatalf("user=%+v err=%v", u, err)
	}
}
