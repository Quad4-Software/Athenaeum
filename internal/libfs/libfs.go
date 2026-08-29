// Package libfs provides filesystem backends for library mounts (local and S3).
package libfs

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
)

const (
	// BackendLocal is a directory on the host filesystem.
	BackendLocal = "local"
	// BackendS3 is a MinIO-compatible object store prefix.
	BackendS3 = "s3"
)

// ErrNotExist reports a missing object or path.
var ErrNotExist = errors.New("libfs: path does not exist")

// FileInfo describes one object or file under a library mount.
type FileInfo struct {
	RelPath string
	Size    int64
	ModTime time.Time
	IsDir   bool
}

// LibraryFS is the I/O surface for scanning and serving library content.
type LibraryFS interface {
	Backend() string
	RootLabel() string
	Stat(ctx context.Context, relPath string) (FileInfo, error)
	Open(ctx context.Context, relPath string) (io.ReadSeekCloser, error)
	Walk(ctx context.Context, fn func(FileInfo) error) error
	Write(ctx context.Context, relPath string, r io.Reader, size int64) error
	Remove(ctx context.Context, relPath string) error
	MkdirAll(ctx context.Context, relPath string) error
}

// Config selects and configures a library backend.
type Config struct {
	Backend string
	Path    string
	S3      S3Config
}

// S3Config holds MinIO-compatible connection settings.
type S3Config struct {
	Endpoint     string `json:"endpoint"`
	Region       string `json:"region"`
	Bucket       string `json:"bucket"`
	Prefix       string `json:"prefix"`
	AccessKey    string `json:"accessKey"`
	SecretKey    string `json:"secretKey"`
	UsePathStyle bool   `json:"usePathStyle"`
	TLS          bool   `json:"tls"`
}

// Open constructs a LibraryFS for the given mount config.
func Open(cfg Config) (LibraryFS, error) {
	backend := strings.TrimSpace(strings.ToLower(cfg.Backend))
	if backend == "" {
		backend = BackendLocal
	}
	switch backend {
	case BackendLocal:
		return newLocalFS(cfg.Path)
	case BackendS3:
		return newS3FS(cfg.S3)
	default:
		return nil, fmt.Errorf("unsupported library backend %q", cfg.Backend)
	}
}

// MountLabel builds the display mount_path for an S3 library.
func MountLabel(bucket, prefix string) string {
	bucket = strings.TrimSpace(bucket)
	prefix = strings.Trim(strings.TrimSpace(prefix), "/")
	if prefix == "" {
		return "s3://" + bucket
	}
	return "s3://" + bucket + "/" + prefix
}

// NormalizeRelPath converts OS paths to slash-separated relative keys.
func NormalizeRelPath(p string) string {
	p = strings.ReplaceAll(p, "\\", "/")
	p = strings.TrimPrefix(p, "/")
	return p
}
