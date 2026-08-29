package cli

import (
	"fmt"
	"io"
	"os"

	"athenaeum/internal/brand"
	"athenaeum/internal/term"
	"athenaeum/internal/version"
)

// PrintVersion writes the binary version to w.
func PrintVersion(w io.Writer) {
	fmt.Fprintf(w, "%s %s (web %s)\n", brand.Name, version.Version, version.WebVersion)
}

// Successf prints a green success line to stdout when color is enabled.
func Successf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintln(os.Stdout, term.Success(os.Stdout, msg))
}

// binaryName returns the lowercase CLI binary name.
func binaryName() string {
	return "athenaeum"
}
