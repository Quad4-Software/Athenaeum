package storage

import (
	"context"
	"testing"
	"time"

	"athenaeum/internal/auth"
	"athenaeum/internal/models"
)

func TestAuthAndSessions(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	hash, err := auth.HashPassword("secretpass")
	if err != nil {
		t.Fatal(err)
	}
	id, err := s.CreateUser(ctx, "alice", hash, true)
	if err != nil {
		t.Fatal(err)
	}

	required, err := s.AuthRequired(ctx)
	if err != nil || !required {
		t.Fatalf("auth required = %v err=%v", required, err)
	}

	u, gotHash, err := s.GetUserByUsername(ctx, "alice")
	if err != nil || !auth.CheckPassword(gotHash, "secretpass") {
		t.Fatalf("lookup failed: %+v err=%v", u, err)
	}
	if u.ID != id || !u.IsAdmin {
		t.Errorf("user = %+v", u)
	}

	if err := s.CreateSession(ctx, "tok123", id, time.Now().Add(24*time.Hour)); err != nil {
		t.Fatal(err)
	}
	su, err := s.SessionUser(ctx, "tok123")
	if err != nil || su.Username != "alice" {
		t.Errorf("session user = %+v err=%v", su, err)
	}
	if err := s.DeleteSession(ctx, "tok123"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.SessionUser(ctx, "tok123"); err != ErrNotFound {
		t.Errorf("expected expired session, got %v", err)
	}

	if err := s.CreateRefreshToken(ctx, "ref123", id, time.Now().Add(24*time.Hour)); err != nil {
		t.Fatal(err)
	}
	ru, err := s.RefreshTokenUser(ctx, "ref123")
	if err != nil || ru.Username != "alice" {
		t.Errorf("refresh user = %+v err=%v", ru, err)
	}
	if err := s.DeleteRefreshToken(ctx, "ref123"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.RefreshTokenUser(ctx, "ref123"); err != ErrNotFound {
		t.Errorf("expected expired refresh, got %v", err)
	}
}

func TestCollectionsCRUD(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	book := &models.Book{Title: "One", Format: models.FormatEPUB, RelPath: "one.epub"}
	bookID, err := s.UpsertBook(ctx, book, 1)
	if err != nil {
		t.Fatal(err)
	}

	c, err := s.CreateCollection(ctx, models.AnonymousUserID, "Favorites", "best picks")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.AddToCollection(ctx, models.AnonymousUserID, c.ID, bookID); err != nil {
		t.Fatal(err)
	}

	page, err := s.ListBooks(ctx, models.BookQuery{CollectionID: c.ID})
	if err != nil || page.Total != 1 {
		t.Fatalf("collection list total=%d err=%v", page.Total, err)
	}

	c2, err := s.UpdateCollection(ctx, models.AnonymousUserID, c.ID, "Top", "updated")
	if err != nil || c2.Name != "Top" {
		t.Fatalf("update: %+v err=%v", c2, err)
	}

	if err := s.RemoveFromCollection(ctx, models.AnonymousUserID, c.ID, bookID); err != nil {
		t.Fatal(err)
	}
	if err := s.DeleteCollection(ctx, models.AnonymousUserID, c.ID); err != nil {
		t.Fatal(err)
	}
}

func TestFTSSearch(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	books := []models.Book{
		{Title: "Neuromancer", Author: "Gibson", Description: "cyberpunk classic", Format: models.FormatEPUB, RelPath: "a.epub"},
		{Title: "Snow Crash", Author: "Stephenson", Series: "Metaverse", Format: models.FormatEPUB, RelPath: "b.epub"},
	}
	for i := range books {
		if _, err := s.UpsertBook(ctx, &books[i], int64(i)); err != nil {
			t.Fatal(err)
		}
	}

	page, err := s.ListBooks(ctx, models.BookQuery{Search: "cyberpunk"})
	if err != nil {
		t.Fatal(err)
	}
	if page.Total != 1 || page.Items[0].Title != "Neuromancer" {
		t.Errorf("fts search: total=%d title=%q", page.Total, page.Items[0].Title)
	}
}

func TestListSeries(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)
	if _, err := s.UpsertBook(ctx, &models.Book{Title: "A", Series: "Dune", Format: models.FormatEPUB, RelPath: "1.epub"}, 1); err != nil {
		t.Fatal(err)
	}
	if _, err := s.UpsertBook(ctx, &models.Book{Title: "B", Series: "Dune", Format: models.FormatEPUB, RelPath: "2.epub"}, 1); err != nil {
		t.Fatal(err)
	}
	series, err := s.ListSeries(ctx, 0, nil)
	if err != nil || len(series) != 1 || series[0].Count != 2 {
		t.Fatalf("series = %+v err=%v", series, err)
	}
}

func TestProgressPerUser(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)
	id, _ := s.UpsertBook(ctx, &models.Book{Title: "X", Format: models.FormatPDF, RelPath: "x.pdf"}, 1)

	_ = s.SaveProgress(ctx, 1, models.Progress{BookID: id, Location: "p1", Percent: 0.2})
	_ = s.SaveProgress(ctx, 2, models.Progress{BookID: id, Location: "p2", Percent: 0.8})

	p1, _ := s.GetProgress(ctx, 1, id)
	p2, _ := s.GetProgress(ctx, 2, id)
	if p1.Percent != 0.2 || p2.Percent != 0.8 {
		t.Errorf("p1=%+v p2=%+v", p1, p2)
	}
}
