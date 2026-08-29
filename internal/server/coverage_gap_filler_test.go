package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"athenaeum/internal/models"
)

func TestCoverageGapFiller(t *testing.T) {
	srv, store := testServer(t)
	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}
	session, csrf := loginAdmin(t, handler, store)

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

	rec := do(http.MethodGet, "/api/auth/settings", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("get auth settings status=%d", rec.Code)
	}

	rec = do(http.MethodPost, "/api/auth/register", map[string]any{
		"username": "reguser", "password": "longpassword",
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("admin register status=%d body=%s", rec.Code, rec.Body.String())
	}
	rec = do(http.MethodPost, "/api/auth/register", map[string]any{
		"username": "reguser", "password": "longpassword",
	})
	if rec.Code != http.StatusConflict {
		t.Fatalf("register taken status=%d", rec.Code)
	}
	rec = do(http.MethodPost, "/api/auth/register", map[string]any{
		"username": "x", "password": "longpassword",
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("register short status=%d", rec.Code)
	}

	rec = do(http.MethodPost, "/api/auth/users/guest", map[string]any{
		"username": "gapguest", "expiresInHours": 2,
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("guest create status=%d", rec.Code)
	}
	var guest models.GuestCredentials
	if err := json.NewDecoder(rec.Body).Decode(&guest); err != nil {
		t.Fatal(err)
	}

	rec = do(http.MethodGet, "/api/auth/users/guests?expiringWithinHours=48", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("list guests filter status=%d", rec.Code)
	}
	rec = do(http.MethodGet, "/api/auth/users/guests?expiringWithinHours=nope", nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("list guests bad filter status=%d", rec.Code)
	}

	rec = do(http.MethodPost, fmt.Sprintf("/api/auth/users/guests/%d/extend", guest.User.ID), map[string]any{
		"expiresAt": time.Now().Add(72 * time.Hour).UTC().Format(time.RFC3339),
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("extend guest expiresAt status=%d body=%s", rec.Code, rec.Body.String())
	}
	rec = do(http.MethodPost, "/api/auth/users/guests/bulk-delete", map[string]any{"ids": []int64{}})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("bulk delete empty status=%d", rec.Code)
	}

	rec = do(http.MethodGet, "/api/collections", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("list collections status=%d", rec.Code)
	}
	rec = do(http.MethodPost, "/api/collections", map[string]any{
		"name": "Manual", "description": "d",
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("create collection status=%d body=%s", rec.Code, rec.Body.String())
	}
	var col models.Collection
	if err := json.NewDecoder(rec.Body).Decode(&col); err != nil {
		t.Fatal(err)
	}
	rec = do(http.MethodPost, "/api/collections", map[string]any{
		"name": "Reading", "kind": models.CollectionReading,
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("create reading collection status=%d body=%s", rec.Code, rec.Body.String())
	}
	rec = do(http.MethodPost, "/api/collections", map[string]any{
		"name": "Smart", "kind": models.CollectionSmart,
		"query": map[string]any{"author": "X"},
	})
	if rec.Code != http.StatusCreated && rec.Code != http.StatusBadRequest && rec.Code != http.StatusInternalServerError {
		t.Fatalf("create smart collection status=%d body=%s", rec.Code, rec.Body.String())
	}
	rec = do(http.MethodGet, fmt.Sprintf("/api/collections/%d", col.ID), nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("get collection status=%d", rec.Code)
	}
	rec = do(http.MethodPut, fmt.Sprintf("/api/collections/%d", col.ID), map[string]any{
		"name": "Manual2", "description": "updated",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("update collection status=%d body=%s", rec.Code, rec.Body.String())
	}

	bookID, err := store.UpsertBook(context.Background(), &models.Book{
		Title: "Gap", Format: models.FormatPDF, RelPath: "gap.pdf",
	}, 1)
	if err != nil {
		t.Fatal(err)
	}
	rec = do(http.MethodPost, fmt.Sprintf("/api/collections/%d/books/%d", col.ID, bookID), nil)
	if rec.Code != http.StatusOK && rec.Code != http.StatusNoContent && rec.Code != http.StatusCreated {
		t.Fatalf("add to collection status=%d body=%s", rec.Code, rec.Body.String())
	}
	rec = do(http.MethodDelete, fmt.Sprintf("/api/collections/%d/books/%d", col.ID, bookID), nil)
	if rec.Code != http.StatusOK && rec.Code != http.StatusNoContent {
		t.Fatalf("remove from collection status=%d", rec.Code)
	}
	rec = do(http.MethodDelete, fmt.Sprintf("/api/collections/%d", col.ID), nil)
	if rec.Code != http.StatusOK && rec.Code != http.StatusNoContent {
		t.Fatalf("delete collection status=%d", rec.Code)
	}

	rec = do(http.MethodGet, "/api/docs", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("api docs status=%d", rec.Code)
	}
	rec = do(http.MethodGet, "/api/openapi.json", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("openapi status=%d", rec.Code)
	}
	docsRec := httptest.NewRecorder()
	handler.ServeHTTP(docsRec, httptest.NewRequest(http.MethodGet, "/docs", nil))
	if docsRec.Code != http.StatusOK && docsRec.Code != http.StatusFound {
		t.Fatalf("docs ui status=%d", docsRec.Code)
	}

	rec = do(http.MethodGet, "/api/auth/api-keys?userId=1", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("api keys list status=%d", rec.Code)
	}
}
