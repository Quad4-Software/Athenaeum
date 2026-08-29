package library

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
)

// FileSHA256 returns the hex-encoded SHA-256 digest of a file.
func FileSHA256(path string) (string, error) {
	f, err := os.Open(path) // #nosec G304 -- path is library-relative from scanner
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
