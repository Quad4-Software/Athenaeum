//go:build linux

package sandbox

import (
	"bytes"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestLandlockABIDetail(t *testing.T) {
	got := landlockABIDetail(6, nil, ModeTry)
	if got == "" || !containsAll(got, "request=V9", "effective=V6", "best-effort") {
		t.Fatalf("got %q", got)
	}
	got = landlockABIDetail(9, nil, ModeStrict)
	if !containsAll(got, "strict", "effective=V9") {
		t.Fatalf("got %q", got)
	}
	got = landlockABIDetail(0, nil, ModeTry)
	if !containsAll(got, "kernel=none") {
		t.Fatalf("got %q", got)
	}
	got = landlockABIDetail(0, errors.New("no abi"), ModeTry)
	if !containsAll(got, "kernel=unavailable") {
		t.Fatalf("got %q", got)
	}
}

func TestClassifyAndUniquePaths(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "f.txt")
	if err := os.WriteFile(file, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	dirs, files := classifyPaths([]string{"", dir, file, filepath.Join(dir, "missing")})
	if len(dirs) < 2 {
		t.Fatalf("dirs=%v", dirs)
	}
	if len(files) != 1 {
		t.Fatalf("files=%v", files)
	}
	out := uniqueAbs([]string{dir, dir, "", file})
	if len(out) != 2 {
		t.Fatalf("uniqueAbs=%v", out)
	}
}

func TestHandleModes(t *testing.T) {
	log := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
	if err := handle(Config{Mode: ModeTry}, log, "landlock", errors.New("skip")); err != nil {
		t.Fatal(err)
	}
	if err := handle(Config{Mode: ModeStrict}, log, "landlock", errors.New("boom")); err == nil {
		t.Fatal("expected strict error")
	}
}

func TestApplyDisabledToggles(t *testing.T) {
	st, err := Apply(Config{Mode: ModeTry}, slog.Default())
	if err != nil {
		t.Fatal(err)
	}
	if st.Landlock.State != StateDisabled || st.Seccomp.State != StateDisabled {
		t.Fatalf("status=%+v", st)
	}
}

func TestSeccompHelpers(t *testing.T) {
	if nativeAuditArch() == 0 && (runtime.GOARCH == "amd64" || runtime.GOARCH == "arm64") {
		t.Fatalf("unexpected zero audit arch on %s", runtime.GOARCH)
	}
	denied := deniedSyscalls()
	if len(denied) == 0 {
		t.Fatal("expected denied syscalls")
	}
	if nr, ok := syscallNumber("mount"); !ok || nr == 0 {
		t.Fatalf("mount nr=%d ok=%v", nr, ok)
	}
	if _, ok := syscallNumber("not-a-syscall"); ok {
		t.Fatal("unknown syscall should miss")
	}
	f := bpfLoad(4)
	if f.K != 4 {
		t.Fatalf("bpfLoad %#v", f)
	}
	j := bpfJumpIf(1, 0, 1)
	if j.K != 1 || j.Jf != 1 {
		t.Fatalf("bpfJumpIf %#v", j)
	}
	r := bpfReturn(seccompRetAllow)
	if r.K != seccompRetAllow {
		t.Fatalf("bpfReturn %#v", r)
	}
}

func containsAll(s string, parts ...string) bool {
	for _, p := range parts {
		if !strings.Contains(s, p) {
			return false
		}
	}
	return true
}
