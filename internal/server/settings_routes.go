package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"athenaeum/internal/auth"
	"athenaeum/internal/library"
	"athenaeum/internal/models"
)

func (s *Server) registerSettingsRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/admin/server", s.handleGetServerConfig)
	mux.HandleFunc("PUT /api/admin/server", s.handlePutServerConfig)
}

func (s *Server) handleGetServerConfig(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	cfg, err := s.store.GetServerConfig(r.Context(), false)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, cfg.Public())
}

func (s *Server) handlePutServerConfig(w http.ResponseWriter, r *http.Request) {
	actor, ok := requireAdmin(w, r)
	if !ok {
		return
	}
	var req struct {
		MetricsEnabled   bool   `json:"metricsEnabled"`
		MetricsAuth      bool   `json:"metricsAuth"`
		MetricsUsername  string `json:"metricsUsername"`
		MetricsPassword  string `json:"metricsPassword"`
		TrustedProxies   string `json:"trustedProxies"`
		CORSEnabled      bool   `json:"corsEnabled"`
		CORSOrigins      string `json:"corsOrigins"`
		CSPEnabled       bool   `json:"cspEnabled"`
		CSPPolicy        string `json:"cspPolicy"`
		AutoScanEnabled  bool   `json:"autoScanEnabled"`
		AutoScanInterval int    `json:"autoScanIntervalSec"`
		ScanWorkers      int    `json:"scanWorkers"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	existing, err := s.store.GetServerConfig(r.Context(), true)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	cfg := models.ServerConfig{
		MetricsEnabled:   req.MetricsEnabled,
		MetricsAuth:      req.MetricsAuth,
		MetricsUsername:  strings.TrimSpace(req.MetricsUsername),
		TrustedProxies:   strings.TrimSpace(req.TrustedProxies),
		CORSEnabled:      req.CORSEnabled,
		CORSOrigins:      strings.TrimSpace(req.CORSOrigins),
		CSPEnabled:       req.CSPEnabled,
		CSPPolicy:        strings.TrimSpace(req.CSPPolicy),
		AutoScanEnabled:  req.AutoScanEnabled,
		AutoScanInterval: req.AutoScanInterval,
		ScanWorkers:      existing.ScanWorkers,
	}
	if cfg.AutoScanInterval < 60 {
		cfg.AutoScanInterval = 300
	}
	if req.ScanWorkers > 0 {
		cfg.ScanWorkers = library.ClampScanWorkers(req.ScanWorkers)
	}
	if req.MetricsPassword != "" {
		hash, err := auth.HashPassword(req.MetricsPassword)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		cfg.MetricsPassword = hash
	} else {
		cfg.MetricsPassword = existing.MetricsPassword
	}
	if cfg.MetricsEnabled && cfg.MetricsAuth {
		if cfg.MetricsUsername == "" {
			writeError(w, http.StatusBadRequest, errors.New("metrics username required when auth is enabled"))
			return
		}
		if cfg.MetricsPassword == "" && !existing.PasswordSet {
			writeError(w, http.StatusBadRequest, errors.New("metrics password required when auth is enabled"))
			return
		}
	}
	if err := s.store.SaveServerConfig(r.Context(), cfg); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	s.applyServerConfig(cfg)
	s.logAudit(r, actor.ID, actor.Username, 0, "", "server.config", "")
	out, _ := s.store.GetServerConfig(r.Context(), false)
	writeJSON(w, http.StatusOK, out.Public())
}

func (s *Server) applyServerConfig(cfg models.ServerConfig) {
	s.proxies.set(cfg.TrustedProxies)
	s.serverCfgMu.Lock()
	s.serverCfg = cfg.Public()
	s.serverCfgMu.Unlock()
}

func (s *Server) loadServerConfig(ctx context.Context) error {
	cfg, err := s.store.GetServerConfig(ctx, true)
	if err != nil {
		return err
	}
	s.applyServerConfig(cfg)
	return nil
}

func (s *Server) currentServerConfig() models.ServerConfigPublic {
	s.serverCfgMu.RLock()
	defer s.serverCfgMu.RUnlock()
	return s.serverCfg
}
