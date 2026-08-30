package storage

import (
	"context"
	"testing"

	"athenaeum/internal/auth"
)

func TestAuthRequiredCacheInvalidates(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	required, err := s.AuthRequired(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if required {
		t.Fatal("expected no auth before users exist")
	}

	// Second call should hit cache (still false).
	required, err = s.AuthRequired(ctx)
	if err != nil || required {
		t.Fatalf("cached empty = %v err=%v", required, err)
	}

	hash, err := auth.HashPassword("secretpass")
	if err != nil {
		t.Fatal(err)
	}
	id, err := s.CreateUser(ctx, "cacheuser", hash, false)
	if err != nil {
		t.Fatal(err)
	}

	required, err = s.AuthRequired(ctx)
	if err != nil || !required {
		t.Fatalf("after create = %v err=%v", required, err)
	}

	if err := s.DeleteUser(ctx, id); err != nil {
		t.Fatal(err)
	}
	required, err = s.AuthRequired(ctx)
	if err != nil || required {
		t.Fatalf("after delete = %v err=%v", required, err)
	}
}
