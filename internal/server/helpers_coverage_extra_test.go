package server

import (
	"net/http/httptest"
	"testing"
)

func TestStatusWriterFlushAndWebhookDeliveryID(t *testing.T) {
	rec := httptest.NewRecorder()
	sw := &statusWriter{ResponseWriter: rec, status: 200}
	sw.WriteHeader(201)
	if sw.status != 201 {
		t.Fatalf("status=%d", sw.status)
	}
	sw.Flush()

	id, err := newWebhookDeliveryID()
	if err != nil {
		t.Fatal(err)
	}
	if len(id) != 32 {
		t.Fatalf("id=%q", id)
	}
	id2, err := newWebhookDeliveryID()
	if err != nil {
		t.Fatal(err)
	}
	if id == id2 {
		t.Fatal("expected unique delivery ids")
	}
}
