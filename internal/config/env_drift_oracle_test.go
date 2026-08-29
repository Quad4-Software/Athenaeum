package config_test

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"testing"
)

// PROVED_ENV_EXAMPLE_DRIFT: every ATHENAEUM_* key in config.go appears in .env.example.

func TestEnvExampleCoversConfigKeys(t *testing.T) {
	root := repoRoot(t)
	cfgSrc, err := os.ReadFile(filepath.Join(root, "internal/config/config.go"))
	if err != nil {
		t.Fatal(err)
	}
	example, err := os.ReadFile(filepath.Join(root, ".env.example"))
	if err != nil {
		t.Fatal(err)
	}

	re := regexp.MustCompile(`ATHENAEUM_[A-Z0-9_]+`)
	want := map[string]struct{}{}
	for _, m := range re.FindAllString(string(cfgSrc), -1) {
		want[m] = struct{}{}
	}
	got := map[string]struct{}{}
	for _, m := range re.FindAllString(string(example), -1) {
		got[m] = struct{}{}
	}

	skip := map[string]struct{}{
		// Runtime-only / not meant for dotenv templates.
		"ATHENAEUM_PASSWORD": {},
		"ATHENAEUM_ENV_FILE": {},
	}

	var missing []string
	for k := range want {
		if _, ok := skip[k]; ok {
			continue
		}
		if _, ok := got[k]; !ok {
			missing = append(missing, k)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		t.Fatalf(".env.example missing config keys:\n%s", strings.Join(missing, "\n"))
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "../.."))
}

func TestEnvExampleHasNoReaderPrefix(t *testing.T) {
	root := repoRoot(t)
	example, err := os.ReadFile(filepath.Join(root, ".env.example"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(example), "READER_") {
		t.Fatal(".env.example still contains READER_ keys")
	}
}
