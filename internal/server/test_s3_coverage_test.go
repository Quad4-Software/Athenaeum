package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/johannesboyne/gofakes3"
	"github.com/johannesboyne/gofakes3/backend/s3mem"

	"athenaeum/internal/models"
)

func TestHandleTestS3(t *testing.T) {
	srv, store := testServer(t)
	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}
	session, csrf := loginAdmin(t, handler, store)

	backend := s3mem.New()
	if err := backend.CreateBucket("books"); err != nil {
		t.Fatal(err)
	}
	faker := gofakes3.New(backend)
	ts := httptest.NewServer(faker.Server())
	t.Cleanup(ts.Close)
	endpoint := strings.TrimPrefix(ts.URL, "http://")

	body, _ := json.Marshal(models.LibraryS3Input{
		Endpoint:     endpoint,
		Region:       "us-east-1",
		Bucket:       "books",
		AccessKey:    "AKIAIOSFODNN7EXAMPLE",
		SecretKey:    "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		UsePathStyle: true,
		TLS:          false,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/libraries/test-s3", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(session)
	withCSRF(req, csrf)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("test-s3 status=%d body=%s", rec.Code, rec.Body.String())
	}

	bad, _ := json.Marshal(models.LibraryS3Input{
		Endpoint:     endpoint,
		Bucket:       "missing-bucket",
		AccessKey:    "AKIAIOSFODNN7EXAMPLE",
		SecretKey:    "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		UsePathStyle: true,
	})
	req = httptest.NewRequest(http.MethodPost, "/api/libraries/test-s3", bytes.NewReader(bad))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(session)
	withCSRF(req, csrf)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("test-s3 missing bucket status=%d", rec.Code)
	}

	_ = store
}
