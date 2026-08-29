package server

import (
	"context"
	"errors"
	"net/http"
	"slices"

	"athenaeum/internal/models"
	"athenaeum/internal/storage"
)

func (s *Server) applyBookAccess(ctx context.Context, q models.BookQuery) (models.BookQuery, error) {
	user, ok := UserFromContext(ctx)
	if !ok {
		return q, nil
	}
	acc, err := s.store.AccessibleLibraries(ctx, user)
	if err != nil {
		return q, err
	}
	if !acc.Restricted {
		return q, nil
	}
	if q.LibraryID > 0 {
		allowed := slices.Contains(acc.LibraryIDs, q.LibraryID)
		if !allowed {
			q.LibraryIDs = []int64{-1}
			return q, nil
		}
		return q, nil
	}
	q.LibraryIDs = acc.LibraryIDs
	return q, nil
}

func (s *Server) libraryFilterIDs(ctx context.Context, libraryID int64) (int64, []int64, error) {
	user, ok := UserFromContext(ctx)
	if !ok {
		return libraryID, nil, nil
	}
	acc, err := s.store.AccessibleLibraries(ctx, user)
	if err != nil {
		return 0, nil, err
	}
	if !acc.Restricted {
		return libraryID, nil, nil
	}
	if libraryID > 0 {
		if slices.Contains(acc.LibraryIDs, libraryID) {
			return libraryID, nil, nil
		}
		return 0, []int64{-1}, nil
	}
	return 0, acc.LibraryIDs, nil
}

func (s *Server) requireLibraryAccess(w http.ResponseWriter, r *http.Request, libraryID int64) bool {
	user, ok := UserFromContext(r.Context())
	if !ok {
		return true
	}
	allowed, err := s.store.UserCanAccessLibrary(r.Context(), user, libraryID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return false
	}
	if !allowed {
		writeError(w, http.StatusForbidden, errors.New("library access denied"))
		return false
	}
	return true
}

func (s *Server) requireBookAccess(w http.ResponseWriter, r *http.Request, book models.Book) bool {
	return s.requireLibraryAccess(w, r, book.LibraryID)
}

func (s *Server) bookByIDChecked(w http.ResponseWriter, r *http.Request) (models.Book, error) {
	book, err := s.bookByID(w, r)
	if err != nil {
		return book, err
	}
	if !s.requireBookAccess(w, r, book) {
		return models.Book{}, storage.ErrNotFound
	}
	return book, nil
}
