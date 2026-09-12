package server

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"athenaeum/internal/backup"
	"athenaeum/internal/brand"
	"athenaeum/internal/models"
)

func (s *Server) registerAdminBackupRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/admin/backup", s.handleBackup)
	mux.HandleFunc("POST /api/admin/restore", s.handleRestore)
	mux.HandleFunc("GET /api/admin/config/export", s.handleConfigExport)
	mux.HandleFunc("POST /api/admin/config/import", s.handleConfigImport)
}

func (s *Server) handleBackup(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	w.Header().Set("Content-Type", mimeZip)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s-%s.zip"`, brand.BackupPrefix, time.Now().UTC().Format("20060102-150405")))

	var items []backup.Item
	if !s.cfg.UsesPostgres() {
		items = append(items, backup.Item{Name: brand.DBFilename, Path: s.cfg.DBPath()})
	} else {
		items = append(items, backup.Item{Name: "DATABASE.txt", Data: []byte(backup.PostgresNotice)})
	}
	items = append(items,
		backup.Item{Name: "covers", Path: s.cfg.CoverDir()},
		backup.Item{Name: "i18n", Path: s.cfg.I18nDir()},
	)
	cfg, err := s.buildConfigExport(r)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	items = append(items, backup.Item{Name: "config.json", Data: data})
	if err := backup.Write(w, items); err != nil {
		writeError(w, http.StatusInternalServerError, err)
	}
}

func addZipFile(zw *zip.Writer, name, path string) error {
	return backup.AddFile(zw, name, path)
}

func (s *Server) handleRestore(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	if err := r.ParseMultipartForm(512 << 20); err != nil { // #nosec G120 -- admin-only restore; 512MB cap
		writeError(w, http.StatusBadRequest, err)
		return
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, errors.New("file required"))
		return
	}
	defer file.Close()
	if err := backup.Restore(file, s.cfg.DataDir, brand.DBFilename); err != nil {
		var cerr *backup.ClientError
		if errors.As(err, &cerr) {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "restored", "message": "restart Athenaeum to load restored database"})
}

type configExport struct {
	ExportedAt time.Time                 `json:"exportedAt"`
	Server     models.ServerConfigPublic `json:"server"`
	OIDC       models.OIDCConfig         `json:"oidc"`
	Libraries  []models.Library          `json:"libraries"`
}

func (s *Server) buildConfigExport(r *http.Request) (configExport, error) {
	var out configExport
	out.ExportedAt = time.Now().UTC()
	srv, err := s.store.GetServerConfig(r.Context(), false)
	if err != nil {
		return out, err
	}
	out.Server = srv.Public()
	oidc, err := s.store.GetOIDCConfig(r.Context(), false)
	if err != nil {
		return out, err
	}
	oidc.ClientSecret = ""
	out.OIDC = oidc
	libs, err := s.store.ListLibraries(r.Context())
	if err != nil {
		return out, err
	}
	out.Libraries = libs
	return out, nil
}

func (s *Server) handleConfigExport(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	cfg, err := s.buildConfigExport(r)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	w.Header().Set("Content-Disposition", `attachment; filename="`+brand.ConfigExportName+`"`)
	writeJSON(w, http.StatusOK, cfg)
}

func (s *Server) handleConfigImport(w http.ResponseWriter, r *http.Request) {
	actor, ok := requireAdmin(w, r)
	if !ok {
		return
	}
	var req struct {
		Server    *models.ServerConfigPublic `json:"server"`
		OIDC      *models.OIDCConfig         `json:"oidc"`
		Libraries []models.Library           `json:"libraries"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if req.Server != nil {
		existing, err := s.store.GetServerConfig(r.Context(), true)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		cfg := models.ServerConfig{
			MetricsEnabled:   req.Server.MetricsEnabled,
			MetricsAuth:      req.Server.MetricsAuth,
			MetricsUsername:  req.Server.MetricsUsername,
			TrustedProxies:   req.Server.TrustedProxies,
			CORSEnabled:      req.Server.CORSEnabled,
			CORSOrigins:      req.Server.CORSOrigins,
			CSPEnabled:       req.Server.CSPEnabled,
			CSPPolicy:        req.Server.CSPPolicy,
			AutoScanEnabled:  req.Server.AutoScanEnabled,
			AutoScanInterval: req.Server.AutoScanInterval,
			MetricsPassword:  existing.MetricsPassword,
		}
		if err := s.store.SaveServerConfig(r.Context(), cfg); err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		s.applyServerConfig(cfg)
	}
	if req.OIDC != nil {
		if err := s.store.SaveOIDCConfig(r.Context(), *req.OIDC); err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
	}
	if len(req.Libraries) > 0 {
		existing, err := s.store.ListLibraries(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		byID := map[int64]models.Library{}
		for _, lib := range existing {
			byID[lib.ID] = lib
		}
		for _, lib := range req.Libraries {
			if lib.ID == 0 {
				continue
			}
			if cur, ok := byID[lib.ID]; ok {
				name := lib.Name
				if name == "" {
					name = cur.Name
				}
				mount := cur.MountPath
				if _, err := s.store.UpdateLibrary(r.Context(), lib.ID, name, mount); err != nil {
					writeError(w, http.StatusInternalServerError, err)
					return
				}
			}
		}
	}
	s.logAudit(r, actor.ID, actor.Username, 0, "", "config.import", "")
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
