package libfs

import (
	"bytes"
	"context"
	"io"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/johannesboyne/gofakes3"
	"github.com/johannesboyne/gofakes3/backend/s3mem"
)

func TestS3FSRoundTrip(t *testing.T) {
	backend := s3mem.New()
	if err := backend.CreateBucket("books"); err != nil {
		t.Fatal(err)
	}
	faker := gofakes3.New(backend)
	ts := httptest.NewServer(faker.Server())
	t.Cleanup(ts.Close)

	endpoint := strings.TrimPrefix(ts.URL, "http://")
	fs, err := Open(Config{
		Backend: BackendS3,
		S3: S3Config{
			Endpoint:     endpoint,
			Bucket:       "books",
			Prefix:       "lib",
			AccessKey:    "AKIAIOSFODNN7EXAMPLE",
			SecretKey:    "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
			Region:       "us-east-1",
			UsePathStyle: true,
			TLS:          false,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	payload := []byte("s3-round-trip-bytes")
	if err := fs.Write(ctx, "dir/a.pdf", bytes.NewReader(payload), int64(len(payload))); err != nil {
		t.Fatal(err)
	}

	info, err := fs.Stat(ctx, "dir/a.pdf")
	if err != nil {
		t.Fatal(err)
	}
	if info.Size != int64(len(payload)) || info.RelPath != "dir/a.pdf" {
		t.Fatalf("stat=%+v", info)
	}
	if _, err := fs.Stat(ctx, "missing.pdf"); err != ErrNotExist {
		t.Fatalf("missing stat err=%v", err)
	}

	rc, err := fs.Open(ctx, "dir/a.pdf")
	if err != nil {
		t.Fatal(err)
	}
	data, err := io.ReadAll(rc)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != string(payload) {
		t.Fatalf("read=%q", data)
	}
	if _, err := rc.Seek(3, io.SeekStart); err != nil {
		t.Fatal(err)
	}
	part := make([]byte, 4)
	n, err := rc.Read(part)
	if err != nil && err != io.EOF {
		t.Fatal(err)
	}
	if n != 4 || string(part) != "roun" {
		t.Fatalf("seek read n=%d part=%q", n, part)
	}
	if _, err := rc.Seek(0, io.SeekEnd); err != nil {
		t.Fatal(err)
	}
	if _, err := rc.Seek(-2, io.SeekCurrent); err != nil {
		t.Fatal(err)
	}
	if err := rc.Close(); err != nil {
		t.Fatal(err)
	}

	seen := 0
	if err := fs.Walk(ctx, func(fi FileInfo) error {
		seen++
		if fi.RelPath != "dir/a.pdf" {
			t.Fatalf("walk=%q", fi.RelPath)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if seen != 1 {
		t.Fatalf("seen=%d", seen)
	}

	tmp, err := Materialize(ctx, fs, "dir/a.pdf", t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Remove(tmp) })
	got, err := os.ReadFile(tmp)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(payload) {
		t.Fatalf("materialize=%q", got)
	}

	if err := fs.Remove(ctx, "dir/a.pdf"); err != nil {
		t.Fatal(err)
	}
	// S3 delete is idempotent; missing keys typically succeed without ErrNotExist.
	_ = fs.Remove(ctx, "dir/a.pdf")
	if _, err := fs.Stat(ctx, "dir/a.pdf"); err != ErrNotExist {
		t.Fatalf("stat after remove err=%v", err)
	}

	if err := TestS3(ctx, S3Config{
		Endpoint:     endpoint,
		Bucket:       "books",
		AccessKey:    "AKIAIOSFODNN7EXAMPLE",
		SecretKey:    "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		Region:       "us-east-1",
		UsePathStyle: true,
	}); err != nil {
		t.Fatal(err)
	}
	if err := TestS3(ctx, S3Config{
		Endpoint:     endpoint,
		Bucket:       "no-such-bucket",
		AccessKey:    "AKIAIOSFODNN7EXAMPLE",
		SecretKey:    "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		UsePathStyle: true,
	}); err == nil {
		t.Fatal("expected missing bucket error")
	}
}

func TestS3SeekReaderEdges(t *testing.T) {
	t.Parallel()
	r := &s3SeekReader{size: 10, pos: 0}
	if _, err := r.Seek(-1, io.SeekStart); err == nil {
		t.Fatal("negative seek")
	}
	if _, err := r.Seek(0, 99); err == nil {
		t.Fatal("bad whence")
	}
	pos, err := r.Seek(4, io.SeekStart)
	if err != nil || pos != 4 {
		t.Fatalf("seek start pos=%d err=%v", pos, err)
	}
	pos, err = r.Seek(1, io.SeekCurrent)
	if err != nil || pos != 5 {
		t.Fatalf("seek cur pos=%d err=%v", pos, err)
	}
	pos, err = r.Seek(-2, io.SeekEnd)
	if err != nil || pos != 8 {
		t.Fatalf("seek end pos=%d err=%v", pos, err)
	}
	if err := r.openRange(9, 8); err != io.EOF {
		t.Fatalf("openRange inverted err=%v", err)
	}
	if err := r.openRange(10, 10); err != io.EOF {
		t.Fatalf("openRange past end err=%v", err)
	}
	if r.Size() != 10 {
		t.Fatalf("size=%d", r.Size())
	}
	_ = r.ModTime()
	if err := r.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestMaterializeErrors(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	fs, err := Open(Config{Backend: BackendLocal, Path: root})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Materialize(context.Background(), fs, "missing.pdf", t.TempDir()); err == nil {
		t.Fatal("expected missing file error")
	}
}
