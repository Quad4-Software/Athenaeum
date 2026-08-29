package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"athenaeum/internal/auth"
	"athenaeum/internal/models"
	"athenaeum/internal/storage"
)

func TestAuthUsersCoverage(t *testing.T) {
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

	rec := do(http.MethodPost, "/api/auth/users/guest", map[string]any{
		"expiresInHours": 24,
		"permissions":    []string{"read", "edit_metadata"},
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("create guest auto name status=%d body=%s", rec.Code, rec.Body.String())
	}
	var autoGuest models.GuestCredentials
	if err := json.NewDecoder(rec.Body).Decode(&autoGuest); err != nil {
		t.Fatal(err)
	}
	if autoGuest.Password == "" || autoGuest.User.ID == 0 {
		t.Fatalf("auto guest=%+v", autoGuest)
	}

	rec = do(http.MethodPost, "/api/auth/users/guest", map[string]any{
		"username": "covguest", "expiresInHours": 48,
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("create guest named status=%d body=%s", rec.Code, rec.Body.String())
	}
	var namedGuest models.GuestCredentials
	if err := json.NewDecoder(rec.Body).Decode(&namedGuest); err != nil {
		t.Fatal(err)
	}

	rec = do(http.MethodPost, "/api/auth/users/guest", map[string]any{
		"username": "covguest", "expiresInHours": 1,
	})
	if rec.Code != http.StatusConflict {
		t.Fatalf("duplicate guest status=%d", rec.Code)
	}
	rec = do(http.MethodPost, "/api/auth/users/guest", map[string]any{
		"username": "x", "expiresInHours": 1,
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("short guest name status=%d", rec.Code)
	}
	rec = do(http.MethodPost, "/api/auth/users/guest", map[string]any{
		"expiresInHours": 9000,
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("guest expires too large status=%d", rec.Code)
	}

	hash, err := auth.HashPassword("longpassword")
	if err != nil {
		t.Fatal(err)
	}
	victimID, err := store.CreateUser(ctx, "permuser", hash, false)
	if err != nil {
		t.Fatal(err)
	}
	adminMe, _, err := store.GetUserByUsername(ctx, "admin")
	if err != nil {
		t.Fatal(err)
	}

	rec = do(http.MethodPut, fmt.Sprintf("/api/auth/users/%d/permissions", victimID), map[string]any{
		"permissions": []string{"read", "edit_metadata"},
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("set perms status=%d body=%s", rec.Code, rec.Body.String())
	}
	victim, err := store.GetUser(ctx, victimID)
	if err != nil {
		t.Fatal(err)
	}
	wantPerm := models.PermRead | models.PermEditMetadata
	if victim.Permissions != wantPerm {
		t.Fatalf("store permissions=%d want %d", victim.Permissions, wantPerm)
	}

	rec = do(http.MethodPut, fmt.Sprintf("/api/auth/users/%d/permissions", adminMe.ID), map[string]any{
		"permissions": []string{"read"},
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("admin perms noop status=%d", rec.Code)
	}

	rec = do(http.MethodPut, "/api/auth/users/999999/permissions", map[string]any{
		"permissions": []string{"read"},
	})
	if rec.Code != http.StatusNotFound {
		t.Fatalf("perms missing status=%d", rec.Code)
	}

	rec = do(http.MethodPut, fmt.Sprintf("/api/auth/users/%d/password", victimID), map[string]any{
		"password": "newlongpassword",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("reset password status=%d body=%s", rec.Code, rec.Body.String())
	}
	_, hashAfter, err := store.GetUserByUsername(ctx, "permuser")
	if err != nil {
		t.Fatal(err)
	}
	if !auth.CheckPassword(hashAfter, "newlongpassword") {
		t.Fatal("reset password did not update hash")
	}
	if auth.CheckPassword(hashAfter, "longpassword") {
		t.Fatal("old password should no longer match")
	}
	rec = do(http.MethodPut, fmt.Sprintf("/api/auth/users/%d/password", victimID), map[string]any{
		"password": "short",
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("weak reset password status=%d", rec.Code)
	}
	rec = do(http.MethodPut, "/api/auth/users/999999/password", map[string]any{
		"password": "longpassword",
	})
	if rec.Code != http.StatusNotFound {
		t.Fatalf("reset missing status=%d", rec.Code)
	}

	rec = do(http.MethodPut, fmt.Sprintf("/api/auth/users/%d/admin", victimID), map[string]any{"admin": true})
	if rec.Code != http.StatusOK {
		t.Fatalf("grant admin status=%d", rec.Code)
	}
	victim, err = store.GetUser(ctx, victimID)
	if err != nil || !victim.IsAdmin {
		t.Fatalf("expected admin after grant: %+v err=%v", victim, err)
	}
	rec = do(http.MethodPut, fmt.Sprintf("/api/auth/users/%d/admin", victimID), map[string]any{"admin": false})
	if rec.Code != http.StatusOK {
		t.Fatalf("revoke admin status=%d", rec.Code)
	}
	victim, err = store.GetUser(ctx, victimID)
	if err != nil || victim.IsAdmin {
		t.Fatalf("expected non-admin after revoke: %+v err=%v", victim, err)
	}
	rec = do(http.MethodPut, fmt.Sprintf("/api/auth/users/%d/admin", adminMe.ID), map[string]any{"admin": false})
	if rec.Code != http.StatusConflict {
		t.Fatalf("last admin revoke status=%d body=%s", rec.Code, rec.Body.String())
	}
	rec = do(http.MethodPut, "/api/auth/users/999999/admin", map[string]any{"admin": true})
	if rec.Code != http.StatusNotFound {
		t.Fatalf("admin missing status=%d", rec.Code)
	}

	rec = do(http.MethodDelete, fmt.Sprintf("/api/auth/users/%d", adminMe.ID), nil)
	if rec.Code != http.StatusConflict {
		t.Fatalf("delete self status=%d", rec.Code)
	}
	rec = do(http.MethodDelete, "/api/auth/users/999999", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("delete missing status=%d", rec.Code)
	}
	rec = do(http.MethodDelete, fmt.Sprintf("/api/auth/users/%d", victimID), nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("delete victim status=%d body=%s", rec.Code, rec.Body.String())
	}
	if _, err := store.GetUser(ctx, victimID); err != storage.ErrNotFound {
		t.Fatalf("deleted user still present: %v", err)
	}
	rec = do(http.MethodDelete, fmt.Sprintf("/api/auth/users/%d", namedGuest.User.ID), nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("delete guest status=%d", rec.Code)
	}

	rec = do(http.MethodGet, "/api/auth/users", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("list users status=%d", rec.Code)
	}

	rec = do(http.MethodGet, "/api/auth/audit?limit=10&offset=0&action=user.guest&q=cov", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("list audit status=%d body=%s", rec.Code, rec.Body.String())
	}
}
