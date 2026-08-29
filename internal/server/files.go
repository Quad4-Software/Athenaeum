package server

import (
	"os"
	"strconv"
)

// openCoverFile opens a cached cover image by book id, scoped to the cover
// directory so path traversal is not possible even if ids were tampered with.
func openCoverFile(coverDir string, id int64) (*os.File, os.FileInfo, error) {
	root, err := os.OpenRoot(coverDir)
	if err != nil {
		return nil, nil, err
	}
	defer root.Close()

	name := strconv.FormatInt(id, 10) + ".img"
	f, err := root.Open(name)
	if err != nil {
		return nil, nil, err
	}
	info, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return nil, nil, err
	}
	return f, info, nil
}

func writeCoverFile(coverDir string, id int64, data []byte) error {
	root, err := os.OpenRoot(coverDir)
	if err != nil {
		return err
	}
	defer root.Close()
	return root.WriteFile(strconv.FormatInt(id, 10)+".img", data, 0o600)
}

func removeCoverFile(coverDir string, id int64) error {
	root, err := os.OpenRoot(coverDir)
	if err != nil {
		return err
	}
	defer root.Close()
	err = root.Remove(strconv.FormatInt(id, 10) + ".img")
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
