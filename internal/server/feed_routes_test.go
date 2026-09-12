package server

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"athenaeum/internal/auth"
	"athenaeum/internal/models"
	"athenaeum/internal/storage"
)

// feedRSSDoc mirrors the RSS shape the feed handler emits, for assertions.
type feedRSSDoc struct {
	XMLName xml.Name `xml:"rss"`
	Version string   `xml:"version,attr"`
	Channel struct {
		Title string `xml:"title"`
		Link  string `xml:"link"`
		Items []struct {
			Title     string `xml:"title"`
			GUID      string `xml:"guid"`
			PubDate   string `xml:"pubDate"`
			Enclosure struct {
				URL    string `xml:"url,attr"`
				Length int64  `xml:"length,attr"`
				Type   string `xml:"type,attr"`
			} `xml:"enclosure"`
			Author string `xml:"author"`
			Image  *struct {
				Href string `xml:"href,attr"`
			} `xml:"image"`
		} `xml:"item"`
	} `xml:"channel"`
}

func seedAudioLibrary(t *testing.T, srv *Server, store *storage.Store, name string) (int64, string) {
	t.Helper()
	ctx := context.Background()
	libDir := filepath.Join(srv.cfg.DataDir, "lib-"+name)
	if err := os.MkdirAll(libDir, 0o750); err != nil {
		t.Fatal(err)
	}
	lib, err := store.CreateLibrary(ctx, name, libDir)
	if err != nil {
		t.Fatal(err)
	}
	return lib.ID, libDir
}

func seedAudioBook(t *testing.T, store *storage.Store, libID int64, libDir, rel, title, format string, payload []byte) int64 {
	t.Helper()
	ctx := context.Background()
	full := filepath.Join(libDir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(full), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, payload, 0o640); err != nil {
		t.Fatal(err)
	}
	id, err := store.UpsertBook(ctx, &models.Book{
		LibraryID: libID,
		Title:     title,
		Author:    "Feed Author",
		Format:    format,
		RelPath:   rel,
		FileSize:  int64(len(payload)),
	}, 1)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func createFeed(t *testing.T, handler http.Handler, session, csrf *http.Cookie, body map[string]any) models.FeedToken {
	t.Helper()
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/feeds", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	if session != nil {
		req.AddCookie(session)
	}
	if csrf != nil {
		withCSRF(req, csrf)
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create feed status=%d body=%s", rec.Code, rec.Body.String())
	}
	var ft models.FeedToken
	if err := json.NewDecoder(rec.Body).Decode(&ft); err != nil {
		t.Fatal(err)
	}
	return ft
}

func getFeedXML(t *testing.T, handler http.Handler, token string) (*httptest.ResponseRecorder, feedRSSDoc) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/feed/"+token, nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	var doc feedRSSDoc
	if rec.Code == http.StatusOK {
		if err := xml.Unmarshal(rec.Body.Bytes(), &doc); err != nil {
			t.Fatalf("feed XML parse: %v\n%s", err, rec.Body.String())
		}
	}
	return rec, doc
}

func TestFeedCreateListAndServeAudioOnly(t *testing.T) {
	srv, store := testServer(t)
	ctx := context.Background()
	hash, _ := auth.HashPassword("longpassword")
	if _, err := store.CreateUser(ctx, "admin", hash, true); err != nil {
		t.Fatal(err)
	}
	libID, libDir := seedAudioLibrary(t, srv, store, "main")
	mp3Payload := []byte("ID3-fake-mp3-bytes")
	m4bPayload := []byte("m4b-audio-payload")
	mp3ID := seedAudioBook(t, store, libID, libDir, "one.mp3", "Audio One", models.FormatMP3, mp3Payload)
	m4bID := seedAudioBook(t, store, libID, libDir, "two.m4b", "Audio Two", models.FormatM4B, m4bPayload)
	epubID := seedAudioBook(t, store, libID, libDir, "three.epub", "Epub Three", models.FormatEPUB, []byte("epub"))

	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}
	session, _, csrf := loginUser(t, handler, "admin", "longpassword")

	ft := createFeed(t, handler, session, csrf, map[string]any{"name": "My Podcasts"})
	if ft.Token == "" || ft.ID <= 0 {
		t.Fatalf("feed=%+v", ft)
	}
	wantURLSuffix := "/feed/" + ft.Token
	if ft.URL == "" || len(ft.URL) <= len(wantURLSuffix) || ft.URL[len(ft.URL)-len(wantURLSuffix):] != wantURLSuffix {
		t.Fatalf("feed url=%q want suffix %q", ft.URL, wantURLSuffix)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/feeds", nil)
	req.AddCookie(session)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list feeds status=%d body=%s", rec.Code, rec.Body.String())
	}
	var feeds []models.FeedToken
	if err := json.NewDecoder(rec.Body).Decode(&feeds); err != nil {
		t.Fatal(err)
	}
	if len(feeds) != 1 || feeds[0].ID != ft.ID || feeds[0].Name != "My Podcasts" || feeds[0].Token != ft.Token {
		t.Fatalf("feeds=%+v", feeds)
	}
	if feeds[0].LibraryID != 0 || feeds[0].CollectionID != 0 {
		t.Fatalf("expected unscoped feed, got %+v", feeds[0])
	}

	// The feed is public: no session cookie on this request.
	feedRec, doc := getFeedXML(t, handler, ft.Token)
	if feedRec.Code != http.StatusOK {
		t.Fatalf("feed status=%d body=%s", feedRec.Code, feedRec.Body.String())
	}
	if ct := feedRec.Header().Get("Content-Type"); ct != "application/rss+xml; charset=utf-8" {
		t.Fatalf("content-type=%q", ct)
	}
	if doc.Version != "2.0" {
		t.Fatalf("rss version=%q", doc.Version)
	}
	if doc.Channel.Title != "My Podcasts" {
		t.Fatalf("channel title=%q", doc.Channel.Title)
	}
	if len(doc.Channel.Items) != 2 {
		t.Fatalf("items=%d want 2 (audio only): %+v", len(doc.Channel.Items), doc.Channel.Items)
	}
	byTitle := map[string]int{}
	for _, it := range doc.Channel.Items {
		byTitle[it.Title]++
		if it.Author != "Feed Author" {
			t.Errorf("itunes author=%q", it.Author)
		}
		if _, err := time.Parse(time.RFC1123Z, it.PubDate); err != nil {
			t.Errorf("pubDate %q not RFC1123Z: %v", it.PubDate, err)
		}
	}
	if byTitle["Audio One"] != 1 || byTitle["Audio Two"] != 1 {
		t.Fatalf("item titles=%v", byTitle)
	}
	if _, ok := byTitle["Epub Three"]; ok {
		t.Fatal("epub leaked into audio feed")
	}
	for _, it := range doc.Channel.Items {
		var wantID int64
		var wantLen int64
		var wantType string
		switch it.Title {
		case "Audio One":
			wantID, wantLen, wantType = mp3ID, int64(len(mp3Payload)), "audio/mpeg"
		case "Audio Two":
			wantID, wantLen, wantType = m4bID, int64(len(m4bPayload)), "audio/mp4"
		}
		wantEnc := fmt.Sprintf("/feed/%s/item/%d/file", ft.Token, wantID)
		if !bytes.HasSuffix([]byte(it.Enclosure.URL), []byte(wantEnc)) {
			t.Errorf("enclosure url=%q want suffix %q", it.Enclosure.URL, wantEnc)
		}
		if it.Enclosure.Length != wantLen {
			t.Errorf("enclosure length=%d want %d", it.Enclosure.Length, wantLen)
		}
		if it.Enclosure.Type != wantType {
			t.Errorf("enclosure type=%q want %q", it.Enclosure.Type, wantType)
		}
		if it.GUID != fmt.Sprintf("book-%d", wantID) {
			t.Errorf("guid=%q", it.GUID)
		}
	}

	// A non-audio book is not a feed item even with a valid token.
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/feed/%s/item/%d/file", ft.Token, epubID), nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("epub item status=%d want 404", rec.Code)
	}
}

func TestFeedTokenUnknownAndPublicAccess(t *testing.T) {
	srv, store := testServer(t)
	ctx := context.Background()
	hash, _ := auth.HashPassword("longpassword")
	if _, err := store.CreateUser(ctx, "admin", hash, true); err != nil {
		t.Fatal(err)
	}
	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}
	// No cookies at all: the route must still be reachable.
	req := httptest.NewRequest(http.MethodGet, "/feed/not-a-real-token", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("bad token status=%d want 404", rec.Code)
	}
}

func TestFeedLibraryRestriction(t *testing.T) {
	srv, store := testServer(t)
	ctx := context.Background()
	hash, _ := auth.HashPassword("longpassword")
	if _, err := store.CreateUser(ctx, "admin", hash, true); err != nil {
		t.Fatal(err)
	}
	readerID, err := store.CreateUser(ctx, "reader", hash, false)
	if err != nil {
		t.Fatal(err)
	}
	libA, dirA := seedAudioLibrary(t, srv, store, "liba")
	libB, dirB := seedAudioLibrary(t, srv, store, "libb")
	seedAudioBook(t, store, libA, dirA, "a.mp3", "Allowed Audio", models.FormatMP3, []byte("aaa"))
	seedAudioBook(t, store, libB, dirB, "b.mp3", "Secret Audio", models.FormatMP3, []byte("bbb"))
	if err := store.SetUserLibraries(ctx, readerID, []int64{libA}); err != nil {
		t.Fatal(err)
	}

	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}
	session, _, csrf := loginUser(t, handler, "reader", "longpassword")

	// Scoping a feed to a library the user cannot reach is refused at creation.
	req := httptest.NewRequest(http.MethodPost, "/api/feeds", bytes.NewBufferString(`{"name":"x","libraryId":`+strconv.FormatInt(libB, 10)+`}`))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(session)
	withCSRF(req, csrf)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden && rec.Code != http.StatusNotFound {
		t.Fatalf("restricted create status=%d body=%s", rec.Code, rec.Body.String())
	}

	// An unscoped feed only ever emits books inside the owner's access.
	ft := createFeed(t, handler, session, csrf, map[string]any{"name": "mine"})
	_, doc := getFeedXML(t, handler, ft.Token)
	if len(doc.Channel.Items) != 1 || doc.Channel.Items[0].Title != "Allowed Audio" {
		t.Fatalf("restricted feed items=%+v", doc.Channel.Items)
	}

	// A token scoped to a forbidden library (created out of band) stays empty.
	sneaky, err := store.CreateFeedToken(ctx, readerID, "sneaky", libB, 0)
	if err != nil {
		t.Fatal(err)
	}
	sneakyRec, sneakyDoc := getFeedXML(t, handler, sneaky.Token)
	if sneakyRec.Code != http.StatusOK {
		t.Fatalf("sneaky feed status=%d", sneakyRec.Code)
	}
	if len(sneakyDoc.Channel.Items) != 0 {
		t.Fatalf("forbidden library leaked items=%+v", sneakyDoc.Channel.Items)
	}
}

func TestFeedEnclosureAndCoverRoutes(t *testing.T) {
	srv, store := testServer(t)
	ctx := context.Background()
	hash, _ := auth.HashPassword("longpassword")
	if _, err := store.CreateUser(ctx, "admin", hash, true); err != nil {
		t.Fatal(err)
	}
	libA, dirA := seedAudioLibrary(t, srv, store, "liba")
	libB, dirB := seedAudioLibrary(t, srv, store, "libb")
	payload := []byte("feed-mp3-payload")
	inID := seedAudioBook(t, store, libA, dirA, "in.mp3", "In Scope", models.FormatMP3, payload)
	outID := seedAudioBook(t, store, libB, dirB, "out.mp3", "Out Scope", models.FormatMP3, []byte("other"))
	epubID := seedAudioBook(t, store, libA, dirA, "doc.epub", "Doc", models.FormatEPUB, []byte("epub"))

	// Multi-file set in scope: two track files merged into one set book.
	setTracks := []models.AudiobookTrack{
		{Index: 0, Title: "Chapter 1", RelPath: "set/p1.mp3", Format: models.FormatMP3, FileSize: 4},
		{Index: 1, Title: "Chapter 2", RelPath: "set/p2.mp3", Format: models.FormatMP3, FileSize: 4},
	}
	if err := os.MkdirAll(filepath.Join(dirA, "set"), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dirA, "set", "p1.mp3"), []byte("trk1"), 0o640); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dirA, "set", "p2.mp3"), []byte("trk2"), 0o640); err != nil {
		t.Fatal(err)
	}
	setID, err := store.UpsertAudiobookSet(ctx, &models.Book{
		LibraryID: libA, Title: "Set Book", Author: "Feed Author",
		Format: models.FormatAudiobook, RelPath: "set/", FileSize: 8,
	}, setTracks)
	if err != nil {
		t.Fatal(err)
	}

	coverBytes := []byte{0xFF, 0xD8, 0xFF, 0xE0, 1, 2, 3}
	if err := os.MkdirAll(srv.cfg.CoverDir(), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := writeCoverFile(srv.cfg.CoverDir(), inID, coverBytes); err != nil {
		t.Fatal(err)
	}
	if err := store.SetBookCover(ctx, inID, true); err != nil {
		t.Fatal(err)
	}

	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}
	session, _, csrf := loginUser(t, handler, "admin", "longpassword")
	ft := createFeed(t, handler, session, csrf, map[string]any{
		"name": "scoped", "libraryId": libA,
	})

	// The set produces one enclosure per track; the single file produces one.
	_, doc := getFeedXML(t, handler, ft.Token)
	if len(doc.Channel.Items) != 3 {
		t.Fatalf("items=%d want 3 (2 tracks + 1 single)", len(doc.Channel.Items))
	}
	encURLs := map[string]bool{}
	for _, it := range doc.Channel.Items {
		encURLs[it.Enclosure.URL] = true
	}
	for _, want := range []string{
		fmt.Sprintf("/feed/%s/item/%d/file?track=0", ft.Token, setID),
		fmt.Sprintf("/feed/%s/item/%d/file?track=1", ft.Token, setID),
		fmt.Sprintf("/feed/%s/item/%d/file", ft.Token, inID),
	} {
		found := false
		for u := range encURLs {
			if len(u) >= len(want) && u[len(u)-len(want):] == want {
				found = true
			}
		}
		if !found {
			t.Errorf("missing enclosure %q in %v", want, encURLs)
		}
	}

	// Enclosure serves real bytes with audio Content-Type and length.
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/feed/%s/item/%d/file", ft.Token, inID), nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("file status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Equal(rec.Body.Bytes(), payload) {
		t.Fatalf("file body=%q", rec.Body.Bytes())
	}
	if ct := rec.Header().Get("Content-Type"); ct != "audio/mpeg" {
		t.Fatalf("file content-type=%q", ct)
	}
	if cl := rec.Header().Get("Content-Length"); cl != strconv.Itoa(len(payload)) {
		t.Fatalf("content-length=%q", cl)
	}

	// Range requests get partial content like the regular media route.
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/feed/%s/item/%d/file", ft.Token, inID), nil)
	req.Header.Set("Range", "bytes=0-3")
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusPartialContent {
		t.Fatalf("range status=%d", rec.Code)
	}
	if !bytes.Equal(rec.Body.Bytes(), payload[:4]) {
		t.Fatalf("range body=%q", rec.Body.Bytes())
	}

	// Track selection serves the requested file of the set.
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/feed/%s/item/%d/file?track=1", ft.Token, setID), nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || !bytes.Equal(rec.Body.Bytes(), []byte("trk2")) {
		t.Fatalf("track=1 status=%d body=%q", rec.Code, rec.Body.Bytes())
	}
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/feed/%s/item/%d/file?track=9", ft.Token, setID), nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("track=9 status=%d want 404", rec.Code)
	}

	// Books outside the feed scope or non-audio are 404 on both item routes.
	for _, bookID := range []int64{outID, epubID} {
		for _, suffix := range []string{"file", "cover"} {
			req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/feed/%s/item/%d/%s", ft.Token, bookID, suffix), nil)
			rec = httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			if rec.Code != http.StatusNotFound {
				t.Fatalf("item %d %s status=%d want 404", bookID, suffix, rec.Code)
			}
		}
	}

	// Cover serves the stored image bytes for an in-scope book.
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/feed/%s/item/%d/cover", ft.Token, inID), nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("cover status=%d", rec.Code)
	}
	if !bytes.Equal(rec.Body.Bytes(), coverBytes) {
		t.Fatalf("cover body=%v", rec.Body.Bytes())
	}
}

func TestFeedDeleteScoping(t *testing.T) {
	srv, store := testServer(t)
	ctx := context.Background()
	hash, _ := auth.HashPassword("longpassword")
	if _, err := store.CreateUser(ctx, "admin", hash, true); err != nil {
		t.Fatal(err)
	}
	if _, err := store.CreateUser(ctx, "other", hash, false); err != nil {
		t.Fatal(err)
	}
	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}
	adminSess, _, adminCSRF := loginUser(t, handler, "admin", "longpassword")
	otherSess, _, otherCSRF := loginUser(t, handler, "other", "longpassword")

	ft := createFeed(t, handler, adminSess, adminCSRF, map[string]any{"name": "admins"})
	del := fmt.Sprintf("/api/feeds/%d", ft.ID)

	// Another user cannot delete it.
	req := httptest.NewRequest(http.MethodDelete, del, nil)
	req.AddCookie(otherSess)
	withCSRF(req, otherCSRF)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound && rec.Code != http.StatusForbidden {
		t.Fatalf("cross-user delete status=%d", rec.Code)
	}

	// A non-admin cannot abuse ?userId= either.
	req = httptest.NewRequest(http.MethodDelete, del+"?userId=1", nil)
	req.AddCookie(otherSess)
	withCSRF(req, otherCSRF)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden && rec.Code != http.StatusUnauthorized {
		t.Fatalf("non-admin userId delete status=%d", rec.Code)
	}

	// Admin can delete another user's feed via ?userId=.
	otherUser, _, err := store.GetUserByUsername(ctx, "other")
	if err != nil {
		t.Fatal(err)
	}
	otherFeed := createFeed(t, handler, otherSess, otherCSRF, map[string]any{"name": "others"})
	req = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/feeds/%d?userId=%d", otherFeed.ID, otherUser.ID), nil)
	req.AddCookie(adminSess)
	withCSRF(req, adminCSRF)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("admin userId delete status=%d body=%s", rec.Code, rec.Body.String())
	}

	// The owner deletes their own feed: 204, then the public URL is gone.
	req = httptest.NewRequest(http.MethodDelete, del, nil)
	req.AddCookie(adminSess)
	withCSRF(req, adminCSRF)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("delete status=%d body=%s", rec.Code, rec.Body.String())
	}
	feedRec, _ := getFeedXML(t, handler, ft.Token)
	if feedRec.Code != http.StatusNotFound {
		t.Fatalf("deleted feed status=%d", feedRec.Code)
	}
}

func TestFeedCollectionScoping(t *testing.T) {
	srv, store := testServer(t)
	ctx := context.Background()
	hash, _ := auth.HashPassword("longpassword")
	if _, err := store.CreateUser(ctx, "admin", hash, true); err != nil {
		t.Fatal(err)
	}
	if _, err := store.CreateUser(ctx, "other", hash, false); err != nil {
		t.Fatal(err)
	}
	libID, libDir := seedAudioLibrary(t, srv, store, "main")
	inID := seedAudioBook(t, store, libID, libDir, "in.mp3", "Picked", models.FormatMP3, []byte("in"))
	seedAudioBook(t, store, libID, libDir, "out.mp3", "Skipped", models.FormatMP3, []byte("out"))

	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}
	session, _, csrf := loginUser(t, handler, "admin", "longpassword")
	otherSess, _, otherCSRF := loginUser(t, handler, "other", "longpassword")

	adminUser, _, err := store.GetUserByUsername(ctx, "admin")
	if err != nil {
		t.Fatal(err)
	}
	col, err := store.CreateCollection(ctx, adminUser.ID, "Picks", "")
	if err != nil {
		t.Fatal(err)
	}
	if err := store.AddToCollection(ctx, adminUser.ID, col.ID, inID); err != nil {
		t.Fatal(err)
	}

	ft := createFeed(t, handler, session, csrf, map[string]any{
		"name": "coll", "collectionId": col.ID,
	})
	_, doc := getFeedXML(t, handler, ft.Token)
	if len(doc.Channel.Items) != 1 || doc.Channel.Items[0].Title != "Picked" {
		t.Fatalf("collection feed items=%+v", doc.Channel.Items)
	}

	// Enclosure for a book outside the collection is denied.
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/feed/%s/item/%d/file", ft.Token, inID), nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("in-collection file status=%d", rec.Code)
	}
	var skippedID int64
	if err := store.DB().QueryRowContext(ctx, `SELECT id FROM books WHERE title='Skipped'`).Scan(&skippedID); err != nil {
		t.Fatal(err)
	}
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/feed/%s/item/%d/file", ft.Token, skippedID), nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("out-of-collection file status=%d", rec.Code)
	}

	// Another user cannot create a feed scoped to a collection they do not own.
	req = httptest.NewRequest(http.MethodPost, "/api/feeds",
		bytes.NewBufferString(fmt.Sprintf(`{"name":"x","collectionId":%d}`, col.ID)))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(otherSess)
	withCSRF(req, otherCSRF)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code == http.StatusCreated {
		t.Fatal("cross-user collection feed created")
	}
}
