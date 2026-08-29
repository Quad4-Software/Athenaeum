package pprofserve

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"
)

func TestIsLoopbackAddr(t *testing.T) {
	cases := []struct {
		addr string
		ok   bool
	}{
		{"127.0.0.1:6060", true},
		{"localhost:6060", true},
		{"[::1]:6060", true},
		{":6060", false},
		{"0.0.0.0:6060", false},
		{"192.168.1.1:6060", false},
		{"", false},
	}
	for _, tc := range cases {
		if got := isLoopbackAddr(tc.addr); got != tc.ok {
			t.Fatalf("isLoopbackAddr(%q)=%v want %v", tc.addr, got, tc.ok)
		}
	}
}

func TestStartEmptyAndLoopback(t *testing.T) {
	stop, err := Start(context.Background(), "", nil)
	if err != nil {
		t.Fatal(err)
	}
	stop()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	stop, err = Start(ctx, "127.0.0.1:0", log)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(20 * time.Millisecond)
	stop()

	_, err = Start(context.Background(), "0.0.0.0:6060", log)
	if err == nil {
		t.Fatal("expected non-loopback error")
	}
	_, err = Start(context.Background(), "192.168.1.10:6060", nil)
	if err == nil {
		t.Fatal("expected non-loopback error")
	}
}
