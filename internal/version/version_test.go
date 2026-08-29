package version_test

import (
	"testing"

	"athenaeum/internal/version"
)

func TestVersionDefaults(t *testing.T) {
	if version.Version == "" {
		t.Fatal("Version should be set")
	}
	if version.WebVersion == "" {
		t.Fatal("WebVersion should be set")
	}
}
