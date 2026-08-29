package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"athenaeum/internal/auth"
)

const (
	totpPendingTTL     = 5 * time.Minute
	totpMaxVerifyTries = 8
)

// pendingTOTP tracks a user who passed the password check and is waiting on
// a second factor before a session is issued.
type pendingTOTP struct {
	userID    int64
	expiresAt time.Time
	attempts  int
}

func (s *Server) registerTOTPRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/auth/totp/setup", s.handleTOTPSetup)
	mux.HandleFunc("POST /api/auth/totp/enable", s.handleTOTPEnable)
	mux.HandleFunc("POST /api/auth/totp/disable", s.handleTOTPDisable)
	mux.HandleFunc("POST /api/auth/totp/verify", s.handleTOTPVerifyLogin)
}

// newPendingTOTP issues a one-time token that identifies a user who has
// passed password verification but still needs to submit a TOTP code.
func (s *Server) newPendingTOTP(userID int64) (string, error) {
	token, err := auth.NewToken()
	if err != nil {
		return "", err
	}
	s.totpMu.Lock()
	defer s.totpMu.Unlock()
	s.pruneExpiredTOTPLocked()
	s.totpPending[token] = pendingTOTP{userID: userID, expiresAt: time.Now().Add(totpPendingTTL)}
	return token, nil
}

func (s *Server) pruneExpiredTOTPLocked() {
	now := time.Now()
	for token, entry := range s.totpPending {
		if now.After(entry.expiresAt) {
			delete(s.totpPending, token)
		}
	}
}

var errTOTPTokenInvalid = errors.New("totp challenge expired or invalid, sign in again")

// consumeTOTPPending validates and records an attempt against a pending
// token, removing it once it is used up or exceeds its attempt budget.
func (s *Server) consumeTOTPPending(token string, success bool) (int64, error) {
	s.totpMu.Lock()
	defer s.totpMu.Unlock()
	entry, ok := s.totpPending[token]
	if !ok || time.Now().After(entry.expiresAt) {
		delete(s.totpPending, token)
		return 0, errTOTPTokenInvalid
	}
	if success {
		delete(s.totpPending, token)
		return entry.userID, nil
	}
	entry.attempts++
	if entry.attempts >= totpMaxVerifyTries {
		delete(s.totpPending, token)
		return 0, errTOTPTokenInvalid
	}
	s.totpPending[token] = entry
	return 0, errBadTOTPCode
}

var errBadTOTPCode = errors.New("invalid authenticator code")

type totpSetupResponse struct {
	Secret     string `json:"secret"`
	OtpauthURL string `json:"otpauthUrl"`
}

func (s *Server) handleTOTPSetup(w http.ResponseWriter, r *http.Request) {
	u, ok := UserFromContext(r.Context())
	if !ok {
		writeErrorReq(w, r, http.StatusUnauthorized, errUnauthorized)
		return
	}
	secret, otpauthURL, err := auth.GenerateTOTPSecret(u.Username)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if err := s.store.SetUserTOTPSecret(r.Context(), u.ID, secret); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, totpSetupResponse{Secret: secret, OtpauthURL: otpauthURL})
}

type totpCodeRequest struct {
	Code string `json:"code"`
}

func (s *Server) handleTOTPEnable(w http.ResponseWriter, r *http.Request) {
	u, ok := UserFromContext(r.Context())
	if !ok {
		writeErrorReq(w, r, http.StatusUnauthorized, errUnauthorized)
		return
	}
	var req totpCodeRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	secret, err := s.store.GetUserTOTPSecret(r.Context(), u.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if secret == "" {
		writeError(w, http.StatusConflict, errors.New("run totp setup first"))
		return
	}
	if !auth.ValidateTOTPCode(secret, req.Code) {
		writeError(w, http.StatusUnauthorized, errBadTOTPCode)
		return
	}
	if err := s.store.EnableUserTOTP(r.Context(), u.ID); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	s.logAudit(r, u.ID, u.Username, u.ID, u.Username, "totp.enable", "")
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

type totpDisableRequest struct {
	Password string `json:"password"`
	Code     string `json:"code"`
}

func (s *Server) handleTOTPDisable(w http.ResponseWriter, r *http.Request) {
	u, ok := UserFromContext(r.Context())
	if !ok {
		writeErrorReq(w, r, http.StatusUnauthorized, errUnauthorized)
		return
	}
	var req totpDisableRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if u.LocalAuth {
		_, hash, err := s.store.GetUserByUsername(r.Context(), u.Username)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		if !auth.CheckPassword(hash, req.Password) {
			writeError(w, http.StatusUnauthorized, errors.New("current password is incorrect"))
			return
		}
	}
	secret, err := s.store.GetUserTOTPSecret(r.Context(), u.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if !auth.ValidateTOTPCode(secret, req.Code) {
		writeError(w, http.StatusUnauthorized, errBadTOTPCode)
		return
	}
	if err := s.store.DisableUserTOTP(r.Context(), u.ID); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	s.logAudit(r, u.ID, u.Username, u.ID, u.Username, "totp.disable", "")
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

type totpVerifyRequest struct {
	TOTPToken string `json:"totpToken"`
	Code      string `json:"code"`
}

func (s *Server) handleTOTPVerifyLogin(w http.ResponseWriter, r *http.Request) {
	var req totpVerifyRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	s.totpMu.Lock()
	entry, ok := s.totpPending[req.TOTPToken]
	s.totpMu.Unlock()
	if !ok || time.Now().After(entry.expiresAt) {
		writeError(w, http.StatusUnauthorized, errTOTPTokenInvalid)
		return
	}
	secret, err := s.store.GetUserTOTPSecret(r.Context(), entry.userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if !auth.ValidateTOTPCode(secret, req.Code) {
		if _, err := s.consumeTOTPPending(req.TOTPToken, false); err != nil {
			writeError(w, http.StatusUnauthorized, err)
			return
		}
		writeError(w, http.StatusUnauthorized, errBadTOTPCode)
		return
	}
	userID, err := s.consumeTOTPPending(req.TOTPToken, true)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err)
		return
	}
	u, err := s.store.GetUser(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if err := s.issueAuthTokens(w, r, u.ID, "local"); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	s.logAudit(r, u.ID, u.Username, 0, "", "auth.login", "totp")
	writeJSON(w, http.StatusOK, u)
}
