package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"athenaeum/internal/storage"
)

func (s *Server) registerGuestRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/auth/users/guests", s.handleListGuests)
	mux.HandleFunc("POST /api/auth/users/guests/bulk-delete", s.handleBulkDeleteGuests)
	mux.HandleFunc("POST /api/auth/users/guests/{id}/extend", s.handleExtendGuest)
}

func (s *Server) handleListGuests(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	hours := 0
	if v := r.URL.Query().Get("expiringWithinHours"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			writeError(w, http.StatusBadRequest, errors.New("invalid expiringWithinHours"))
			return
		}
		hours = n
	}
	users, err := s.store.ListGuestUsers(r.Context(), hours)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	public := make([]any, len(users))
	for i, u := range users {
		public[i] = u.Public()
	}
	writeJSON(w, http.StatusOK, public)
}

func (s *Server) handleBulkDeleteGuests(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	var req struct {
		IDs []int64 `json:"ids"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if len(req.IDs) == 0 {
		writeError(w, http.StatusBadRequest, errors.New("ids required"))
		return
	}
	n, err := s.store.DeleteGuestUsers(r.Context(), req.IDs)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]int64{"deleted": n})
}

func (s *Server) handleExtendGuest(w http.ResponseWriter, r *http.Request) {
	actor, ok := requireAdmin(w, r)
	if !ok {
		return
	}
	id, ok := userPathID(w, r)
	if !ok {
		return
	}
	var req struct {
		ExpiresInHours int       `json:"expiresInHours"`
		ExpiresAt      time.Time `json:"expiresAt"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	var expires time.Time
	if !req.ExpiresAt.IsZero() {
		expires = req.ExpiresAt
	} else {
		hours := req.ExpiresInHours
		if hours <= 0 {
			hours = 24
		}
		u, err := s.store.GetUser(r.Context(), id)
		if errors.Is(err, storage.ErrNotFound) {
			writeError(w, http.StatusNotFound, err)
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		base := time.Now()
		if u.ExpiresAt != nil && u.ExpiresAt.After(base) {
			base = *u.ExpiresAt
		}
		expires = base.Add(time.Duration(hours) * time.Hour)
	}
	if err := s.store.ExtendGuestExpiry(r.Context(), id, expires); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			writeError(w, http.StatusNotFound, err)
		} else {
			writeError(w, http.StatusInternalServerError, err)
		}
		return
	}
	u, _ := s.store.GetUser(r.Context(), id)
	s.logAudit(r, actor.ID, actor.Username, id, u.Username, "user.guest.extend", expires.Format(time.RFC3339))
	writeJSON(w, http.StatusOK, u.Public())
}
