package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"athenaeum/internal/brand"
	"athenaeum/internal/models"
	"athenaeum/internal/opds"
	"athenaeum/internal/storage"
)

func (s *Server) registerOPDSRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /opds/", s.handleOPDSRoot)
	mux.HandleFunc("GET /opds/recent", s.handleOPDSRecent)
	mux.HandleFunc("GET /opds/search", s.handleOPDSSearch)
	mux.HandleFunc("GET /opds/series", s.handleOPDSSeriesNav)
	mux.HandleFunc("GET /opds/series/{name}", s.handleOPDSSeriesBooks)
	mux.HandleFunc("GET /opds/comics", s.handleOPDSComics)
	mux.HandleFunc("GET /opds/kindle", s.handleOPDSKindle)
	mux.HandleFunc("GET /opds/papers", s.handleOPDSPapers)
	mux.HandleFunc("GET /opds/v2/", s.handleOPDS2Root)
	mux.HandleFunc("GET /opds/v2/recent", s.handleOPDS2Recent)
}

func (s *Server) opdsWriter(r *http.Request) opds.FeedWriter {
	base := s.requestBaseURL(r)
	return opds.FeedWriter{BaseURL: base, Title: brand.Name}
}

func (s *Server) opdsBookQuery(r *http.Request, q models.BookQuery) (models.BookQuery, error) {
	return s.applyBookAccess(r.Context(), q)
}

func (s *Server) handleOPDSRoot(w http.ResponseWriter, r *http.Request) {
	opds.SetXMLHeaders(w)
	if err := s.opdsWriter(r).WriteRootCatalog(w); err != nil {
		s.log.Error("opds root", "err", err)
	}
}

func (s *Server) handleOPDSRecent(w http.ResponseWriter, r *http.Request) {
	query := models.BookQuery{Sort: "recent", Limit: 50}
	var err error
	query, err = s.opdsBookQuery(r, query)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	page, err := s.store.ListBooks(r.Context(), query)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	opds.SetXMLHeaders(w)
	progress := s.opdsProgressMap(r, page.Items)
	if err := s.opdsWriter(r).WriteAcquisitionFeed(w, page.Items, "/opds/recent", progress); err != nil {
		s.log.Error("opds recent", "err", err)
	}
}

func (s *Server) handleOPDSSearch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	query := models.BookQuery{Search: q, Limit: 50}
	var err error
	query, err = s.opdsBookQuery(r, query)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	page, err := s.store.ListBooks(r.Context(), query)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	opds.SetXMLHeaders(w)
	path := "/opds/search?q=" + url.QueryEscape(q)
	progress := s.opdsProgressMap(r, page.Items)
	if err := s.opdsWriter(r).WriteAcquisitionFeed(w, page.Items, path, progress); err != nil {
		s.log.Error("opds search", "err", err)
	}
}

func (s *Server) handleOPDSSeriesNav(w http.ResponseWriter, r *http.Request) {
	libID, libIDs, err := s.libraryFilterIDs(r.Context(), 0)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	series, err := s.store.ListSeries(r.Context(), libID, libIDs)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	opds.SetXMLHeaders(w)
	if err := s.opdsWriter(r).WriteSeriesNavigation(w, series); err != nil {
		s.log.Error("opds series nav", "err", err)
	}
}

func (s *Server) handleOPDSSeriesBooks(w http.ResponseWriter, r *http.Request) {
	name, err := url.PathUnescape(r.PathValue("name"))
	if err != nil || name == "" {
		writeError(w, http.StatusBadRequest, errors.New("invalid series name"))
		return
	}
	query := models.BookQuery{Series: name, Sort: "author", Limit: 200}
	query, err = s.opdsBookQuery(r, query)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	page, err := s.store.ListBooks(r.Context(), query)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	opds.SetXMLHeaders(w)
	path := "/opds/series/" + url.PathEscape(name)
	progress := s.opdsProgressMap(r, page.Items)
	if err := s.opdsWriter(r).WriteAcquisitionFeed(w, page.Items, path, progress); err != nil {
		s.log.Error("opds series books", "err", err)
	}
}

type userLibrariesBody struct {
	LibraryIDs []int64 `json:"libraryIds"`
}

func (s *Server) registerUserLibraryRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/auth/users/{id}/libraries", s.handleGetUserLibraries)
	mux.HandleFunc("PUT /api/auth/users/{id}/libraries", s.handleSetUserLibraries)
}

func (s *Server) handleGetUserLibraries(w http.ResponseWriter, r *http.Request) {
	actor, ok := UserFromContext(r.Context())
	if !ok || !actor.IsAdmin {
		writeError(w, http.StatusForbidden, errForbidden)
		return
	}
	userID, ok := userPathID(w, r)
	if !ok {
		return
	}
	ids, err := s.store.ListUserLibraryIDs(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if ids == nil {
		ids = []int64{}
	}
	writeJSON(w, http.StatusOK, userLibrariesBody{LibraryIDs: ids})
}

func (s *Server) handleSetUserLibraries(w http.ResponseWriter, r *http.Request) {
	actor, ok := UserFromContext(r.Context())
	if !ok || !actor.IsAdmin {
		writeError(w, http.StatusForbidden, errForbidden)
		return
	}
	userID, ok := userPathID(w, r)
	if !ok {
		return
	}
	if _, err := s.store.GetUser(r.Context(), userID); errors.Is(err, storage.ErrNotFound) {
		writeError(w, http.StatusNotFound, err)
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	var body userLibrariesBody
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := s.store.SetUserLibraries(r.Context(), userID, body.LibraryIDs); errors.Is(err, storage.ErrNotFound) {
		writeError(w, http.StatusBadRequest, err)
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	target, _ := s.store.GetUser(r.Context(), userID)
	s.logAudit(r, actor.ID, actor.Username, userID, target.Username, "user.libraries",
		"libraries="+formatInt64List(body.LibraryIDs))
	writeJSON(w, http.StatusOK, userLibrariesBody{LibraryIDs: body.LibraryIDs})
}

func userPathID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, errors.New("invalid user id"))
		return 0, false
	}
	return id, true
}

func formatInt64List(ids []int64) string {
	if len(ids) == 0 {
		return ""
	}
	var out strings.Builder
	out.WriteString(strconv.FormatInt(ids[0], 10))
	for _, id := range ids[1:] {
		out.WriteString("," + strconv.FormatInt(id, 10))
	}
	return out.String()
}

func (s *Server) handleOPDSComics(w http.ResponseWriter, r *http.Request) {
	query := models.BookQuery{Format: models.FormatComic, Sort: "recent", Limit: 50}
	var err error
	query, err = s.opdsBookQuery(r, query)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	page, err := s.store.ListBooks(r.Context(), query)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	opds.SetXMLHeaders(w)
	progress := s.opdsProgressMap(r, page.Items)
	if err := s.opdsWriter(r).WriteAcquisitionFeed(w, page.Items, "/opds/comics", progress); err != nil {
		s.log.Error("opds comics", "err", err)
	}
}

func (s *Server) handleOPDSKindle(w http.ResponseWriter, r *http.Request) {
	query := models.BookQuery{Format: models.FormatKindle, Sort: "recent", Limit: 50}
	var err error
	query, err = s.opdsBookQuery(r, query)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	page, err := s.store.ListBooks(r.Context(), query)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	opds.SetXMLHeaders(w)
	progress := s.opdsProgressMap(r, page.Items)
	if err := s.opdsWriter(r).WriteAcquisitionFeed(w, page.Items, "/opds/kindle", progress); err != nil {
		s.log.Error("opds kindle", "err", err)
	}
}

func (s *Server) handleOPDSPapers(w http.ResponseWriter, r *http.Request) {
	query := models.BookQuery{Format: models.FormatPapers, Sort: "recent", Limit: 50}
	var err error
	query, err = s.opdsBookQuery(r, query)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	page, err := s.store.ListBooks(r.Context(), query)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	opds.SetXMLHeaders(w)
	progress := s.opdsProgressMap(r, page.Items)
	if err := s.opdsWriter(r).WriteAcquisitionFeed(w, page.Items, "/opds/papers", progress); err != nil {
		s.log.Error("opds papers", "err", err)
	}
}

func (s *Server) opdsProgressMap(r *http.Request, books []models.Book) map[int64]models.Progress {
	userID := UserIDFromContext(r.Context())
	ids := make([]int64, len(books))
	for i, b := range books {
		ids[i] = b.ID
	}
	m, err := s.store.ProgressMap(r.Context(), userID, ids)
	if err != nil {
		return map[int64]models.Progress{}
	}
	return m
}

func (s *Server) handleOPDS2Root(w http.ResponseWriter, r *http.Request) {
	base := s.requestBaseURL(r)
	feed := map[string]any{
		"metadata": map[string]any{
			"title": brand.Name,
		},
		"links": []map[string]string{
			{"rel": "self", "href": base + "/opds/v2/", "type": "application/opds+json"},
			{"rel": "http://opds-spec.org/sort/new", "href": base + "/opds/v2/recent", "type": "application/opds+json"},
			{"rel": "search", "href": base + "/opds/search?q={searchTerms}", "type": "application/atom+xml"},
		},
		"navigation": []map[string]any{
			{"title": "Recently Added", "href": base + "/opds/v2/recent", "type": "application/opds+json"},
			{"title": "All (OPDS 1.2)", "href": base + "/opds/", "type": "application/atom+xml;profile=opds-catalog"},
		},
	}
	w.Header().Set("Content-Type", "application/opds+json")
	_ = json.NewEncoder(w).Encode(feed)
}

func (s *Server) handleOPDS2Recent(w http.ResponseWriter, r *http.Request) {
	query := models.BookQuery{Sort: "recent", Limit: 50}
	var err error
	query, err = s.opdsBookQuery(r, query)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	page, err := s.store.ListBooks(r.Context(), query)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	base := s.requestBaseURL(r)
	pubs := make([]map[string]any, 0, len(page.Items))
	for _, b := range page.Items {
		pubs = append(pubs, map[string]any{
			"metadata": map[string]any{
				"title":      b.Title,
				"author":     b.Author,
				"identifier": strconv.FormatInt(b.ID, 10),
			},
			"links": []map[string]string{
				{"rel": "http://opds-spec.org/acquisition", "href": base + "/api/books/" + strconv.FormatInt(b.ID, 10) + "/download", "type": contentType(b.Format)},
				{"rel": "http://opds-spec.org/image", "href": base + "/api/books/" + strconv.FormatInt(b.ID, 10) + "/cover", "type": "image/jpeg"},
			},
		})
	}
	feed := map[string]any{
		"metadata":     map[string]any{"title": "Recently Added"},
		"links":        []map[string]string{{"rel": "self", "href": base + "/opds/v2/recent", "type": "application/opds+json"}},
		"publications": pubs,
	}
	w.Header().Set("Content-Type", "application/opds+json")
	_ = json.NewEncoder(w).Encode(feed)
}
