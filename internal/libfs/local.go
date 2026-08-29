package libfs

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type localFS struct {
	root string
}

func newLocalFS(path string) (*localFS, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(abs)
	if err != nil {
		return nil, errors.New("mount path must exist and be readable")
	}
	if !info.IsDir() {
		return nil, errors.New("mount path must be a directory")
	}
	return &localFS{root: abs}, nil
}

func (f *localFS) Backend() string   { return BackendLocal }
func (f *localFS) RootLabel() string { return f.root }

func (f *localFS) resolve(relPath string) (string, error) {
	relPath = NormalizeRelPath(relPath)
	if relPath == "" {
		return f.root, nil
	}
	if strings.Contains(relPath, "..") {
		clean := filepath.Clean(filepath.FromSlash(relPath))
		if clean == ".." || strings.HasPrefix(clean, ".."+string(os.PathSeparator)) {
			return "", errors.New("path escapes mount")
		}
	}
	full := filepath.Join(f.root, filepath.FromSlash(relPath))
	rel, err := filepath.Rel(f.root, full)
	if err != nil || strings.HasPrefix(rel, "..") {
		return "", errors.New("path escapes mount")
	}
	return full, nil
}

func (f *localFS) Stat(ctx context.Context, relPath string) (FileInfo, error) {
	if err := ctx.Err(); err != nil {
		return FileInfo{}, err
	}
	full, err := f.resolve(relPath)
	if err != nil {
		return FileInfo{}, err
	}
	info, err := os.Stat(full)
	if err != nil {
		if os.IsNotExist(err) {
			return FileInfo{}, ErrNotExist
		}
		return FileInfo{}, err
	}
	return FileInfo{
		RelPath: NormalizeRelPath(relPath),
		Size:    info.Size(),
		ModTime: info.ModTime(),
		IsDir:   info.IsDir(),
	}, nil
}

func (f *localFS) Open(ctx context.Context, relPath string) (io.ReadSeekCloser, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	full, err := f.resolve(relPath)
	if err != nil {
		return nil, err
	}
	file, err := os.Open(full) // #nosec G304 -- path jails via resolve
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotExist
		}
		return nil, err
	}
	return file, nil
}

func (f *localFS) Walk(ctx context.Context, fn func(FileInfo) error) error {
	return filepath.WalkDir(f.root, func(p string, d os.DirEntry, werr error) error {
		if werr != nil {
			return nil
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if d.IsDir() {
			if strings.HasPrefix(d.Name(), ".") && p != f.root {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasPrefix(d.Name(), ".") {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		rel, err := filepath.Rel(f.root, p)
		if err != nil {
			return nil
		}
		return fn(FileInfo{
			RelPath: filepath.ToSlash(rel),
			Size:    info.Size(),
			ModTime: info.ModTime(),
			IsDir:   false,
		})
	})
}

func (f *localFS) Write(ctx context.Context, relPath string, r io.Reader, size int64) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	full, err := f.resolve(relPath)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(full), 0o750); err != nil {
		return err
	}
	tmp := full + ".tmp"
	out, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o640) // #nosec G304 G302 -- path jails via resolve library files group-readable
	if err != nil {
		return err
	}
	var written int64
	if size > 0 {
		written, err = io.Copy(out, io.LimitReader(r, size))
	} else {
		written, err = io.Copy(out, r)
	}
	closeErr := out.Close()
	if err != nil {
		_ = os.Remove(tmp)
		return err
	}
	if closeErr != nil {
		_ = os.Remove(tmp)
		return closeErr
	}
	_ = written
	return os.Rename(tmp, full)
}

func (f *localFS) Remove(ctx context.Context, relPath string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	full, err := f.resolve(relPath)
	if err != nil {
		return err
	}
	err = os.Remove(full)
	if os.IsNotExist(err) {
		return ErrNotExist
	}
	return err
}

func (f *localFS) MkdirAll(ctx context.Context, relPath string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	full, err := f.resolve(relPath)
	if err != nil {
		return err
	}
	return os.MkdirAll(full, 0o750)
}
