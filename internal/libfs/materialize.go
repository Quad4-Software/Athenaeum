package libfs

import (
	"context"
	"io"
	"os"
	"path/filepath"
)

// Materialize copies relPath from fs into a temp file under dir and returns its path.
// The caller must remove the file when finished.
func Materialize(ctx context.Context, fs LibraryFS, relPath, dir string) (string, error) {
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return "", err
	}
	src, err := fs.Open(ctx, relPath)
	if err != nil {
		return "", err
	}
	defer src.Close()

	ext := filepath.Ext(relPath)
	tmp, err := os.CreateTemp(dir, "athenaeum-lib-*"+ext)
	if err != nil {
		return "", err
	}
	tmpPath := tmp.Name()
	if _, err := io.Copy(tmp, src); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return "", err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return "", err
	}
	return tmpPath, nil
}
