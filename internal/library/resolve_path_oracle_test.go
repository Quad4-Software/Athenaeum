package library

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// PROVED_RESOLVE_BOOK_PATH_JAIL
// Guarantee: ResolveBookAbsPath stays under mount for hostile RelPath.
// Expected: empty string (denied). Actual before fix: filepath.Join escaped.

func TestResolveBookAbsPathDotDotOracle(t *testing.T) {
	base := t.TempDir()
	mount := filepath.Join(base, "mount")
	if err := os.Mkdir(mount, 0o755); err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(base, "outside-secret")
	if err := os.WriteFile(outside, []byte("pwned"), 0o600); err != nil {
		t.Fatal(err)
	}

	got := ResolveBookAbsPath(mount, "../outside-secret")
	if got != "" {
		t.Fatalf("escape allowed: got %q", got)
	}
	fmt.Println("PROVED_RESOLVE_BOOK_PATH_JAIL: traversal RelPath rejected")
}
