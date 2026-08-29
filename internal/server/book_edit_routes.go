package server

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"athenaeum/internal/library"
	"athenaeum/internal/models"
	"athenaeum/internal/storage"
)

const maxCoverUpload = 8 << 20

func (s *Server) registerBookEditRoutes(mux *http.ServeMux) {
	mux.HandleFunc("PUT /api/books/{id}", s.handlePutBook)
	mux.HandleFunc("DELETE /api/books/{id}", s.handleDeleteBook)
	mux.HandleFunc("PUT /api/books/{id}/cover", s.handlePutCover)
	mux.HandleFunc("DELETE /api/books/{id}/cover", s.handleDeleteCover)
	mux.HandleFunc("GET /api/metadata/providers", s.handleMetadataProviders)
	mux.HandleFunc("POST /api/books/{id}/metadata/search", s.handleMetadataSearch)
	mux.HandleFunc("POST /api/books/{id}/metadata/apply", s.handleMetadataApply)
	mux.HandleFunc("PUT /api/books/{id}/cover-from-url", s.handleCoverFromURL)
}

func (s *Server) handleMetadataProviders(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, library.MetadataProviders())
}

func (s *Server) handleMetadataSearch(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.requirePermission(w, r, models.PermEditMetadata); !ok {
		return
	}
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	book, err := s.store.GetBook(r.Context(), id)
	if errors.Is(err, storage.ErrNotFound) {
		writeError(w, http.StatusNotFound, errors.New("book not found"))
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if !s.requireBookAccess(w, r, book) {
		return
	}

	var q models.MetadataSearchQuery
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&q); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	matches := library.SearchMetadata(r.Context(), q)
	if matches == nil {
		matches = []models.MetadataMatch{}
	}
	writeJSON(w, http.StatusOK, models.MetadataSearchResult{Matches: matches})
}

func (s *Server) handleMetadataApply(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.requirePermission(w, r, models.PermEditMetadata); !ok {
		return
	}
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	book, err := s.store.GetBook(r.Context(), id)
	if errors.Is(err, storage.ErrNotFound) {
		writeError(w, http.StatusNotFound, errors.New("book not found"))
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if !s.requireBookAccess(w, r, book) {
		return
	}

	var req models.MetadataApplyRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<18)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	update := library.MatchToBookUpdate(req.Match)
	if update.Title == "" {
		writeError(w, http.StatusBadRequest, errors.New("match title is required"))
		return
	}

	book, err = s.store.UpdateBookMetadata(r.Context(), id, update)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			writeError(w, http.StatusNotFound, errors.New("book not found"))
		} else {
			writeError(w, http.StatusInternalServerError, err)
		}
		return
	}

	if req.ApplyCover && strings.TrimSpace(req.Match.CoverURL) != "" {
		if data, err := library.FetchCoverImage(r.Context(), req.Match.CoverURL); err == nil {
			if err := writeCoverFile(s.cfg.CoverDir(), id, data); err == nil {
				if err := s.store.SetBookCover(r.Context(), id, true); err == nil {
					book, _ = s.store.GetBook(r.Context(), id)
				}
			}
		}
	}

	writeJSON(w, http.StatusOK, book)
}

func (s *Server) handleCoverFromURL(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	if _, err := s.store.GetBook(r.Context(), id); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			writeError(w, http.StatusNotFound, errors.New("book not found"))
		} else {
			writeError(w, http.StatusInternalServerError, err)
		}
		return
	}

	var body struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	data, err := library.FetchCoverImage(r.Context(), body.URL)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := writeCoverFile(s.cfg.CoverDir(), id, data); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if err := s.store.SetBookCover(r.Context(), id, true); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	book, err := s.store.GetBook(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, book)
}

func (s *Server) handlePutBook(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.requirePermission(w, r, models.PermEditMetadata); !ok {
		return
	}
	book, err := s.bookByIDChecked(w, r)
	if err != nil {
		return
	}
	var u models.BookUpdate
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&u); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	u.Title = strings.TrimSpace(u.Title)
	if u.Title == "" {
		writeError(w, http.StatusBadRequest, errors.New("title is required"))
		return
	}
	u.Author = strings.TrimSpace(u.Author)
	u.Series = library.CleanSeriesName(u.Series)
	u.Language = strings.TrimSpace(u.Language)
	u.Description = strings.TrimSpace(u.Description)

	book, err = s.store.UpdateBookMetadata(r.Context(), book.ID, u)
	if errors.Is(err, storage.ErrNotFound) {
		writeError(w, http.StatusNotFound, errors.New("book not found"))
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, book)
}

func (s *Server) handlePutCover(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	if _, err := s.store.GetBook(r.Context(), id); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			writeError(w, http.StatusNotFound, errors.New("book not found"))
		} else {
			writeError(w, http.StatusInternalServerError, err)
		}
		return
	}

	data, err := readCoverUpload(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if len(data) == 0 {
		writeError(w, http.StatusBadRequest, errors.New("empty cover image"))
		return
	}
	if len(data) > maxCoverUpload {
		writeError(w, http.StatusBadRequest, errors.New("cover image too large"))
		return
	}

	if err := writeCoverFile(s.cfg.CoverDir(), id, data); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if err := s.store.SetBookCover(r.Context(), id, true); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	book, err := s.store.GetBook(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, book)
}

func (s *Server) handleDeleteCover(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	if _, err := s.store.GetBook(r.Context(), id); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			writeError(w, http.StatusNotFound, errors.New("book not found"))
		} else {
			writeError(w, http.StatusInternalServerError, err)
		}
		return
	}
	_ = removeCoverFile(s.cfg.CoverDir(), id)
	if err := s.store.SetBookCover(r.Context(), id, false); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	book, err := s.store.GetBook(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, book)
}

func (s *Server) handleDeleteBook(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.requirePermission(w, r, models.PermDeleteBooks); !ok {
		return
	}
	book, err := s.bookByIDChecked(w, r)
	if err != nil {
		return
	}

	_ = removeCoverFile(s.cfg.CoverDir(), book.ID)

	if book.RelPath != "" {
		if fs, err := s.openLibraryFS(r.Context(), book.LibraryID); err == nil {
			_ = fs.Remove(r.Context(), book.RelPath)
		}
	}

	if err := s.store.DeleteBook(r.Context(), book.ID); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			writeError(w, http.StatusNotFound, errors.New("book not found"))
		} else {
			writeError(w, http.StatusInternalServerError, err)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func readCoverUpload(r *http.Request) ([]byte, error) {
	ct := r.Header.Get("Content-Type")
	if strings.HasPrefix(ct, "multipart/form-data") {
		mr, err := r.MultipartReader()
		if err != nil {
			return nil, err
		}
		for {
			part, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, err
			}
			if part.FormName() == "cover" {
				return io.ReadAll(io.LimitReader(part, maxCoverUpload+1))
			}
		}
		return nil, errors.New("missing cover file")
	}
	return io.ReadAll(io.LimitReader(r.Body, maxCoverUpload+1))
}
