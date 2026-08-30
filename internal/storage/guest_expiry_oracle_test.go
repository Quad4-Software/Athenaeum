package storage

import (
	"context"
	"fmt"
	"testing"
	"time"

	"athenaeum/internal/models"
)

// PROVED_GUEST_SESSION_EXPIRED
// Guarantee: expired guest accounts cannot authenticate via SessionUser.

func TestGuestExpiredSessionOracle(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	expired := time.Now().Add(-time.Hour)
	id, err := s.CreateGuestUser(ctx, "expguest", "hash", expired, models.DefaultUserPermissions)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.CreateSession(ctx, "guest-tok", id, time.Now().Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if _, err := s.SessionUser(ctx, "guest-tok"); err != ErrNotFound {
		t.Fatalf("expired guest session=%v want not found", err)
	}
	fmt.Println("PROVED_GUEST_SESSION_EXPIRED: expired guest denied via SessionUser")
}
