package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"athenaeum/internal/auth"
	"athenaeum/internal/models"
	"athenaeum/internal/pocketid"
)

func (s *Server) registerPocketIDRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/admin/pocketid", s.handleGetPocketID)
	mux.HandleFunc("PUT /api/admin/pocketid", s.handlePutPocketID)
	mux.HandleFunc("POST /api/admin/pocketid/test", s.handleTestPocketID)
	mux.HandleFunc("POST /api/admin/pocketid/apply-oidc", s.handleApplyPocketIDOIDC)
	mux.HandleFunc("POST /api/admin/pocketid/signup-tokens", s.handleCreatePocketIDSignupToken)
	mux.HandleFunc("GET /api/admin/pocketid/signup-tokens", s.handleListPocketIDSignupTokens)
	mux.HandleFunc("DELETE /api/admin/pocketid/signup-tokens/{id}", s.handleDeletePocketIDSignupToken)
}

func (s *Server) handleGetPocketID(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	cfg, err := s.store.GetPocketIDSettings(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, cfg.Public())
}

func (s *Server) handlePutPocketID(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	var cfg models.PocketIDSettings
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&cfg); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if cfg.APIKey == "" {
		existing, err := s.store.GetPocketIDSettings(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		cfg.APIKey = existing.APIKey
	}
	cfg.BaseURL = strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if cfg.DefaultGroupIDs == nil {
		cfg.DefaultGroupIDs = []string{}
	}
	if err := s.store.SavePocketIDSettings(r.Context(), cfg); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, cfg.Public())
}

func (s *Server) pocketIDClient(r *http.Request) (*pocketid.Client, error) {
	cfg, err := s.store.GetPocketIDSettings(r.Context())
	if err != nil {
		return nil, err
	}
	if cfg.BaseURL == "" || cfg.APIKey == "" {
		return nil, errors.New("pocket id is not configured")
	}
	return pocketid.NewClient(cfg.BaseURL, cfg.APIKey), nil
}

func (s *Server) handleTestPocketID(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	client, err := s.pocketIDClient(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := client.ListUsers(r.Context()); err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"ok": "connected"})
}

func (s *Server) handleApplyPocketIDOIDC(w http.ResponseWriter, r *http.Request) {
	actor, ok := requireAdmin(w, r)
	if !ok {
		return
	}
	cfg, err := s.store.GetPocketIDSettings(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if cfg.BaseURL == "" {
		writeError(w, http.StatusBadRequest, errors.New("pocket id base url is required"))
		return
	}
	endpoints, err := auth.DiscoverOIDC(r.Context(), cfg.BaseURL)
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	existing, err := s.store.GetOIDCConfig(r.Context(), true)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	existing.IssuerURL = endpoints.Issuer
	existing.AuthorizeURL = endpoints.AuthURL
	existing.TokenURL = endpoints.TokenURL
	existing.UserinfoURL = endpoints.UserinfoURL
	existing.JWKSURL = endpoints.JWKSURL
	existing.LogoutURL = endpoints.LogoutURL
	existing.Enabled = true
	existing.MatchBy = models.OIDCMatchEmail
	existing.AutoRegister = true
	if existing.ButtonText == "" {
		existing.ButtonText = "Sign in with Pocket ID"
	}
	if existing.GroupClaim == "" {
		existing.GroupClaim = "groups"
	}
	if err := s.store.SaveOIDCConfig(r.Context(), existing); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	s.logAudit(r, actor.ID, actor.Username, 0, "", "oidc.config", "pocketid-apply")
	out, err := s.store.GetOIDCConfig(r.Context(), false)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleCreatePocketIDSignupToken(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	client, err := s.pocketIDClient(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	var req struct {
		TTL          string   `json:"ttl"`
		UsageLimit   int      `json:"usageLimit"`
		UserGroupIDs []string `json:"userGroupIds"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if req.TTL == "" {
		req.TTL = "24h"
	}
	if req.UsageLimit <= 0 {
		req.UsageLimit = 1
	}
	tok, err := client.CreateSignupToken(r.Context(), req.TTL, req.UsageLimit, req.UserGroupIDs)
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusCreated, tok)
}

func (s *Server) handleListPocketIDSignupTokens(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	client, err := s.pocketIDClient(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	list, err := client.ListSignupTokens(r.Context())
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	if list == nil {
		list = []pocketid.SignupToken{}
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) handleDeletePocketIDSignupToken(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	client, err := s.pocketIDClient(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, errors.New("id required"))
		return
	}
	if err := client.DeleteSignupToken(r.Context(), id); err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"ok": "deleted"})
}
