package cli

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"athenaeum/internal/auth"
	"athenaeum/internal/brand"
	"athenaeum/internal/storage"
)

func testCLIStore(t *testing.T) (*storage.Store, string) {
	t.Helper()
	dir := t.TempDir()
	db := filepath.Join(dir, brand.DBFilename)
	store, err := storage.Open(db)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { store.Close() })
	return store, dir
}

func seedUser(t *testing.T, store *storage.Store, username, password string, admin bool) int64 {
	t.Helper()
	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatal(err)
	}
	id, err := store.CreateUser(context.Background(), username, hash, admin)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	t.Cleanup(func() { os.Stdout = old })
	fn()
	_ = w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	_ = r.Close()
	return buf.String()
}

func TestUsersListAndShow(t *testing.T) {
	store, dir := testCLIStore(t)
	seedUser(t, store, "alice", "longpassword1", true)

	out := captureStdout(t, func() {
		if err := usersList([]string{"--data", dir}); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(out, "alice") {
		t.Fatalf("list output missing alice: %q", out)
	}

	out = captureStdout(t, func() {
		if err := usersShow([]string{"--data", dir, "alice"}); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(out, "username:    alice") {
		t.Fatalf("show output: %q", out)
	}
}

func TestUsersAddAndResetPassword(t *testing.T) {
	_, dir := testCLIStore(t)

	if err := usersAdd([]string{"--data", dir, "--password", "longpassword1", "bob"}); err != nil {
		t.Fatal(err)
	}
	if err := usersResetPassword([]string{"--data", dir, "--password", "longpassword2", "bob"}); err != nil {
		t.Fatal(err)
	}

	store, err := storage.Open(filepath.Join(dir, brand.DBFilename))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	_, hash, err := store.GetUserByUsername(context.Background(), "bob")
	if err != nil {
		t.Fatal(err)
	}
	if auth.CheckPassword(hash, "longpassword1") {
		t.Fatal("old password still works")
	}
	if !auth.CheckPassword(hash, "longpassword2") {
		t.Fatal("new password not applied")
	}
}

func TestUsersRenameAndDelete(t *testing.T) {
	store, dir := testCLIStore(t)
	id := seedUser(t, store, "carol", "longpassword1", false)

	if err := usersRename([]string{"--data", dir, "carol", "carol2"}); err != nil {
		t.Fatal(err)
	}
	if err := usersDelete([]string{"--data", dir, "--force", "carol2"}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.GetUser(context.Background(), id); err != storage.ErrNotFound {
		t.Fatalf("expected deleted user, got %v", err)
	}
}

func TestUsersSetAdmin(t *testing.T) {
	_, dir := testCLIStore(t)
	if err := usersAdd([]string{"--data", dir, "--password", "longpassword1", "--admin", "admin"}); err != nil {
		t.Fatal(err)
	}
	if err := usersAdd([]string{"--data", dir, "--password", "longpassword2", "member"}); err != nil {
		t.Fatal(err)
	}
	if err := usersSetAdmin([]string{"--data", dir, "--admin", "member"}); err != nil {
		t.Fatal(err)
	}
	if err := usersSetAdmin([]string{"--data", dir, "--no-admin", "member"}); err != nil {
		t.Fatal(err)
	}
}

func TestUsersCannotDeleteLastAdmin(t *testing.T) {
	_, dir := testCLIStore(t)
	if err := usersAdd([]string{"--data", dir, "--password", "longpassword1", "--admin", "onlyadmin"}); err != nil {
		t.Fatal(err)
	}
	if err := usersDelete([]string{"--data", dir, "--force", "onlyadmin"}); err == nil {
		t.Fatal("expected error deleting last admin")
	}
}

func TestResolveUserByID(t *testing.T) {
	store, _ := testCLIStore(t)
	id := seedUser(t, store, "dave", "longpassword1", false)

	u, err := resolveUser(context.Background(), store, "99999")
	if err == nil {
		t.Fatalf("expected missing user, got %+v", u)
	}

	u, err = resolveUser(context.Background(), store, fmt.Sprintf("%d", id))
	if err != nil {
		t.Fatal(err)
	}
	if u.ID != id || u.Username != "dave" {
		t.Fatalf("user = %+v", u)
	}
}

func TestReadPasswordFromEnv(t *testing.T) {
	t.Setenv("ATHENAEUM_PASSWORD", "from-env")
	pass, err := readPassword("")
	if err != nil || pass != "from-env" {
		t.Fatalf("pass=%q err=%v", pass, err)
	}
}
