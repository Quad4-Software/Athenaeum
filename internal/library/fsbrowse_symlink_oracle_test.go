package library

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// PROVED_FSBROWSE_SYMLINK_JAIL
// Guarantee claimed: BrowseDirs must not accept paths that are only inside
// roots via a symlink whose target is outside those roots.
// Oracle: expected deny. Actual: accepts and returns the symlink path.

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

	res, err := BrowseDirs([]string{root}, link)
	if err != nil {
		t.Fatalf("not vulnerable (symlink denied): %v", err)
	}
	if res.Path != link {
		t.Fatalf("unexpected path %q", res.Path)
	}
	fmt.Println("PROVED_FSBROWSE_SYMLINK_JAIL: BrowseDirs followed symlink out of jail path=", res.Path)
}
