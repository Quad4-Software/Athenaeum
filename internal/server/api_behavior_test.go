package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/pquerna/otp/totp"

	"athenaeum/internal/auth"
	"athenaeum/internal/models"
	"athenaeum/internal/storage"
)

// adminClient is a small authenticated API helper for behavior tests.
type adminClient struct {
	t       *testing.T
	handler http.Handler
	store   *storage.Store
	session *http.Cookie
	csrf    *http.Cookie
}

func newAdminClient(t *testing.T) (*Server, *adminClient) {
	t.Helper()
	srv, store := testServer(t)
	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}
	session, csrf := loginAdmin(t, handler, store)
	return srv, &adminClient{t: t, handler: handler, store: store, session: session, csrf: csrf}
}

func (c *adminClient) do(method, path string, body any) *httptest.ResponseRecorder {
	c.t.Helper()
	var rdr io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			c.t.Fatal(err)
		}
		rdr = bytes.NewReader(b)
	}
	req := httptest.NewRequest(method, path, rdr)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.AddCookie(c.session)
	if method != http.MethodGet && method != http.MethodHead {
		withCSRF(req, c.csrf)
	}
	rec := httptest.NewRecorder()
	c.handler.ServeHTTP(rec, req)
	return rec
}

func (c *adminClient) mustStatus(rec *httptest.ResponseRecorder, want int) {
	c.t.Helper()
	if rec.Code != want {
		c.t.Fatalf("status=%d want %d body=%s", rec.Code, want, rec.Body.String())
	}
}

func seedLibraryBook(t *testing.T, srv *Server, store *storage.Store, name string, payload []byte) (libID, bookID int64) {
	t.Helper()
	ctx := context.Background()
	libDir := filepath.Join(srv.cfg.DataDir, "lib-"+name)
	if err := os.MkdirAll(libDir, 0o750); err != nil {
		t.Fatal(err)
	}
	rel := name + ".pdf"
	if err := os.WriteFile(filepath.Join(libDir, rel), payload, 0o640); err != nil {
		t.Fatal(err)
	}
	lib, err := store.CreateLibrary(ctx, name, libDir)
	if err != nil {
		t.Fatal(err)
	}
	id, err := store.UpsertBook(ctx, &models.Book{
		LibraryID: lib.ID,
		Title:     name,
		Author:    "Author",
		Format:    models.FormatPDF,
		RelPath:   rel,
		FileSize:  int64(len(payload)),
	}, 1)
	if err != nil {
		t.Fatal(err)
	}
	return lib.ID, id
}

func TestFavoriteRoundTrip(t *testing.T) {
	srv, c := newAdminClient(t)
	_, bookID := seedLibraryBook(t, srv, c.store, "fav", []byte("%PDF-1.4 fav"))

	c.mustStatus(c.do(http.MethodPut, fmt.Sprintf("/api/books/%d/favorite", bookID), map[string]any{"favorite": true}), http.StatusOK)

	rec := c.do(http.MethodGet, fmt.Sprintf("/api/books/%d/favorite", bookID), nil)
	c.mustStatus(rec, http.StatusOK)
	var fav struct {
		Favorite bool `json:"favorite"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&fav); err != nil || !fav.Favorite {
		t.Fatalf("favorite get=%v err=%v body=%s", fav, err, rec.Body.String())
	}

	rec = c.do(http.MethodGet, "/api/favorites", nil)
	c.mustStatus(rec, http.StatusOK)
	var favList struct {
		IDs []int64 `json:"ids"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&favList); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, id := range favList.IDs {
		if id == bookID {
			found = true
		}
	}
	if !found {
		t.Fatalf("favorites=%v missing %d", favList.IDs, bookID)
	}

	c.mustStatus(c.do(http.MethodPut, fmt.Sprintf("/api/books/%d/favorite", bookID), map[string]any{"favorite": false}), http.StatusOK)
	rec = c.do(http.MethodGet, fmt.Sprintf("/api/books/%d/favorite", bookID), nil)
	c.mustStatus(rec, http.StatusOK)
	if err := json.NewDecoder(rec.Body).Decode(&fav); err != nil || fav.Favorite {
		t.Fatalf("expected unfavorited got %+v", fav)
	}
}

func TestOfflineGrantRoundTrip(t *testing.T) {
	srv, c := newAdminClient(t)
	_, bookID := seedLibraryBook(t, srv, c.store, "off", []byte("%PDF-1.4 off"))

	c.mustStatus(c.do(http.MethodPost, "/api/offline", map[string]any{"bookIds": []int64{bookID}}), http.StatusOK)
	rec := c.do(http.MethodGet, "/api/offline", nil)
	c.mustStatus(rec, http.StatusOK)
	var offline struct {
		BookIDs []int64 `json:"bookIds"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&offline); err != nil {
		t.Fatal(err)
	}
	if len(offline.BookIDs) != 1 || offline.BookIDs[0] != bookID {
		t.Fatalf("offline=%v", offline.BookIDs)
	}
	c.mustStatus(c.do(http.MethodDelete, "/api/offline", map[string]any{"bookIds": []int64{bookID}}), http.StatusOK)
	rec = c.do(http.MethodGet, "/api/offline", nil)
	c.mustStatus(rec, http.StatusOK)
	offline.BookIDs = nil
	if err := json.NewDecoder(rec.Body).Decode(&offline); err != nil {
		t.Fatal(err)
	}
	if len(offline.BookIDs) != 0 {
		t.Fatalf("expected empty offline got %v", offline.BookIDs)
	}
}

func TestBookmarkAndProgressPersist(t *testing.T) {
	srv, c := newAdminClient(t)
	_, bookID := seedLibraryBook(t, srv, c.store, "bm", []byte("%PDF-1.4 bm"))

	rec := c.do(http.MethodPost, fmt.Sprintf("/api/books/%d/bookmarks", bookID), map[string]any{
		"location": "cfi(/6/2)", "label": "ch1",
	})
	c.mustStatus(rec, http.StatusCreated)
	var bm models.Bookmark
	if err := json.NewDecoder(rec.Body).Decode(&bm); err != nil {
		t.Fatal(err)
	}
	if bm.Location != "cfi(/6/2)" || bm.Label != "ch1" || bm.ID <= 0 {
		t.Fatalf("bookmark=%+v", bm)
	}
	rec = c.do(http.MethodGet, fmt.Sprintf("/api/books/%d/bookmarks", bookID), nil)
	c.mustStatus(rec, http.StatusOK)
	var list []models.Bookmark
	if err := json.NewDecoder(rec.Body).Decode(&list); err != nil || len(list) != 1 || list[0].ID != bm.ID {
		t.Fatalf("list=%v err=%v", list, err)
	}
	c.mustStatus(c.do(http.MethodDelete, fmt.Sprintf("/api/books/%d/bookmarks/%d", bookID, bm.ID), nil), http.StatusOK)
	rec = c.do(http.MethodGet, fmt.Sprintf("/api/books/%d/bookmarks", bookID), nil)
	c.mustStatus(rec, http.StatusOK)
	list = nil
	if err := json.NewDecoder(rec.Body).Decode(&list); err != nil || len(list) != 0 {
		t.Fatalf("after delete list=%v err=%v", list, err)
	}

	c.mustStatus(c.do(http.MethodPut, fmt.Sprintf("/api/books/%d/progress", bookID), map[string]any{
		"location": "loc-9", "percent": 0.42,
	}), http.StatusOK)
	rec = c.do(http.MethodGet, fmt.Sprintf("/api/books/%d/progress", bookID), nil)
	c.mustStatus(rec, http.StatusOK)
	var prog models.Progress
	if err := json.NewDecoder(rec.Body).Decode(&prog); err != nil {
		t.Fatal(err)
	}
	if prog.Location != "loc-9" || prog.Percent < 0.41 || prog.Percent > 0.43 {
		t.Fatalf("progress=%+v", prog)
	}
}

func TestShareDownloadServesBookBytes(t *testing.T) {
	srv, c := newAdminClient(t)
	payload := []byte("%PDF-1.4 share-bytes-unique")
	_, bookID := seedLibraryBook(t, srv, c.store, "share", payload)

	rec := c.do(http.MethodPost, fmt.Sprintf("/api/books/%d/share", bookID), map[string]any{
		"expiresInHours": 24, "maxDownloads": 2,
	})
	c.mustStatus(rec, http.StatusCreated)
	var created struct {
		ID    int64  `json:"id"`
		Token string `json:"token"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&created); err != nil || created.Token == "" {
		t.Fatalf("share=%+v err=%v", created, err)
	}

	metaReq := httptest.NewRequest(http.MethodGet, "/api/share/"+created.Token, nil)
	metaRec := httptest.NewRecorder()
	c.handler.ServeHTTP(metaRec, metaReq)
	c.mustStatus(metaRec, http.StatusOK)

	dlReq := httptest.NewRequest(http.MethodGet, "/share/"+created.Token+"/download", nil)
	dlRec := httptest.NewRecorder()
	c.handler.ServeHTTP(dlRec, dlReq)
	c.mustStatus(dlRec, http.StatusOK)
	if !bytes.Equal(dlRec.Body.Bytes(), payload) {
		t.Fatalf("download body=%q want %q", dlRec.Body.Bytes(), payload)
	}

	fileRec := c.do(http.MethodGet, fmt.Sprintf("/api/books/%d/file", bookID), nil)
	c.mustStatus(fileRec, http.StatusOK)
	if !bytes.Equal(fileRec.Body.Bytes(), payload) {
		t.Fatalf("file body mismatch")
	}
}

func TestInviteAcceptCreatesLoginableUser(t *testing.T) {
	_, c := newAdminClient(t)

	rec := c.do(http.MethodPost, "/api/invites", map[string]any{
		"kind": models.InviteKindPermanent, "email": "newuser@example.com", "expiresInHours": 24,
		"permissions": []string{"read"},
	})
	c.mustStatus(rec, http.StatusCreated)
	var inv models.InviteCreateResult
	if err := json.NewDecoder(rec.Body).Decode(&inv); err != nil || inv.Invite.Token == "" {
		t.Fatalf("invite=%+v err=%v", inv, err)
	}

	metaReq := httptest.NewRequest(http.MethodGet, "/api/invite/"+inv.Invite.Token, nil)
	metaRec := httptest.NewRecorder()
	c.handler.ServeHTTP(metaRec, metaReq)
	c.mustStatus(metaRec, http.StatusOK)
	var meta models.InviteMeta
	if err := json.NewDecoder(metaRec.Body).Decode(&meta); err != nil || !meta.Valid {
		t.Fatalf("meta=%+v err=%v", meta, err)
	}

	acceptCSRF := fetchCSRF(t, c.handler)
	body, _ := json.Marshal(map[string]string{"username": "invitee1", "password": "longpassword9"})
	acceptReq := httptest.NewRequest(http.MethodPost, "/api/invite/"+inv.Invite.Token+"/accept", bytes.NewReader(body))
	acceptReq.Header.Set("Content-Type", "application/json")
	withCSRF(acceptReq, acceptCSRF)
	acceptRec := httptest.NewRecorder()
	c.handler.ServeHTTP(acceptRec, acceptReq)
	if acceptRec.Code != http.StatusCreated && acceptRec.Code != http.StatusOK {
		t.Fatalf("accept status=%d body=%s", acceptRec.Code, acceptRec.Body.String())
	}

	u, hash, err := c.store.GetUserByUsername(context.Background(), "invitee1")
	if err != nil || u.Username != "invitee1" {
		t.Fatalf("user=%+v err=%v", u, err)
	}
	if !auth.CheckPassword(hash, "longpassword9") {
		t.Fatal("password hash does not match accepted password")
	}

	session, _, csrf := loginUser(t, c.handler, "invitee1", "longpassword9")
	meReq := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	meReq.AddCookie(session)
	meRec := httptest.NewRecorder()
	c.handler.ServeHTTP(meRec, meReq)
	if meRec.Code != http.StatusOK {
		t.Fatalf("me status=%d", meRec.Code)
	}
	_ = csrf

	metaReq = httptest.NewRequest(http.MethodGet, "/api/invite/"+inv.Invite.Token, nil)
	metaRec = httptest.NewRecorder()
	c.handler.ServeHTTP(metaRec, metaReq)
	c.mustStatus(metaRec, http.StatusOK)
	meta = models.InviteMeta{}
	if err := json.NewDecoder(metaRec.Body).Decode(&meta); err != nil || meta.Valid || meta.Reason != "accepted" {
		t.Fatalf("accepted meta=%+v", meta)
	}
}

func TestUserPermissionsPersistInStore(t *testing.T) {
	_, c := newAdminClient(t)
	ctx := context.Background()
	hash, err := auth.HashPassword("longpassword")
	if err != nil {
		t.Fatal(err)
	}
	victimID, err := c.store.CreateUser(ctx, "permuser", hash, false)
	if err != nil {
		t.Fatal(err)
	}

	c.mustStatus(c.do(http.MethodPut, fmt.Sprintf("/api/auth/users/%d/permissions", victimID), map[string]any{
		"permissions": []string{"read", "manage_library"},
	}), http.StatusOK)

	u, err := c.store.GetUser(ctx, victimID)
	if err != nil {
		t.Fatal(err)
	}
	want := models.PermRead | models.PermManageLibrary
	if u.Permissions != want {
		t.Fatalf("permissions=%d want %d", u.Permissions, want)
	}
	if models.HasPermission(u.Permissions, models.PermManageUsers) {
		t.Fatal("manage_users should not be set")
	}
}

func TestBookEditPersistsMetadata(t *testing.T) {
	srv, c := newAdminClient(t)
	_, bookID := seedLibraryBook(t, srv, c.store, "edit", []byte("%PDF-1.4 edit"))

	c.mustStatus(c.do(http.MethodPut, fmt.Sprintf("/api/books/%d", bookID), map[string]any{
		"title": "Edited Title", "author": "Edited Author", "series": "Edited Series", "seriesIndex": 3,
	}), http.StatusOK)

	rec := c.do(http.MethodGet, fmt.Sprintf("/api/books/%d", bookID), nil)
	c.mustStatus(rec, http.StatusOK)
	var book models.Book
	if err := json.NewDecoder(rec.Body).Decode(&book); err != nil {
		t.Fatal(err)
	}
	if book.Title != "Edited Title" || book.Author != "Edited Author" || book.Series != "Edited Series" {
		t.Fatalf("book=%+v", book)
	}
}

func TestUploadRejectsPathTraversal(t *testing.T) {
	srv, c := newAdminClient(t)
	libID, _ := seedLibraryBook(t, srv, c.store, "up", []byte("%PDF-1.4 up"))

	rec := c.do(http.MethodPost, fmt.Sprintf("/api/libraries/%d/uploads", libID), map[string]any{
		"relPath": "../escape.pdf", "totalSize": 12,
	})
	if rec.Code == http.StatusCreated || rec.Code == http.StatusOK {
		t.Fatalf("traversal should fail, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestNonAdminDeniedAdminRoutes(t *testing.T) {
	srv, store := testServer(t)
	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	adminHash, _ := auth.HashPassword("longpassword")
	if _, err := store.CreateUser(ctx, "admin", adminHash, true); err != nil {
		t.Fatal(err)
	}
	userHash, _ := auth.HashPassword("longpassword")
	if _, err := store.CreateUser(ctx, "reader", userHash, false); err != nil {
		t.Fatal(err)
	}
	session, _, csrf := loginUser(t, handler, "reader", "longpassword")

	req := httptest.NewRequest(http.MethodGet, "/api/admin/server", nil)
	req.AddCookie(session)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden && rec.Code != http.StatusUnauthorized {
		t.Fatalf("non-admin admin route status=%d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/api/invites", bytes.NewBufferString(`{"kind":"permanent"}`))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(session)
	withCSRF(req, csrf)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden && rec.Code != http.StatusUnauthorized {
		t.Fatalf("non-admin invite status=%d", rec.Code)
	}
}

func TestWebhookTestDeliversSignedPayload(t *testing.T) {
	_, c := newAdminClient(t)
	webhookAllowLocal = true
	t.Cleanup(func() { webhookAllowLocal = false })
	var gotBody []byte
	var gotSig string
	hook := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		gotSig = r.Header.Get("X-Athenaeum-Signature")
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(hook.Close)

	rec := c.do(http.MethodPost, "/api/admin/webhooks", map[string]any{
		"url": hook.URL, "secret": "hook-secret", "events": []string{"user.create"},
	})
	c.mustStatus(rec, http.StatusCreated)
	var wh models.WebhookPublic
	if err := json.NewDecoder(rec.Body).Decode(&wh); err != nil {
		t.Fatal(err)
	}

	c.mustStatus(c.do(http.MethodPost, fmt.Sprintf("/api/admin/webhooks/%d/test", wh.ID), nil), http.StatusOK)
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) && len(gotBody) == 0 {
		time.Sleep(20 * time.Millisecond)
	}
	if len(gotBody) == 0 {
		t.Fatal("webhook never delivered")
	}
	if gotSig == "" {
		t.Fatal("missing webhook signature header")
	}
	if !VerifyWebhookSignature("hook-secret", gotSig, gotBody) {
		t.Fatalf("signature invalid body=%s sig=%s", gotBody, gotSig)
	}
}

func TestTOTPChallengeRequiredAfterEnable(t *testing.T) {
	_, c := newAdminClient(t)

	rec := c.do(http.MethodPost, "/api/auth/totp/setup", nil)
	c.mustStatus(rec, http.StatusOK)
	var setup totpSetupResponse
	if err := json.NewDecoder(rec.Body).Decode(&setup); err != nil || setup.Secret == "" {
		t.Fatalf("setup=%+v err=%v", setup, err)
	}
	code, err := totp.GenerateCode(setup.Secret, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	c.mustStatus(c.do(http.MethodPost, "/api/auth/totp/enable", map[string]any{"code": code}), http.StatusOK)

	logout := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	logout.AddCookie(c.session)
	withCSRF(logout, c.csrf)
	logoutRec := httptest.NewRecorder()
	c.handler.ServeHTTP(logoutRec, logout)

	loginCSRF := fetchCSRF(t, c.handler)
	loginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(`{"username":"admin","password":"longpassword"}`))
	loginReq.Header.Set("Content-Type", "application/json")
	withCSRF(loginReq, loginCSRF)
	loginRec := httptest.NewRecorder()
	c.handler.ServeHTTP(loginRec, loginReq)
	c.mustStatus(loginRec, http.StatusOK)
	var challenge models.LoginChallenge
	if err := json.NewDecoder(loginRec.Body).Decode(&challenge); err != nil {
		t.Fatal(err)
	}
	if !challenge.NeedsTOTP || challenge.TOTPToken == "" {
		t.Fatalf("expected totp challenge got %+v", challenge)
	}

	badCSRF := fetchCSRF(t, c.handler)
	badBody, _ := json.Marshal(map[string]string{"totpToken": challenge.TOTPToken, "code": "000000"})
	badReq := httptest.NewRequest(http.MethodPost, "/api/auth/totp/verify", bytes.NewReader(badBody))
	badReq.Header.Set("Content-Type", "application/json")
	withCSRF(badReq, badCSRF)
	badRec := httptest.NewRecorder()
	c.handler.ServeHTTP(badRec, badReq)
	if badRec.Code == http.StatusOK {
		t.Fatal("wrong totp code should not succeed")
	}

	goodCode, err := totp.GenerateCode(setup.Secret, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	goodCSRF := fetchCSRF(t, c.handler)
	goodBody, _ := json.Marshal(map[string]string{"totpToken": challenge.TOTPToken, "code": goodCode})
	goodReq := httptest.NewRequest(http.MethodPost, "/api/auth/totp/verify", bytes.NewReader(goodBody))
	goodReq.Header.Set("Content-Type", "application/json")
	withCSRF(goodReq, goodCSRF)
	goodRec := httptest.NewRecorder()
	c.handler.ServeHTTP(goodRec, goodReq)
	c.mustStatus(goodRec, http.StatusOK)
	gotSession := false
	for _, cookie := range goodRec.Result().Cookies() {
		if cookie.Name == auth.SessionCookie && cookie.Value != "" {
			gotSession = true
		}
	}
	if !gotSession {
		t.Fatal("expected session cookie after totp verify")
	}
}

func TestKosyncProgressRoundTrip(t *testing.T) {
	_, c := newAdminClient(t)

	putBody, _ := json.Marshal(map[string]any{
		"document": "doc-behavior", "progress": "33", "percentage": 0.33, "device": "test", "device_id": "d1",
	})
	putReq := httptest.NewRequest(http.MethodPut, "/kosync/syncs/progress", bytes.NewReader(putBody))
	putReq.Header.Set("Content-Type", "application/json")
	putReq.SetBasicAuth("admin", "longpassword")
	putRec := httptest.NewRecorder()
	c.handler.ServeHTTP(putRec, putReq)
	c.mustStatus(putRec, http.StatusOK)

	getReq := httptest.NewRequest(http.MethodGet, "/kosync/syncs/progress/doc-behavior", nil)
	getReq.SetBasicAuth("admin", "longpassword")
	getRec := httptest.NewRecorder()
	c.handler.ServeHTTP(getRec, getReq)
	c.mustStatus(getRec, http.StatusOK)
	var got map[string]any
	if err := json.NewDecoder(getRec.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got["document"] != "doc-behavior" {
		t.Fatalf("got=%v", got)
	}
	pct, _ := got["percentage"].(float64)
	if pct < 0.32 || pct > 0.34 {
		t.Fatalf("percentage=%v", got["percentage"])
	}
}
