package server

import (
	"errors"
	"net/http"
	"strconv"

	"athenaeum/internal/auth"
	"athenaeum/internal/storage"
)

func (s *Server) registerSessionRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/auth/sessions", s.handleListSessions)
	mux.HandleFunc("DELETE /api/auth/sessions", s.handleRevokeOtherSessions)
	mux.HandleFunc("DELETE /api/auth/sessions/{id}", s.handleRevokeSession)
	mux.HandleFunc("GET /api/auth/users/{id}/sessions", s.handleListUserSessions)
}

func (s *Server) handleListSessions(w http.ResponseWriter, r *http.Request) {
	u, ok := UserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, errUnauthorized)
		return
	}
	current := ""
	if c, err := r.Cookie(auth.SessionCookie); err == nil {
		current = c.Value
	}
	sessions, err := s.store.ListUserSessions(r.Context(), u.ID, current)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, sessions)
}

func (s *Server) handleListUserSessions(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	targetID, ok := pathID(w, r)
	if !ok {
		return
	}
	if _, err := s.store.GetUser(r.Context(), targetID); errors.Is(err, storage.ErrNotFound) {
		writeError(w, http.StatusNotFound, errors.New("user not found"))
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	sessions, err := s.store.ListUserSessions(r.Context(), targetID, "")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, sessions)
}

func (s *Server) handleRevokeSession(w http.ResponseWriter, r *http.Request) {
	u, ok := UserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, errUnauthorized)
		return
	}
	sessionID := r.PathValue("id")
	if sessionID == "" {
		writeError(w, http.StatusBadRequest, errors.New("invalid session id"))
		return
	}

	targetUserID := u.ID
	if u.IsAdmin {
		if q := r.URL.Query().Get("userId"); q != "" {
			if id, err := strconv.ParseInt(q, 10, 64); err == nil && id > 0 {
				targetUserID = id
			}
		}
	}

	currentToken := ""
	if c, err := r.Cookie(auth.SessionCookie); err == nil {
		currentToken = c.Value
	}
	currentSessionID, _ := s.store.SessionIDForToken(r.Context(), currentToken)

	if err := s.store.RevokeSessionByID(r.Context(), targetUserID, sessionID); errors.Is(err, storage.ErrNotFound) {
		writeError(w, http.StatusNotFound, errors.New("session not found"))
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	target, _ := s.store.GetUser(r.Context(), targetUserID)
	details := sessionID
	if targetUserID != u.ID {
		details = "user " + target.Username + " session " + sessionID
	}
	s.logAudit(r, u.ID, u.Username, targetUserID, target.Username, "session.revoke", details)

	if sessionID == currentSessionID {
		s.clearAuthCookies(w, r)
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleRevokeOtherSessions(w http.ResponseWriter, r *http.Request) {
	u, ok := UserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, errUnauthorized)
		return
	}
	current := ""
	if c, err := r.Cookie(auth.SessionCookie); err == nil {
		current = c.Value
	}
	n, err := s.store.RevokeOtherSessions(r.Context(), u.ID, current)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	s.logAudit(r, u.ID, u.Username, u.ID, u.Username, "session.revoke", "revoked other sessions")
	writeJSON(w, http.StatusOK, map[string]any{"revoked": n})
}
