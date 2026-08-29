package system_test

import (
	"testing"

	"athenaeum/internal/system"
)

func TestReadStatsAndDiskFree(t *testing.T) {
	dir := t.TempDir()
	st := system.ReadStats([]string{"", dir, "/nonexistent/path/athenaeum-coverage"})
	if st.Disks == nil {
		t.Fatal("disks should be non-nil")
	}
	found := false
	for _, d := range st.Disks {
		if d.Path == dir {
			found = true
			if d.Total <= 0 || d.Available < 0 {
				t.Fatalf("unexpected disk usage: %+v", d)
			}
		}
	}
	if !found {
		t.Fatalf("expected disk entry for %s in %+v", dir, st.Disks)
	}
	if free := system.DiskFree(dir); free < 0 {
		t.Fatalf("DiskFree=%d", free)
	}
	if system.DiskFree("/nonexistent/path/athenaeum-coverage") != -1 {
		t.Fatal("expected -1 for missing path")
	}
}
