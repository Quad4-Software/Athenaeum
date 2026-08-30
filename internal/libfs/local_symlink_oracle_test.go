package libfs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// PROVED_LOCALFS_SYMLINK_JAIL
// Guarantee: localFS Open/Stat/LocalAbsPath refuse symlink targets outside root.

func TestLocalFSSymlinkEscapeOracle(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "mount")
	outside := filepath.Join(base, "secret.epub")
	if err := os.Mkdir(root, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(outside, []byte("secret"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "leak.epub")); err != nil {
		t.Fatal(err)
	}

	fs, err := Open(Config{Backend: BackendLocal, Path: root})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fs.LocalAbsPath("leak.epub"); err == nil {
		t.Fatal("LocalAbsPath allowed symlink escape")
	}
	if _, err := fs.Stat(context.Background(), "leak.epub"); err == nil {
		t.Fatal("Stat allowed symlink escape")
	}
	if _, err := fs.Open(context.Background(), "leak.epub"); err == nil {
		t.Fatal("Open allowed symlink escape")
	}
	fmt.Println("PROVED_LOCALFS_SYMLINK_JAIL: symlink escape denied")
}
