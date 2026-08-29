package library

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"athenaeum/internal/storage"
)

func TestScannerScanning(t *testing.T) {
	store, err := storage.Open(t.TempDir() + "/s.db")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	sc := New(store, t.TempDir(), t.TempDir(), log, 1)
	if sc.Scanning() {
		t.Fatal("expected not scanning")
	}
	st := sc.Status()
	if st.Scanning || st.Indexed != 0 {
		t.Fatalf("idle status=%+v", st)
	}
	called := false
	sc.SetOnComplete(func(ScanCompleteEvent) { called = true })
	if err := sc.Wait(context.Background()); err != nil {
		t.Fatalf("idle wait: %v", err)
	}
	if called {
		t.Fatal("onComplete should not fire without a scan")
	}
}
