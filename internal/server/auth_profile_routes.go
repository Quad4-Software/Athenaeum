package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"athenaeum/internal/auth"
	"athenaeum/internal/models"
)

type profileRequest struct {
	Username string `json:"username"`
}

type changePasswordRequest struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

func (s *Server) handleGetAuthSettings(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	settings, err := s.store.GetAuthSettings(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, settings)
}

func (s *Server) handlePutAuthSettings(w http.ResponseWriter, r *http.Request) {
	actor, ok := requireAdmin(w, r)
	if !ok {
		return
	}
	var settings models.AuthSettings
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&settings); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := s.store.SaveAuthSettings(r.Context(), settings); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	s.logAudit(r, actor.ID, actor.Username, 0, "", "auth.settings", mapEnabled(settings.AllowRegistration))
	writeJSON(w, http.StatusOK, settings)
}

func (s *Server) handleUpdateProfile(w http.ResponseWriter, r *http.Request) {
	u, ok := UserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, errUnauthorized)
		return
	}
	var req profileRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	username := strings.TrimSpace(req.Username)
	if len(username) < 2 {
		writeError(w, http.StatusBadRequest, errors.New("username must be at least 2 characters"))
		return
	}
	if strings.EqualFold(username, u.Username) {
		writeJSON(w, http.StatusOK, u)
		return
	}
	taken, err := s.store.UsernameTaken(r.Context(), username, u.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if taken {
		writeError(w, http.StatusConflict, errUsernameTaken)
		return
	}
	oldName := u.Username
	if err := s.store.UpdateUsername(r.Context(), u.ID, username); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	u.Username = username
	s.logAudit(r, u.ID, username, u.ID, oldName, "user.rename", "renamed from "+oldName)
	writeJSON(w, http.StatusOK, u)
}

func (s *Server) handleChangePassword(w http.ResponseWriter, r *http.Request) {
	u, ok := UserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, errUnauthorized)
		return
	}
	var req changePasswordRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := auth.ValidatePassword(req.NewPassword); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	_, hash, err := s.store.GetUserByUsername(r.Context(), u.Username)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if !auth.CheckPassword(hash, req.CurrentPassword) {
		writeError(w, http.StatusUnauthorized, errors.New("current password is incorrect"))
		return
	}
	newHash, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if err := s.store.UpdateUserPassword(r.Context(), u.ID, newHash); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if _, err := s.store.RevokeUserSessions(r.Context(), u.ID); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if err := s.issueAuthTokens(w, r, u.ID, "local"); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	s.logAudit(r, u.ID, u.Username, u.ID, u.Username, "password.change", "")
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
