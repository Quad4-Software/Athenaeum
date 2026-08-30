package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"athenaeum/internal/auth"
	"athenaeum/internal/models"
	"athenaeum/internal/storage"
)

type guestRequest struct {
	Username       string   `json:"username"`
	ExpiresInHours int      `json:"expiresInHours"`
	Permissions    []string `json:"permissions"`
}

func (s *Server) handleCreateGuest(w http.ResponseWriter, r *http.Request) {
	actor, ok := requireAdmin(w, r)
	if !ok {
		return
	}
	required, err := s.store.AuthRequired(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if !required {
		writeError(w, http.StatusForbidden, errors.New("initial setup required"))
		return
	}
	var req guestRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	hours := req.ExpiresInHours
	if hours <= 0 {
		hours = 24
	}
	if hours > 24*365 {
		writeError(w, http.StatusBadRequest, errors.New("expiresInHours must be at most 8760"))
		return
	}
	username := strings.TrimSpace(req.Username)
	if username == "" {
		token, err := auth.NewToken()
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		username = "guest-" + token[:8]
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
	perms := models.DefaultUserPermissions
	if len(req.Permissions) > 0 {
		perms = models.ParsePermissions(req.Permissions)
	}
	expiresAt := time.Now().Add(time.Duration(hours) * time.Hour)
	id, err := s.store.CreateGuestUser(r.Context(), username, hash, expiresAt, perms)
	if err != nil {
		writeError(w, http.StatusConflict, errUsernameTaken)
		return
	}
	u, err := s.store.GetUser(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	s.logAudit(r, actor.ID, actor.Username, u.ID, u.Username, "user.guest", fmt.Sprintf("expires in %dh", hours))
	writeJSON(w, http.StatusCreated, models.GuestCredentials{User: u.Public(), Password: plain})
}

func (s *Server) handleListUsers(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	users, err := s.store.ListUsers(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	out := make([]models.UserPublic, len(users))
	for i, u := range users {
		out[i] = u.Public()
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleSetUserPermissions(w http.ResponseWriter, r *http.Request) {
	actor, ok := requireAdmin(w, r)
	if !ok {
		return
	}
	targetID, ok := pathID(w, r)
	if !ok {
		return
	}
	var req struct {
		Permissions []string `json:"permissions"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	target, err := s.store.GetUser(r.Context(), targetID)
	if errors.Is(err, storage.ErrNotFound) {
		writeError(w, http.StatusNotFound, errors.New("user not found"))
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if target.IsAdmin {
		writeJSON(w, http.StatusOK, target.Public())
		return
	}
	mask := models.ParsePermissions(req.Permissions)
	if err := s.store.SetUserPermissions(r.Context(), targetID, mask); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			writeError(w, http.StatusNotFound, errors.New("user not found"))
		} else {
			writeError(w, http.StatusInternalServerError, err)
		}
		return
	}
	target, _ = s.store.GetUser(r.Context(), targetID)
	s.logAudit(r, actor.ID, actor.Username, target.ID, target.Username, "user.permissions", strings.Join(req.Permissions, ","))
	writeJSON(w, http.StatusOK, target.Public())
}

func (s *Server) handleResetUserPassword(w http.ResponseWriter, r *http.Request) {
	actor, ok := requireAdmin(w, r)
	if !ok {
		return
	}
	targetID, ok := pathID(w, r)
	if !ok {
		return
	}
	var req resetPasswordRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := auth.ValidatePassword(req.Password); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	target, err := s.store.GetUser(r.Context(), targetID)
	if errors.Is(err, storage.ErrNotFound) {
		writeError(w, http.StatusNotFound, errors.New("user not found"))
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if err := s.store.UpdateUserPassword(r.Context(), targetID, hash); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	_, _ = s.store.RevokeUserSessions(r.Context(), targetID)
	s.logAudit(r, actor.ID, actor.Username, target.ID, target.Username, "password.reset", "")
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleSetUserAdmin(w http.ResponseWriter, r *http.Request) {
	actor, ok := requireAdmin(w, r)
	if !ok {
		return
	}
	targetID, ok := pathID(w, r)
	if !ok {
		return
	}
	var req struct {
		Admin bool `json:"admin"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	target, err := s.store.GetUser(r.Context(), targetID)
	if errors.Is(err, storage.ErrNotFound) {
		writeError(w, http.StatusNotFound, errors.New("user not found"))
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if target.IsAdmin && !req.Admin {
		n, err := s.store.AdminCount(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		if n <= 1 {
			writeError(w, http.StatusConflict, errors.New("cannot remove the last admin"))
			return
		}
	}
	if err := s.store.SetUserAdmin(r.Context(), targetID, req.Admin); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	details := "granted"
	if !req.Admin {
		details = "revoked"
	}
	s.logAudit(r, actor.ID, actor.Username, target.ID, target.Username, "user.admin", details)
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleDeleteUser(w http.ResponseWriter, r *http.Request) {
	actor, ok := requireAdmin(w, r)
	if !ok {
		return
	}
	targetID, ok := pathID(w, r)
	if !ok {
		return
	}
	if targetID == actor.ID {
		writeError(w, http.StatusConflict, errors.New("cannot delete your own account"))
		return
	}
	target, err := s.store.GetUser(r.Context(), targetID)
	if errors.Is(err, storage.ErrNotFound) {
		writeError(w, http.StatusNotFound, errors.New("user not found"))
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if target.IsAdmin {
		n, err := s.store.AdminCount(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		if n <= 1 {
			writeError(w, http.StatusConflict, errors.New("cannot delete the last admin"))
			return
		}
	}
	if err := s.store.DeleteUser(r.Context(), targetID); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	s.logAudit(r, actor.ID, actor.Username, target.ID, target.Username, "user.delete", "")
	s.emitWebhook(models.WebhookEventUserDelete, map[string]any{"userId": target.ID, "username": target.Username})
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleListAudit(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	limit := atoiDefault(r.URL.Query().Get("limit"), 50)
	offset := atoiDefault(r.URL.Query().Get("offset"), 0)
	action := r.URL.Query().Get("action")
	q := r.URL.Query().Get("q")
	page, err := s.store.ListAudit(r.Context(), limit, offset, action, q)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, page)
}

func (s *Server) logAudit(r *http.Request, actorID int64, actorName string, targetID int64, targetName, action, details string) {
	_ = s.store.InsertAudit(r.Context(), models.AuditEntry{
		ActorID:      actorID,
		ActorName:    actorName,
		TargetUserID: targetID,
		TargetName:   targetName,
		Action:       action,
		Details:      details,
		IP:           s.clientIP(r),
	})
}
