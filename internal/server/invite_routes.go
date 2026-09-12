package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"athenaeum/internal/invite"
	"athenaeum/internal/models"
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
	res, err := s.inviteService().Create(r.Context(), invite.CreateInput{
		Kind:                req.Kind,
		Email:               req.Email,
		Username:            req.Username,
		Permissions:         req.Permissions,
		ExpiresInHours:      req.ExpiresInHours,
		GuestExpiresInHours: req.GuestExpiresInHours,
		ProvisionPocketID:   req.ProvisionPocketID,
		ActorID:             actor.ID,
		ActorName:           actor.Username,
		BaseURL:             s.requestBaseURL(r),
	}, s.auditFunc(r))
	if err != nil {
		writeInviteError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, res)
}

// inviteService builds the invite orchestration service with this server's
// dependencies.
func (s *Server) inviteService() *invite.Service {
	return &invite.Service{
		Store:          s.store,
		Log:            s.log,
		SendMail:       sendSMTPText,
		Emit:           s.emitWebhook,
		PocketTokenTTL: pocketIDTokenTTL,
	}
}

// auditFunc binds the request to logAudit so services can record audit
// entries without seeing http.Request.
func (s *Server) auditFunc(r *http.Request) invite.AuditFunc {
	return func(actorID int64, actorName string, targetID int64, targetName, action, details string) {
		s.logAudit(r, actorID, actorName, targetID, targetName, action, details)
	}
}

// writeInviteError maps invite.StatusError codes to HTTP responses; anything
// else is a 500.
func writeInviteError(w http.ResponseWriter, err error) {
	var se *invite.StatusError
	if errors.As(err, &se) {
		writeError(w, se.Code, se.Err)
		return
	}
	writeError(w, http.StatusInternalServerError, err)
}

func emailLocalPart(email string) string {
	return invite.EmailLocalPart(email)
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
		writeError(w, http.StatusBadRequest, errInvalidID)
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
	res, err := s.inviteService().Accept(r.Context(), r.PathValue("token"), http.MaxBytesReader(w, r.Body, 1<<16), s.auditFunc(r))
	if err != nil {
		writeInviteError(w, err)
		return
	}
	if res.SSORedirect {
		http.SetCookie(w, s.invitePendingCookie(r, res.InviteToken))
		writeJSON(w, http.StatusOK, map[string]string{
			"ok":       "continue",
			"redirect": "/auth/oidc/login",
		})
		return
	}
	if res.Guest != nil {
		writeJSON(w, http.StatusCreated, *res.Guest)
		return
	}
	writeJSON(w, http.StatusCreated, *res.User)
}
