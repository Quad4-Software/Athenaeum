package server

import (
	"errors"
	"net/http"
)

type csrfResponse struct {
	CSRFToken string `json:"csrfToken"`
}

func (s *Server) registerAuthRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/auth/csrf", s.handleCSRF)
	mux.HandleFunc("GET /api/auth/setup", s.handleAuthSetup)
	mux.HandleFunc("POST /api/auth/setup", s.handleAuthSetupPost)
	mux.HandleFunc("POST /api/auth/login", s.handleLogin)
	mux.HandleFunc("POST /api/auth/logout", s.handleLogout)
	mux.HandleFunc("POST /api/auth/refresh", s.handleRefresh)
	mux.HandleFunc("GET /api/auth/me", s.handleMe)
	mux.HandleFunc("POST /api/auth/register", s.handleRegister)
	mux.HandleFunc("POST /api/auth/register-public", s.handleRegisterPublic)
	mux.HandleFunc("GET /api/auth/settings", s.handleGetAuthSettings)
	mux.HandleFunc("PUT /api/auth/settings", s.handlePutAuthSettings)
	mux.HandleFunc("PUT /api/auth/profile", s.handleUpdateProfile)
	mux.HandleFunc("PUT /api/auth/password", s.handleChangePassword)
	mux.HandleFunc("GET /api/auth/reader-prefs", s.handleGetReaderPrefs)
	mux.HandleFunc("PUT /api/auth/reader-prefs", s.handlePutReaderPrefs)
	mux.HandleFunc("GET /api/auth/users", s.handleListUsers)
	mux.HandleFunc("PUT /api/auth/users/{id}/password", s.handleResetUserPassword)
	mux.HandleFunc("PUT /api/auth/users/{id}/admin", s.handleSetUserAdmin)
	mux.HandleFunc("PUT /api/auth/users/{id}/permissions", s.handleSetUserPermissions)
	mux.HandleFunc("DELETE /api/auth/users/{id}", s.handleDeleteUser)
	mux.HandleFunc("POST /api/auth/users/guest", s.handleCreateGuest)
	mux.HandleFunc("GET /api/auth/audit", s.handleListAudit)
	mux.HandleFunc("POST /api/auth/password/check", s.handleCheckPassword)
	s.registerAltchaRoutes(mux)
	s.registerSessionRoutes(mux)
	s.registerAPIKeyRoutes(mux)
	s.registerOIDCRoutes(mux)
	s.registerTOTPRoutes(mux)
}

func (s *Server) handleCSRF(w http.ResponseWriter, r *http.Request) {
	token, err := s.ensureCSRFCookie(w, r)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, csrfResponse{CSRFToken: token})
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	u, ok := UserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, errUnauthorized)
		return
	}
	writeJSON(w, http.StatusOK, u.Public())
}

var (
	errBadCredentials  = errors.New("invalid username or password")
	errWeakCredentials = errors.New("username must be at least 2 characters and password at least 8")
	errUsernameTaken   = errors.New("username already taken")
	errForbidden       = errors.New("forbidden")
)
