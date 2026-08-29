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
