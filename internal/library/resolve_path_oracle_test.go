package library

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// PROVED_RESOLVE_BOOK_PATH_JAIL
// Guarantee claimed: ResolveBookAbsPath (used by smtp ReadFile and
// content-index zip open) stays under mount for hostile RelPath.
// Oracle: expected confined. Actual: filepath.Join follows .. out of mount.

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
	if filepath.Clean(got) != filepath.Clean(outside) {
		t.Fatalf("not vulnerable: got %q want escape to %q", got, outside)
	}
	rel, err := filepath.Rel(mount, got)
	if err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		t.Fatalf("unexpected confined rel=%q", rel)
	}
	fmt.Println("PROVED_RESOLVE_BOOK_PATH_JAIL: Join escaped mount to", got)
}
