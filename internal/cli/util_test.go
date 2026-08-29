package cli_test

import (
	"bytes"
	"strings"
	"testing"

	"athenaeum/internal/cli"
)

func TestPrintVersion(t *testing.T) {
	var buf bytes.Buffer
	cli.PrintVersion(&buf)
	out := buf.String()
	if !strings.Contains(out, "web") || !strings.Contains(out, "\n") {
		t.Fatalf("unexpected version output: %q", out)
	}
}
