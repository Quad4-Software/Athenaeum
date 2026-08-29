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
)

func TestPutBookMetadata(t *testing.T) {
	srv, store := testServer(t)
	ctx := context.Background()
	id, err := store.UpsertBook(ctx, &models.Book{Title: "Old", Author: "A", Format: models.FormatEPUB, RelPath: "a.epub"}, 1)
	if err != nil {
		t.Fatal(err)
	}
	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}
	csrf := fetchCSRF(t, handler)
	body, _ := json.Marshal(models.BookUpdate{Title: "New Title", Author: "B"})
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/books/%d", id), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	withCSRF(req, csrf)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("put status=%d body=%s", rec.Code, rec.Body.String())
	}
	var book models.Book
	if err := json.NewDecoder(rec.Body).Decode(&book); err != nil {
		t.Fatal(err)
	}
	if book.Title != "New Title" {
		t.Fatalf("title=%q", book.Title)
	}
}

func TestPutBookMetadataAuthUser(t *testing.T) {
	srv, store := testServer(t)
	ctx := context.Background()
	hash, _ := auth.HashPassword("longpassword")
	if _, err := store.CreateUser(ctx, "reader", hash, false); err != nil {
		t.Fatal(err)
	}
	id, err := store.UpsertBook(ctx, &models.Book{Title: "Old", Author: "A", Format: models.FormatEPUB, RelPath: "a.epub"}, 1)
	if err != nil {
		t.Fatal(err)
	}
	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}
	csrf := fetchCSRF(t, handler)
	loginBody := bytes.NewBufferString(`{"username":"reader","password":"longpassword"}`)
	loginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", loginBody)
	withCSRF(loginReq, csrf)
	loginRec := httptest.NewRecorder()
	handler.ServeHTTP(loginRec, loginReq)
	if loginRec.Code != http.StatusOK {
		t.Fatalf("login status=%d", loginRec.Code)
	}
	var sessionCookie *http.Cookie
	for _, c := range loginRec.Result().Cookies() {
		if c.Name == auth.SessionCookie {
			sessionCookie = c
		}
	}
	body, _ := json.Marshal(models.BookUpdate{Title: "Auth New", Author: "B"})
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/books/%d", id), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(sessionCookie)
	withCSRF(req, csrf)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("put status=%d body=%s", rec.Code, rec.Body.String())
	}
}
