package server

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"athenaeum/internal/library"
)

// PROVED_SMTP_READFILE_PATH_JAIL
// Guarantee claimed: library RelPath with .. must not let smtp/content-index
// ReadFile/Open escape the mount (incomplete OpenRoot sibling).

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
	data, err := os.ReadFile(joined)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if string(data) != "SECRET-BYTES" {
		t.Fatalf("unexpected data %q", data)
	}
	fmt.Println("PROVED_SMTP_READFILE_PATH_JAIL: ReadFile via Join read outside mount")
}
