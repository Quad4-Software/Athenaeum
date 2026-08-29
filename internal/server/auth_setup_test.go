package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"athenaeum/internal/auth"
	"athenaeum/internal/models"
)

func TestAuthSetupAndInProgress(t *testing.T) {
	srv, store := testServer(t)
	ctx := context.Background()
	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/auth/setup", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("setup status=%d", rec.Code)
	}
	var setup map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&setup); err != nil {
		t.Fatal(err)
	}
	if setup["needed"] != true || setup["authEnabled"] != false {
		t.Fatalf("setup=%v", setup)
	}
	if _, ok := setup["passwordPolicy"].(map[string]any); !ok {
		t.Fatalf("expected passwordPolicy object, got %T", setup["passwordPolicy"])
	}

	csrf := fetchCSRF(t, handler)
	body := bytes.NewBufferString(`{"username":"admin","password":"longpassword1"}`)
	req = httptest.NewRequest(http.MethodPost, "/api/auth/setup", body)
	req.Header.Set("Content-Type", "application/json")
	withCSRF(req, csrf)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("setup post status=%d body=%s", rec.Code, rec.Body.String())
	}

	id, _ := store.UpsertBook(ctx, &models.Book{Title: "A", Format: models.FormatEPUB, RelPath: "a.epub"}, 1)
	hash, _ := auth.HashPassword("pass")
	uid, _ := store.CreateUser(ctx, "reader", hash, false)
	_ = store.SaveProgress(ctx, uid, models.Progress{BookID: id, Location: "x", Percent: 0.5})

	req = httptest.NewRequest(http.MethodGet, "/api/books?inProgress=1", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without login, got %d", rec.Code)
	}
}
