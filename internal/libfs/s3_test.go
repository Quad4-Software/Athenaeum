package libfs

import (
	"context"
	"errors"
	"net/http"
	"os"
	"testing"

	"github.com/minio/minio-go/v7"
)

func TestValidateS3Config(t *testing.T) {
	t.Parallel()
	err := ValidateS3Config(S3Config{}, true)
	if err == nil {
		t.Fatal("expected error")
	}
	cfg := S3Config{
		Endpoint:  "localhost:9000",
		Bucket:    "books",
		AccessKey: "ak",
		SecretKey: "sk",
		TLS:       false,
	}
	if err := ValidateS3Config(cfg, true); err != nil {
		t.Fatal(err)
	}
	cfg.SecretKey = ""
	if err := ValidateS3Config(cfg, true); err == nil {
		t.Fatal("expected secret required")
	}
	if err := ValidateS3Config(cfg, false); err != nil {
		t.Fatal(err)
	}
}

func TestNormalizeS3Prefix(t *testing.T) {
	t.Parallel()
	cfg := normalizeS3Config(S3Config{
		Endpoint: "https://minio.example/",
		Bucket:   "b",
		Prefix:   "/lib/",
	})
	if cfg.Endpoint != "minio.example" {
		t.Fatalf("endpoint %q", cfg.Endpoint)
	}
	if cfg.Prefix != "lib/" {
		t.Fatalf("prefix %q", cfg.Prefix)
	}
}

func TestS3ObjectKey(t *testing.T) {
	t.Parallel()
	f := &s3FS{prefix: "lib/"}
	key, err := f.objectKey("a/b.epub")
	if err != nil {
		t.Fatal(err)
	}
	if key != "lib/a/b.epub" {
		t.Fatalf("key %q", key)
	}
	if _, err := f.objectKey("../x"); err == nil {
		t.Fatal("expected escape error")
	}
	if _, err := f.objectKey(""); err == nil {
		t.Fatal("expected empty path error")
	}
}

func TestS3FSMetaAndMkdirAll(t *testing.T) {
	t.Parallel()
	f := &s3FS{label: "s3://books/lib", prefix: "lib/"}
	if f.Backend() != BackendS3 {
		t.Fatalf("backend %q", f.Backend())
	}
	if f.RootLabel() != "s3://books/lib" {
		t.Fatalf("label %q", f.RootLabel())
	}
	if err := f.MkdirAll(context.Background(), "ignored"); err != nil {
		t.Fatal(err)
	}
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	if err := f.MkdirAll(canceled, "x"); err == nil {
		t.Fatal("expected canceled context error")
	}
}

func TestIsS3NotFound(t *testing.T) {
	t.Parallel()
	if isS3NotFound(nil) {
		t.Fatal("nil should not be not-found")
	}
	if isS3NotFound(errors.New("other")) {
		t.Fatal("generic error")
	}
	if !isS3NotFound(errors.New("The specified key does not exist")) {
		t.Fatal("string match")
	}
	if !isS3NotFound(minio.ErrorResponse{StatusCode: http.StatusNotFound, Code: "NoSuchKey"}) {
		t.Fatal("status not found")
	}
	if !isS3NotFound(minio.ErrorResponse{Code: "NotFound"}) {
		t.Fatal("NotFound code")
	}
	if isS3NotFound(minio.ErrorResponse{StatusCode: http.StatusForbidden, Code: "AccessDenied"}) {
		t.Fatal("access denied is not not-found")
	}
}

func TestS3OpenAndTestValidation(t *testing.T) {
	t.Parallel()
	if err := TestS3(context.Background(), S3Config{}); err == nil {
		t.Fatal("expected validation error")
	}
	if _, err := Open(Config{Backend: BackendS3, S3: S3Config{}}); err == nil {
		t.Fatal("expected newS3FS validation error")
	}
	if endpoint := os.Getenv("TEST_S3_ENDPOINT"); endpoint == "" {
		t.Log("TEST_S3_* unset; skipping live S3 round-trip")
		return
	}
}
