package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"athenaeum/internal/auth"
	"athenaeum/internal/models"
	"athenaeum/internal/pocketid"
	"athenaeum/internal/storage"
)

func (s *Server) registerInviteRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/invites", s.handleCreateInvite)
	mux.HandleFunc("GET /api/invites", s.handleListInvites)
	mux.HandleFunc("DELETE /api/invites/{id}", s.handleRevokeInvite)
	mux.HandleFunc("GET /api/invite/{token}", s.handleInviteMeta)
	mux.HandleFunc("POST /api/invite/{token}/accept", s.handleAcceptInvite)
}

type createInviteRequest struct {
	Kind                string   `json:"kind"`
	Email               string   `json:"email"`
	Username            string   `json:"username"`
	Permissions         []string `json:"permissions"`
	ExpiresInHours      int      `json:"expiresInHours"`
	GuestExpiresInHours int      `json:"guestExpiresInHours"`
	ProvisionPocketID   bool     `json:"provisionPocketId"`
}

func (s *Server) handleCreateInvite(w http.ResponseWriter, r *http.Request) {
	actor, ok := requireAdmin(w, r)
	if !ok {
		return
	}
	var req createInviteRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	kind := strings.TrimSpace(req.Kind)
	if kind == "" {
		kind = models.InviteKindPermanent
	}
	if kind != models.InviteKindPermanent && kind != models.InviteKindGuest {
		writeError(w, http.StatusBadRequest, errors.New("kind must be permanent or guest"))
		return
	}
	hours := req.ExpiresInHours
	if hours <= 0 {
		hours = 168
	}
	if hours > 24*365 {
		writeError(w, http.StatusBadRequest, errors.New("expiresInHours must be at most 8760"))
		return
	}
	email := strings.TrimSpace(strings.ToLower(req.Email))
	perms := models.DefaultUserPermissions
	if len(req.Permissions) > 0 {
		perms = models.ParsePermissions(req.Permissions)
	}
	expiresAt := time.Now().Add(time.Duration(hours) * time.Hour)
	inv := models.Invite{
		Kind:        kind,
		Email:       email,
		Permissions: perms,
		CreatedBy:   actor.ID,
		ExpiresAt:   &expiresAt,
	}
	if kind == models.InviteKindGuest {
		guestHours := req.GuestExpiresInHours
		if guestHours <= 0 {
			guestHours = 24
		}
		if guestHours > 24*365 {
			writeError(w, http.StatusBadRequest, errors.New("guestExpiresInHours must be at most 8760"))
			return
		}
		guestExp := time.Now().Add(time.Duration(guestHours) * time.Hour)
		inv.GuestExpiresAt = &guestExp
	}

	var pocketSetupURL string
	if req.ProvisionPocketID {
		if kind != models.InviteKindPermanent {
			writeError(w, http.StatusBadRequest, errors.New("pocket id provisioning requires permanent invites"))
			return
		}
		if email == "" {
			writeError(w, http.StatusBadRequest, errors.New("email is required for pocket id provisioning"))
			return
		}
		setupURL, pocketUserID, err := s.provisionPocketIDInvite(r, email, strings.TrimSpace(req.Username))
		if err != nil {
			writeError(w, http.StatusBadGateway, err)
			return
		}
		inv.PocketIDUserID = pocketUserID
		pocketSetupURL = setupURL
	}

	created, err := s.store.CreateInvite(r.Context(), inv)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	pathURL := "/invite/" + created.Token
	absURL := s.requestBaseURL(r) + pathURL
	emailSent := false
	if email != "" {
		smtpCfg, err := s.store.GetSMTPSettings(r.Context())
		if err == nil && smtpCfg.Enabled && smtpCfg.Host != "" {
			body := "You have been invited to Athenaeum.\n\nAccept your invite:\n" + absURL + "\n"
			if pocketSetupURL != "" {
				body += "\nSet up your Pocket ID passkey:\n" + pocketSetupURL + "\n"
			}
			if err := sendSMTPText(smtpCfg, email, "Athenaeum invitation", body); err != nil {
				s.log.Warn("invite email failed", "err", err)
			} else {
				emailSent = true
			}
		} else if pocketSetupURL != "" && created.PocketIDUserID != "" {
			pidCfg, err := s.store.GetPocketIDSettings(r.Context())
			if err == nil && pidCfg.Enabled {
				client := pocketid.NewClient(pidCfg.BaseURL, pidCfg.APIKey)
				if err := client.RequestOneTimeAccessEmail(r.Context(), created.PocketIDUserID, "24h"); err != nil {
					s.log.Warn("pocket id one-time access email failed", "err", err)
				} else {
					emailSent = true
				}
			}
		}
	}

	s.logAudit(r, actor.ID, actor.Username, 0, email, "invite.created", kind)
	s.emitWebhook(models.WebhookEventInviteCreated, map[string]any{
		"inviteId": created.ID,
		"kind":     created.Kind,
		"email":    created.Email != "",
	})

	writeJSON(w, http.StatusCreated, models.InviteCreateResult{
		Invite:           created.Public(),
		URL:              pathURL,
		PocketIDSetupURL: pocketSetupURL,
		EmailSent:        emailSent,
	})
}

func (s *Server) provisionPocketIDInvite(r *http.Request, email, username string) (setupURL, userID string, err error) {
	cfg, err := s.store.GetPocketIDSettings(r.Context())
	if err != nil {
		return "", "", err
	}
	if !cfg.Enabled || cfg.BaseURL == "" || cfg.APIKey == "" {
		return "", "", errors.New("pocket id is not configured")
	}
	if username == "" {
		username = emailLocalPart(email)
	}
	if len(username) < 2 {
		return "", "", errors.New("username is required for pocket id provisioning")
	}
	client := pocketid.NewClient(cfg.BaseURL, cfg.APIKey)
	display := username
	u, err := client.CreateUser(r.Context(), pocketid.UserCreate{
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
		if err := client.UpdateUserGroups(r.Context(), u.ID, cfg.DefaultGroupIDs); err != nil {
			s.log.Warn("pocket id group assign failed", "err", err)
		}
	}
	tok, err := client.CreateOneTimeAccessToken(r.Context(), u.ID, "24h")
	if err != nil {
		return "", "", err
	}
	return client.SetupURL(tok), u.ID, nil
}

func emailLocalPart(email string) string {
	at := strings.IndexByte(email, '@')
	if at <= 0 {
		return email
	}
	return email[:at]
}

func (s *Server) handleListInvites(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	list, err := s.store.ListInvites(r.Context(), status)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	out := make([]models.InvitePublic, 0, len(list))
	for _, inv := range list {
		out = append(out, inv.Public())
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleRevokeInvite(w http.ResponseWriter, r *http.Request) {
	actor, ok := requireAdmin(w, r)
	if !ok {
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, errors.New("invalid id"))
		return
	}
	if err := s.store.RevokeInvite(r.Context(), id); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			writeError(w, http.StatusNotFound, err)
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	s.logAudit(r, actor.ID, actor.Username, 0, "", "invite.revoke", strconv.FormatInt(id, 10))
	writeJSON(w, http.StatusOK, map[string]string{"ok": "revoked"})
}

func (s *Server) handleInviteMeta(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	s.issueCSRFForAuthPages(w, r)
	inv, err := s.store.GetInviteByToken(r.Context(), token)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			writeJSON(w, http.StatusOK, models.InviteMeta{Valid: false, Reason: "not_found"})
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	meta := models.InviteMeta{
		Kind:               inv.Kind,
		EmailPresent:       inv.Email != "",
		ExpiresAt:          inv.ExpiresAt,
		PocketIDConfigured: inv.PocketIDUserID != "",
		Valid:              true,
	}
	now := time.Now()
	if inv.RevokedAt != nil {
		meta.Valid = false
		meta.Reason = "revoked"
	} else if inv.AcceptedAt != nil {
		meta.Valid = false
		meta.Reason = "accepted"
	} else if inv.ExpiresAt != nil && now.After(*inv.ExpiresAt) {
		meta.Valid = false
		meta.Reason = "expired"
	}
	writeJSON(w, http.StatusOK, meta)
}

func (s *Server) handleAcceptInvite(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	inv, err := s.store.GetInviteByToken(r.Context(), token)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			writeError(w, http.StatusNotFound, errors.New("invite not found"))
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	now := time.Now()
	if inv.RevokedAt != nil {
		writeError(w, http.StatusGone, errors.New("invite revoked"))
		return
	}
	if inv.AcceptedAt != nil {
		writeError(w, http.StatusConflict, errors.New("invite already accepted"))
		return
	}
	if inv.ExpiresAt != nil && now.After(*inv.ExpiresAt) {
		writeError(w, http.StatusGone, errors.New("invite expired"))
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&req); err != nil && err.Error() != "EOF" {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if inv.Kind == models.InviteKindPermanent && inv.PocketIDUserID != "" {
		if err := s.store.AcceptInviteSSO(r.Context(), inv.ID); err != nil {
			writeError(w, http.StatusConflict, errors.New("invite already accepted"))
			return
		}
		s.logAudit(r, 0, inv.Email, 0, "", "invite.accepted", "pocketid")
		s.emitWebhook(models.WebhookEventInviteAccepted, map[string]any{
			"inviteId": inv.ID,
			"kind":     inv.Kind,
			"via":      "pocketid",
		})
		writeJSON(w, http.StatusOK, map[string]string{
			"ok":       "accepted",
			"redirect": "/auth/oidc/login",
		})
		return
	}

	if inv.Kind == models.InviteKindGuest {
		username := strings.TrimSpace(req.Username)
		if username == "" {
			tok, err := auth.NewToken()
			if err != nil {
				writeError(w, http.StatusInternalServerError, err)
				return
			}
			username = "guest-" + tok[:8]
		}
		if len(username) < 2 {
			writeError(w, http.StatusBadRequest, errWeakCredentials)
			return
		}
		taken, err := s.store.UsernameTaken(r.Context(), username, 0)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		if taken {
			writeError(w, http.StatusConflict, errUsernameTaken)
			return
		}
		plain, err := auth.NewToken()
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		plain = plain[:16]
		hash, err := auth.HashPassword(plain)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		guestExp := now.Add(24 * time.Hour)
		if inv.GuestExpiresAt != nil {
			guestExp = *inv.GuestExpiresAt
		}
		id, err := s.store.CreateGuestUser(r.Context(), username, hash, guestExp, inv.Permissions)
		if err != nil {
			writeError(w, http.StatusConflict, errUsernameTaken)
			return
		}
		if err := s.store.AcceptInvite(r.Context(), inv.ID, id); err != nil {
			writeError(w, http.StatusConflict, errors.New("invite already accepted"))
			return
		}
		u, err := s.store.GetUser(r.Context(), id)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		s.logAudit(r, 0, username, u.ID, u.Username, "invite.accepted", "guest")
		s.logAudit(r, 0, username, u.ID, u.Username, "user.guest", fmt.Sprintf("invite %d", inv.ID))
		s.emitWebhook(models.WebhookEventInviteAccepted, map[string]any{"inviteId": inv.ID, "kind": inv.Kind, "userId": u.ID})
		s.emitWebhook(models.WebhookEventUserCreate, map[string]any{"userId": u.ID, "username": u.Username, "guest": true})
		writeJSON(w, http.StatusCreated, models.GuestCredentials{User: u.Public(), Password: plain})
		return
	}

	username := strings.TrimSpace(req.Username)
	if len(username) < 2 {
		writeError(w, http.StatusBadRequest, errWeakCredentials)
		return
	}
	if err := auth.ValidatePassword(req.Password); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	taken, err := s.store.UsernameTaken(r.Context(), username, 0)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if taken {
		writeError(w, http.StatusConflict, errUsernameTaken)
		return
	}
	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	id, err := s.store.CreateInvitedUser(r.Context(), username, hash, inv.Email, inv.Permissions)
	if err != nil {
		writeError(w, http.StatusConflict, errUsernameTaken)
		return
	}
	if err := s.store.AcceptInvite(r.Context(), inv.ID, id); err != nil {
		writeError(w, http.StatusConflict, errors.New("invite already accepted"))
		return
	}
	u, err := s.store.GetUser(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	s.logAudit(r, 0, username, u.ID, u.Username, "invite.accepted", "permanent")
	s.logAudit(r, 0, username, u.ID, u.Username, "user.create", "invite")
	s.emitWebhook(models.WebhookEventInviteAccepted, map[string]any{"inviteId": inv.ID, "kind": inv.Kind, "userId": u.ID})
	s.emitWebhook(models.WebhookEventUserCreate, map[string]any{"userId": u.ID, "username": u.Username})
	writeJSON(w, http.StatusCreated, u.Public())
}
