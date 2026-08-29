package library

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func calibreAvailable() bool {
	_, err := exec.LookPath("ebook-convert")
	return err == nil
}

func calibreConvert(src, dest, target string) error {
	if !calibreAvailable() {
		return errors.New("calibre not installed")
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0o750); err != nil {
		return err
	}
	cmd := exec.Command("ebook-convert", src, dest) // #nosec G204
	out, err := cmd.CombinedOutput()
	if err != nil {
		return errors.New(strings.TrimSpace(string(out)))
	}
	return nil
}

// ConvertBook converts a library file to target format (epub or pdf).
func ConvertBook(srcPath, outDir, target string) (string, error) {
	base := strings.TrimSuffix(filepath.Base(srcPath), filepath.Ext(srcPath))
	dest := filepath.Join(outDir, base+"."+target)
	if err := calibreConvert(srcPath, dest, target); err == nil {
		return dest, nil
	}
	if target == "epub" {
		if err := ConvertToEPUB(srcPath, dest); err != nil {
			return "", err
		}
		return dest, nil
	}
	return "", errors.New("conversion requires calibre for this target format")
}
