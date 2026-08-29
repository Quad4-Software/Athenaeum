// Package term provides TTY-aware ANSI color helpers for CLI output.
package term

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/mattn/go-isatty"
)

// Mode controls whether ANSI color is used.
type Mode int

const (
	// ModeAuto enables color when stdout/stderr is a TTY and NO_COLOR is unset.
	ModeAuto Mode = iota
	// ModeAlways forces color on.
	ModeAlways
	// ModeNever forces color off.
	ModeNever
)

var (
	stdoutColored bool
	stderrColored bool
)

func init() {
	Apply(ModeAuto)
}

// Apply configures color for stdout and stderr from mode and environment.
// Respects NO_COLOR (https://no-color.org/) and FORCE_COLOR when mode is Auto.
func Apply(mode Mode) {
	switch mode {
	case ModeAlways:
		stdoutColored = true
		stderrColored = true
	case ModeNever:
		stdoutColored = false
		stderrColored = false
	default:
		force := envTruthy("FORCE_COLOR")
		noColor := envTruthy("NO_COLOR") || envTruthy("ATHENAEUM_NO_COLOR")
		stdoutColored = force || (!noColor && isTerminal(os.Stdout))
		stderrColored = force || (!noColor && isTerminal(os.Stderr))
	}
}

// Enabled reports whether color is active for the given writer.
func Enabled(w io.Writer) bool {
	if f, ok := w.(*os.File); ok {
		if f == os.Stdout {
			return stdoutColored
		}
		if f == os.Stderr {
			return stderrColored
		}
		return isTerminal(f) && stdoutColored
	}
	return false
}

// Bold returns text in bold when color is enabled for w.
func Bold(w io.Writer, text string) string {
	return wrap(w, "1", text)
}

// Dim returns dimmed text when color is enabled for w.
func Dim(w io.Writer, text string) string {
	return wrap(w, "2", text)
}

// Green returns green text when color is enabled for w.
func Green(w io.Writer, text string) string {
	return wrap(w, "32", text)
}

// Yellow returns yellow text when color is enabled for w.
func Yellow(w io.Writer, text string) string {
	return wrap(w, "33", text)
}

// Red returns red text when color is enabled for w.
func Red(w io.Writer, text string) string {
	return wrap(w, "31", text)
}

// Cyan returns cyan text when color is enabled for w.
func Cyan(w io.Writer, text string) string {
	return wrap(w, "36", text)
}

// Magenta returns magenta text when color is enabled for w.
func Magenta(w io.Writer, text string) string {
	return wrap(w, "35", text)
}

// Header styles a section heading.
func Header(w io.Writer, text string) string {
	return Bold(w, text)
}

// Flag styles a CLI flag name.
func Flag(w io.Writer, text string) string {
	return Cyan(w, text)
}

// Success styles a success message.
func Success(w io.Writer, text string) string {
	return Green(w, text)
}

// Error styles an error message.
func Error(w io.Writer, text string) string {
	return Red(w, text)
}

// Warn styles a warning message.
func Warn(w io.Writer, text string) string {
	return Yellow(w, text)
}

// Fprint wraps fmt.Fprint and ignores write errors (CLI help/status output).
func Fprint(w io.Writer, a ...any) {
	_, _ = fmt.Fprint(w, a...)
}

// Fprintln wraps fmt.Fprintln and ignores write errors (CLI help/status output).
func Fprintln(w io.Writer, a ...any) {
	_, _ = fmt.Fprintln(w, a...)
}

// Fprintf wraps fmt.Fprintf and ignores write errors (CLI help/status output).
func Fprintf(w io.Writer, format string, a ...any) {
	_, _ = fmt.Fprintf(w, format, a...)
}

func wrap(w io.Writer, code, text string) string {
	if !Enabled(w) || text == "" {
		return text
	}
	return "\x1b[" + code + "m" + text + "\x1b[0m"
}

func isTerminal(f *os.File) bool {
	if f == nil {
		return false
	}
	return isatty.IsTerminal(f.Fd()) || isatty.IsCygwinTerminal(f.Fd())
}

func envTruthy(key string) bool {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return false
	}
	switch strings.ToLower(v) {
	case "0", "false", "no", "off":
		return false
	default:
		return true
	}
}

// ParseMode interprets a --color flag value: auto, always, never, true, false.
func ParseMode(s string) (Mode, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "auto":
		return ModeAuto, nil
	case "always", "on", "yes", "true", "1", "force":
		return ModeAlways, nil
	case "never", "off", "no", "false", "0":
		return ModeNever, nil
	default:
		return ModeAuto, fmt.Errorf("invalid color mode %q (want auto, always, or never)", s)
	}
}
