package server

import (
	"net/http"
	"time"

	"athenaeum/internal/auth"
	"athenaeum/internal/models"
)

func (s *Server) rotateFromRefreshCookie(w http.ResponseWriter, r *http.Request, clearOnFailure bool) (models.User, error) {
	c, err := r.Cookie(auth.RefreshCookie)
	if err != nil || c.Value == "" {
		if clearOnFailure {
			s.clearAuthCookiesUnlessSessionValid(w, r)
		}
		return models.User{}, errUnauthorized
	}

	u, err := s.store.RefreshTokenUser(r.Context(), c.Value)
	if err != nil {
		if clearOnFailure {
			s.clearAuthCookiesUnlessSessionValid(w, r)
		}
		return models.User{}, errUnauthorized
	}

	sessionID, _ := s.store.RefreshTokenSessionID(r.Context(), c.Value)

	newAccess, err := auth.NewSessionToken()
	if err != nil {
		return models.User{}, err
	}
	newRefresh, err := auth.NewSessionToken()
	if err != nil {
		return models.User{}, err
	}
	accessExp := time.Now().Add(auth.AccessTTL)
	refreshExp := time.Now().Add(auth.RefreshTTL)

	if sessionID != "" {
		if err := s.store.RotateSessionTokens(r.Context(), sessionID, c.Value, newAccess, newRefresh, accessExp, refreshExp); err != nil {
			return models.User{}, err
		}
	} else {
		if err := s.store.DeleteRefreshToken(r.Context(), c.Value); err != nil {
			return models.User{}, err
		}
		if sessionCookie, err := r.Cookie(auth.SessionCookie); err == nil && sessionCookie.Value != "" {
			_ = s.store.DeleteSession(r.Context(), sessionCookie.Value)
		}
		sessID, err := auth.NewToken()
		if err != nil {
			return models.User{}, err
		}
		if err := s.store.CreateUserSession(r.Context(), models.SessionCreate{
			SessionID:      sessID,
			AccessToken:    newAccess,
			RefreshToken:   newRefresh,
			UserID:         u.ID,
			IP:             s.clientIP(r),
			UserAgent:      r.UserAgent(),
			Device:         auth.ParseDevice(r.UserAgent()),
			AuthMethod:     "local",
			AccessExpires:  accessExp,
			RefreshExpires: refreshExp,
		}); err != nil {
			return models.User{}, err
		}
	}

	s.setAuthCookies(w, r, newAccess, newRefresh, accessExp, refreshExp)
	if _, err := s.rotateCSRFCookie(w, r); err != nil {
		return models.User{}, err
	}
	return u, nil
}

func (s *Server) trySilentRefresh(w http.ResponseWriter, r *http.Request) (models.User, bool) {
	u, err := s.rotateFromRefreshCookie(w, r, false)
	if err != nil {
		return models.User{}, false
	}
	return u, true
}
