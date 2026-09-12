// Package backup builds and restores zip archives of the Athenaeum data
// directory for the admin backup routes.
package backup

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// PostgresNotice is written as DATABASE.txt in place of a database file when
// Athenaeum runs on PostgreSQL.
const PostgresNotice = "Athenaeum is using PostgreSQL. Back up the database with pg_dump separately.\n"

// Item is a single archive member: a filesystem file or directory when Path
// is set, or literal bytes when Path is empty.
type Item struct {
	Name string
	Path string
	Data []byte
}

// ClientError marks archive problems caused by the uploaded file rather than
// the server. Handlers should map it to 400.
type ClientError struct {
	msg string
}

func (e *ClientError) Error() string { return e.msg }

func clientError(msg string) *ClientError { return &ClientError{msg: msg} }

// Write streams a zip archive of items to w in order. Items whose Path does
// not exist are skipped.
func Write(w io.Writer, items []Item) error {
	zw := zip.NewWriter(w)
	for _, it := range items {
		var err error
		if it.Path != "" {
			err = addPath(zw, it.Name, it.Path)
		} else {
			err = addData(zw, it.Name, it.Data)
		}
		if err != nil {
			_ = zw.Close()
			return err
		}
	}
	return zw.Close()
}

func addPath(zw *zip.Writer, name, path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if info.IsDir() {
		return filepath.Walk(path, func(p string, fi os.FileInfo, err error) error {
			if err != nil || fi.IsDir() {
				return err
			}
			rel, err := filepath.Rel(path, p)
			if err != nil {
				return err
			}
			return AddFile(zw, filepath.ToSlash(filepath.Join(name, rel)), p)
		})
	}
	return AddFile(zw, name, path)
}

// AddFile writes a single filesystem file into the archive under name.
func AddFile(zw *zip.Writer, name, path string) error {
	f, err := os.Open(path) // #nosec G304 -- admin backup reads fixed server paths
	if err != nil {
		return err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return err
	}
	hdr, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	hdr.Name = name
	hdr.Method = zip.Deflate
	w, err := zw.CreateHeader(hdr)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, f)
	return err
}

func addData(zw *zip.Writer, name string, data []byte) error {
	fw, err := zw.Create(name)
	if err != nil {
		return err
	}
	_, _ = fw.Write(data)
	return nil
}

// Restore extracts a zip archive read from src into dataDir. Entries that
// would escape dataDir are skipped. The archive must contain dbFilename at
// the top level or inside a single directory.
func Restore(src io.Reader, dataDir, dbFilename string) error {
	tmp, err := os.CreateTemp(dataDir, "restore-*.zip")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	if _, err := io.Copy(tmp, src); err != nil {
		_ = tmp.Close()
		return err
	}
	_ = tmp.Close()

	zr, err := zip.OpenReader(tmpPath)
	if err != nil {
		return clientError("invalid zip archive")
	}
	defer zr.Close()
	hasDB := false
	for _, f := range zr.File {
		if f.Name == dbFilename || strings.HasSuffix(f.Name, "/"+dbFilename) {
			hasDB = true
			break
		}
	}
	if !hasDB {
		return clientError("archive must contain " + dbFilename)
	}
	for _, f := range zr.File {
		if f.FileInfo().IsDir() {
			continue
		}
		name := filepath.Clean(filepath.FromSlash(f.Name))
		if name == "." || name == "" || strings.HasPrefix(name, ".."+string(os.PathSeparator)) || name == ".." {
			continue
		}
		if filepath.IsAbs(name) {
			continue
		}
		dest := filepath.Join(dataDir, name)
		rel, err := filepath.Rel(dataDir, dest)
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
			continue
		}
		if err := os.MkdirAll(filepath.Dir(dest), 0o750); err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		out, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600) // #nosec G304 -- dest validated under dataDir
		if err != nil {
			_ = rc.Close()
			return err
		}
		_, err = io.Copy(out, rc) // #nosec G110 -- admin-only restore; entries validated under dataDir
		_ = out.Close()
		_ = rc.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
