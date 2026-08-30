package storage

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"athenaeum/internal/auth"
	"athenaeum/internal/models"
)

func TestListLibrariesForUserAndAPIKeyAdmin(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)
	if err := s.EnsureDefaultLibrary(ctx, t.TempDir()); err != nil {
		t.Fatal(err)
	}
	lib2, err := s.CreateLibrary(ctx, "Second", t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	hash, err := auth.HashPassword("secretpass12")
	if err != nil {
		t.Fatal(err)
	}
	uid, err := s.CreateUser(ctx, "libuser", hash, false)
	if err != nil {
		t.Fatal(err)
	}
	user, err := s.GetUser(ctx, uid)
	if err != nil {
		t.Fatal(err)
	}

	all, err := s.ListLibrariesForUser(ctx, user)
	if err != nil || len(all) < 2 {
		t.Fatalf("unrestricted=%d err=%v", len(all), err)
	}

	if err := s.SetUserLibraries(ctx, uid, []int64{lib2.ID}); err != nil {
		t.Fatal(err)
	}
	restricted, err := s.ListLibrariesForUser(ctx, user)
	if err != nil || len(restricted) != 1 || restricted[0].ID != lib2.ID {
		t.Fatalf("restricted=%v err=%v", restricted, err)
	}

	admin := models.User{ID: uid, IsAdmin: true}
	adminLibs, err := s.ListLibrariesForUser(ctx, admin)
	if err != nil || len(adminLibs) < 2 {
		t.Fatalf("admin=%d err=%v", len(adminLibs), err)
	}

	key, err := s.CreateAPIKey(ctx, uid, "admin-del")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.DeleteAPIKeyAdmin(ctx, key.ID); err != nil {
		t.Fatal(err)
	}
	if err := s.DeleteAPIKeyAdmin(ctx, key.ID); err != ErrNotFound {
		t.Fatalf("second delete=%v", err)
	}
}

func TestHideBookAndAudiobookTracks(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	trackBookID, err := s.UpsertBook(ctx, &models.Book{
		LibraryID: 1, Title: "Track1", Format: models.FormatMP3, RelPath: "set/01.mp3",
	}, 1)
	if err != nil {
		t.Fatal(err)
	}
	setID, err := s.UpsertAudiobookSet(ctx, &models.Book{
		LibraryID: 1, Title: "Set", Author: "A", Format: models.FormatAudiobook,
		RelPath: "set/.athenaeum-set", AbsPath: "/lib/set",
	}, []models.AudiobookTrack{
		{Index: 0, RelPath: "set/01.mp3", Title: "One", Format: models.FormatMP3, FileSize: 10},
		{Index: 1, RelPath: "set/02.mp3", Title: "Two", Format: models.FormatMP3, FileSize: 20},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.HideBook(ctx, trackBookID, setID); err != nil {
		t.Fatal(err)
	}
	b, err := s.GetBook(ctx, trackBookID)
	if err != nil || !b.Hidden {
		t.Fatalf("hidden book=%+v err=%v", b, err)
	}
	tracks, err := s.ListAudiobookTracks(ctx, setID)
	if err != nil || len(tracks) != 2 || tracks[0].Title != "One" {
		t.Fatalf("tracks=%v err=%v", tracks, err)
	}
	empty, err := s.ListAudiobookTracks(ctx, setID+99)
	if err != nil || len(empty) != 0 {
		t.Fatalf("empty tracks=%v err=%v", empty, err)
	}
}

func TestListAuditFilters(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	entries := []models.AuditEntry{
		{ActorID: 1, ActorName: "admin", Action: "user.create", Details: "alice", IP: "10.0.0.1"},
		{ActorID: 1, ActorName: "admin", Action: "user.delete", Details: "bob", TargetName: "bob", IP: "10.0.0.2"},
		{ActorID: 2, ActorName: "other", Action: "login", Details: "ok", IP: "192.168.1.1"},
	}
	for _, e := range entries {
		if err := s.InsertAudit(ctx, e); err != nil {
			t.Fatal(err)
		}
	}

	page, err := s.ListAudit(ctx, 0, 0, "", "")
	if err != nil || page.Total != 3 || page.Limit != 50 {
		t.Fatalf("all=%+v err=%v", page, err)
	}
	byAction, err := s.ListAudit(ctx, 10, 0, "user.create", "")
	if err != nil || byAction.Total != 1 || byAction.Items[0].Action != "user.create" {
		t.Fatalf("action=%+v err=%v", byAction, err)
	}
	byQ, err := s.ListAudit(ctx, 10, 0, "", "bob")
	if err != nil || byQ.Total != 1 {
		t.Fatalf("q=%+v err=%v", byQ, err)
	}
	both, err := s.ListAudit(ctx, 10, 0, "user.delete", "10.0.0.2")
	if err != nil || both.Total != 1 {
		t.Fatalf("both=%+v err=%v", both, err)
	}
	capped, err := s.ListAudit(ctx, 500, 0, "", "")
	if err != nil || capped.Limit != 200 {
		t.Fatalf("cap=%+v err=%v", capped, err)
	}
	offset, err := s.ListAudit(ctx, 1, 1, "", "")
	if err != nil || len(offset.Items) != 1 || offset.Total != 3 {
		t.Fatalf("offset=%+v err=%v", offset, err)
	}
}

func TestCreateReadingCollection(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)
	uid, err := s.CreateUser(ctx, "reader", "hash", false)
	if err != nil {
		t.Fatal(err)
	}
	c, err := s.CreateReadingCollection(ctx, uid, "Currently Reading", "desc")
	if err != nil {
		t.Fatal(err)
	}
	if c.Kind != models.CollectionReading || c.Name != "Currently Reading" {
		t.Fatalf("collection=%+v", c)
	}
}

func TestReplaceAndSearchBookContent(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)
	id, err := s.UpsertBook(ctx, &models.Book{
		Title: "Dune", Format: models.FormatEPUB, RelPath: "dune.epub",
	}, 1)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.ReplaceBookContent(ctx, id, []string{
		"The spice must flow across Arrakis",
		"Paul Atreides walked the desert",
	}); err != nil {
		t.Fatal(err)
	}
	ids, err := s.SearchBookContentIDs(ctx, "spice")
	if err != nil || len(ids) != 1 || ids[0] != id {
		t.Fatalf("search spice=%v err=%v", ids, err)
	}
	ids, err = s.SearchBookContentIDs(ctx, "atreides desert")
	if err != nil || len(ids) != 1 {
		t.Fatalf("search multi=%v err=%v", ids, err)
	}
	ids, err = s.SearchBookContentIDs(ctx, "")
	if err != nil || ids != nil {
		t.Fatalf("empty query=%v err=%v", ids, err)
	}
	if err := s.ReplaceBookContent(ctx, id, []string{"replacement only"}); err != nil {
		t.Fatal(err)
	}
	ids, err = s.SearchBookContentIDs(ctx, "spice")
	if err != nil || len(ids) != 0 {
		t.Fatalf("after replace=%v err=%v", ids, err)
	}
	ids, err = s.SearchBookContentIDs(ctx, "replacement")
	if err != nil || len(ids) != 1 {
		t.Fatalf("replacement=%v err=%v", ids, err)
	}
}

func TestGuestUsersLifecycle(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	soon := time.Now().Add(2 * time.Hour)
	later := time.Now().Add(48 * time.Hour)
	expired := time.Now().Add(-time.Hour)

	g1, err := s.CreateGuestUser(ctx, "guest1", "hash", soon, models.DefaultUserPermissions)
	if err != nil {
		t.Fatal(err)
	}
	g2, err := s.CreateGuestUser(ctx, "guest2", "hash", later, models.DefaultUserPermissions)
	if err != nil {
		t.Fatal(err)
	}
	g3, err := s.CreateGuestUser(ctx, "guest3", "hash", expired, models.DefaultUserPermissions)
	if err != nil {
		t.Fatal(err)
	}
	permanent, err := s.CreateUser(ctx, "perm", "hash", false)
	if err != nil {
		t.Fatal(err)
	}

	all, err := s.ListGuestUsers(ctx, 0)
	if err != nil || len(all) != 3 {
		t.Fatalf("all guests=%d err=%v", len(all), err)
	}
	expiring, err := s.ListGuestUsers(ctx, 6)
	if err != nil || len(expiring) != 1 || expiring[0].ID != g1 {
		t.Fatalf("expiring=%v err=%v", expiring, err)
	}

	newExp := time.Now().Add(72 * time.Hour)
	if err := s.ExtendGuestExpiry(ctx, g1, newExp); err != nil {
		t.Fatal(err)
	}
	if err := s.ExtendGuestExpiry(ctx, permanent, newExp); err != ErrNotFound {
		t.Fatalf("extend permanent=%v", err)
	}

	n, err := s.PurgeExpiredGuests(ctx)
	if err != nil || n != 1 {
		t.Fatalf("purge=%d err=%v", n, err)
	}
	if _, err := s.GetUser(ctx, g3); err != ErrNotFound {
		t.Fatalf("expired still present: %v", err)
	}

	deleted, err := s.DeleteGuestUsers(ctx, []int64{g2, permanent, 99999})
	if err != nil || deleted != 1 {
		t.Fatalf("delete=%d err=%v", deleted, err)
	}
}

func TestInvitesCoverage(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)
	adminID, err := s.CreateUser(ctx, "admin", "hash", true)
	if err != nil {
		t.Fatal(err)
	}
	expires := time.Now().Add(24 * time.Hour)
	past := time.Now().Add(-time.Hour)

	pending, err := s.CreateInvite(ctx, models.Invite{
		Kind: models.InviteKindPermanent, Email: "p@example.com",
		Permissions: models.DefaultUserPermissions, CreatedBy: adminID, ExpiresAt: &expires,
	})
	if err != nil {
		t.Fatal(err)
	}
	got, err := s.GetInvite(ctx, pending.ID)
	if err != nil || got.Email != "p@example.com" {
		t.Fatalf("get=%+v err=%v", got, err)
	}
	if _, err := s.GetInvite(ctx, 99999); err != ErrNotFound {
		t.Fatalf("missing=%v", err)
	}

	if err := s.SetInvitePocketID(ctx, pending.ID, "pocket-xyz"); err != nil {
		t.Fatal(err)
	}
	got, err = s.GetInvite(ctx, pending.ID)
	if err != nil || got.PocketIDUserID != "pocket-xyz" {
		t.Fatalf("pocket=%+v err=%v", got, err)
	}

	sso, err := s.CreateInvite(ctx, models.Invite{
		Kind: models.InviteKindGuest, CreatedBy: adminID, ExpiresAt: &expires,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.AcceptInviteSSO(ctx, sso.ID); err != nil {
		t.Fatal(err)
	}
	if err := s.AcceptInviteSSO(ctx, sso.ID); err != ErrConflict {
		t.Fatalf("sso conflict=%v", err)
	}

	revoked, err := s.CreateInvite(ctx, models.Invite{
		Kind: models.InviteKindGuest, CreatedBy: adminID, ExpiresAt: &expires,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.RevokeInvite(ctx, revoked.ID); err != nil {
		t.Fatal(err)
	}

	expiredInv, err := s.CreateInvite(ctx, models.Invite{
		Kind: models.InviteKindPermanent, CreatedBy: adminID, ExpiresAt: &past,
	})
	if err != nil {
		t.Fatal(err)
	}

	all, err := s.ListInvites(ctx, "")
	if err != nil || len(all) < 4 {
		t.Fatalf("all=%d err=%v", len(all), err)
	}
	for _, status := range []string{"pending", "accepted", "revoked", "expired"} {
		list, err := s.ListInvites(ctx, status)
		if err != nil || len(list) < 1 {
			t.Fatalf("status %s=%d err=%v", status, len(list), err)
		}
	}
	_ = expiredInv

	now := time.Now()
	if inviteStatus(pending, now) != "pending" {
		t.Fatal("pending status")
	}
	acc := pending
	tAcc := now
	acc.AcceptedAt = &tAcc
	if inviteStatus(acc, now) != "accepted" {
		t.Fatal("accepted status")
	}
	rev := pending
	tRev := now
	rev.RevokedAt = &tRev
	if inviteStatus(rev, now) != "revoked" {
		t.Fatal("revoked status")
	}
	exp := pending
	exp.ExpiresAt = &past
	if inviteStatus(exp, now) != "expired" {
		t.Fatal("expired status")
	}
}

func TestKosyncProgress(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)
	uid, err := s.CreateUser(ctx, "ko", "hash", false)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.GetKosyncProgress(ctx, uid, "doc"); err != ErrNotFound {
		t.Fatalf("missing=%v", err)
	}
	if err := s.SaveKosyncProgress(ctx, uid, "doc", "/3/2", 0.4, "kobo", "dev1", 100); err != nil {
		t.Fatal(err)
	}
	got, err := s.GetKosyncProgress(ctx, uid, "doc")
	if err != nil || got.Progress != "/3/2" || got.Percentage != 0.4 || got.Device != "kobo" {
		t.Fatalf("got=%+v err=%v", got, err)
	}
	if err := s.SaveKosyncProgress(ctx, uid, "doc", "/4/2", 0.8, "kindle", "dev2", 200); err != nil {
		t.Fatal(err)
	}
	got, err = s.GetKosyncProgress(ctx, uid, "doc")
	if err != nil || got.Progress != "/4/2" || got.Timestamp != 200 {
		t.Fatalf("upsert=%+v err=%v", got, err)
	}
}

func TestLibraryMountUpdateAndS3Input(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)
	root := t.TempDir()
	if err := s.EnsureDefaultLibrary(ctx, root); err != nil {
		t.Fatal(err)
	}

	path, err := s.LibraryMountPath(ctx, 1)
	if err != nil || path != root {
		t.Fatalf("path=%q err=%v", path, err)
	}
	if _, err := s.LibraryMountPath(ctx, 999); err != ErrNotFound {
		t.Fatalf("missing path=%v", err)
	}
	backend, err := s.LibraryBackend(ctx, 1)
	if err != nil || backend != models.LibraryBackendLocal {
		t.Fatalf("backend=%q err=%v", backend, err)
	}
	if _, err := s.LibraryBackend(ctx, 999); err != ErrNotFound {
		t.Fatalf("missing backend=%v", err)
	}
	if _, err := s.OpenLibraryFS(ctx, 1); err != nil {
		t.Fatal(err)
	}
	if _, err := s.OpenLibraryFS(ctx, 999); err != ErrNotFound {
		t.Fatalf("missing fs=%v", err)
	}
	lcfg, err := libfsConfigFromRow(libraryBackendRow{Backend: "", Path: root, Config: "{}"})
	if err != nil || lcfg.Backend == "" {
		t.Fatalf("libfs cfg=%+v err=%v", lcfg, err)
	}
	if _, err := libfsConfigFromRow(libraryBackendRow{Backend: "s3", Path: "s3://b", Config: "{"}); err == nil {
		t.Fatal("expected invalid s3 config")
	}

	newMount := filepath.Join(t.TempDir(), "lib2")
	if err := os.MkdirAll(newMount, 0o755); err != nil {
		t.Fatal(err)
	}
	updated, err := s.UpdateLibrary(ctx, 1, "Renamed", newMount)
	if err != nil || updated.Name != "Renamed" || updated.MountPath != newMount {
		t.Fatalf("update=%+v err=%v", updated, err)
	}

	full, err := s.UpdateLibraryFull(ctx, 1, models.LibraryCreate{
		Name: "Full", MountPath: newMount, Backend: models.LibraryBackendLocal,
	})
	if err != nil || full.Name != "Full" {
		t.Fatalf("full=%+v err=%v", full, err)
	}
	if _, err := s.UpdateLibraryFull(ctx, 1, models.LibraryCreate{Name: ""}); err == nil {
		t.Fatal("expected name required")
	}
	if _, err := s.UpdateLibraryFull(ctx, 1, models.LibraryCreate{
		Name: "X", Backend: "ftp",
	}); err == nil {
		t.Fatal("expected unsupported backend")
	}
	if _, err := s.UpdateLibraryFull(ctx, 1, models.LibraryCreate{
		Name: "X", Backend: models.LibraryBackendS3,
	}); err == nil {
		t.Fatal("expected s3 config required")
	}
	if _, err := s.UpdateLibraryFull(ctx, 999, models.LibraryCreate{
		Name: "X", MountPath: newMount,
	}); err != ErrNotFound {
		t.Fatalf("missing lib=%v", err)
	}

	s3cfg := s3InputToLibfs(models.LibraryS3Input{
		Endpoint: "http://localhost:9000", Region: "us-east-1", Bucket: "b",
		Prefix: "p", AccessKey: "ak", SecretKey: "sk", UsePathStyle: true, TLS: false,
	})
	if s3cfg.Bucket != "b" || s3cfg.AccessKey != "ak" || !s3cfg.UsePathStyle {
		t.Fatalf("s3 cfg=%+v", s3cfg)
	}
}

func TestTouchSessionEmailAndOIDCLink(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)
	uid, err := s.CreateUser(ctx, "sess", "hash", false)
	if err != nil {
		t.Fatal(err)
	}
	access, _ := auth.NewSessionToken()
	exp := time.Now().Add(time.Hour)
	if err := s.CreateUserSession(ctx, models.SessionCreate{
		SessionID: "sid1", AccessToken: access, RefreshToken: "r1",
		UserID: uid, AccessExpires: exp, RefreshExpires: exp,
	}); err != nil {
		t.Fatal(err)
	}
	if err := s.TouchSession(ctx, access); err != nil {
		t.Fatal(err)
	}

	if _, err := s.FindUserByEmail(ctx, ""); err != ErrNotFound {
		t.Fatalf("empty email=%v", err)
	}
	if _, err := s.FindUserByEmail(ctx, "nobody@example.com"); err != ErrNotFound {
		t.Fatalf("missing email=%v", err)
	}
	invited, err := s.CreateInvitedUser(ctx, "emailed", "hash", "User@Example.com", models.DefaultUserPermissions)
	if err != nil {
		t.Fatal(err)
	}
	u, err := s.FindUserByEmail(ctx, "user@example.com")
	if err != nil || u.ID != invited {
		t.Fatalf("find email=%+v err=%v", u, err)
	}

	if err := s.LinkOIDCSub(ctx, uid, "sub-abc", "linked@example.com"); err != nil {
		t.Fatal(err)
	}
	bySub, err := s.FindUserByOIDCSub(ctx, "sub-abc")
	if err != nil || bySub.ID != uid {
		t.Fatalf("oidc=%+v err=%v", bySub, err)
	}
	if err := s.LinkOIDCSub(ctx, 99999, "x", "y"); err != ErrConflict {
		t.Fatalf("link missing user=%v want conflict", err)
	}
}

func TestShareLinksCoverage(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)
	uid, err := s.CreateUser(ctx, "share", "hash", true)
	if err != nil {
		t.Fatal(err)
	}
	bookID, err := s.UpsertBook(ctx, &models.Book{
		Title: "Shared", Format: models.FormatEPUB, RelPath: "s.epub",
	}, 1)
	if err != nil {
		t.Fatal(err)
	}
	exp := time.Now().Add(time.Hour)
	sl, err := s.CreateShareLink(ctx, bookID, uid, &exp, 5)
	if err != nil {
		t.Fatal(err)
	}
	list, err := s.ListSharesForBook(ctx, bookID)
	if err != nil || len(list) != 1 || list[0].ID != sl.ID {
		t.Fatalf("list=%v err=%v", list, err)
	}
	if err := s.IncrementShareDownload(ctx, sl.ID); err != nil {
		t.Fatal(err)
	}
	got, err := s.GetShareLinkByToken(ctx, sl.Token)
	if err != nil || got.DownloadCount != 1 {
		t.Fatalf("inc=%+v err=%v", got, err)
	}
	ok, err := s.TryIncrementShareDownload(ctx, sl.ID, 5)
	if err != nil || !ok {
		t.Fatalf("try=%v err=%v", ok, err)
	}
	if err := s.DeleteShareLink(ctx, bookID, sl.ID); err != nil {
		t.Fatal(err)
	}
	if err := s.DeleteShareLink(ctx, bookID, sl.ID); err != ErrNotFound {
		t.Fatalf("second delete=%v", err)
	}
	empty, err := s.ListSharesForBook(ctx, bookID)
	if err != nil || len(empty) != 0 {
		t.Fatalf("empty=%v err=%v", empty, err)
	}
}

func TestUploadSessionsAndDuplicate(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)
	uid, err := s.CreateUser(ctx, "up", "hash", false)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.CreateUploadSession(ctx, models.UploadSession{
		ID: "up1", LibraryID: 1, UserID: uid, RelPath: "new.epub", TotalSize: 100,
	}); err != nil {
		t.Fatal(err)
	}
	sess, err := s.GetUploadSession(ctx, "up1")
	if err != nil || sess.RelPath != "new.epub" || sess.Offset != 0 || sess.Done {
		t.Fatalf("get=%+v err=%v", sess, err)
	}
	if _, err := s.GetUploadSession(ctx, "missing"); err != ErrNotFound {
		t.Fatalf("missing=%v", err)
	}
	if err := s.UpdateUploadOffset(ctx, "up1", 50); err != nil {
		t.Fatal(err)
	}
	if err := s.UpdateUploadOffset(ctx, "missing", 1); err != ErrNotFound {
		t.Fatalf("offset missing=%v", err)
	}
	bookID, err := s.UpsertBook(ctx, &models.Book{
		Title: "Uploaded", Format: models.FormatEPUB, RelPath: "new.epub",
	}, 1)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.CompleteUploadSession(ctx, "up1", bookID); err != nil {
		t.Fatal(err)
	}
	done, err := s.GetUploadSession(ctx, "up1")
	if err != nil || !done.Done || done.BookID != bookID {
		t.Fatalf("done=%+v err=%v", done, err)
	}
	if err := s.CompleteUploadSession(ctx, "missing", 1); err != ErrNotFound {
		t.Fatalf("complete missing=%v", err)
	}
	if err := s.UpdateUploadOffset(ctx, "up1", 60); err != ErrNotFound {
		t.Fatalf("offset after done=%v", err)
	}
	if err := s.DeleteUploadSession(ctx, "up1"); err != nil {
		t.Fatal(err)
	}
	if err := s.DeleteUploadSession(ctx, "up1"); err != ErrNotFound {
		t.Fatalf("second delete=%v", err)
	}

	other, err := s.UpsertBook(ctx, &models.Book{
		Title: "Other", Format: models.FormatEPUB, RelPath: "other.epub",
	}, 2)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.SetBookDuplicateOf(ctx, other, bookID); err != nil {
		t.Fatal(err)
	}
	b, err := s.GetBook(ctx, other)
	if err != nil || b.DuplicateOf != bookID {
		t.Fatalf("dup=%+v err=%v", b, err)
	}
}

func TestPurgeExpiredSessions(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)
	uid, err := s.CreateUser(ctx, "purge", "hash", false)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.CreateSession(ctx, "alive", uid, time.Now().Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := s.CreateSession(ctx, "dead", uid, time.Now().Add(-time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := s.CreateRefreshToken(ctx, "r-dead", uid, time.Now().Add(-time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := s.CreateRefreshToken(ctx, "r-alive", uid, time.Now().Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := s.PurgeExpiredSessions(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := s.SessionUser(ctx, "alive"); err != nil {
		t.Fatalf("alive gone: %v", err)
	}
	if _, err := s.SessionUser(ctx, "dead"); err != ErrNotFound {
		t.Fatalf("dead still present: %v", err)
	}
}

func TestWebhookGetAndUpdate(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)
	wh, err := s.CreateWebhook(ctx, models.Webhook{
		URL: "http://example.com/h", Secret: "s", Events: []string{models.WebhookEventPing}, Enabled: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	got, err := s.GetWebhook(ctx, wh.ID)
	if err != nil || got.URL != wh.URL {
		t.Fatalf("get=%+v err=%v", got, err)
	}
	if _, err := s.GetWebhook(ctx, 999); err != ErrNotFound {
		t.Fatalf("missing=%v", err)
	}
	got.URL = "http://example.com/new"
	got.Enabled = false
	got.Events = []string{models.WebhookEventUserCreate}
	if err := s.UpdateWebhook(ctx, got); err != nil {
		t.Fatal(err)
	}
	updated, err := s.GetWebhook(ctx, wh.ID)
	if err != nil || updated.URL != "http://example.com/new" || updated.Enabled {
		t.Fatalf("updated=%+v err=%v", updated, err)
	}
	if err := s.UpdateWebhook(ctx, models.Webhook{ID: 999, URL: "x", Events: []string{}}); err != ErrNotFound {
		t.Fatalf("update missing=%v", err)
	}
}

func TestStoreDriverDBParseAndRebind(t *testing.T) {
	s := newTestStore(t)
	if s.Driver() != DriverSQLite {
		t.Fatalf("driver=%q", s.Driver())
	}
	if s.DB() == nil {
		t.Fatal("db nil")
	}

	for _, tc := range []struct {
		in   string
		want Driver
	}{
		{"", DriverSQLite},
		{"sqlite", DriverSQLite},
		{"sqlite3", DriverSQLite},
		{"postgres", DriverPostgres},
		{"postgresql", DriverPostgres},
		{"pg", DriverPostgres},
	} {
		d, err := ParseDriver(tc.in)
		if err != nil || d != tc.want {
			t.Fatalf("ParseDriver(%q)=%q err=%v", tc.in, d, err)
		}
	}
	_, err := ParseDriver("mysql")
	if err == nil || err.Error() == "" {
		t.Fatal("expected unsupported driver")
	}
	if got := errUnsupportedDriver("foo").Error(); got == "" {
		t.Fatal("empty error")
	}

	if got := rebindDollar("SELECT a FROM t WHERE x=? AND y=?"); got != "SELECT a FROM t WHERE x=$1 AND y=$2" {
		t.Fatalf("rebind=%q", got)
	}
	if got := rebindDollar("SELECT '?' AS q, ?"); got != "SELECT '?' AS q, $1" {
		t.Fatalf("quoted=%q", got)
	}
	if got := rebindDollar("SELECT 'it''s' WHERE a=?"); got != "SELECT 'it''s' WHERE a=$1" {
		t.Fatalf("escaped=%q", got)
	}
	pg := &Store{driver: DriverPostgres}
	if got := pg.rebind("SELECT ?"); got != "SELECT $1" {
		t.Fatalf("store rebind=%q", got)
	}
	sqlite := &Store{driver: DriverSQLite}
	if got := sqlite.rebind("SELECT ?"); got != "SELECT ?" {
		t.Fatalf("sqlite rebind=%q", got)
	}
}
