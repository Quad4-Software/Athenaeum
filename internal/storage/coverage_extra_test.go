package storage

import (
	"context"
	"testing"

	"athenaeum/internal/auth"
	"athenaeum/internal/models"
)

func TestBookmarksHighlightsAndReadingStats(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	hash, err := auth.HashPassword("secretpass12")
	if err != nil {
		t.Fatal(err)
	}
	userID, err := s.CreateUser(ctx, "reader", hash, false)
	if err != nil {
		t.Fatal(err)
	}
	bookID, err := s.UpsertBook(ctx, &models.Book{
		Title: "Dune", Author: "Herbert", Format: models.FormatEPUB, RelPath: "dune.epub",
	}, 1)
	if err != nil {
		t.Fatal(err)
	}

	bmID, err := s.CreateBookmark(ctx, userID, models.Bookmark{
		BookID: bookID, Location: "epubcfi(/6/2)", Label: "ch1",
	})
	if err != nil {
		t.Fatal(err)
	}
	list, err := s.ListBookmarks(ctx, userID, bookID)
	if err != nil || len(list) != 1 || list[0].ID != bmID {
		t.Fatalf("bookmarks=%v err=%v", list, err)
	}
	if err := s.DeleteBookmark(ctx, userID, bmID); err != nil {
		t.Fatal(err)
	}
	if err := s.DeleteBookmark(ctx, userID, bmID); err != ErrNotFound {
		t.Fatalf("delete again: %v", err)
	}

	hlID, err := s.CreateHighlight(ctx, userID, models.Highlight{
		BookID: bookID, Location: "loc", Excerpt: "quote", Note: "n",
	})
	if err != nil {
		t.Fatal(err)
	}
	hls, err := s.ListHighlights(ctx, userID, bookID)
	if err != nil || len(hls) != 1 || hls[0].Color != "yellow" {
		t.Fatalf("highlights=%v err=%v", hls, err)
	}
	if err := s.DeleteHighlight(ctx, userID, hlID); err != nil {
		t.Fatal(err)
	}

	if err := s.AddReadSeconds(ctx, userID, bookID, 0); err != nil {
		t.Fatal(err)
	}
	if err := s.AddReadSeconds(ctx, userID, bookID, 120); err != nil {
		t.Fatal(err)
	}
	if err := s.SaveProgress(ctx, userID, models.Progress{
		BookID: bookID, Location: "x", Percent: 0.5,
	}); err != nil {
		t.Fatal(err)
	}
	st, err := s.ReadingStats(ctx, userID)
	if err != nil || st.TotalReadSeconds < 120 || st.BooksInProgress < 1 {
		t.Fatalf("stats=%+v err=%v", st, err)
	}
	pm, err := s.ProgressMap(ctx, userID, []int64{bookID, 999})
	if err != nil || pm[bookID].Percent != 0.5 {
		t.Fatalf("progress map=%v err=%v", pm, err)
	}
	empty, err := s.ProgressMap(ctx, userID, nil)
	if err != nil || len(empty) != 0 {
		t.Fatalf("empty map=%v err=%v", empty, err)
	}
}

func TestFavoritesOfflineSMTPAndAuthors(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	hash, err := auth.HashPassword("secretpass12")
	if err != nil {
		t.Fatal(err)
	}
	userID, err := s.CreateUser(ctx, "fav", hash, false)
	if err != nil {
		t.Fatal(err)
	}
	bookID, err := s.UpsertBook(ctx, &models.Book{
		Title: "Book", Author: "Ada", Format: models.FormatPDF, RelPath: "a.pdf",
	}, 1)
	if err != nil {
		t.Fatal(err)
	}

	if err := s.SetFavorite(ctx, userID, bookID, true); err != nil {
		t.Fatal(err)
	}
	ok, err := s.IsFavorite(ctx, userID, bookID)
	if err != nil || !ok {
		t.Fatalf("favorite=%v err=%v", ok, err)
	}
	ids, err := s.ListFavoriteIDs(ctx, userID)
	if err != nil || len(ids) != 1 || ids[0] != bookID {
		t.Fatalf("ids=%v err=%v", ids, err)
	}
	n, err := s.CountFavorites(ctx, userID)
	if err != nil || n != 1 {
		t.Fatalf("count=%d err=%v", n, err)
	}
	if err := s.SetFavorite(ctx, userID, bookID, false); err != nil {
		t.Fatal(err)
	}

	if err := s.AddOfflineGrants(ctx, userID, []int64{bookID}); err != nil {
		t.Fatal(err)
	}
	offline, err := s.ListOfflineGrants(ctx, userID)
	if err != nil || len(offline) != 1 {
		t.Fatalf("offline=%v err=%v", offline, err)
	}
	if err := s.RemoveOfflineGrants(ctx, userID, []int64{bookID}); err != nil {
		t.Fatal(err)
	}

	smtp := models.SMTPSettings{
		Enabled: true, Host: "smtp.example", Port: 587, Username: "u",
		Password: "p", FromAddr: "a@b.c", UseTLS: true,
	}
	if err := s.SaveSMTPSettings(ctx, smtp); err != nil {
		t.Fatal(err)
	}
	got, err := s.GetSMTPSettings(ctx)
	if err != nil || !got.Enabled || got.Host != "smtp.example" || got.Password != "p" {
		t.Fatalf("smtp=%+v err=%v", got, err)
	}
	if err := s.SaveKindleEmail(ctx, userID, "kindle@example.com"); err != nil {
		t.Fatal(err)
	}
	email, err := s.GetKindleEmail(ctx, userID)
	if err != nil || email != "kindle@example.com" {
		t.Fatalf("kindle=%q err=%v", email, err)
	}
	email, err = s.GetKindleEmail(ctx, userID+99)
	if err != nil || email != "" {
		t.Fatalf("missing kindle=%q err=%v", email, err)
	}

	authors, err := s.ListAuthors(ctx, 0, nil)
	if err != nil || len(authors) == 0 || authors[0].Name != "Ada" {
		t.Fatalf("authors=%v err=%v", authors, err)
	}
}

func TestChaptersAuthSettingsPermissionsTOTP(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	hash, err := auth.HashPassword("secretpass12")
	if err != nil {
		t.Fatal(err)
	}
	userID, err := s.CreateUser(ctx, "perms", hash, false)
	if err != nil {
		t.Fatal(err)
	}
	bookID, err := s.UpsertBook(ctx, &models.Book{
		Title: "Audio", Format: models.FormatM4B, RelPath: "a.m4b",
	}, 1)
	if err != nil {
		t.Fatal(err)
	}

	if err := s.ReplaceChapters(ctx, bookID, []models.Chapter{
		{Index: 0, Title: "Intro", StartSec: 1.5},
		{Index: 1, Title: "One", StartSec: 10},
	}); err != nil {
		t.Fatal(err)
	}
	chs, err := s.ListChapters(ctx, bookID)
	if err != nil || len(chs) != 2 || chs[0].StartSec != 1.5 {
		t.Fatalf("chapters=%v err=%v", chs, err)
	}

	if err := s.SaveAuthSettings(ctx, models.AuthSettings{AllowRegistration: true, RequireTOTP: true}); err != nil {
		t.Fatal(err)
	}
	as, err := s.GetAuthSettings(ctx)
	if err != nil || !as.AllowRegistration || !as.RequireTOTP {
		t.Fatalf("auth settings=%+v err=%v", as, err)
	}

	if err := s.SetUserPermissions(ctx, userID, models.PermRead|models.PermEditMetadata); err != nil {
		t.Fatal(err)
	}
	u, err := s.GetUser(ctx, userID)
	if err != nil || u.Permissions != models.PermRead|models.PermEditMetadata {
		t.Fatalf("user=%+v err=%v", u, err)
	}

	if err := s.SetUserTOTPSecret(ctx, userID, "SECRET"); err != nil {
		t.Fatal(err)
	}
	secret, err := s.GetUserTOTPSecret(ctx, userID)
	if err != nil || secret != "SECRET" {
		t.Fatalf("secret=%q err=%v", secret, err)
	}
	if err := s.EnableUserTOTP(ctx, userID); err != nil {
		t.Fatal(err)
	}
	u, err = s.GetUser(ctx, userID)
	if err != nil || !u.TOTPEnabled {
		t.Fatalf("totp enabled=%+v err=%v", u, err)
	}
	if err := s.DisableUserTOTP(ctx, userID); err != nil {
		t.Fatal(err)
	}
	secret, err = s.GetUserTOTPSecret(ctx, userID)
	if err != nil || secret != "" {
		t.Fatalf("cleared secret=%q err=%v", secret, err)
	}
}
