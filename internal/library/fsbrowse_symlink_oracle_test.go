package library

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// PROVED_FSBROWSE_SYMLINK_JAIL
// Guarantee: BrowseDirs must not accept paths that only appear inside roots
// via a symlink whose resolved target is outside those roots.
// Expected: deny. Actual before fix: accepted the symlink path.

func TestBrowseDirsSymlinkEscapeOracle(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "mount")
	secret := filepath.Join(base, "secret")
	if err := os.MkdirAll(filepath.Join(root, "books"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(secret, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(secret, "private.txt"), []byte("secret"), 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "escape")
	if err := os.Symlink(secret, link); err != nil {
		t.Fatal(err)
	}

	_, err := BrowseDirs([]string{root}, link)
	if err == nil {
		t.Fatal("BrowseDirs accepted symlink escape")
	}
	fmt.Println("PROVED_FSBROWSE_SYMLINK_JAIL: symlink escape denied err=", err)
}
