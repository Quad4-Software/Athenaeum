// Command athenaeum is a single-binary EPUB/PDF library server: a fast,
// self-hosted replacement for tools like Calibre-web, Audiobookshelf or
// Stump, with the frontend embedded into the executable (or served from
// --web-dir).
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"athenaeum/internal/auth"
	"athenaeum/internal/cli"
	"athenaeum/internal/config"
	"athenaeum/internal/demo"
	"athenaeum/internal/library"
	"athenaeum/internal/logging"
	"athenaeum/internal/pprofserve"
	"athenaeum/internal/sandbox"
	"athenaeum/internal/server"
	"athenaeum/internal/storage"
	"athenaeum/internal/telemetry"
	"athenaeum/internal/term"
)

func main() {
	if err := config.LoadEnv(); err != nil {
		fmt.Fprintln(os.Stderr, term.Error(os.Stderr, "error:"), err)
		os.Exit(1)
	}
	applyEarlyColor(os.Args[1:])

	if cli.WantsSelfCheck(os.Args[1:]) {
		if err := cli.RunSelfCheck(cli.StripSelfCheck(os.Args[1:])); err != nil {
			fmt.Fprintln(os.Stderr, term.Error(os.Stderr, "error:"), err)
			os.Exit(1)
		}
		return
	}

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "users":
			if err := cli.RunUsers(os.Args[2:]); err != nil {
				fmt.Fprintln(os.Stderr, term.Error(os.Stderr, "error:"), err)
				os.Exit(1)
			}
			return
		case "help", "-h", "--help":
			config.PrintHelp(os.Stdout)
			return
		case "version", "-version", "--version":
			cli.PrintVersion(os.Stdout)
			return
		}
	}

	if err := run(); err != nil {
		slog.Error("fatal", "err", err)
		os.Exit(1)
	}
}

func applyEarlyColor(args []string) {
	mode := term.ModeAuto
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--no-color":
			mode = term.ModeNever
		case a == "--color" && i+1 < len(args):
			if m, err := term.ParseMode(args[i+1]); err == nil {
				mode = m
			}
			i++
		case strings.HasPrefix(a, "--color="):
			if m, err := term.ParseMode(strings.TrimPrefix(a, "--color=")); err == nil {
				mode = m
			}
		}
	}
	term.Apply(mode)
}

func run() error {
	cfg, err := config.Parse(os.Args[1:])
	if err != nil {
		return err
	}

	auth.SetPasswordPolicy(auth.PasswordPolicy{
		MinLength:     cfg.PasswordMinLength,
		LongLength:    cfg.PasswordLongLength,
		MinKinds:      cfg.PasswordMinKinds,
		RequireLower:  cfg.PasswordRequireLower,
		RequireUpper:  cfg.PasswordRequireUpper,
		RequireDigit:  cfg.PasswordRequireDigit,
		RequireSymbol: cfg.PasswordRequireSymbol,
	})

	if m, err := term.ParseMode(cfg.ColorMode); err == nil {
		term.Apply(m)
	} else {
		return err
	}

	logCloser, err := logging.Setup(logging.Options{
		Level: cfg.LogLevel,
		File:  cfg.LogFile,
	})
	if err != nil {
		return err
	}
	defer logCloser.Close()
	log := slog.Default()

	if err := telemetry.Init(cfg); err != nil {
		log.Warn("sentry init failed", "err", err)
	}
	defer telemetry.Flush(2 * time.Second)

	if err := os.MkdirAll(cfg.DataDir, 0o750); err != nil {
		return err
	}
	if err := os.MkdirAll(cfg.LibraryDir, 0o750); err != nil {
		return err
	}
	if err := os.MkdirAll(cfg.CoverDir(), 0o750); err != nil { // #nosec G703 -- CoverDir is under configured DataDir
		return err
	}
	if err := os.MkdirAll(cfg.UploadDir(), 0o750); err != nil { // #nosec G703 -- UploadDir is under configured DataDir
		return err
	}
	if err := os.MkdirAll(cfg.TempDir(), 0o750); err != nil { // #nosec G703 -- TempDir is under configured DataDir
		return err
	}

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
		return err
	}
	defer store.Close()

	if err := store.EnsureDefaultLibrary(context.Background(), cfg.LibraryDir); err != nil {
		return err
	}

	if cfg.Demo {
		if err := demo.Seed(context.Background(), store, cfg.LibraryDir, cfg.CoverDir(), log); err != nil {
			return fmt.Errorf("demo seed: %w", err)
		}
	}

	if err := ensureAdmin(context.Background(), store, cfg, log); err != nil {
		return err
	}

	libs, err := store.ListLibraries(context.Background())
	if err != nil {
		return err
	}
	roPaths := make([]string, 0, len(libs)+2)
	roPaths = append(roPaths, cfg.LibraryDir)
	for _, lib := range libs {
		if lib.Backend != "" && lib.Backend != "local" {
			continue
		}
		if lib.MountPath != "" && !strings.HasPrefix(lib.MountPath, "s3://") {
			roPaths = append(roPaths, lib.MountPath)
		}
	}
	if cfg.WebDir != "" {
		roPaths = append(roPaths, cfg.WebDir)
	}
	rwPaths := []string{cfg.DataDir}
	if d := cfg.LogFileDir(); d != "" {
		rwPaths = append(rwPaths, d)
	}

	sandboxMode, err := sandbox.ParseMode(cfg.Sandbox)
	if err != nil {
		return err
	}
	sandboxStatus, err := sandbox.Apply(sandbox.Config{
		Mode:           sandboxMode,
		ReadWritePaths: rwPaths,
		ReadOnlyPaths:  roPaths,
		EnableLandlock: cfg.SandboxLandlock,
		EnableSeccomp:  cfg.SandboxSeccomp,
	}, log)
	if err != nil {
		return err
	}
	sandboxStatus.Announce()

	scanner := library.New(store, cfg.CoverDir(), cfg.TempDir(), log, cfg.ScanWorkers)
	watcher := library.NewWatcher(store, scanner, log)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	stopPprof, err := pprofserve.Start(ctx, cfg.PprofAddr, log)
	if err != nil {
		return err
	}
	defer stopPprof()

	var bg sync.WaitGroup
	bg.Go(func() { sessionJanitor(ctx, store, log) })
	bg.Go(func() { guestJanitor(ctx, store, log) })
	bg.Go(func() { autoScanJanitor(ctx, store, scanner, log) })
	bg.Go(func() { watcher.Run(ctx) })
	if !cfg.Demo {
		bg.Go(func() {
			if err := scanner.Scan(ctx); err != nil && ctx.Err() == nil {
				log.Error("initial scan failed", "err", err)
			}
		})
	} else {
		log.Info("demo mode: skipping initial filesystem scan (catalog already seeded)")
	}

	startArgs := []any{
		"addr", cfg.Addr,
		"library", cfg.LibraryDir,
		"data", cfg.DataDir,
		"database", string(driver),
		"log_level", cfg.LogLevel,
		"demo", cfg.Demo,
	}
	if cfg.WebDir != "" {
		startArgs = append(startArgs, "web_dir", cfg.WebDir)
	} else {
		startArgs = append(startArgs, "web", "embedded")
	}
	startArgs = append(startArgs, sandboxStatus.LogArgs()...)
	log.Info("starting server", startArgs...)

	srv, err := server.New(ctx, cfg, store, scanner, log)
	if err != nil {
		return err
	}
	if cfg.AltchaEnabled {
		log.Info("altcha enabled", "mode", cfg.AltchaMode)
	}
	err = srv.Run(ctx)
	stop()

	waitCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	done := make(chan struct{})
	go func() {
		bg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-waitCtx.Done():
		log.Warn("background workers did not stop before timeout")
	}

	if err != nil {
		return err
	}
	log.Info("shutdown complete")
	return nil
}

func ensureAdmin(ctx context.Context, store *storage.Store, cfg config.Config, log *slog.Logger) error {
	required, err := store.AuthRequired(ctx)
	if err != nil {
		return err
	}
	if required || cfg.AdminUser == "" || cfg.AdminPass == "" {
		return nil
	}
	hash, err := auth.HashPassword(cfg.AdminPass)
	if err != nil {
		return err
	}
	id, err := store.CreateUser(ctx, cfg.AdminUser, hash, true)
	if err != nil {
		return err
	}
	log.Info("created admin user", "username", cfg.AdminUser, "id", id)
	return nil
}

func sessionJanitor(ctx context.Context, store *storage.Store, log *slog.Logger) {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := store.PurgeExpiredSessions(ctx); err != nil {
				log.Warn("session purge failed", "err", err)
			}
		}
	}
}

func guestJanitor(ctx context.Context, store *storage.Store, log *slog.Logger) {
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			n, err := store.PurgeExpiredGuests(ctx)
			if err != nil {
				log.Warn("guest purge failed", "err", err)
			} else if n > 0 {
				log.Info("purged expired guest accounts", "count", n)
			}
		}
	}
}

func autoScanJanitor(ctx context.Context, store *storage.Store, scanner *library.Scanner, log *slog.Logger) {
	var lastScan time.Time
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			cfg, err := store.GetServerConfig(ctx, false)
			if err != nil || !cfg.AutoScanEnabled {
				continue
			}
			interval := time.Duration(cfg.AutoScanInterval) * time.Second
			if interval < 60*time.Second {
				interval = 5 * time.Minute
			}
			if !lastScan.IsZero() && time.Since(lastScan) < interval {
				continue
			}
			if scanner.Status().Scanning {
				continue
			}
			lastScan = time.Now()
			if err := scanner.Scan(ctx); err != nil && ctx.Err() == nil {
				log.Warn("auto scan failed", "err", err)
			}
		}
	}
}
