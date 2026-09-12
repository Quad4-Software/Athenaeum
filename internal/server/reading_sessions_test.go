package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"athenaeum/internal/models"
)

func TestReadingSessionsRoute(t *testing.T) {
	srv, store := testServer(t)
	ctx := context.Background()
	uid := models.AnonymousUserID

	bookA, err := store.UpsertBook(ctx, &models.Book{Title: "Alpha", Format: models.FormatEPUB, RelPath: "a.epub"}, 1)
	if err != nil {
		t.Fatal(err)
	}
	bookB, err := store.UpsertBook(ctx, &models.Book{Title: "Beta", Format: models.FormatEPUB, RelPath: "b.epub"}, 1)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.AddReadSeconds(ctx, uid, bookA, 120); err != nil {
		t.Fatal(err)
	}
	if err := store.AddReadSeconds(ctx, uid, bookB, 60); err != nil {
		t.Fatal(err)
	}

	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/me/reading-sessions", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	var sessions []models.ReadingSession
	if err := json.NewDecoder(rec.Body).Decode(&sessions); err != nil {
		t.Fatal(err)
	}
	if len(sessions) != 2 {
		t.Fatalf("sessions = %d, want 2", len(sessions))
	}
	// Newest first: Beta was recorded after Alpha.
	if sessions[0].Title != "Beta" || sessions[0].Seconds != 60 {
		t.Errorf("first session = %+v, want Beta/60s", sessions[0])
	}
	if sessions[1].Title != "Alpha" || sessions[1].BookID != bookA {
		t.Errorf("second session = %+v, want Alpha book %d", sessions[1], bookA)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/me/reading-sessions?limit=1", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("limit status=%d", rec.Code)
	}
	var limited []models.ReadingSession
	if err := json.NewDecoder(rec.Body).Decode(&limited); err != nil {
		t.Fatal(err)
	}
	if len(limited) != 1 || limited[0].Title != "Beta" {
		t.Fatalf("limit=1 got %+v", limited)
	}
}
