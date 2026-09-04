package storage

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
)

// PROVED_SETUP_SINGLE_ADMIN
// Guarantee: concurrent CreateInitialAdmin calls create at most one admin.

func TestCreateInitialAdminRaceOracle(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	var okCount atomic.Int64
	var wg sync.WaitGroup
	for i := range 16 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, err := s.CreateInitialAdmin(ctx, fmt.Sprintf("admin%d", i), "hash")
			if err == nil {
				okCount.Add(1)
			}
		}(i)
	}
	wg.Wait()
	if okCount.Load() != 1 {
		t.Fatalf("CreateInitialAdmin successes=%d want 1", okCount.Load())
	}
	n, err := s.UserCount(ctx)
	if err != nil || n != 1 {
		t.Fatalf("users=%d err=%v", n, err)
	}
	fmt.Println("PROVED_SETUP_SINGLE_ADMIN: concurrent setup created exactly one admin")
}
