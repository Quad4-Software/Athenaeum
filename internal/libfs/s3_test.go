package libfs

import (
	"testing"
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
}
