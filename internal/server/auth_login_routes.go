package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"athenaeum/internal/auth"
	"athenaeum/internal/models"
)

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Altcha   string `json:"altcha"`
}

type registerPublicRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Altcha   string `json:"altcha"`
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if !s.requireAltcha(w, r, "login", req.Altcha) {
		return
	}
	if req.Username == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, errBadCredentials)
		return
	}

	u, hash, err := s.store.GetUserByUsername(r.Context(), req.Username)
	if err != nil {
		writeError(w, http.StatusUnauthorized, errBadCredentials)
		return
	}
	if !auth.CheckPassword(hash, req.Password) {
		writeError(w, http.StatusUnauthorized, errBadCredentials)
		return
	}
	if hash == "" {
		writeError(w, http.StatusUnauthorized, errBadCredentials)
		return
	}

	if u.TOTPEnabled {
		token, err := s.newPendingTOTP(u.ID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, models.LoginChallenge{NeedsTOTP: true, TOTPToken: token})
		return
	}

	if err := s.issueAuthTokens(w, r, u.ID, "local"); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	s.logAudit(r, u.ID, u.Username, 0, "", "auth.login", "")
	writeJSON(w, http.StatusOK, u)
}

func (s *Server) handleRefresh(w http.ResponseWriter, r *http.Request) {
	u, err := s.rotateFromRefreshCookie(w, r, true)
	if err != nil {
		if err == errUnauthorized {
			writeError(w, http.StatusUnauthorized, errUnauthorized)
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, u)
}

func (s *Server) issueAuthTokens(w http.ResponseWriter, r *http.Request, userID int64, authMethod string) error {
	sessionID, err := auth.NewToken()
	if err != nil {
		return err
	}
	accessToken, err := auth.NewSessionToken()
	if err != nil {
		return err
	}
	refreshToken, err := auth.NewSessionToken()
	if err != nil {
		return err
	}
	accessExp := time.Now().Add(auth.AccessTTL)
	refreshExp := time.Now().Add(auth.RefreshTTL)
	if err := s.store.CreateUserSession(r.Context(), models.SessionCreate{
		SessionID:      sessionID,
		AccessToken:    accessToken,
		RefreshToken:   refreshToken,
		UserID:         userID,
		IP:             s.clientIP(r),
		UserAgent:      r.UserAgent(),
		Device:         auth.ParseDevice(r.UserAgent()),
		AuthMethod:     authMethod,
		AccessExpires:  accessExp,
		RefreshExpires: refreshExp,
	}); err != nil {
		return err
	}
	s.setAuthCookies(w, r, accessToken, refreshToken, accessExp, refreshExp)
	if _, err := s.rotateCSRFCookie(w, r); err != nil {
		return err
	}
	return nil
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	if u, ok := UserFromContext(r.Context()); ok {
		s.logAudit(r, u.ID, u.Username, 0, "", "auth.logout", "")
	}
	if c, err := r.Cookie(auth.SessionCookie); err == nil && c.Value != "" {
		_ = s.store.DeleteSessionByAccessToken(r.Context(), c.Value)
	}
	if c, err := r.Cookie(auth.RefreshCookie); err == nil && c.Value != "" {
		_ = s.store.DeleteRefreshByToken(r.Context(), c.Value)
	}
	s.clearAuthCookies(w, r)
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	required, err := s.store.AuthRequired(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	var req registerRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
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
	if !required {
		writeError(w, http.StatusForbidden, errors.New("initial setup required"))
		return
	}

	admin := false
	var actorID int64
	var actorName string
	actor, ok := requireAdmin(w, r)
	if !ok {
		return
	}
	actorID = actor.ID
	actorName = actor.Username

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	id, err := s.store.CreateUser(r.Context(), req.Username, hash, admin)
	if err != nil {
		writeError(w, http.StatusConflict, errUsernameTaken)
		return
	}
	u, err := s.store.GetUser(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if actorID > 0 {
		s.logAudit(r, actorID, actorName, u.ID, u.Username, "user.create", "")
	}
	s.emitWebhook(models.WebhookEventUserCreate, map[string]any{"userId": u.ID, "username": u.Username})
	writeJSON(w, http.StatusCreated, u)
}

func (s *Server) handleRegisterPublic(w http.ResponseWriter, r *http.Request) {
	authSettings, err := s.store.GetAuthSettings(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if !authSettings.AllowRegistration {
		writeError(w, http.StatusForbidden, errors.New("self-registration is disabled"))
		return
	}
	var req registerPublicRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if !s.requireAltcha(w, r, "register-public", req.Altcha) {
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
	id, err := s.store.CreateUser(r.Context(), req.Username, hash, false)
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
	s.logAudit(r, u.ID, u.Username, u.ID, u.Username, "user.create", "self-registration")
	s.emitWebhook(models.WebhookEventUserCreate, map[string]any{"userId": u.ID, "username": u.Username})
	writeJSON(w, http.StatusCreated, u)
}
