// Package logging configures structured slog handlers for stderr and optional files.
package logging

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
)

// Options configure the process logger.
type Options struct {
	Level     string
	File      string
	AddSource bool
}

// Setup installs a default slog logger writing to stderr and optionally a file.
// The returned closer should be deferred by the caller.
func Setup(opts Options) (io.Closer, error) {
	level, err := ParseLevel(opts.Level)
	if err != nil {
		return nil, err
	}

	writers := []io.Writer{os.Stderr}
	var file *os.File
	if path := strings.TrimSpace(opts.File); path != "" {
		f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o640) // #nosec G304 G302 -- operator-configured log path
		if err != nil {
			return nil, fmt.Errorf("open log file %s: %w", path, err)
		}
		file = f
		writers = append(writers, f)
	}

	var w io.Writer
	if len(writers) == 1 {
		w = writers[0]
	} else {
		w = io.MultiWriter(writers...)
	}

	handler := slog.NewTextHandler(w, &slog.HandlerOptions{
		Level:     level,
		AddSource: opts.AddSource || level <= slog.LevelDebug,
	})
	slog.SetDefault(slog.New(handler))
	if file == nil {
		return nopCloser{}, nil
	}
	return file, nil
}

// ParseLevel maps common level names to slog.Level.
func ParseLevel(s string) (slog.Level, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "info":
		return slog.LevelInfo, nil
	case "debug":
		return slog.LevelDebug, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("invalid log level %q (want debug, info, warn, or error)", s)
	}
}

type nopCloser struct{}

func (nopCloser) Close() error { return nil }
