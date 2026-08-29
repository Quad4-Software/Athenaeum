package sandbox

import (
	"fmt"
	"io"
	"os"
	"strings"

	"athenaeum/internal/term"
)

// State is the outcome for one sandbox component.
type State string

const (
	// StateOff means the sandbox mode is off so the component was not attempted.
	StateOff State = "off"
	// StateDisabled means sandbox mode is on but the component toggle is false.
	StateDisabled State = "disabled"
	// StateApplied means the restriction was installed successfully.
	StateApplied State = "applied"
	// StateSkipped means setup failed and was skipped (try mode).
	StateSkipped State = "skipped"
	// StateUnsupported means the platform cannot apply the component.
	StateUnsupported State = "unsupported"
)

// Component reports one of Landlock or seccomp.
type Component struct {
	State  State
	Detail string
	Err    string
}

// Status summarizes sandbox setup after Apply.
type Status struct {
	Mode     Mode
	Landlock Component
	Seccomp  Component
}

// Summary returns a plain single-line description for logs and files.
func (s Status) Summary() string {
	return fmt.Sprintf("mode=%s landlock=%s seccomp=%s",
		s.Mode,
		formatComponent(s.Landlock),
		formatComponent(s.Seccomp),
	)
}

// LogArgs returns slog attributes for structured startup logging.
func (s Status) LogArgs() []any {
	return []any{
		"sandbox", string(s.Mode),
		"landlock", formatComponent(s.Landlock),
		"seccomp", formatComponent(s.Seccomp),
	}
}

// Print writes a human-readable status line to w.
// Color follows stderr TTY / --color settings even when w is not a terminal.
func (s Status) Print(w io.Writer) {
	if w == nil {
		w = os.Stderr
	}
	cw := os.Stderr
	label := term.Bold(cw, "sandbox")
	mode := colorMode(cw, s.Mode)
	ll := colorComponent(cw, "landlock", s.Landlock)
	sc := colorComponent(cw, "seccomp", s.Seccomp)
	fmt.Fprintf(w, "%s  mode=%s  %s  %s\n", label, mode, ll, sc)
}

// Announce writes a colored status line to stderr when color is enabled.
func (s Status) Announce() {
	if term.Enabled(os.Stderr) {
		s.Print(os.Stderr)
	}
}

func formatComponent(c Component) string {
	parts := []string{string(c.State)}
	if c.Detail != "" {
		parts = append(parts, c.Detail)
	}
	if c.Err != "" {
		parts = append(parts, "err="+c.Err)
	}
	return strings.Join(parts, " ")
}

func colorMode(w io.Writer, mode Mode) string {
	switch mode {
	case ModeStrict:
		return term.Green(w, string(mode))
	case ModeTry:
		return term.Cyan(w, string(mode))
	default:
		return term.Yellow(w, string(mode))
	}
}

func colorComponent(w io.Writer, name string, c Component) string {
	body := formatComponent(c)
	var colored string
	switch c.State {
	case StateApplied:
		colored = term.Green(w, body)
	case StateSkipped, StateUnsupported:
		colored = term.Yellow(w, body)
	case StateDisabled, StateOff:
		colored = term.Dim(w, body)
	default:
		colored = body
	}
	return name + "=" + colored
}
