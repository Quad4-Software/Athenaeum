package storage

import (
	"context"
	"path/filepath"
	"testing"

	"athenaeum/internal/auth"
)

func TestAPIKeys(t *testing.T) {
	dir := t.TempDir()
	store, err := Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	ctx := context.Background()

	hash, _ := auth.HashPassword("longpassword")
	userID, err := store.CreateUser(ctx, "apiuser", hash, false)
	if err != nil {
		t.Fatal(err)
	}

	created, err := store.CreateAPIKey(ctx, userID, "script")
	if err != nil {
		t.Fatal(err)
	}
	if created.Key == "" || created.Prefix == "" {
		t.Fatal("expected key and prefix")
	}

	keys, err := store.ListAPIKeys(ctx, userID)
	if err != nil {
		t.Fatal(err)
	}
	if len(keys) != 1 || keys[0].Name != "script" {
		t.Fatalf("keys=%+v", keys)
	}

	u, keyID, err := store.UserFromAPIKey(ctx, created.Key)
	if err != nil {
		t.Fatal(err)
	}
	if u.ID != userID || keyID != created.ID {
		t.Fatalf("user=%+v keyID=%d", u, keyID)
	}

	if err := store.DeleteAPIKey(ctx, userID, created.ID); err != nil {
		t.Fatal(err)
	}
	if _, _, err := store.UserFromAPIKey(ctx, created.Key); err != ErrNotFound {
		t.Fatalf("expected not found after delete, got %v", err)
	}
}
