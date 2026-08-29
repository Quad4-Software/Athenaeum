package server

import (
	"encoding/json"
	"errors"
	"net/http"

	"athenaeum/internal/auth"
	"athenaeum/internal/models"
)

type registerRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Altcha   string `json:"altcha"`
}

type resetPasswordRequest struct {
	Password string `json:"password"`
}

func (s *Server) handleCheckPassword(w http.ResponseWriter, r *http.Request) {
	var req resetPasswordRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, auth.ScorePassword(req.Password))
}

func (s *Server) passwordPolicyPublic() *models.PasswordPolicy {
	p := auth.GetPasswordPolicy()
	return &models.PasswordPolicy{
		MinLength:     p.MinLength,
		LongLength:    p.LongLength,
		MinKinds:      p.MinKinds,
		RequireLower:  p.RequireLower,
		RequireUpper:  p.RequireUpper,
		RequireDigit:  p.RequireDigit,
		RequireSymbol: p.RequireSymbol,
	}
}

func (s *Server) handleAuthSetup(w http.ResponseWriter, r *http.Request) {
	required, err := s.store.AuthRequired(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	authSettings, err := s.store.GetAuthSettings(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	out := map[string]any{
		"needed":            !required,
		"authEnabled":       required,
		"allowRegistration": authSettings.AllowRegistration,
		"passwordPolicy":    s.passwordPolicyPublic(),
	}
	if pub := s.altchaPublic(); pub.Enabled {
		out["altcha"] = pub
	}
	s.issueCSRFForAuthPages(w, r)
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleAuthSetupPost(w http.ResponseWriter, r *http.Request) {
	required, err := s.store.AuthRequired(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if required {
		writeError(w, http.StatusConflict, errors.New("setup already completed"))
		return
	}
	var req registerRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if !s.requireAltcha(w, r, "setup", req.Altcha) {
		return
	}
	if len(req.Username) < 2 {
		writeError(w, http.StatusBadRequest, errWeakCredentials)
		return
	}
	if err := auth.ValidatePassword(req.Password); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	id, err := s.store.CreateUser(r.Context(), req.Username, hash, true)
	if err != nil {
		writeError(w, http.StatusConflict, errUsernameTaken)
		return
	}
	u, err := s.store.GetUser(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if err := s.issueAuthTokens(w, r, u.ID, "local"); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	s.logAudit(r, u.ID, u.Username, u.ID, u.Username, "user.create", "initial admin")
	s.emitWebhook(models.WebhookEventUserCreate, map[string]any{"userId": u.ID, "username": u.Username})
	writeJSON(w, http.StatusCreated, u)
}
