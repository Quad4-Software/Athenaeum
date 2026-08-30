package library

import (
	"context"
	"os"
	"sync/atomic"

	"athenaeum/internal/libfs"
)

// IndexFile indexes a single file already present under a library mount.
func (s *Scanner) IndexFile(ctx context.Context, libraryID int64, relPath string) (int64, error) {
	fs, err := s.store.OpenLibraryFS(ctx, libraryID)
	if err != nil {
		return 0, err
	}
	info, err := fs.Stat(ctx, relPath)
	if err != nil {
		return 0, err
	}
	if info.IsDir {
		return 0, os.ErrInvalid
	}
	var abs string
	if fs.Backend() == libfs.BackendLocal {
		full, err := fs.LocalAbsPath(relPath)
		if err != nil {
			return 0, err
		}
		abs = full
	} else {
		abs = stringsTrimJoin(fs.RootLabel(), relPath)
	}
	j := job{
		libraryID: libraryID,
		root:      fs.RootLabel(),
		absPath:   abs,
		relPath:   relPath,
		size:      info.Size,
		mtime:     info.ModTime,
		backend:   fs.Backend(),
	}
	var idx, skip atomic.Int64
	if !s.process(ctx, fs, j, &idx, &skip) {
		return 0, os.ErrInvalid
	}
	book, err := s.store.GetBookByPath(ctx, libraryID, relPath)
	if err != nil {
		return 0, err
	}
	return book.ID, nil
}

func stringsTrimJoin(root, rel string) string {
	for len(root) > 0 && root[len(root)-1] == '/' {
		root = root[:len(root)-1]
	}
	return root + "/" + rel
}
