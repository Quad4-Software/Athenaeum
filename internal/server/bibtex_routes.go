package server

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"athenaeum/internal/library"
	"athenaeum/internal/models"
	"athenaeum/internal/storage"
)

const maxBibTeXBytes = 4 << 20
const maxBibTeXExportIDs = 500

func (s *Server) registerBibTeXRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/books/{id}/bibtex", s.handleBookBibTeX)
	mux.HandleFunc("GET /api/books/bibtex", s.handleBooksBibTeX)
	mux.HandleFunc("POST /api/library/bibtex/import", s.handleBibTeXImport)
}

func (s *Server) handleBookBibTeX(w http.ResponseWriter, r *http.Request) {
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
	body := library.FormatBibTeX(book)
	w.Header().Set("Content-Type", "application/x-bibtex; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="`+bibFilename(book)+`"`)
	_, _ = w.Write([]byte(body)) // #nosec G705 -- BibTeX attachment, not HTML
}

func (s *Server) handleBooksBibTeX(w http.ResponseWriter, r *http.Request) {
	idsParam := strings.TrimSpace(r.URL.Query().Get("ids"))
	format := strings.TrimSpace(r.URL.Query().Get("format"))
	var books []models.Book
	if idsParam != "" {
		n := 0
		for part := range strings.SplitSeq(idsParam, ",") {
			if n >= maxBibTeXExportIDs {
				break
			}
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			id, err := strconv.ParseInt(part, 10, 64)
			if err != nil || id <= 0 {
				continue
			}
			b, err := s.store.GetBook(r.Context(), id)
			if err != nil {
				continue
			}
			if !s.canAccessBook(r, b) {
				continue
			}
			books = append(books, b)
			n++
		}
	} else {
		q := models.BookQuery{
			Format: format,
			Limit:  500,
			Offset: 0,
		}
		if q.Format == "" {
			q.Format = models.FormatPapers
		}
		q, err := s.applyBookAccess(r.Context(), q)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		page, err := s.store.ListBooks(r.Context(), q)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		books = page.Items
	}
	if len(books) == 0 {
		writeError(w, http.StatusNotFound, errors.New("no books to export"))
		return
	}
	body := library.FormatBibTeXMulti(books)
	w.Header().Set("Content-Type", "application/x-bibtex; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="athenaeum-export.bib"`)
	_, _ = w.Write([]byte(body)) // #nosec G705 -- BibTeX attachment, not HTML
}

func (s *Server) handleBibTeXImport(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.requirePermission(w, r, models.PermEditMetadata); !ok {
		return
	}
	data, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxBibTeXBytes))
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	ct := r.Header.Get("Content-Type")
	text := string(data)
	if strings.Contains(ct, "application/json") {
		var req struct {
			BibTeX string `json:"bibtex"`
		}
		if err := json.Unmarshal(data, &req); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		text = req.BibTeX
	}
	if len(text) > maxBibTeXBytes {
		writeError(w, http.StatusBadRequest, errors.New("BibTeX too large"))
		return
	}
	if strings.TrimSpace(text) == "" {
		writeError(w, http.StatusBadRequest, errors.New("empty BibTeX"))
		return
	}
	opts := library.BibImportOptions{}
	if user, ok := UserFromContext(r.Context()); ok {
		acc, err := s.store.AccessibleLibraries(r.Context(), user)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		if acc.Restricted {
			if len(acc.LibraryIDs) == 0 {
				writeError(w, http.StatusForbidden, errors.New("library access denied"))
				return
			}
			opts.LibraryIDs = acc.LibraryIDs
		}
	}
	result, err := library.ImportBibTeX(r.Context(), s.store, text, opts)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func bibFilename(b models.Book) string {
	base := "citation"
	if b.DOI != "" {
		parts := strings.Split(b.DOI, "/")
		base = parts[len(parts)-1]
	} else if b.ArxivID != "" {
		base = "arxiv-" + strings.ReplaceAll(b.ArxivID, "/", "-")
	} else if b.PubmedID != "" {
		base = "pmid-" + b.PubmedID
	} else if b.ID > 0 {
		base = "book-" + strconv.FormatInt(b.ID, 10)
	}
	base = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			return r
		}
		return '-'
	}, base)
	return base + ".bib"
}

func (s *Server) canAccessBook(r *http.Request, book models.Book) bool {
	user, ok := UserFromContext(r.Context())
	if !ok {
		return true
	}
	allowed, err := s.store.UserCanAccessLibrary(r.Context(), user, book.LibraryID)
	if err != nil {
		return false
	}
	return allowed
}
