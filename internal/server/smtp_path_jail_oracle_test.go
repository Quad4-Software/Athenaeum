package server

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"athenaeum/internal/library"
)

// PROVED_SMTP_READFILE_PATH_JAIL
// Guarantee: library RelPath with .. must not let smtp/content-index
// resolve a path outside the mount.

func TestSMTPPathJoinEscapesMountOracle(t *testing.T) {
	base := t.TempDir()
	mount := filepath.Join(base, "mount")
	if err := os.Mkdir(mount, 0o755); err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(base, "secret.epub")
	if err := os.WriteFile(outside, []byte("SECRET-BYTES"), 0o600); err != nil {
		t.Fatal(err)
	}

	rel := "../secret.epub"
	joined := library.ResolveBookAbsPath(mount, rel)
	if joined != "" {
		t.Fatalf("escape path returned: %q", joined)
	}
	fmt.Println("PROVED_SMTP_READFILE_PATH_JAIL: ResolveBookAbsPath denied traversal")
}
