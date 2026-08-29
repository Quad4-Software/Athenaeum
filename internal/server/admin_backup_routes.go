package server

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

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
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s-%s.zip"`, brand.BackupPrefix, time.Now().UTC().Format("20060102-150405")))
	zw := zip.NewWriter(w)
	defer zw.Close()

	addFile := func(name, path string) error {
		info, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if info.IsDir() {
			return filepath.Walk(path, func(p string, fi os.FileInfo, err error) error {
				if err != nil || fi.IsDir() {
					return err
				}
				rel, err := filepath.Rel(path, p)
				if err != nil {
					return err
				}
				return addZipFile(zw, filepath.ToSlash(filepath.Join(name, rel)), p)
			})
		}
		return addZipFile(zw, name, path)
	}

	if !s.cfg.UsesPostgres() {
		if err := addFile(brand.DBFilename, s.cfg.DBPath()); err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
	} else {
		fw, err := zw.Create("DATABASE.txt")
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		_, _ = io.WriteString(fw, "Athenaeum is using PostgreSQL. Back up the database with pg_dump separately.\n")
	}
	if err := addFile("covers", s.cfg.CoverDir()); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if err := addFile("i18n", s.cfg.I18nDir()); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
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
	fw, err := zw.Create("config.json")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	_, _ = fw.Write(data)
}

func addZipFile(zw *zip.Writer, name, path string) error {
	f, err := os.Open(path) // #nosec G304 -- admin backup reads fixed server paths
	if err != nil {
		return err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return err
	}
	hdr, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	hdr.Name = name
	hdr.Method = zip.Deflate
	w, err := zw.CreateHeader(hdr)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, f)
	return err
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
	tmp, err := os.CreateTemp(s.cfg.DataDir, "restore-*.zip")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	if _, err := io.Copy(tmp, file); err != nil {
		_ = tmp.Close()
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	_ = tmp.Close()

	zr, err := zip.OpenReader(tmpPath)
	if err != nil {
		writeError(w, http.StatusBadRequest, errors.New("invalid zip archive"))
		return
	}
	defer zr.Close()
	hasDB := false
	for _, f := range zr.File {
		if f.Name == brand.DBFilename || strings.HasSuffix(f.Name, "/"+brand.DBFilename) {
			hasDB = true
			break
		}
	}
	if !hasDB {
		writeError(w, http.StatusBadRequest, errors.New("archive must contain "+brand.DBFilename))
		return
	}
	for _, f := range zr.File {
		if f.FileInfo().IsDir() {
			continue
		}
		name := filepath.Clean(filepath.FromSlash(f.Name))
		if name == "." || name == "" || strings.HasPrefix(name, ".."+string(os.PathSeparator)) || name == ".." {
			continue
		}
		if filepath.IsAbs(name) {
			continue
		}
		dest := filepath.Join(s.cfg.DataDir, name)
		rel, err := filepath.Rel(s.cfg.DataDir, dest)
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
			continue
		}
		if err := os.MkdirAll(filepath.Dir(dest), 0o750); err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		rc, err := f.Open()
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		out, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600) // #nosec G304 -- dest validated under dataDir
		if err != nil {
			_ = rc.Close()
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		_, err = io.Copy(out, rc) // #nosec G110 -- admin-only restore; entries validated under dataDir
		_ = out.Close()
		_ = rc.Close()
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
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
