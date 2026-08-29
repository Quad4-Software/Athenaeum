package libfs

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type s3FS struct {
	client *minio.Client
	bucket string
	prefix string
	label  string
	cfg    S3Config
}

func newS3FS(cfg S3Config) (*s3FS, error) {
	cfg = normalizeS3Config(cfg)
	if err := ValidateS3Config(cfg, true); err != nil {
		return nil, err
	}
	client, err := minioClient(cfg)
	if err != nil {
		return nil, err
	}
	return &s3FS{
		client: client,
		bucket: cfg.Bucket,
		prefix: cfg.Prefix,
		label:  MountLabel(cfg.Bucket, cfg.Prefix),
		cfg:    cfg,
	}, nil
}

func normalizeS3Config(cfg S3Config) S3Config {
	cfg.Endpoint = strings.TrimSpace(cfg.Endpoint)
	cfg.Endpoint = strings.TrimPrefix(cfg.Endpoint, "https://")
	cfg.Endpoint = strings.TrimPrefix(cfg.Endpoint, "http://")
	cfg.Endpoint = strings.TrimSuffix(cfg.Endpoint, "/")
	cfg.Region = strings.TrimSpace(cfg.Region)
	if cfg.Region == "" {
		cfg.Region = "us-east-1"
	}
	cfg.Bucket = strings.TrimSpace(cfg.Bucket)
	cfg.Prefix = strings.Trim(strings.TrimSpace(cfg.Prefix), "/")
	if cfg.Prefix != "" {
		cfg.Prefix += "/"
	}
	cfg.AccessKey = strings.TrimSpace(cfg.AccessKey)
	return cfg
}

// ValidateS3Config checks required S3 fields. requireSecret enforces secretKey.
func ValidateS3Config(cfg S3Config, requireSecret bool) error {
	cfg = normalizeS3Config(cfg)
	if cfg.Endpoint == "" {
		return errors.New("s3 endpoint is required")
	}
	if cfg.Bucket == "" {
		return errors.New("s3 bucket is required")
	}
	if cfg.AccessKey == "" {
		return errors.New("s3 accessKey is required")
	}
	if requireSecret && cfg.SecretKey == "" {
		return errors.New("s3 secretKey is required")
	}
	return nil
}

func minioClient(cfg S3Config) (*minio.Client, error) {
	cfg = normalizeS3Config(cfg)
	opts := &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.TLS,
		Region: cfg.Region,
	}
	if cfg.UsePathStyle {
		opts.BucketLookup = minio.BucketLookupPath
	} else {
		opts.BucketLookup = minio.BucketLookupDNS
	}
	return minio.New(cfg.Endpoint, opts)
}

// TestS3 verifies credentials can reach the bucket.
func TestS3(ctx context.Context, cfg S3Config) error {
	cfg = normalizeS3Config(cfg)
	if err := ValidateS3Config(cfg, true); err != nil {
		return err
	}
	client, err := minioClient(cfg)
	if err != nil {
		return err
	}
	exists, err := client.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		return fmt.Errorf("s3 bucket check failed: %w", err)
	}
	if !exists {
		return fmt.Errorf("s3 bucket %q does not exist or is not accessible", cfg.Bucket)
	}
	return nil
}

func (f *s3FS) Backend() string   { return BackendS3 }
func (f *s3FS) RootLabel() string { return f.label }

func (f *s3FS) objectKey(relPath string) (string, error) {
	relPath = NormalizeRelPath(relPath)
	if relPath == "" {
		return "", errors.New("empty object path")
	}
	if strings.Contains(relPath, "..") {
		return "", errors.New("path escapes mount")
	}
	return f.prefix + relPath, nil
}

func (f *s3FS) Stat(ctx context.Context, relPath string) (FileInfo, error) {
	key, err := f.objectKey(relPath)
	if err != nil {
		return FileInfo{}, err
	}
	info, err := f.client.StatObject(ctx, f.bucket, key, minio.StatObjectOptions{})
	if err != nil {
		if isS3NotFound(err) {
			return FileInfo{}, ErrNotExist
		}
		return FileInfo{}, err
	}
	return FileInfo{
		RelPath: NormalizeRelPath(relPath),
		Size:    info.Size,
		ModTime: info.LastModified,
		IsDir:   false,
	}, nil
}

func (f *s3FS) Open(ctx context.Context, relPath string) (io.ReadSeekCloser, error) {
	info, err := f.Stat(ctx, relPath)
	if err != nil {
		return nil, err
	}
	key, err := f.objectKey(relPath)
	if err != nil {
		return nil, err
	}
	return &s3SeekReader{
		ctx:    ctx,
		client: f.client,
		bucket: f.bucket,
		key:    key,
		size:   info.Size,
		mod:    info.ModTime,
	}, nil
}

func (f *s3FS) Walk(ctx context.Context, fn func(FileInfo) error) error {
	opts := minio.ListObjectsOptions{
		Prefix:    f.prefix,
		Recursive: true,
	}
	for obj := range f.client.ListObjects(ctx, f.bucket, opts) {
		if obj.Err != nil {
			return obj.Err
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if strings.HasSuffix(obj.Key, "/") {
			continue
		}
		rel := strings.TrimPrefix(obj.Key, f.prefix)
		rel = NormalizeRelPath(rel)
		if rel == "" {
			continue
		}
		base := rel
		if i := strings.LastIndex(rel, "/"); i >= 0 {
			base = rel[i+1:]
		}
		if strings.HasPrefix(base, ".") {
			continue
		}
		if err := fn(FileInfo{
			RelPath: rel,
			Size:    obj.Size,
			ModTime: obj.LastModified,
			IsDir:   false,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (f *s3FS) Write(ctx context.Context, relPath string, r io.Reader, size int64) error {
	key, err := f.objectKey(relPath)
	if err != nil {
		return err
	}
	if size < 0 {
		size = -1
	}
	_, err = f.client.PutObject(ctx, f.bucket, key, r, size, minio.PutObjectOptions{})
	return err
}

func (f *s3FS) Remove(ctx context.Context, relPath string) error {
	key, err := f.objectKey(relPath)
	if err != nil {
		return err
	}
	err = f.client.RemoveObject(ctx, f.bucket, key, minio.RemoveObjectOptions{})
	if isS3NotFound(err) {
		return ErrNotExist
	}
	return err
}

func (f *s3FS) MkdirAll(ctx context.Context, relPath string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	_ = relPath
	return nil
}

func isS3NotFound(err error) bool {
	if err == nil {
		return false
	}
	var resp minio.ErrorResponse
	if errors.As(err, &resp) {
		return resp.StatusCode == http.StatusNotFound || resp.Code == "NoSuchKey" || resp.Code == "NotFound"
	}
	return strings.Contains(err.Error(), "The specified key does not exist")
}

// s3SeekReader implements io.ReadSeekCloser via ranged GetObject calls.
type s3SeekReader struct {
	ctx    context.Context
	client *minio.Client
	bucket string
	key    string
	size   int64
	mod    time.Time
	pos    int64
	body   io.ReadCloser
	bodyAt int64
	bodyTo int64
}

func (r *s3SeekReader) Read(p []byte) (int, error) {
	if r.pos >= r.size {
		return 0, io.EOF
	}
	if r.body == nil || r.pos < r.bodyAt || r.pos >= r.bodyTo {
		if err := r.openRange(r.pos, r.size-1); err != nil {
			return 0, err
		}
	}
	n, err := r.body.Read(p)
	r.pos += int64(n)
	r.bodyAt += int64(n)
	if errors.Is(err, io.EOF) && r.pos < r.size {
		err = nil
	}
	return n, err
}

func (r *s3SeekReader) Seek(offset int64, whence int) (int64, error) {
	var abs int64
	switch whence {
	case io.SeekStart:
		abs = offset
	case io.SeekCurrent:
		abs = r.pos + offset
	case io.SeekEnd:
		abs = r.size + offset
	default:
		return 0, errors.New("invalid seek whence")
	}
	if abs < 0 {
		return 0, errors.New("negative seek")
	}
	if abs != r.pos {
		_ = r.closeBody()
		r.pos = abs
	}
	return r.pos, nil
}

func (r *s3SeekReader) Close() error {
	return r.closeBody()
}

func (r *s3SeekReader) ModTime() time.Time { return r.mod }
func (r *s3SeekReader) Size() int64        { return r.size }

func (r *s3SeekReader) openRange(start, end int64) error {
	_ = r.closeBody()
	if start > end || start >= r.size {
		return io.EOF
	}
	opts := minio.GetObjectOptions{}
	if err := opts.SetRange(start, end); err != nil {
		return err
	}
	obj, err := r.client.GetObject(r.ctx, r.bucket, r.key, opts)
	if err != nil {
		return err
	}
	r.body = obj
	r.bodyAt = start
	r.bodyTo = end + 1
	return nil
}

func (r *s3SeekReader) closeBody() error {
	if r.body == nil {
		return nil
	}
	err := r.body.Close()
	r.body = nil
	return err
}
