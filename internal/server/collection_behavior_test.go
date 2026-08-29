package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"athenaeum/internal/models"
)

func TestCollectionMembershipRoundTrip(t *testing.T) {
	srv, c := newAdminClient(t)
	_, bookID := seedLibraryBook(t, srv, c.store, "coll", []byte("%PDF-1.4 coll"))

	rec := c.do(http.MethodPost, "/api/collections", map[string]any{
		"name": "Shelf", "description": "d",
	})
	c.mustStatus(rec, http.StatusCreated)
	var coll models.Collection
	if err := json.NewDecoder(rec.Body).Decode(&coll); err != nil || coll.ID == 0 || coll.Name != "Shelf" {
		t.Fatalf("create=%+v err=%v", coll, err)
	}

	c.mustStatus(c.do(http.MethodPut, fmt.Sprintf("/api/collections/%d", coll.ID), map[string]any{
		"name": "Shelf 2", "description": "d2",
	}), http.StatusOK)

	rec = c.do(http.MethodGet, fmt.Sprintf("/api/collections/%d", coll.ID), nil)
	c.mustStatus(rec, http.StatusOK)
	coll = models.Collection{}
	if err := json.NewDecoder(rec.Body).Decode(&coll); err != nil || coll.Name != "Shelf 2" {
		t.Fatalf("get after rename=%+v err=%v", coll, err)
	}

	c.mustStatus(c.do(http.MethodPost, fmt.Sprintf("/api/collections/%d/books/%d", coll.ID, bookID), nil), http.StatusNoContent)
	rec = c.do(http.MethodGet, fmt.Sprintf("/api/collections/%d", coll.ID), nil)
	c.mustStatus(rec, http.StatusOK)
	if err := json.NewDecoder(rec.Body).Decode(&coll); err != nil {
		t.Fatal(err)
	}
	if coll.BookCount < 1 {
		t.Fatalf("expected book in collection, count=%d", coll.BookCount)
	}

	c.mustStatus(c.do(http.MethodDelete, fmt.Sprintf("/api/collections/%d/books/%d", coll.ID, bookID), nil), http.StatusNoContent)
	rec = c.do(http.MethodGet, fmt.Sprintf("/api/collections/%d", coll.ID), nil)
	c.mustStatus(rec, http.StatusOK)
	if err := json.NewDecoder(rec.Body).Decode(&coll); err != nil {
		t.Fatal(err)
	}
	if coll.BookCount != 0 {
		t.Fatalf("expected empty collection after remove, count=%d", coll.BookCount)
	}

	c.mustStatus(c.do(http.MethodDelete, fmt.Sprintf("/api/collections/%d", coll.ID), nil), http.StatusNoContent)
	rec = c.do(http.MethodGet, fmt.Sprintf("/api/collections/%d", coll.ID), nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("deleted collection status=%d", rec.Code)
	}
}
