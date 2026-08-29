package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"athenaeum/internal/auth"
	"athenaeum/internal/models"
)

func TestInviteAcceptPermanent(t *testing.T) {
	srv, store := testServer(t)
	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	hash, err := auth.HashPassword("longpassword")
	if err != nil {
		t.Fatal(err)
	}
	adminID, err := store.CreateUser(ctx, "admin", hash, true)
	if err != nil {
		t.Fatal(err)
	}
	expires := time.Now().Add(time.Hour)
	inv, err := store.CreateInvite(ctx, models.Invite{
		Kind:        models.InviteKindPermanent,
		Email:       "new@example.com",
		Permissions: models.DefaultUserPermissions,
		CreatedBy:   adminID,
		ExpiresAt:   &expires,
	})
	if err != nil {
		t.Fatal(err)
	}

	csrf := fetchCSRF(t, handler)
	body, _ := json.Marshal(map[string]string{
		"username": "newbie",
		"password": "longpassword",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/invite/"+inv.Token+"/accept", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	withCSRF(req, csrf)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	u, _, err := store.GetUserByUsername(ctx, "newbie")
	if err != nil {
		t.Fatal(err)
	}
	if u.Email != "new@example.com" {
		t.Fatalf("email=%q", u.Email)
	}
}

func TestInviteRejectExpired(t *testing.T) {
	srv, store := testServer(t)
	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	hash, err := auth.HashPassword("longpassword")
	if err != nil {
		t.Fatal(err)
	}
	adminID, err := store.CreateUser(ctx, "admin", hash, true)
	if err != nil {
		t.Fatal(err)
	}
	expires := time.Now().Add(-time.Hour)
	inv, err := store.CreateInvite(ctx, models.Invite{
		Kind:      models.InviteKindPermanent,
		CreatedBy: adminID,
		ExpiresAt: &expires,
	})
	if err != nil {
		t.Fatal(err)
	}
	csrf := fetchCSRF(t, handler)
	body, _ := json.Marshal(map[string]string{"username": "x", "password": "longpassword"})
	req := httptest.NewRequest(http.MethodPost, "/api/invite/"+inv.Token+"/accept", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	withCSRF(req, csrf)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusGone {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
}
