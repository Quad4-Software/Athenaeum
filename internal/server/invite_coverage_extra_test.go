package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"athenaeum/internal/models"
	"athenaeum/internal/pocketid"
)

func mockPocketIDProvision(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-API-KEY") == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		if r.Method == http.MethodPost {
			var body pocketid.UserCreate
			_ = json.NewDecoder(r.Body).Decode(&body)
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(pocketid.User{
				ID: "pid-user-1", Username: body.Username, Email: body.Email, DisplayName: body.DisplayName,
			})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": []any{}})
	})
	mux.HandleFunc("/api/users/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		switch {
		case strings.HasSuffix(path, "/user-groups") && r.Method == http.MethodPut:
			w.WriteHeader(http.StatusNoContent)
		case strings.HasSuffix(path, "/one-time-access-token") && r.Method == http.MethodPost:
			_ = json.NewEncoder(w).Encode(map[string]string{"token": "ota-token-xyz"})
		case strings.HasSuffix(path, "/one-time-access-email") && r.Method == http.MethodPost:
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func TestInviteCoverageExtra(t *testing.T) {
	srv, store := testServer(t)
	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}
	session, csrf := loginAdmin(t, handler, store)
	ctx := context.Background()

	do := func(method, path string, body any) *httptest.ResponseRecorder {
		t.Helper()
		var rdr *bytes.Reader
		if body != nil {
			b, err := json.Marshal(body)
			if err != nil {
				t.Fatal(err)
			}
			rdr = bytes.NewReader(b)
		} else {
			rdr = bytes.NewReader(nil)
		}
		req := httptest.NewRequest(method, path, rdr)
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		req.AddCookie(session)
		if method != http.MethodGet && method != http.MethodHead {
			withCSRF(req, csrf)
		}
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		return rec
	}

	rec := do(http.MethodPost, "/api/invites", map[string]any{
		"kind": models.InviteKindGuest, "expiresInHours": 48, "guestExpiresInHours": 12,
		"permissions": []string{"read"},
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("create guest invite status=%d body=%s", rec.Code, rec.Body.String())
	}
	var guestInv models.InviteCreateResult
	if err := json.NewDecoder(rec.Body).Decode(&guestInv); err != nil {
		t.Fatal(err)
	}

	acceptCSRF := fetchCSRF(t, handler)
	acceptReq := httptest.NewRequest(http.MethodPost, "/api/invite/"+guestInv.Invite.Token+"/accept",
		bytes.NewBufferString(`{}`))
	acceptReq.Header.Set("Content-Type", "application/json")
	withCSRF(acceptReq, acceptCSRF)
	acceptRec := httptest.NewRecorder()
	handler.ServeHTTP(acceptRec, acceptReq)
	if acceptRec.Code != http.StatusCreated {
		t.Fatalf("accept guest no user status=%d body=%s", acceptRec.Code, acceptRec.Body.String())
	}
	var guestCred models.GuestCredentials
	if err := json.NewDecoder(acceptRec.Body).Decode(&guestCred); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(guestCred.User.Username, "guest-") || guestCred.Password == "" {
		t.Fatalf("guest creds=%+v", guestCred)
	}

	rec = do(http.MethodPost, "/api/invites", map[string]any{
		"kind": models.InviteKindGuest, "guestExpiresInHours": 6,
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("create guest invite2 status=%d", rec.Code)
	}
	var guestInv2 models.InviteCreateResult
	if err := json.NewDecoder(rec.Body).Decode(&guestInv2); err != nil {
		t.Fatal(err)
	}
	acceptCSRF = fetchCSRF(t, handler)
	acceptReq = httptest.NewRequest(http.MethodPost, "/api/invite/"+guestInv2.Invite.Token+"/accept",
		bytes.NewBufferString(`{"username":"namedguest"}`))
	acceptReq.Header.Set("Content-Type", "application/json")
	withCSRF(acceptReq, acceptCSRF)
	acceptRec = httptest.NewRecorder()
	handler.ServeHTTP(acceptRec, acceptReq)
	if acceptRec.Code != http.StatusCreated {
		t.Fatalf("accept guest named status=%d body=%s", acceptRec.Code, acceptRec.Body.String())
	}

	mock := mockPocketIDProvision(t)
	if err := store.SavePocketIDSettings(ctx, models.PocketIDSettings{
		Enabled: true, BaseURL: mock.URL, APIKey: "test-key", DefaultGroupIDs: []string{"g1"},
	}); err != nil {
		t.Fatal(err)
	}
	rec = do(http.MethodPost, "/api/invites", map[string]any{
		"kind": models.InviteKindPermanent, "email": "sso@example.com",
		"username": "ssoUser", "provisionPocketId": true, "expiresInHours": 24,
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("pocket provision status=%d body=%s", rec.Code, rec.Body.String())
	}
	var pocketInv models.InviteCreateResult
	if err := json.NewDecoder(rec.Body).Decode(&pocketInv); err != nil {
		t.Fatal(err)
	}
	if pocketInv.Invite.PocketIDUserID == "" || pocketInv.PocketIDSetupURL == "" {
		t.Fatalf("pocket result=%+v", pocketInv)
	}

	acceptCSRF = fetchCSRF(t, handler)
	acceptReq = httptest.NewRequest(http.MethodPost, "/api/invite/"+pocketInv.Invite.Token+"/accept",
		bytes.NewBufferString(`{}`))
	acceptReq.Header.Set("Content-Type", "application/json")
	withCSRF(acceptReq, acceptCSRF)
	acceptRec = httptest.NewRecorder()
	handler.ServeHTTP(acceptRec, acceptReq)
	if acceptRec.Code != http.StatusOK {
		t.Fatalf("accept pocket sso status=%d body=%s", acceptRec.Code, acceptRec.Body.String())
	}
	var ssoOut map[string]string
	if err := json.NewDecoder(acceptRec.Body).Decode(&ssoOut); err != nil {
		t.Fatal(err)
	}
	if ssoOut["redirect"] != "/auth/oidc/login" {
		t.Fatalf("sso out=%v", ssoOut)
	}

	rec = do(http.MethodPost, "/api/invites", map[string]any{
		"kind": models.InviteKindPermanent, "email": "localpart@example.com",
		"provisionPocketId": true,
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("pocket username from email status=%d body=%s", rec.Code, rec.Body.String())
	}

	rec = do(http.MethodPost, "/api/invites", map[string]any{"kind": "nope"})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("bad kind status=%d", rec.Code)
	}
	rec = do(http.MethodPost, "/api/invites", map[string]any{"expiresInHours": 9000})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expires too large status=%d", rec.Code)
	}
	rec = do(http.MethodPost, "/api/invites", map[string]any{
		"kind": models.InviteKindGuest, "provisionPocketId": true, "email": "g@example.com",
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("guest pocket status=%d", rec.Code)
	}
	rec = do(http.MethodPost, "/api/invites", map[string]any{
		"kind": models.InviteKindPermanent, "provisionPocketId": true,
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("pocket without email status=%d", rec.Code)
	}
	rec = do(http.MethodPost, "/api/invites", map[string]any{
		"kind": models.InviteKindGuest, "guestExpiresInHours": 9000,
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("guest expires too large status=%d", rec.Code)
	}

	expires := time.Now().Add(time.Hour)
	revoked, err := store.CreateInvite(ctx, models.Invite{
		Kind: models.InviteKindPermanent, CreatedBy: 1, ExpiresAt: &expires,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := store.RevokeInvite(ctx, revoked.ID); err != nil {
		t.Fatal(err)
	}
	metaRec := httptest.NewRecorder()
	handler.ServeHTTP(metaRec, httptest.NewRequest(http.MethodGet, "/api/invite/"+revoked.Token, nil))
	if metaRec.Code != http.StatusOK {
		t.Fatalf("meta revoked status=%d", metaRec.Code)
	}
	var meta models.InviteMeta
	if err := json.NewDecoder(metaRec.Body).Decode(&meta); err != nil {
		t.Fatal(err)
	}
	if meta.Valid || meta.Reason != "revoked" {
		t.Fatalf("meta revoked=%+v", meta)
	}

	acceptCSRF = fetchCSRF(t, handler)
	acceptReq = httptest.NewRequest(http.MethodPost, "/api/invite/"+revoked.Token+"/accept",
		bytes.NewBufferString(`{"username":"x","password":"longpassword"}`))
	acceptReq.Header.Set("Content-Type", "application/json")
	withCSRF(acceptReq, acceptCSRF)
	acceptRec = httptest.NewRecorder()
	handler.ServeHTTP(acceptRec, acceptReq)
	if acceptRec.Code != http.StatusGone {
		t.Fatalf("accept revoked status=%d", acceptRec.Code)
	}

	accepted, err := store.CreateInvite(ctx, models.Invite{
		Kind: models.InviteKindPermanent, CreatedBy: 1, ExpiresAt: &expires,
	})
	if err != nil {
		t.Fatal(err)
	}
	uid, err := store.CreateInvitedUser(ctx, "alreadyaccepted", "hash", "a@example.com", models.DefaultUserPermissions)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.AcceptInvite(ctx, accepted.ID, uid); err != nil {
		t.Fatal(err)
	}
	metaRec = httptest.NewRecorder()
	handler.ServeHTTP(metaRec, httptest.NewRequest(http.MethodGet, "/api/invite/"+accepted.Token, nil))
	_ = json.NewDecoder(metaRec.Body).Decode(&meta)
	if meta.Valid || meta.Reason != "accepted" {
		t.Fatalf("meta accepted=%+v", meta)
	}
	acceptCSRF = fetchCSRF(t, handler)
	acceptReq = httptest.NewRequest(http.MethodPost, "/api/invite/"+accepted.Token+"/accept",
		bytes.NewBufferString(`{"username":"y","password":"longpassword"}`))
	acceptReq.Header.Set("Content-Type", "application/json")
	withCSRF(acceptReq, acceptCSRF)
	acceptRec = httptest.NewRecorder()
	handler.ServeHTTP(acceptRec, acceptReq)
	if acceptRec.Code != http.StatusConflict {
		t.Fatalf("accept already status=%d", acceptRec.Code)
	}

	past := time.Now().Add(-time.Hour)
	expired, err := store.CreateInvite(ctx, models.Invite{
		Kind: models.InviteKindPermanent, CreatedBy: 1, ExpiresAt: &past,
	})
	if err != nil {
		t.Fatal(err)
	}
	metaRec = httptest.NewRecorder()
	handler.ServeHTTP(metaRec, httptest.NewRequest(http.MethodGet, "/api/invite/"+expired.Token, nil))
	_ = json.NewDecoder(metaRec.Body).Decode(&meta)
	if meta.Valid || meta.Reason != "expired" {
		t.Fatalf("meta expired=%+v", meta)
	}
	acceptCSRF = fetchCSRF(t, handler)
	acceptReq = httptest.NewRequest(http.MethodPost, "/api/invite/"+expired.Token+"/accept",
		bytes.NewBufferString(`{"username":"z","password":"longpassword"}`))
	acceptReq.Header.Set("Content-Type", "application/json")
	withCSRF(acceptReq, acceptCSRF)
	acceptRec = httptest.NewRecorder()
	handler.ServeHTTP(acceptRec, acceptReq)
	if acceptRec.Code != http.StatusGone {
		t.Fatalf("accept expired status=%d", acceptRec.Code)
	}

	metaRec = httptest.NewRecorder()
	handler.ServeHTTP(metaRec, httptest.NewRequest(http.MethodGet, "/api/invite/no-such-token", nil))
	_ = json.NewDecoder(metaRec.Body).Decode(&meta)
	if meta.Valid || meta.Reason != "not_found" {
		t.Fatalf("meta not_found=%+v", meta)
	}
	acceptCSRF = fetchCSRF(t, handler)
	acceptReq = httptest.NewRequest(http.MethodPost, "/api/invite/no-such-token/accept",
		bytes.NewBufferString(`{}`))
	acceptReq.Header.Set("Content-Type", "application/json")
	withCSRF(acceptReq, acceptCSRF)
	acceptRec = httptest.NewRecorder()
	handler.ServeHTTP(acceptRec, acceptReq)
	if acceptRec.Code != http.StatusNotFound {
		t.Fatalf("accept not found status=%d", acceptRec.Code)
	}

	for _, status := range []string{"pending", "revoked", "accepted", "expired"} {
		rec = do(http.MethodGet, "/api/invites?status="+status, nil)
		if rec.Code != http.StatusOK {
			t.Fatalf("list status=%q code=%d", status, rec.Code)
		}
	}

	rec = do(http.MethodDelete, fmt.Sprintf("/api/invites/%d", 999999), nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("revoke missing status=%d", rec.Code)
	}
	rec = do(http.MethodDelete, "/api/invites/bad", nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("revoke bad id status=%d", rec.Code)
	}
}
