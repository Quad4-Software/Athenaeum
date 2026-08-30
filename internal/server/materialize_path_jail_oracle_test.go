package server

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"athenaeum/internal/models"
)

// PROVED_MATERIALIZE_PATH_JAIL
// Guarantee: materializeBookFile for local mounts must refuse RelPath with ..
// so tools (comics, convert, content-index) cannot open files outside the mount.
// Expected: error. Actual before fix: returned abs path outside the mount.

func TestMaterializeBookFileRejectsTraversalOracle(t *testing.T) {
	srv, store := testServer(t)
	ctx := context.Background()

	base := srv.cfg.DataDir
	mount := filepath.Join(base, "libmount")
	outside := filepath.Join(base, "outside-secret.epub")
	if err := os.MkdirAll(mount, 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(outside, []byte("secret-bytes"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(mount, "ok.epub"), []byte("ok"), 0o600); err != nil {
		t.Fatal(err)
	}

	lib, err := store.CreateLibraryFull(ctx, models.LibraryCreate{
		Name:      "jail",
		MountPath: mount,
		Backend:   models.LibraryBackendLocal,
	})
	if err != nil {
		t.Fatal(err)
	}

	path, cleanup, err := srv.materializeBookFile(ctx, lib.ID, "../outside-secret.epub")
	if cleanup != nil {
		cleanup()
	}
	if err == nil {
		t.Fatalf("materialize allowed escape path=%q", path)
	}
	fmt.Println("PROVED_MATERIALIZE_PATH_JAIL: traversal RelPath rejected err=", err)
}

// PROVED_MATERIALIZE_SYMLINK_JAIL
// Guarantee: a RelPath that is a symlink inside the mount must not resolve
// to a target outside the mount.

func TestMaterializeBookFileRejectsSymlinkEscapeOracle(t *testing.T) {
	srv, store := testServer(t)
	ctx := context.Background()

	base := srv.cfg.DataDir
	mount := filepath.Join(base, "symlink-mount")
	outside := filepath.Join(base, "symlink-secret.epub")
	if err := os.MkdirAll(mount, 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(outside, []byte("secret-bytes"), 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(mount, "leak.epub")
	if err := os.Symlink(outside, link); err != nil {
		t.Fatal(err)
	}

	lib, err := store.CreateLibraryFull(ctx, models.LibraryCreate{
		Name:      "symlink-jail",
		MountPath: mount,
		Backend:   models.LibraryBackendLocal,
	})
	if err != nil {
		t.Fatal(err)
	}

	path, cleanup, err := srv.materializeBookFile(ctx, lib.ID, "leak.epub")
	if cleanup != nil {
		cleanup()
	}
	if err == nil {
		t.Fatalf("materialize followed symlink escape path=%q", path)
	}
	fmt.Println("PROVED_MATERIALIZE_SYMLINK_JAIL: symlink RelPath rejected err=", err)
}
