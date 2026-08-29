package pocketid

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCreateUserAndOTA(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/users", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-API-KEY") != "test-key" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		_ = json.NewEncoder(w).Encode(User{ID: "u1", Username: "alice", Email: "a@example.com"})
	})
	mux.HandleFunc("POST /api/users/u1/one-time-access-token", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"token": "abc123token"})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := NewClient(srv.URL, "test-key")
	u, err := c.CreateUser(context.Background(), UserCreate{
		Username:    "alice",
		Email:       "a@example.com",
		FirstName:   "Alice",
		DisplayName: "Alice",
	})
	if err != nil {
		t.Fatal(err)
	}
	if u.ID != "u1" {
		t.Fatalf("id = %q", u.ID)
	}
	tok, err := c.CreateOneTimeAccessToken(context.Background(), u.ID, "24h")
	if err != nil {
		t.Fatal(err)
	}
	if tok != "abc123token" {
		t.Fatalf("token = %q", tok)
	}
	setup := c.SetupURL(tok)
	want := srv.URL + "/lc/abc123token"
	if setup != want {
		t.Fatalf("setup url = %q want %q", setup, want)
	}
}

func TestErrorMapping(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "user already exists", http.StatusConflict)
	}))
	defer srv.Close()
	c := NewClient(srv.URL, "key")
	_, err := c.CreateUser(context.Background(), UserCreate{Username: "x", FirstName: "X", DisplayName: "X"})
	if err == nil || !strings.Contains(err.Error(), "user already exists") {
		t.Fatalf("err = %v", err)
	}
}

func TestClientAdminAPISurface(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("PUT /api/users/u1/user-groups", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("POST /api/users/u1/one-time-access-email", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("POST /api/signup-tokens", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(SignupToken{ID: "st1", Token: "tok", UsageLimit: 1})
	})
	mux.HandleFunc("GET /api/signup-tokens", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []SignupToken{{ID: "st1", Token: "tok"}},
		})
	})
	mux.HandleFunc("DELETE /api/signup-tokens/st1", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("GET /api/users", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":[]}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := NewClient(srv.URL, "test-key")
	ctx := context.Background()
	if err := c.UpdateUserGroups(ctx, "u1", []string{"g1"}); err != nil {
		t.Fatal(err)
	}
	if err := c.RequestOneTimeAccessEmail(ctx, "u1", "1h"); err != nil {
		t.Fatal(err)
	}
	tok, err := c.CreateSignupToken(ctx, "24h", 1, []string{"g1"})
	if err != nil {
		t.Fatal(err)
	}
	if tok.ID != "st1" {
		t.Fatalf("signup token %+v", tok)
	}
	list, err := c.ListSignupTokens(ctx)
	if err != nil || len(list) != 1 {
		t.Fatalf("list=%v err=%v", list, err)
	}
	if err := c.DeleteSignupToken(ctx, "st1"); err != nil {
		t.Fatal(err)
	}
	if err := c.ListUsers(ctx); err != nil {
		t.Fatal(err)
	}
}

func TestClientMissingConfig(t *testing.T) {
	c := NewClient("", "")
	if err := c.ListUsers(context.Background()); err == nil {
		t.Fatal("expected base url error")
	}
	c = NewClient("http://example.invalid", "")
	if err := c.ListUsers(context.Background()); err == nil {
		t.Fatal("expected api key error")
	}
}
