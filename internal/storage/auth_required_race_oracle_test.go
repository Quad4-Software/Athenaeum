package storage

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"athenaeum/internal/auth"
)

// PROVED_AUTH_REQUIRED_CACHE_RACE
// Guarantee: after CreateUser, AuthRequired must not stay false due to a
// stale cache fill that raced with invalidateAuthRequired.

func TestAuthRequiredCacheRaceOracle(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	hash, err := auth.HashPassword("secretpass")
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	errCh := make(chan error, 64)
	for i := range 32 {
		wg.Add(2)
		go func() {
			defer wg.Done()
			if _, err := s.AuthRequired(ctx); err != nil {
				errCh <- err
			}
		}()
		go func(i int) {
			defer wg.Done()
			name := fmt.Sprintf("raceuser-%d", i)
			id, err := s.CreateUser(ctx, name, hash, false)
			if err != nil {
				errCh <- err
				return
			}
			_ = s.DeleteUser(ctx, id)
		}(i)
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		t.Fatal(err)
	}

	required, err := s.AuthRequired(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if required {
		t.Fatal("expected no users after cleanup")
	}

	if _, err := s.CreateUser(ctx, "final", hash, false); err != nil {
		t.Fatal(err)
	}
	for i := range 100 {
		required, err = s.AuthRequired(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if !required {
			t.Fatalf("stale AuthRequired=false after CreateUser on iteration %d", i)
		}
	}
	fmt.Println("PROVED_AUTH_REQUIRED_CACHE_RACE: cache stays coherent under create/delete races")
}
