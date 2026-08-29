package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"athenaeum/internal/models"
	"athenaeum/internal/storage"
)

func (s *Server) registerCollectionRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/collections", s.handleListCollections)
	mux.HandleFunc("POST /api/collections", s.handleCreateCollection)
	mux.HandleFunc("GET /api/collections/{id}", s.handleGetCollection)
	mux.HandleFunc("PUT /api/collections/{id}", s.handleUpdateCollection)
	mux.HandleFunc("DELETE /api/collections/{id}", s.handleDeleteCollection)
	mux.HandleFunc("POST /api/collections/{id}/books/{bookId}", s.handleAddToCollection)
	mux.HandleFunc("DELETE /api/collections/{id}/books/{bookId}", s.handleRemoveFromCollection)
}

type collectionBody struct {
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Kind        string             `json:"kind"`
	Query       *models.SmartQuery `json:"query"`
}

func (s *Server) handleListCollections(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	items, err := s.store.ListCollections(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if items == nil {
		items = []models.Collection{}
	}
	writeJSON(w, http.StatusOK, items)
}

func (s *Server) handleCreateCollection(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	var body collectionBody
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if body.Name == "" {
		writeError(w, http.StatusBadRequest, errors.New("name is required"))
		return
	}
	var c models.Collection
	var err error
	switch body.Kind {
	case models.CollectionSmart:
		if body.Query == nil {
			writeError(w, http.StatusBadRequest, errors.New("query is required for smart collections"))
			return
		}
		c, err = s.store.CreateSmartCollection(r.Context(), userID, body.Name, body.Description, *body.Query)
	case models.CollectionReading:
		c, err = s.store.CreateReadingCollection(r.Context(), userID, body.Name, body.Description)
	default:
		c, err = s.store.CreateCollection(r.Context(), userID, body.Name, body.Description)
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, c)
}

func (s *Server) handleGetCollection(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	id, ok := collectionPathID(w, r)
	if !ok {
		return
	}
	c, err := s.store.GetCollection(r.Context(), userID, id)
	if errors.Is(err, storage.ErrNotFound) {
		writeError(w, http.StatusNotFound, err)
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, c)
}

func (s *Server) handleUpdateCollection(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	id, ok := collectionPathID(w, r)
	if !ok {
		return
	}
	var body collectionBody
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if body.Name == "" {
		writeError(w, http.StatusBadRequest, errors.New("name is required"))
		return
	}
	c, err := s.store.UpdateCollection(r.Context(), userID, id, body.Name, body.Description)
	if errors.Is(err, storage.ErrNotFound) {
		writeError(w, http.StatusNotFound, err)
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, c)
}

func (s *Server) handleDeleteCollection(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	id, ok := collectionPathID(w, r)
	if !ok {
		return
	}
	if err := s.store.DeleteCollection(r.Context(), userID, id); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			writeError(w, http.StatusNotFound, err)
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleAddToCollection(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	cid, ok := collectionPathID(w, r)
	if !ok {
		return
	}
	bid, ok := bookPathID(w, r)
	if !ok {
		return
	}
	if err := s.store.AddToCollection(r.Context(), userID, cid, bid); errors.Is(err, storage.ErrNotFound) {
		writeError(w, http.StatusNotFound, err)
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleRemoveFromCollection(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	cid, ok := collectionPathID(w, r)
	if !ok {
		return
	}
	bid, ok := bookPathID(w, r)
	if !ok {
		return
	}
	if err := s.store.RemoveFromCollection(r.Context(), userID, cid, bid); errors.Is(err, storage.ErrNotFound) {
		writeError(w, http.StatusNotFound, err)
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func collectionPathID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, errors.New("invalid collection id"))
		return 0, false
	}
	return id, true
}

func bookPathID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("bookId"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, errors.New("invalid book id"))
		return 0, false
	}
	return id, true
}
