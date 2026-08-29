package cli

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"athenaeum/internal/config"
	"athenaeum/internal/library"
	"athenaeum/internal/server"
	"athenaeum/internal/storage"
	"athenaeum/internal/term"
	"athenaeum/internal/version"
)

// WantsSelfCheck reports whether args request a runtime self-check.
func WantsSelfCheck(args []string) bool {
	if len(args) > 0 {
		switch args[0] {
		case "self-check", "doctor":
			return true
		}
	}
	for _, a := range args {
		if a == "--self-check" || a == "-self-check" {
			return true
		}
	}
	return false
}

// StripSelfCheck removes the self-check command or flag from args.
func StripSelfCheck(args []string) []string {
	out := make([]string, 0, len(args))
	if len(args) > 0 {
		switch args[0] {
		case "self-check", "doctor":
			args = args[1:]
		}
	}
	for _, a := range args {
		if a == "--self-check" || a == "-self-check" {
			continue
		}
		out = append(out, a)
	}
	return out
}

// RunSelfCheck verifies the binary can create dirs, read and write them,
// open the database, serve HTTP, and answer /api/health. It exits the process
// path by returning nil on success. Temporary dirs are used unless --data /
// --library are provided.
func RunSelfCheck(args []string) error {
	cfg, err := config.Parse(args)
	if err != nil {
		return err
	}

	tmpRoot := ""
	cleanup := func() {}
	if usingDefaultDataAndLibrary(args) {
		tmpRoot, err = os.MkdirTemp("", "athenaeum-self-check-*")
		if err != nil {
			return fmt.Errorf("create temp root: %w", err)
		}
		cleanup = func() { _ = os.RemoveAll(tmpRoot) }
		cfg.DataDir = filepath.Join(tmpRoot, "data")
		cfg.LibraryDir = filepath.Join(tmpRoot, "library")
	}
	defer cleanup()

	cfg.Addr = "127.0.0.1:0"
	cfg.Demo = false
	cfg.Sandbox = "off"
	cfg.LogLevel = "error"
	cfg.PprofAddr = ""

	w := os.Stdout
	pass := func(msg string) {
		fmt.Fprintln(w, term.Success(w, "OK  "+msg))
	}
	info := func(format string, args ...any) {
		fmt.Fprintln(w, term.Dim(w, fmt.Sprintf(format, args...)))
	}

	info("self-check %s (web %s) %s/%s", version.Version, version.WebVersion, runtime.GOOS, runtime.GOARCH)

	if err := probeDir(cfg.LibraryDir, "library"); err != nil {
		return err
	}
	pass("library directory read/write: " + cfg.LibraryDir)

	if err := probeDir(cfg.DataDir, "data"); err != nil {
		return err
	}
	pass("data directory read/write: " + cfg.DataDir)

	for _, sub := range []string{cfg.CoverDir(), cfg.UploadDir(), cfg.TempDir()} {
		if err := os.MkdirAll(sub, 0o750); err != nil {
			return fmt.Errorf("mkdir %s: %w", sub, err)
		}
	}
	pass("data subdirectories created")

	driver, err := storage.ParseDriver(cfg.DatabaseDriver)
	if err != nil {
		return err
	}
	store, err := storage.OpenWith(storage.OpenOptions{
		Driver: driver,
		Path:   cfg.DBPath(),
		URL:    cfg.DatabaseURL,
	})
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer store.Close()

	ctx := context.Background()
	if err := store.EnsureDefaultLibrary(ctx, cfg.LibraryDir); err != nil {
		return fmt.Errorf("ensure default library: %w", err)
	}
	pass("database open and schema ready (" + string(driver) + ")")

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	addr := ln.Addr().String()
	_ = ln.Close()
	cfg.Addr = addr

	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	scanner := library.New(store, cfg.CoverDir(), cfg.TempDir(), log, 1)
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	srv, err := server.New(runCtx, cfg, store, scanner, log)
	if err != nil {
		return fmt.Errorf("create server: %w", err)
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Run(runCtx)
	}()

	healthURL := "http://" + addr + "/api/health"
	if err := waitHealthy(healthURL, 10*time.Second); err != nil {
		cancel()
		<-errCh
		return err
	}
	pass("HTTP server started and /api/health ok on " + addr)

	cancel()
	select {
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("server shutdown: %w", err)
		}
	case <-time.After(10 * time.Second):
		return fmt.Errorf("server did not shut down in time")
	}
	pass("clean shutdown")

	fmt.Fprintln(w, term.Success(w, "self-check passed"))
	return nil
}

func usingDefaultDataAndLibrary(args []string) bool {
	for _, a := range args {
		if a == "--data" || strings.HasPrefix(a, "--data=") {
			return false
		}
		if a == "--library" || strings.HasPrefix(a, "--library=") {
			return false
		}
	}
	return true
}

func probeDir(dir, label string) error {
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("%s mkdir: %w", label, err)
	}
	probe := filepath.Join(dir, ".athenaeum-self-check")
	payload := []byte("athenaeum-self-check\n")
	if err := os.WriteFile(probe, payload, 0o600); err != nil {
		return fmt.Errorf("%s write: %w", label, err)
	}
	got, err := os.ReadFile(probe) // #nosec G304 -- probe is under a directory this function just created
	if err != nil {
		return fmt.Errorf("%s read: %w", label, err)
	}
	if string(got) != string(payload) {
		return fmt.Errorf("%s read mismatch", label)
	}
	if err := os.Remove(probe); err != nil {
		return fmt.Errorf("%s remove: %w", label, err)
	}
	return nil
}

func waitHealthy(url string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	var last error
	for time.Now().Before(deadline) {
		resp, err := http.Get(url) // #nosec G107 -- url is constructed as http://127.0.0.1:<port>/api/health
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
			last = fmt.Errorf("health status %d", resp.StatusCode)
		} else {
			last = err
		}
		time.Sleep(25 * time.Millisecond)
	}
	if last == nil {
		last = fmt.Errorf("timed out")
	}
	return fmt.Errorf("health check failed: %w", last)
}
