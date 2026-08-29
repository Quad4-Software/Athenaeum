package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/pquerna/otp/totp"

	"athenaeum/internal/auth"
	"athenaeum/internal/models"
	"athenaeum/internal/storage"
)

func TestAPICoverageHappyPaths(t *testing.T) {
	srv, store := testServer(t)
	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	session, csrf := loginAdmin(t, handler, store)

	hookHits := 0
	hookSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hookHits++
		io.Copy(io.Discard, r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(hookSrv.Close)

	libDir := filepath.Join(srv.cfg.DataDir, "lib")
	if err := os.MkdirAll(libDir, 0o750); err != nil {
		t.Fatal(err)
	}
	bookPath := filepath.Join(libDir, "coverage.pdf")
	if err := os.WriteFile(bookPath, []byte("%PDF-1.4 coverage"), 0o640); err != nil {
		t.Fatal(err)
	}

	do := func(method, path string, body any) *httptest.ResponseRecorder {
		t.Helper()
		var rdr io.Reader
		if body != nil {
			switch v := body.(type) {
			case []byte:
				rdr = bytes.NewReader(v)
			case string:
				rdr = strings.NewReader(v)
			case io.Reader:
				rdr = v
			default:
				b, err := json.Marshal(v)
				if err != nil {
					t.Fatal(err)
				}
				rdr = bytes.NewReader(b)
			}
		}
		req := httptest.NewRequest(method, path, rdr)
		if body != nil {
			if _, ok := body.(io.Reader); !ok {
				if _, ok := body.([]byte); !ok {
					req.Header.Set("Content-Type", "application/json")
				}
			}
		}
		req.AddCookie(session)
		if method != http.MethodGet && method != http.MethodHead {
			withCSRF(req, csrf)
		}
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		return rec
	}
	expect := func(rec *httptest.ResponseRecorder, codes ...int) {
		t.Helper()
		for _, c := range codes {
			if rec.Code == c {
				return
			}
		}
		t.Fatalf("status=%d want %v body=%s", rec.Code, codes, rec.Body.String())
	}

	rec := do(http.MethodGet, "/api/auth/me", nil)
	expect(rec, http.StatusOK)
	var me models.UserPublic
	if err := json.NewDecoder(rec.Body).Decode(&me); err != nil {
		t.Fatal(err)
	}
	adminID := me.ID

	rec = do(http.MethodPost, "/api/libraries", map[string]any{
		"name": "Main", "mountPath": libDir, "backend": "local",
	})
	expect(rec, http.StatusCreated)
	var lib models.Library
	if err := json.NewDecoder(rec.Body).Decode(&lib); err != nil {
		t.Fatal(err)
	}
	libID := lib.ID

	extraDir := filepath.Join(srv.cfg.DataDir, "lib2")
	if err := os.MkdirAll(extraDir, 0o750); err != nil {
		t.Fatal(err)
	}
	rec = do(http.MethodPost, "/api/libraries", map[string]any{
		"name": "Extra", "mountPath": extraDir, "backend": "local",
	})
	expect(rec, http.StatusCreated)
	var lib2 models.Library
	if err := json.NewDecoder(rec.Body).Decode(&lib2); err != nil {
		t.Fatal(err)
	}

	bookID, err := store.UpsertBook(ctx, &models.Book{
		LibraryID: libID,
		Title:     "Coverage Book",
		Author:    "Tester",
		Series:    "Coverage Series",
		Format:    models.FormatPDF,
		RelPath:   "coverage.pdf",
		FileSize:  int64(len("%PDF-1.4 coverage")),
	}, 1)
	if err != nil {
		t.Fatal(err)
	}

	rec = do(http.MethodGet, "/api/libraries", nil)
	expect(rec, http.StatusOK)
	rec = do(http.MethodGet, fmt.Sprintf("/api/libraries/%d", libID), nil)
	expect(rec, http.StatusOK)
	rec = do(http.MethodPut, fmt.Sprintf("/api/libraries/%d", libID), map[string]any{
		"name": "Main Renamed", "mountPath": libDir, "backend": "local",
	})
	expect(rec, http.StatusOK)
	rec = do(http.MethodPut, "/api/libraries/reorder", map[string]any{"ids": []int64{lib2.ID, libID}})
	expect(rec, http.StatusOK)
	rec = do(http.MethodPost, fmt.Sprintf("/api/libraries/%d/scan", libID), nil)
	expect(rec, http.StatusAccepted)

	rec = do(http.MethodPut, fmt.Sprintf("/api/books/%d/favorite", bookID), map[string]any{"favorite": true})
	expect(rec, http.StatusOK)
	rec = do(http.MethodGet, fmt.Sprintf("/api/books/%d/favorite", bookID), nil)
	expect(rec, http.StatusOK)
	rec = do(http.MethodGet, "/api/favorites", nil)
	expect(rec, http.StatusOK)

	rec = do(http.MethodPost, "/api/offline", map[string]any{"bookIds": []int64{bookID}})
	expect(rec, http.StatusOK)
	rec = do(http.MethodGet, "/api/offline", nil)
	expect(rec, http.StatusOK)
	rec = do(http.MethodDelete, "/api/offline", map[string]any{"bookIds": []int64{bookID}})
	expect(rec, http.StatusOK)

	rec = do(http.MethodPost, fmt.Sprintf("/api/books/%d/bookmarks", bookID), map[string]any{
		"location": "page-1", "label": "start",
	})
	expect(rec, http.StatusCreated)
	var bm models.Bookmark
	if err := json.NewDecoder(rec.Body).Decode(&bm); err != nil {
		t.Fatal(err)
	}
	rec = do(http.MethodGet, fmt.Sprintf("/api/books/%d/bookmarks", bookID), nil)
	expect(rec, http.StatusOK)
	rec = do(http.MethodDelete, fmt.Sprintf("/api/books/%d/bookmarks/%d", bookID, bm.ID), nil)
	expect(rec, http.StatusOK)

	rec = do(http.MethodPost, fmt.Sprintf("/api/books/%d/highlights", bookID), map[string]any{
		"location": "page-2", "excerpt": "hello", "note": "n", "color": "yellow",
	})
	expect(rec, http.StatusCreated)
	var hl models.Highlight
	if err := json.NewDecoder(rec.Body).Decode(&hl); err != nil {
		t.Fatal(err)
	}
	rec = do(http.MethodGet, fmt.Sprintf("/api/books/%d/highlights", bookID), nil)
	expect(rec, http.StatusOK)
	rec = do(http.MethodDelete, fmt.Sprintf("/api/books/%d/highlights/%d", bookID, hl.ID), nil)
	expect(rec, http.StatusOK)

	rec = do(http.MethodPost, fmt.Sprintf("/api/books/%d/reading-time", bookID), map[string]any{"seconds": 120})
	expect(rec, http.StatusOK)
	rec = do(http.MethodGet, "/api/stats/reading", nil)
	expect(rec, http.StatusOK)

	rec = do(http.MethodPost, "/api/tags", map[string]any{"name": "coverage-tag"})
	expect(rec, http.StatusCreated)
	rec = do(http.MethodGet, fmt.Sprintf("/api/books/%d/tags", bookID), nil)
	expect(rec, http.StatusOK)
	rec = do(http.MethodPost, fmt.Sprintf("/api/books/%d/tags", bookID), map[string]any{"name": "book-tag"})
	expect(rec, http.StatusOK)

	rec = do(http.MethodPost, fmt.Sprintf("/api/books/%d/share", bookID), map[string]any{
		"expiresInHours": 24, "maxDownloads": 5,
	})
	expect(rec, http.StatusCreated)
	var shareCreated struct {
		ID    int64  `json:"id"`
		Token string `json:"token"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&shareCreated); err != nil {
		t.Fatal(err)
	}
	rec = do(http.MethodGet, fmt.Sprintf("/api/books/%d/share", bookID), nil)
	expect(rec, http.StatusOK)
	rec = do(http.MethodGet, "/api/share/"+shareCreated.Token, nil)
	expect(rec, http.StatusOK)
	dlReq := httptest.NewRequest(http.MethodGet, "/share/"+shareCreated.Token+"/download", nil)
	dlRec := httptest.NewRecorder()
	handler.ServeHTTP(dlRec, dlReq)
	expect(dlRec, http.StatusOK)
	rec = do(http.MethodDelete, fmt.Sprintf("/api/books/%d/share/%d", bookID, shareCreated.ID), nil)
	expect(rec, http.StatusOK)

	rec = do(http.MethodPost, "/api/invites", map[string]any{
		"kind": models.InviteKindPermanent, "email": "invitee@example.com", "expiresInHours": 48,
	})
	expect(rec, http.StatusCreated)
	var invRes models.InviteCreateResult
	if err := json.NewDecoder(rec.Body).Decode(&invRes); err != nil {
		t.Fatal(err)
	}
	rec = do(http.MethodGet, "/api/invites", nil)
	expect(rec, http.StatusOK)
	metaReq := httptest.NewRequest(http.MethodGet, "/api/invite/"+invRes.Invite.Token, nil)
	metaRec := httptest.NewRecorder()
	handler.ServeHTTP(metaRec, metaReq)
	expect(metaRec, http.StatusOK)

	acceptCSRF := fetchCSRF(t, handler)
	acceptBody, _ := json.Marshal(map[string]string{"username": "inviteduser", "password": "longpassword"})
	acceptReq := httptest.NewRequest(http.MethodPost, "/api/invite/"+invRes.Invite.Token+"/accept", bytes.NewReader(acceptBody))
	acceptReq.Header.Set("Content-Type", "application/json")
	withCSRF(acceptReq, acceptCSRF)
	acceptRec := httptest.NewRecorder()
	handler.ServeHTTP(acceptRec, acceptReq)
	expect(acceptRec, http.StatusCreated)

	rec = do(http.MethodPost, "/api/invites", map[string]any{"kind": models.InviteKindPermanent, "expiresInHours": 12})
	expect(rec, http.StatusCreated)
	var inv2 models.InviteCreateResult
	if err := json.NewDecoder(rec.Body).Decode(&inv2); err != nil {
		t.Fatal(err)
	}
	rec = do(http.MethodDelete, fmt.Sprintf("/api/invites/%d", inv2.Invite.ID), nil)
	expect(rec, http.StatusOK)

	_, _, sess2CSRF := loginUser(t, handler, "admin", "longpassword")
	_ = sess2CSRF
	rec = do(http.MethodGet, "/api/auth/sessions", nil)
	expect(rec, http.StatusOK)
	rec = do(http.MethodDelete, "/api/auth/sessions", nil)
	expect(rec, http.StatusOK)
	rec = do(http.MethodGet, fmt.Sprintf("/api/auth/users/%d/sessions", adminID), nil)
	expect(rec, http.StatusOK)

	rec = do(http.MethodGet, "/api/auth/settings", nil)
	expect(rec, http.StatusOK)
	rec = do(http.MethodPut, "/api/auth/settings", models.AuthSettings{AllowRegistration: true})
	expect(rec, http.StatusOK)
	rec = do(http.MethodPut, "/api/auth/profile", map[string]any{"username": "admin"})
	expect(rec, http.StatusOK)

	checkCSRF := fetchCSRF(t, handler)
	checkReq := httptest.NewRequest(http.MethodPost, "/api/auth/password/check", bytes.NewBufferString(`{"password":"longpassword"}`))
	checkReq.Header.Set("Content-Type", "application/json")
	withCSRF(checkReq, checkCSRF)
	checkRec := httptest.NewRecorder()
	handler.ServeHTTP(checkRec, checkReq)
	expect(checkRec, http.StatusOK)

	rec = do(http.MethodGet, "/api/auth/users", nil)
	expect(rec, http.StatusOK)

	rec = do(http.MethodPost, "/api/auth/users/guest", map[string]any{
		"username": "tempguest", "expiresInHours": 24,
	})
	expect(rec, http.StatusCreated)
	var guestCred models.GuestCredentials
	if err := json.NewDecoder(rec.Body).Decode(&guestCred); err != nil {
		t.Fatal(err)
	}
	guestID := guestCred.User.ID
	rec = do(http.MethodGet, "/api/auth/users/guests", nil)
	expect(rec, http.StatusOK)
	rec = do(http.MethodPost, fmt.Sprintf("/api/auth/users/guests/%d/extend", guestID), map[string]any{"expiresInHours": 48})
	expect(rec, http.StatusOK)

	hash2, err := auth.HashPassword("longpassword")
	if err != nil {
		t.Fatal(err)
	}
	victimID, err := store.CreateUser(ctx, "victim", hash2, false)
	if err != nil {
		t.Fatal(err)
	}
	rec = do(http.MethodPut, fmt.Sprintf("/api/auth/users/%d/permissions", victimID), map[string]any{
		"permissions": []string{"read", "edit_metadata"},
	})
	expect(rec, http.StatusOK)
	rec = do(http.MethodPut, fmt.Sprintf("/api/auth/users/%d/password", victimID), map[string]any{"password": "longpassword2"})
	expect(rec, http.StatusOK)
	rec = do(http.MethodPut, fmt.Sprintf("/api/auth/users/%d/admin", victimID), map[string]any{"admin": true})
	expect(rec, http.StatusOK)
	rec = do(http.MethodPut, fmt.Sprintf("/api/auth/users/%d/admin", victimID), map[string]any{"admin": false})
	expect(rec, http.StatusOK)
	rec = do(http.MethodGet, fmt.Sprintf("/api/auth/users/%d/libraries", victimID), nil)
	expect(rec, http.StatusOK)
	rec = do(http.MethodPut, fmt.Sprintf("/api/auth/users/%d/libraries", victimID), map[string]any{"libraryIds": []int64{libID}})
	expect(rec, http.StatusOK)
	rec = do(http.MethodDelete, fmt.Sprintf("/api/auth/users/%d", victimID), nil)
	expect(rec, http.StatusOK)

	rec = do(http.MethodPost, "/api/auth/users/guests/bulk-delete", map[string]any{"ids": []int64{guestID}})
	expect(rec, http.StatusOK)
	rec = do(http.MethodGet, "/api/auth/audit", nil)
	expect(rec, http.StatusOK)

	regCSRF := fetchCSRF(t, handler)
	regBody, _ := json.Marshal(map[string]string{"username": "publicreg", "password": "longpassword"})
	regReq := httptest.NewRequest(http.MethodPost, "/api/auth/register-public", bytes.NewReader(regBody))
	regReq.Header.Set("Content-Type", "application/json")
	withCSRF(regReq, regCSRF)
	regRec := httptest.NewRecorder()
	handler.ServeHTTP(regRec, regReq)
	expect(regRec, http.StatusCreated)

	rec = do(http.MethodGet, "/api/auth/api-keys", nil)
	expect(rec, http.StatusOK)
	rec = do(http.MethodPost, "/api/auth/api-keys", map[string]any{"name": "coverage-key"})
	expect(rec, http.StatusCreated)
	var key models.APIKeyCreated
	if err := json.NewDecoder(rec.Body).Decode(&key); err != nil {
		t.Fatal(err)
	}
	rec = do(http.MethodDelete, fmt.Sprintf("/api/auth/api-keys/%d", key.ID), nil)
	expect(rec, http.StatusNoContent)

	for _, path := range []string{"/docs", "/docs/app.js", "/api/docs", "/api/openapi.json"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		r := httptest.NewRecorder()
		handler.ServeHTTP(r, req)
		expect(r, http.StatusOK)
	}

	rec = do(http.MethodGet, "/api/admin/smtp", nil)
	expect(rec, http.StatusOK)
	rec = do(http.MethodPut, "/api/admin/smtp", models.SMTPSettings{
		Enabled: false, Host: "smtp.example", Port: 587, Username: "u", FromAddr: "a@b.c",
	})
	expect(rec, http.StatusOK)
	rec = do(http.MethodGet, "/api/auth/kindle-email", nil)
	expect(rec, http.StatusOK)
	rec = do(http.MethodPut, "/api/auth/kindle-email", map[string]any{"email": "kindle@example.com"})
	expect(rec, http.StatusOK)

	rec = do(http.MethodGet, "/api/admin/pocketid", nil)
	expect(rec, http.StatusOK)
	rec = do(http.MethodPut, "/api/admin/pocketid", models.PocketIDSettings{
		Enabled: false, BaseURL: "https://pocket.example", APIKey: "test-key", DefaultGroupIDs: []string{},
	})
	expect(rec, http.StatusOK)

	rec = do(http.MethodPost, "/api/admin/webhooks", map[string]any{
		"url": hookSrv.URL, "secret": "sekrit", "events": []string{"user.create", "ping"},
	})
	expect(rec, http.StatusCreated)
	var wh models.WebhookPublic
	if err := json.NewDecoder(rec.Body).Decode(&wh); err != nil {
		t.Fatal(err)
	}
	rec = do(http.MethodGet, "/api/admin/webhooks", nil)
	expect(rec, http.StatusOK)
	rec = do(http.MethodGet, fmt.Sprintf("/api/admin/webhooks/%d", wh.ID), nil)
	expect(rec, http.StatusOK)
	rec = do(http.MethodPut, fmt.Sprintf("/api/admin/webhooks/%d", wh.ID), map[string]any{
		"url": hookSrv.URL, "events": []string{"user.create"}, "enabled": true,
	})
	expect(rec, http.StatusOK)
	rec = do(http.MethodPost, fmt.Sprintf("/api/admin/webhooks/%d/test", wh.ID), nil)
	expect(rec, http.StatusOK)
	if hookHits < 1 {
		t.Fatal("expected webhook delivery hit")
	}
	rec = do(http.MethodGet, fmt.Sprintf("/api/admin/webhooks/%d/deliveries", wh.ID), nil)
	expect(rec, http.StatusOK)
	rec = do(http.MethodDelete, fmt.Sprintf("/api/admin/webhooks/%d", wh.ID), nil)
	expect(rec, http.StatusOK)

	rec = do(http.MethodGet, "/api/admin/server", nil)
	expect(rec, http.StatusOK)
	rec = do(http.MethodGet, "/api/system/stats", nil)
	expect(rec, http.StatusOK)

	for _, path := range []string{
		"/opds/", "/opds/recent", "/opds/search?q=Coverage", "/opds/series",
		"/opds/series/Coverage%20Series", "/opds/comics", "/opds/kindle", "/opds/v2/", "/opds/v2/recent",
	} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.AddCookie(session)
		r := httptest.NewRecorder()
		handler.ServeHTTP(r, req)
		expect(r, http.StatusOK)
	}

	rec = do(http.MethodGet, "/api/library/stats", nil)
	expect(rec, http.StatusOK)
	rec = do(http.MethodGet, "/api/series", nil)
	expect(rec, http.StatusOK)
	rec = do(http.MethodGet, "/api/authors", nil)
	expect(rec, http.StatusOK)
	rec = do(http.MethodGet, "/api/library/scan/status", nil)
	expect(rec, http.StatusOK)
	rec = do(http.MethodGet, fmt.Sprintf("/api/books/%d/chapters", bookID), nil)
	expect(rec, http.StatusOK)
	rec = do(http.MethodPut, fmt.Sprintf("/api/books/%d/progress", bookID), map[string]any{
		"location": "loc-1", "percent": 12.5,
	})
	expect(rec, http.StatusOK)
	rec = do(http.MethodGet, fmt.Sprintf("/api/books/%d/progress", bookID), nil)
	expect(rec, http.StatusOK)

	rec = do(http.MethodGet, "/api/metadata/providers", nil)
	expect(rec, http.StatusOK)
	rec = do(http.MethodPost, "/api/library/series/cleanup", nil)
	expect(rec, http.StatusOK)
	rec = do(http.MethodGet, "/api/library/metadata/match/status", nil)
	expect(rec, http.StatusOK)
	rec = do(http.MethodPost, "/api/library/metadata/match", map[string]any{
		"bookIds": []int64{bookID}, "applyCover": false,
	})
	expect(rec, http.StatusAccepted, http.StatusConflict)

	rec = do(http.MethodGet, "/api/admin/tasks/status", nil)
	expect(rec, http.StatusOK)
	rec = do(http.MethodPost, "/api/admin/tasks/prune-missing", nil)
	expect(rec, http.StatusOK)
	rec = do(http.MethodPost, "/api/admin/tasks/cleanup-covers", nil)
	expect(rec, http.StatusOK)
	rec = do(http.MethodPost, "/api/admin/tasks/regenerate-covers", map[string]any{"libraryId": libID})
	expect(rec, http.StatusAccepted, http.StatusConflict)
	rec = do(http.MethodPost, "/api/admin/tasks/cleanup-series", nil)
	expect(rec, http.StatusOK)
	rec = do(http.MethodPost, "/api/admin/tasks/cleanup-text", nil)
	expect(rec, http.StatusOK)

	rec = do(http.MethodGet, "/api/fs/browse?path="+libDir, nil)
	expect(rec, http.StatusOK)

	kosyncReq := httptest.NewRequest(http.MethodGet, "/kosync/users/auth", nil)
	kosyncReq.SetBasicAuth("admin", "longpassword")
	kosyncRec := httptest.NewRecorder()
	handler.ServeHTTP(kosyncRec, kosyncReq)
	expect(kosyncRec, http.StatusOK)
	kosyncPutBody, _ := json.Marshal(map[string]any{
		"document": "doc-1", "progress": "50", "percentage": 0.5, "device": "test", "device_id": "d1",
	})
	kosyncPut := httptest.NewRequest(http.MethodPut, "/kosync/syncs/progress", bytes.NewReader(kosyncPutBody))
	kosyncPut.Header.Set("Content-Type", "application/json")
	kosyncPut.SetBasicAuth("admin", "longpassword")
	kosyncPutRec := httptest.NewRecorder()
	handler.ServeHTTP(kosyncPutRec, kosyncPut)
	expect(kosyncPutRec, http.StatusOK)
	kosyncGet := httptest.NewRequest(http.MethodGet, "/kosync/syncs/progress/doc-1", nil)
	kosyncGet.SetBasicAuth("admin", "longpassword")
	kosyncGetRec := httptest.NewRecorder()
	handler.ServeHTTP(kosyncGetRec, kosyncGet)
	expect(kosyncGetRec, http.StatusOK)

	rec = do(http.MethodPost, "/api/collections", map[string]any{
		"name": "My Collection", "description": "d",
	})
	expect(rec, http.StatusCreated)
	var coll models.Collection
	if err := json.NewDecoder(rec.Body).Decode(&coll); err != nil {
		t.Fatal(err)
	}
	rec = do(http.MethodGet, fmt.Sprintf("/api/collections/%d", coll.ID), nil)
	expect(rec, http.StatusOK)
	rec = do(http.MethodPut, fmt.Sprintf("/api/collections/%d", coll.ID), map[string]any{
		"name": "My Collection 2", "description": "d2",
	})
	expect(rec, http.StatusOK)
	rec = do(http.MethodPost, fmt.Sprintf("/api/collections/%d/books/%d", coll.ID, bookID), nil)
	expect(rec, http.StatusNoContent)
	rec = do(http.MethodDelete, fmt.Sprintf("/api/collections/%d/books/%d", coll.ID, bookID), nil)
	expect(rec, http.StatusNoContent)
	rec = do(http.MethodDelete, fmt.Sprintf("/api/collections/%d", coll.ID), nil)
	expect(rec, http.StatusNoContent)

	var coverBuf bytes.Buffer
	mw := multipart.NewWriter(&coverBuf)
	part, err := mw.CreateFormFile("cover", "c.jpg")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write([]byte{0xff, 0xd8, 0xff, 0xd9}); err != nil {
		t.Fatal(err)
	}
	if err := mw.Close(); err != nil {
		t.Fatal(err)
	}
	coverReq := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/books/%d/cover", bookID), &coverBuf)
	coverReq.Header.Set("Content-Type", mw.FormDataContentType())
	coverReq.AddCookie(session)
	withCSRF(coverReq, csrf)
	coverRec := httptest.NewRecorder()
	handler.ServeHTTP(coverRec, coverReq)
	expect(coverRec, http.StatusOK)
	rec = do(http.MethodGet, fmt.Sprintf("/api/books/%d/cover", bookID), nil)
	expect(rec, http.StatusOK)
	rec = do(http.MethodDelete, fmt.Sprintf("/api/books/%d/cover", bookID), nil)
	expect(rec, http.StatusOK)

	rec = do(http.MethodGet, fmt.Sprintf("/api/books/%d/file", bookID), nil)
	expect(rec, http.StatusOK)
	rec = do(http.MethodGet, fmt.Sprintf("/api/books/%d/download", bookID), nil)
	expect(rec, http.StatusOK)

	payload := []byte("hello-upload-pdf-bytes!!")
	rec = do(http.MethodPost, fmt.Sprintf("/api/libraries/%d/uploads", libID), map[string]any{
		"relPath": "uploads/hello.pdf", "totalSize": len(payload),
	})
	expect(rec, http.StatusCreated)
	var up models.UploadSession
	if err := json.NewDecoder(rec.Body).Decode(&up); err != nil {
		t.Fatal(err)
	}
	rec = do(http.MethodGet, fmt.Sprintf("/api/libraries/%d/uploads/%s", libID, up.ID), nil)
	expect(rec, http.StatusOK)
	patchReq := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/libraries/%d/uploads/%s", libID, up.ID), bytes.NewReader(payload[:10]))
	patchReq.Header.Set("Content-Range", fmt.Sprintf("bytes 0-9/%d", len(payload)))
	patchReq.AddCookie(session)
	withCSRF(patchReq, csrf)
	patchRec := httptest.NewRecorder()
	handler.ServeHTTP(patchRec, patchReq)
	expect(patchRec, http.StatusOK)
	rec = do(http.MethodDelete, fmt.Sprintf("/api/libraries/%d/uploads/%s", libID, up.ID), nil)
	expect(rec, http.StatusOK, http.StatusNoContent)

	rec = do(http.MethodGet, "/api/admin/backup", nil)
	expect(rec, http.StatusOK)
	rec = do(http.MethodGet, "/api/admin/config/export", nil)
	expect(rec, http.StatusOK)

	rec = do(http.MethodGet, "/api/auth/methods", nil)
	expect(rec, http.StatusOK)
	rec = do(http.MethodGet, "/api/auth/oidc/config", nil)
	expect(rec, http.StatusOK)
	rec = do(http.MethodPut, "/api/auth/oidc/config", models.OIDCConfig{
		Enabled: false, LoginLocal: true, ButtonText: "SSO",
	})
	expect(rec, http.StatusOK)

	rec = do(http.MethodPost, "/api/admin/content-index", nil)
	expect(rec, http.StatusAccepted)

	rec = do(http.MethodPost, "/api/auth/totp/setup", nil)
	expect(rec, http.StatusOK)
	var setup totpSetupResponse
	if err := json.NewDecoder(rec.Body).Decode(&setup); err != nil {
		t.Fatal(err)
	}
	code, err := totp.GenerateCode(setup.Secret, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	rec = do(http.MethodPost, "/api/auth/totp/enable", map[string]any{"code": code})
	expect(rec, http.StatusOK)

	logoutReq := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	logoutReq.AddCookie(session)
	withCSRF(logoutReq, csrf)
	logoutRec := httptest.NewRecorder()
	handler.ServeHTTP(logoutRec, logoutReq)
	expect(logoutRec, http.StatusOK, http.StatusNoContent)

	loginCSRF := fetchCSRF(t, handler)
	loginBody := bytes.NewBufferString(`{"username":"admin","password":"longpassword"}`)
	loginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", loginBody)
	loginReq.Header.Set("Content-Type", "application/json")
	withCSRF(loginReq, loginCSRF)
	loginRec := httptest.NewRecorder()
	handler.ServeHTTP(loginRec, loginReq)
	expect(loginRec, http.StatusOK)
	var challenge models.LoginChallenge
	if err := json.NewDecoder(loginRec.Body).Decode(&challenge); err != nil {
		t.Fatal(err)
	}
	if !challenge.NeedsTOTP || challenge.TOTPToken == "" {
		t.Fatalf("expected totp challenge, got %+v", challenge)
	}
	verifyCode, err := totp.GenerateCode(setup.Secret, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	verifyCSRF := fetchCSRF(t, handler)
	verifyBody, _ := json.Marshal(map[string]string{"totpToken": challenge.TOTPToken, "code": verifyCode})
	verifyReq := httptest.NewRequest(http.MethodPost, "/api/auth/totp/verify", bytes.NewReader(verifyBody))
	verifyReq.Header.Set("Content-Type", "application/json")
	withCSRF(verifyReq, verifyCSRF)
	verifyRec := httptest.NewRecorder()
	handler.ServeHTTP(verifyRec, verifyReq)
	expect(verifyRec, http.StatusOK)
	for _, c := range verifyRec.Result().Cookies() {
		switch c.Name {
		case auth.SessionCookie:
			session = c
		case auth.CSRFCookie:
			csrf = c
		}
	}
	if session == nil || csrf == nil {
		t.Fatal("missing cookies after totp verify")
	}

	disableCode, err := totp.GenerateCode(setup.Secret, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	rec = do(http.MethodPost, "/api/auth/totp/disable", map[string]any{
		"password": "longpassword", "code": disableCode,
	})
	expect(rec, http.StatusOK)

	rec = do(http.MethodDelete, fmt.Sprintf("/api/libraries/%d", lib2.ID), nil)
	expect(rec, http.StatusNoContent)

	_ = store
	_ = storage.ErrNotFound
}
