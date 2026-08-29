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

func TestRunUsersHelpAndUnknown(t *testing.T) {
	out := captureStdout(t, func() {
		if err := RunUsers(nil); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(out, "Manage local user accounts") {
		t.Fatalf("help missing: %q", out)
	}
	if !strings.Contains(out, binaryName()) {
		t.Fatalf("binary name missing: %q", out)
	}

	out = captureStdout(t, func() {
		if err := RunUsers([]string{"help"}); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(out, "Commands:") {
		t.Fatalf("help command: %q", out)
	}

	if err := RunUsers([]string{"nope"}); err == nil {
		t.Fatal("expected unknown command error")
	}
}

func TestRunUsersCRUD(t *testing.T) {
	_, dir := testCLIStore(t)

	if err := RunUsers([]string{"add", "--data", dir, "--password", "longpassword1", "eve"}); err != nil {
		t.Fatal(err)
	}
	out := captureStdout(t, func() {
		if err := RunUsers([]string{"ls", "--data", dir}); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(out, "eve") {
		t.Fatalf("list: %q", out)
	}
	out = captureStdout(t, func() {
		if err := RunUsers([]string{"show", "--data", dir, "eve"}); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(out, "username:    eve") {
		t.Fatalf("show: %q", out)
	}
	if err := RunUsers([]string{"rename", "--data", dir, "eve", "eve2"}); err != nil {
		t.Fatal(err)
	}
	if err := RunUsers([]string{"set-permissions", "--data", dir, "--set", "read,edit_metadata", "eve2"}); err != nil {
		t.Fatal(err)
	}
	if err := RunUsers([]string{"add", "--data", dir, "--password", "longpassword9", "--admin", "boss"}); err != nil {
		t.Fatal(err)
	}
	if err := RunUsers([]string{"set-admin", "--data", dir, "--admin", "eve2"}); err != nil {
		t.Fatal(err)
	}
	if err := RunUsers([]string{"set-admin", "--data", dir, "--no-admin", "eve2"}); err != nil {
		t.Fatal(err)
	}
	if err := RunUsers([]string{"delete", "--data", dir, "--force", "eve2"}); err != nil {
		t.Fatal(err)
	}
}

func TestRunUsersResetPasswordPiped(t *testing.T) {
	_, dir := testCLIStore(t)
	if err := usersAdd([]string{"--data", dir, "--password", "longpassword1", "piper"}); err != nil {
		t.Fatal(err)
	}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdin
	os.Stdin = r
	t.Cleanup(func() { os.Stdin = old })
	if _, err := w.Write([]byte("pipedpassword1\n")); err != nil {
		t.Fatal(err)
	}
	_ = w.Close()

	t.Setenv("ATHENAEUM_PASSWORD", "")
	if err := RunUsers([]string{"reset-password", "--data", dir, "piper"}); err != nil {
		t.Fatal(err)
	}

	store, err := storage.Open(filepath.Join(dir, brand.DBFilename))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	_, hash, err := store.GetUserByUsername(context.Background(), "piper")
	if err != nil {
		t.Fatal(err)
	}
	if !auth.CheckPassword(hash, "pipedpassword1") {
		t.Fatal("piped password not applied")
	}
}

func TestUsersSetPermissionsErrors(t *testing.T) {
	_, dir := testCLIStore(t)
	if err := usersAdd([]string{"--data", dir, "--password", "longpassword1", "--admin", "admin1"}); err != nil {
		t.Fatal(err)
	}
	if err := usersAdd([]string{"--data", dir, "--password", "longpassword2", "member1"}); err != nil {
		t.Fatal(err)
	}
	if err := usersSetPermissions([]string{"--data", dir, "member1"}); err == nil {
		t.Fatal("expected --set required")
	}
	if err := usersSetPermissions([]string{"--data", dir, "--set", "nope", "member1"}); err == nil {
		t.Fatal("expected invalid permissions")
	}
	if err := usersSetPermissions([]string{"--data", dir, "--set", "read", "admin1"}); err == nil {
		t.Fatal("expected admin rejection")
	}
}

func TestApplyUsersColorAndHelpers(t *testing.T) {
	applyUsersColor([]string{"--no-color"})
	applyUsersColor([]string{"--color", "never"})
	applyUsersColor([]string{"--color=always"})
	t.Setenv("ATHENAEUM_NO_COLOR", "1")
	applyUsersColor([]string{"--color", "always"})

	var buf bytes.Buffer
	printUsersHelp(&buf)
	if !strings.Contains(buf.String(), "Usage:") {
		t.Fatalf("help: %q", buf.String())
	}
	buf.Reset()
	printCmd(&buf, "list", "List users")
	if !strings.Contains(buf.String(), "list") {
		t.Fatalf("printCmd: %q", buf.String())
	}

	got := splitCSV(" a, b ,,c ")
	if len(got) != 3 || got[0] != "a" || got[2] != "c" {
		t.Fatalf("splitCSV = %#v", got)
	}
	if len(splitCSV("")) != 0 {
		t.Fatal("empty csv")
	}
	if binaryName() != "athenaeum" {
		t.Fatalf("binaryName = %q", binaryName())
	}

	t.Setenv("ATHENAEUM_PASSWORD_MIN_LENGTH", "10")
	t.Setenv("ATHENAEUM_PASSWORD_LONG_LENGTH", "14")
	t.Setenv("ATHENAEUM_PASSWORD_MIN_KINDS", "2")
	t.Setenv("ATHENAEUM_PASSWORD_REQUIRE_LOWER", "1")
	configurePasswordPolicyFromEnv()
	t.Cleanup(func() { auth.SetPasswordPolicy(auth.DefaultPasswordPolicy()) })
	if err := auth.ValidatePassword("short"); err == nil {
		t.Fatal("expected short password rejection after policy")
	}
}
