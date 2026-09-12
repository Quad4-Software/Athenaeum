// Package invite contains the non-HTTP orchestration behind the invite
// routes: validation, Pocket ID provisioning, mail delivery, and webhooks.
package invite

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"athenaeum/internal/auth"
	"athenaeum/internal/models"
	"athenaeum/internal/pocketid"
	"athenaeum/internal/storage"
)

// guestTTL is the default lifetime of a guest account created from a guest
// invite when the invite carries no explicit expiry.
const guestTTL = 24 * time.Hour

// Shared invite error values mirrored by the HTTP handlers.
var (
	ErrWeakCredentials = errors.New("username must be at least 2 characters and password at least 8")
	ErrUsernameTaken   = errors.New("username already taken")
)

// StatusError couples an error with the HTTP status the handler should send.
type StatusError struct {
	Code int
	Err  error
}

func (e *StatusError) Error() string { return e.Err.Error() }
func (e *StatusError) Unwrap() error { return e.Err }

func statusError(code int, err error) *StatusError { return &StatusError{Code: code, Err: err} }

// AuditFunc records an audit entry. The server binds one per request so the
// service stays HTTP-free.
type AuditFunc func(actorID int64, actorName string, targetID int64, targetName, action, details string)

// Service orchestrates invite creation and acceptance.
type Service struct {
	Store          *storage.Store
	Log            *slog.Logger
	SendMail       func(cfg models.SMTPSettings, to, subject, body string) error
	Emit           func(event string, data map[string]any)
	PocketTokenTTL string
}

func (s *Service) warn(msg string, args ...any) {
	if s.Log != nil {
		s.Log.Warn(msg, args...)
	}
}

func (s *Service) emit(event string, data map[string]any) {
	if s.Emit != nil {
		s.Emit(event, data)
	}
}

// CreateInput carries the decoded create-invite request plus request-scoped
// values the service needs.
type CreateInput struct {
	Kind                string
	Email               string
	Username            string
	Permissions         []string
	ExpiresInHours      int
	GuestExpiresInHours int
	ProvisionPocketID   bool
	ActorID             int64
	ActorName           string
	BaseURL             string
}

// Create validates the input, provisions Pocket ID when requested, stores the
// invite, and delivers the invite email.
func (s *Service) Create(ctx context.Context, in CreateInput, audit AuditFunc) (models.InviteCreateResult, error) {
	var out models.InviteCreateResult
	kind := strings.TrimSpace(in.Kind)
	if kind == "" {
		kind = models.InviteKindPermanent
	}
	if kind != models.InviteKindPermanent && kind != models.InviteKindGuest {
		return out, statusError(http.StatusBadRequest, errors.New("kind must be permanent or guest"))
	}
	hours := in.ExpiresInHours
	if hours <= 0 {
		hours = 168
	}
	if hours > 24*365 {
		return out, statusError(http.StatusBadRequest, errors.New("expiresInHours must be at most 8760"))
	}
	email := strings.TrimSpace(strings.ToLower(in.Email))
	perms := models.DefaultUserPermissions
	if len(in.Permissions) > 0 {
		perms = models.ParsePermissions(in.Permissions)
	}
	expiresAt := time.Now().Add(time.Duration(hours) * time.Hour)
	inv := models.Invite{
		Kind:        kind,
		Email:       email,
		Permissions: perms,
		CreatedBy:   in.ActorID,
		ExpiresAt:   &expiresAt,
	}
	if kind == models.InviteKindGuest {
		guestHours := in.GuestExpiresInHours
		if guestHours <= 0 {
			guestHours = 24
		}
		if guestHours > 24*365 {
			return out, statusError(http.StatusBadRequest, errors.New("guestExpiresInHours must be at most 8760"))
		}
		guestExp := time.Now().Add(time.Duration(guestHours) * time.Hour)
		inv.GuestExpiresAt = &guestExp
	}

	var pocketSetupURL string
	if in.ProvisionPocketID {
		if kind != models.InviteKindPermanent {
			return out, statusError(http.StatusBadRequest, errors.New("pocket id provisioning requires permanent invites"))
		}
		if email == "" {
			return out, statusError(http.StatusBadRequest, errors.New("email is required for pocket id provisioning"))
		}
		setupURL, pocketUserID, err := s.ProvisionPocketID(ctx, email, strings.TrimSpace(in.Username))
		if err != nil {
			return out, statusError(http.StatusBadGateway, err)
		}
		inv.PocketIDUserID = pocketUserID
		pocketSetupURL = setupURL
	}

	created, err := s.Store.CreateInvite(ctx, inv)
	if err != nil {
		if inv.PocketIDUserID != "" {
			if cfg, cfgErr := s.Store.GetPocketIDSettings(ctx); cfgErr == nil && cfg.Enabled {
				client := pocketid.NewClient(cfg.BaseURL, cfg.APIKey)
				if delErr := client.DeleteUser(ctx, inv.PocketIDUserID); delErr != nil {
					s.warn("pocket id orphan cleanup failed", "userId", inv.PocketIDUserID, "err", delErr)
				}
			}
		}
		return out, err
	}

	pathURL := "/invite/" + created.Token
	absURL := in.BaseURL + pathURL
	emailSent := false
	if email != "" {
		smtpCfg, err := s.Store.GetSMTPSettings(ctx)
		if err == nil && smtpCfg.Enabled && smtpCfg.Host != "" {
			body := "You have been invited to Athenaeum.\n\nAccept your invite:\n" + absURL + "\n"
			if pocketSetupURL != "" {
				body += "\nSet up your Pocket ID passkey:\n" + pocketSetupURL + "\n"
			}
			if err := s.SendMail(smtpCfg, email, "Athenaeum invitation", body); err != nil {
				s.warn("invite email failed", "err", err)
			} else {
				emailSent = true
			}
		} else if pocketSetupURL != "" && created.PocketIDUserID != "" {
			pidCfg, err := s.Store.GetPocketIDSettings(ctx)
			if err == nil && pidCfg.Enabled {
				client := pocketid.NewClient(pidCfg.BaseURL, pidCfg.APIKey)
				if err := client.RequestOneTimeAccessEmail(ctx, created.PocketIDUserID, s.PocketTokenTTL); err != nil {
					s.warn("pocket id one-time access email failed", "err", err)
				} else {
					emailSent = true
				}
			}
		}
	}

	if audit != nil {
		audit(in.ActorID, in.ActorName, 0, email, "invite.created", kind)
	}
	s.emit(models.WebhookEventInviteCreated, map[string]any{
		"inviteId": created.ID,
		"kind":     created.Kind,
		"email":    created.Email != "",
	})

	out = models.InviteCreateResult{
		Invite:           created.Public(),
		URL:              pathURL,
		PocketIDSetupURL: pocketSetupURL,
		EmailSent:        emailSent,
	}
	return out, nil
}

// ProvisionPocketID creates a Pocket ID user for an invite and returns the
// passkey setup URL.
func (s *Service) ProvisionPocketID(ctx context.Context, email, username string) (setupURL, userID string, err error) {
	cfg, err := s.Store.GetPocketIDSettings(ctx)
	if err != nil {
		return "", "", err
	}
	if !cfg.Enabled || cfg.BaseURL == "" || cfg.APIKey == "" {
		return "", "", errors.New("pocket id is not configured")
	}
	if username == "" {
		username = EmailLocalPart(email)
	}
	if len(username) < 2 {
		return "", "", errors.New("username is required for pocket id provisioning")
	}
	client := pocketid.NewClient(cfg.BaseURL, cfg.APIKey)
	display := username
	u, err := client.CreateUser(ctx, pocketid.UserCreate{
		Username:      username,
		Email:         email,
		FirstName:     username,
		DisplayName:   display,
		EmailVerified: true,
	})
	if err != nil {
		return "", "", err
	}
	if len(cfg.DefaultGroupIDs) > 0 {
		if err := client.UpdateUserGroups(ctx, u.ID, cfg.DefaultGroupIDs); err != nil {
			s.warn("pocket id group assign failed", "err", err)
		}
	}
	tok, err := client.CreateOneTimeAccessToken(ctx, u.ID, s.PocketTokenTTL)
	if err != nil {
		return "", "", err
	}
	return client.SetupURL(tok), u.ID, nil
}

// EmailLocalPart returns the part of an email address before the @ sign.
func EmailLocalPart(email string) string {
	at := strings.IndexByte(email, '@')
	if at <= 0 {
		return email
	}
	return email[:at]
}

// AcceptResult describes what the accept handler should write back.
type AcceptResult struct {
	// SSORedirect marks a permanent Pocket ID invite: the handler must set
	// the pending cookie and answer with the OIDC redirect.
	SSORedirect bool
	InviteToken string
	Guest       *models.GuestCredentials
	User        *models.UserPublic
}

// Accept validates the invite token and creates the invited account. body is
// the raw accept request body; it is decoded only after the invite checks so
// invalid invites keep their original status codes.
func (s *Service) Accept(ctx context.Context, token string, body io.Reader, audit AuditFunc) (*AcceptResult, error) {
	inv, err := s.Store.GetInviteByToken(ctx, token)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, statusError(http.StatusNotFound, errors.New("invite not found"))
		}
		return nil, err
	}
	now := time.Now()
	if inv.RevokedAt != nil {
		return nil, statusError(http.StatusGone, errors.New("invite revoked"))
	}
	if inv.AcceptedAt != nil {
		return nil, statusError(http.StatusConflict, errors.New("invite already accepted"))
	}
	if inv.ExpiresAt != nil && now.After(*inv.ExpiresAt) {
		return nil, statusError(http.StatusGone, errors.New("invite expired"))
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(body).Decode(&req); err != nil && err.Error() != "EOF" {
		return nil, statusError(http.StatusBadRequest, err)
	}

	if inv.Kind == models.InviteKindPermanent && inv.PocketIDUserID != "" {
		if audit != nil {
			audit(0, inv.Email, 0, "", "invite.sso_start", "pocketid")
		}
		return &AcceptResult{SSORedirect: true, InviteToken: inv.Token}, nil
	}

	if inv.Kind == models.InviteKindGuest {
		return s.acceptGuest(ctx, inv, req.Username, now, audit)
	}
	return s.acceptPermanent(ctx, inv, req.Username, req.Password, audit)
}

func (s *Service) acceptGuest(ctx context.Context, inv models.Invite, username string, now time.Time, audit AuditFunc) (*AcceptResult, error) {
	username = strings.TrimSpace(username)
	if username == "" {
		tok, err := auth.NewToken()
		if err != nil {
			return nil, err
		}
		username = "guest-" + tok[:8]
	}
	if len(username) < 2 {
		return nil, statusError(http.StatusBadRequest, ErrWeakCredentials)
	}
	taken, err := s.Store.UsernameTaken(ctx, username, 0)
	if err != nil {
		return nil, err
	}
	if taken {
		return nil, statusError(http.StatusConflict, ErrUsernameTaken)
	}
	plain, err := auth.NewToken()
	if err != nil {
		return nil, err
	}
	plain = plain[:16]
	hash, err := auth.HashPassword(plain)
	if err != nil {
		return nil, err
	}
	guestExp := now.Add(guestTTL)
	if inv.GuestExpiresAt != nil {
		guestExp = *inv.GuestExpiresAt
	}
	id, err := s.Store.CreateGuestUser(ctx, username, hash, guestExp, inv.Permissions)
	if err != nil {
		return nil, statusError(http.StatusConflict, ErrUsernameTaken)
	}
	if err := s.Store.AcceptInvite(ctx, inv.ID, id); err != nil {
		_ = s.Store.DeleteUser(ctx, id)
		return nil, statusError(http.StatusConflict, errors.New("invite already accepted"))
	}
	u, err := s.Store.GetUser(ctx, id)
	if err != nil {
		return nil, err
	}
	if audit != nil {
		audit(0, username, u.ID, u.Username, "invite.accepted", "guest")
		audit(0, username, u.ID, u.Username, "user.guest", fmt.Sprintf("invite %d", inv.ID))
	}
	s.emit(models.WebhookEventInviteAccepted, map[string]any{"inviteId": inv.ID, "kind": inv.Kind, "userId": u.ID})
	s.emit(models.WebhookEventUserCreate, map[string]any{"userId": u.ID, "username": u.Username, "guest": true})
	return &AcceptResult{Guest: &models.GuestCredentials{User: u.Public(), Password: plain}}, nil
}

func (s *Service) acceptPermanent(ctx context.Context, inv models.Invite, username, password string, audit AuditFunc) (*AcceptResult, error) {
	username = strings.TrimSpace(username)
	if len(username) < 2 {
		return nil, statusError(http.StatusBadRequest, ErrWeakCredentials)
	}
	if err := auth.ValidatePassword(password); err != nil {
		return nil, statusError(http.StatusBadRequest, err)
	}
	taken, err := s.Store.UsernameTaken(ctx, username, 0)
	if err != nil {
		return nil, err
	}
	if taken {
		return nil, statusError(http.StatusConflict, ErrUsernameTaken)
	}
	hash, err := auth.HashPassword(password)
	if err != nil {
		return nil, err
	}
	id, err := s.Store.CreateInvitedUser(ctx, username, hash, inv.Email, inv.Permissions)
	if err != nil {
		return nil, statusError(http.StatusConflict, ErrUsernameTaken)
	}
	if err := s.Store.AcceptInvite(ctx, inv.ID, id); err != nil {
		_ = s.Store.DeleteUser(ctx, id)
		return nil, statusError(http.StatusConflict, errors.New("invite already accepted"))
	}
	u, err := s.Store.GetUser(ctx, id)
	if err != nil {
		return nil, err
	}
	if audit != nil {
		audit(0, username, u.ID, u.Username, "invite.accepted", "permanent")
		audit(0, username, u.ID, u.Username, "user.create", "invite")
	}
	s.emit(models.WebhookEventInviteAccepted, map[string]any{"inviteId": inv.ID, "kind": inv.Kind, "userId": u.ID})
	s.emit(models.WebhookEventUserCreate, map[string]any{"userId": u.ID, "username": u.Username})
	pub := u.Public()
	return &AcceptResult{User: &pub}, nil
}
