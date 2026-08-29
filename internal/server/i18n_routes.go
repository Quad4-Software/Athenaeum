package server

import (
	"errors"
	"io/fs"
	"net/http"
	"strings"

	"athenaeum/internal/i18n"
)

func (s *Server) registerI18nRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/i18n/locales", s.handleI18nLocales)
	mux.HandleFunc("GET /api/i18n/template", s.handleI18nTemplate)
	mux.HandleFunc("GET /api/i18n/{locale}", s.handleI18nLocale)
}

func (s *Server) i18nLoader() *i18n.Loader {
	return i18n.NewLoader(s.cfg.I18nDir())
}

func (s *Server) handleI18nLocales(w http.ResponseWriter, r *http.Request) {
	cat, err := s.i18nLoader().Catalog()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, cat)
}

func (s *Server) handleI18nTemplate(w http.ResponseWriter, r *http.Request) {
	tmpl, err := i18n.DefaultTemplate()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, tmpl)
}

func (s *Server) handleI18nLocale(w http.ResponseWriter, r *http.Request) {
	code := strings.TrimSpace(r.PathValue("locale"))
	if code == "" || code == "template" {
		writeError(w, http.StatusBadRequest, errors.New("invalid locale"))
		return
	}
	msgs, err := s.i18nLoader().Load(code)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			writeError(w, http.StatusNotFound, errors.New("locale not found"))
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, msgs)
}
