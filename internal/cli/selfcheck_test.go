package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWantsSelfCheck(t *testing.T) {
	cases := []struct {
		args []string
		want bool
	}{
		{nil, false},
		{[]string{"--addr", ":8080"}, false},
		{[]string{"--self-check"}, true},
		{[]string{"-self-check"}, true},
		{[]string{"self-check"}, true},
		{[]string{"doctor"}, true},
		{[]string{"self-check", "--data", "/tmp/x"}, true},
		{[]string{"--library", "./l", "--self-check"}, true},
		{[]string{"users", "list"}, false},
	}
	for _, tc := range cases {
		if got := WantsSelfCheck(tc.args); got != tc.want {
			t.Fatalf("WantsSelfCheck(%v)=%v want %v", tc.args, got, tc.want)
		}
	}
}

func TestStripSelfCheck(t *testing.T) {
	got := StripSelfCheck([]string{"self-check", "--data", "/tmp/d"})
	if len(got) != 2 || got[0] != "--data" || got[1] != "/tmp/d" {
		t.Fatalf("got %#v", got)
	}
	got = StripSelfCheck([]string{"--library", "./l", "--self-check", "--debug"})
	if len(got) != 3 || got[0] != "--library" || got[2] != "--debug" {
		t.Fatalf("got %#v", got)
	}
}

func TestRunSelfCheckTempDirs(t *testing.T) {
	if err := RunSelfCheck(nil); err != nil {
		t.Fatal(err)
	}
}

func TestRunSelfCheckExplicitDirs(t *testing.T) {
	root := t.TempDir()
	data := filepath.Join(root, "data")
	lib := filepath.Join(root, "library")
	if err := RunSelfCheck([]string{"--data", data, "--library", lib}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(data); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(lib); err != nil {
		t.Fatal(err)
	}
}
