package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"athenaeum/internal/brand"
	"athenaeum/internal/models"
	"athenaeum/internal/rss"
	"athenaeum/internal/storage"
)

// feedItemLimit caps how many enclosure items one feed document emits.
const feedItemLimit = 500

// feedPageSize is the largest page ListBooks accepts without clamping.
const feedPageSize = 200

const feedNameMaxLen = 200

func (s *Server) registerFeedRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/feeds", s.handleCreateFeed)
	mux.HandleFunc("GET /api/feeds", s.handleListFeeds)
	mux.HandleFunc("DELETE /api/feeds/{id}", s.handleDeleteFeed)
	mux.HandleFunc("GET /feed/{token}", s.handleFeedRSS)
	mux.HandleFunc("GET /feed/{token}/item/{bookId}/file", s.handleFeedItemFile)
	mux.HandleFunc("GET /feed/{token}/item/{bookId}/cover", s.handleFeedItemCover)
}

func (s *Server) handleCreateFeed(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name         string `json:"name"`
		LibraryID    int64  `json:"libraryId"`
		CollectionID int64  `json:"collectionId"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<12)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" || len(name) > feedNameMaxLen {
		writeError(w, http.StatusBadRequest, errors.New("invalid feed name"))
		return
	}
	ctx := r.Context()
	user, _ := UserFromContext(ctx)
	userID := UserIDFromContext(ctx)
	if req.LibraryID > 0 {
		if _, err := s.store.GetLibrary(ctx, req.LibraryID); errors.Is(err, storage.ErrNotFound) {
			writeError(w, http.StatusNotFound, err)
			return
		} else if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		allowed, err := s.store.UserCanAccessLibrary(ctx, user, req.LibraryID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		if !allowed {
			writeError(w, http.StatusForbidden, errors.New("library access denied"))
			return
		}
	}
	if req.CollectionID > 0 {
		if _, err := s.store.GetCollection(ctx, userID, req.CollectionID); errors.Is(err, storage.ErrNotFound) {
			writeError(w, http.StatusNotFound, errors.New("collection not found"))
			return
		} else if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
	}
	ft, err := s.store.CreateFeedToken(ctx, userID, name, req.LibraryID, req.CollectionID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"id":        ft.ID,
		"name":      ft.Name,
		"url":       s.requestBaseURL(r) + "/feed/" + ft.Token,
		"token":     ft.Token,
		"createdAt": ft.CreatedAt,
	})
}

func (s *Server) handleListFeeds(w http.ResponseWriter, r *http.Request) {
	feeds, err := s.store.ListFeedTokens(r.Context(), UserIDFromContext(r.Context()))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if feeds == nil {
		feeds = []models.FeedToken{}
	}
	base := s.requestBaseURL(r)
	for i := range feeds {
		feeds[i].URL = base + "/feed/" + feeds[i].Token
	}
	writeJSON(w, http.StatusOK, feeds)
}

func (s *Server) handleDeleteFeed(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	ownerID := UserIDFromContext(r.Context())
	if raw := r.URL.Query().Get("userId"); raw != "" {
		if _, ok := requireAdmin(w, r); !ok {
			return
		}
		uid, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || uid < 0 {
			writeError(w, http.StatusBadRequest, errInvalidID)
			return
		}
		ownerID = uid
	}
	if err := s.store.DeleteFeedToken(r.Context(), ownerID, id); errors.Is(err, storage.ErrNotFound) {
		writeError(w, http.StatusNotFound, err)
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// feedByToken resolves a public feed token to the feed row and its owner.
// Feeds created while auth is disabled have no owner row; they resolve to
// the anonymous user, which has unrestricted library access.
func (s *Server) feedByToken(ctx context.Context, token string) (models.FeedToken, models.User, error) {
	feed, err := s.store.GetFeedTokenByToken(ctx, token)
	if err != nil {
		return models.FeedToken{}, models.User{}, err
	}
	if feed.UserID == 0 {
		return feed, models.User{}, nil
	}
	owner, err := s.store.GetUser(ctx, feed.UserID)
	if err != nil {
		return models.FeedToken{}, models.User{}, err
	}
	return feed, owner, nil
}

// feedBooks lists the audio books inside the feed scope under the owner's
// library access, resolved at request time so revoked access still applies.
func (s *Server) feedBooks(ctx context.Context, feed models.FeedToken, owner models.User) ([]models.Book, error) {
	q := models.BookQuery{
		Format:       models.FormatAudio,
		Sort:         "recent",
		LibraryID:    feed.LibraryID,
		CollectionID: feed.CollectionID,
		UserID:       feed.UserID,
	}
	q, err := s.applyBookAccess(WithUser(ctx, owner), q)
	if err != nil {
		return nil, err
	}
	var books []models.Book
	for len(books) < feedItemLimit {
		q.Limit = feedPageSize
		q.Offset = len(books)
		page, err := s.store.ListBooks(ctx, q)
		if errors.Is(err, storage.ErrNotFound) {
			break
		}
		if err != nil {
			return nil, err
		}
		books = append(books, page.Items...)
		if len(page.Items) < feedPageSize {
			break
		}
	}
	if len(books) > feedItemLimit {
		books = books[:feedItemLimit]
	}
	return books, nil
}

// feedScopedBook returns the book only when it is a playable feed item:
// audio, visible, inside the feed scope, and inside a library the feed
// owner can still access.
func (s *Server) feedScopedBook(ctx context.Context, feed models.FeedToken, owner models.User, bookID int64) (models.Book, error) {
	book, err := s.store.GetBook(ctx, bookID)
	if err != nil {
		return models.Book{}, err
	}
	if book.Hidden || !models.IsAudio(book.Format) {
		return models.Book{}, storage.ErrNotFound
	}
	if feed.LibraryID > 0 && book.LibraryID != feed.LibraryID {
		return models.Book{}, storage.ErrNotFound
	}
	allowed, err := s.store.UserCanAccessLibrary(ctx, owner, book.LibraryID)
	if err != nil {
		return models.Book{}, err
	}
	if !allowed {
		return models.Book{}, storage.ErrNotFound
	}
	if feed.CollectionID > 0 {
		in, err := s.store.CollectionContainsBook(ctx, feed.UserID, feed.CollectionID, bookID)
		if err != nil {
			return models.Book{}, err
		}
		if !in {
			return models.Book{}, storage.ErrNotFound
		}
	}
	return book, nil
}

func (s *Server) handleFeedRSS(w http.ResponseWriter, r *http.Request) {
	feed, owner, err := s.feedByToken(r.Context(), r.PathValue("token"))
	if errors.Is(err, storage.ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	books, err := s.feedBooks(r.Context(), feed, owner)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	base := s.requestBaseURL(r)
	title := feed.Name
	if title == "" {
		title = brand.Name
	}
	w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
	items, err := s.feedRSSItems(r.Context(), feed, base, books)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	wr := rss.Writer{Title: title, Link: base, Description: brand.Name + " audio feed"}
	if err := wr.Write(w, items); err != nil {
		s.log.Warn("feed encode failed", "err", err)
	}
}

// feedRSSItems expands books into one enclosure item per playable file:
// a multi-file audiobook set produces one item per track.
func (s *Server) feedRSSItems(ctx context.Context, feed models.FeedToken, base string, books []models.Book) ([]rss.Item, error) {
	itemBase := base + "/feed/" + feed.Token
	var items []rss.Item
	for _, b := range books {
		imageURL := ""
		if b.HasCover {
			imageURL = fmt.Sprintf("%s/item/%d/cover", itemBase, b.ID)
		}
		if b.Format == models.FormatAudiobook {
			tracks, err := s.store.ListAudiobookTracks(ctx, b.ID)
			if err != nil {
				return nil, err
			}
			for i, tr := range tracks {
				items = append(items, rss.Item{
					Title:         feedTrackTitle(b, tr, i),
					GUID:          fmt.Sprintf("book-%d-track-%d", b.ID, i),
					PubDate:       b.AddedAt,
					EnclosureURL:  fmt.Sprintf("%s/item/%d/file?track=%d", itemBase, b.ID, i),
					EnclosureSize: tr.FileSize,
					EnclosureMIME: contentType(tr.Format),
					Author:        b.Author,
					ImageURL:      imageURL,
				})
			}
			continue
		}
		items = append(items, rss.Item{
			Title:         feedItemTitle(b),
			GUID:          fmt.Sprintf("book-%d", b.ID),
			PubDate:       b.AddedAt,
			EnclosureURL:  fmt.Sprintf("%s/item/%d/file", itemBase, b.ID),
			EnclosureSize: b.FileSize,
			EnclosureMIME: contentType(b.Format),
			Author:        b.Author,
			ImageURL:      imageURL,
		})
	}
	if len(items) > feedItemLimit {
		items = items[:feedItemLimit]
	}
	return items, nil
}

func feedItemTitle(b models.Book) string {
	if b.Series == "" {
		return b.Title
	}
	if b.SeriesIndex > 0 {
		return fmt.Sprintf("%s #%g - %s", b.Series, b.SeriesIndex, b.Title)
	}
	return b.Series + " - " + b.Title
}

func feedTrackTitle(b models.Book, tr models.AudiobookTrack, i int) string {
	part := tr.Title
	if part == "" {
		part = fmt.Sprintf("Part %d", i+1)
	}
	return feedItemTitle(b) + " - " + part
}

func (s *Server) handleFeedItemFile(w http.ResponseWriter, r *http.Request) {
	feed, owner, err := s.feedByToken(r.Context(), r.PathValue("token"))
	if errors.Is(err, storage.ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	bookID, err := strconv.ParseInt(r.PathValue("bookId"), 10, 64)
	if err != nil || bookID <= 0 {
		writeError(w, http.StatusBadRequest, errInvalidID)
		return
	}
	book, err := s.feedScopedBook(r.Context(), feed, owner, bookID)
	if errors.Is(err, storage.ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	relPath, format := book.RelPath, book.Format
	if book.Format == models.FormatAudiobook {
		tracks, err := s.store.ListAudiobookTracks(r.Context(), book.ID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		idx := 0
		if raw := r.URL.Query().Get("track"); raw != "" {
			v, convErr := strconv.Atoi(raw)
			if convErr != nil || v < 0 || v >= len(tracks) {
				http.NotFound(w, r)
				return
			}
			idx = v
		}
		if idx >= len(tracks) {
			http.NotFound(w, r)
			return
		}
		relPath, format = tracks[idx].RelPath, tracks[idx].Format
	}
	s.serveLibraryFile(w, r, book.LibraryID, relPath, safeFilename(book), contentType(format), false)
}

func (s *Server) handleFeedItemCover(w http.ResponseWriter, r *http.Request) {
	feed, owner, err := s.feedByToken(r.Context(), r.PathValue("token"))
	if errors.Is(err, storage.ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	bookID, err := strconv.ParseInt(r.PathValue("bookId"), 10, 64)
	if err != nil || bookID <= 0 {
		writeError(w, http.StatusBadRequest, errInvalidID)
		return
	}
	book, err := s.feedScopedBook(r.Context(), feed, owner, bookID)
	if errors.Is(err, storage.ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	f, info, err := openCoverFile(s.cfg.CoverDir(), book.ID)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer f.Close()
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	http.ServeContent(w, r, "cover", info.ModTime(), f)
}
