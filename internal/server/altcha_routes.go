package server

import (
	"errors"
	"net/http"

	"athenaeum/internal/altcha"
)

func (s *Server) registerAltchaRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/altcha/challenge", s.handleAltchaChallenge)
}

func (s *Server) handleAltchaChallenge(w http.ResponseWriter, r *http.Request) {
	if s.altcha == nil || !s.altcha.Enabled() || s.altcha.Mode() != altcha.ModeBuiltin {
		writeError(w, http.StatusNotFound, errors.New("altcha challenge endpoint unavailable"))
		return
	}
	challenge, err := s.altcha.CreateChallenge()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, challenge)
}

func (s *Server) requireAltcha(w http.ResponseWriter, r *http.Request, action, payload string) bool {
	if s.altcha == nil {
		return true
	}
	if err := s.altcha.VerifyPayload(r.Context(), action, payload); err != nil {
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, altcha.ErrMissingPayload):
			status = http.StatusBadRequest
		case errors.Is(err, altcha.ErrExpired), errors.Is(err, altcha.ErrInvalidPayload):
			status = http.StatusUnprocessableEntity
		}
		writeError(w, status, err)
		return false
	}
	return true
}

func (s *Server) altchaPublic() altcha.PublicConfig {
	if s.altcha == nil {
		return altcha.PublicConfig{}
	}
	return s.altcha.Public()
}
