// Package server wires the HTTP API and the single-page application
// (embedded or served from --web-dir) behind a single net/http server,
// keeping the dependency surface limited to the standard library.
package server

import (
	"context"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"athenaeum/internal/altcha"
	"athenaeum/internal/config"
	"athenaeum/internal/library"
	"athenaeum/internal/models"
	"athenaeum/internal/storage"
	"athenaeum/internal/telemetry"
)

const (
	httpShutdownTimeout = 10 * time.Second
	jobsShutdownTimeout = 20 * time.Second
)

// Server holds the dependencies shared by all HTTP handlers.
type Server struct {
	cfg             config.Config
	store           *storage.Store
	scanner         *library.Scanner
	metadataMatcher *library.MetadataMatcher
	maintenance     *library.Maintenance
	altcha          *altcha.Service
	log             *slog.Logger
	proxies         proxyTrust
	serverCfgMu     sync.RWMutex
	serverCfg       models.ServerConfigPublic

	totpMu      sync.Mutex
	totpPending map[string]pendingTOTP

	jobsCtx    context.Context
	jobsCancel context.CancelFunc
}

// New constructs a Server. Background jobs inherit from ctx and stop when it is cancelled.
func New(ctx context.Context, cfg config.Config, store *storage.Store, scanner *library.Scanner, log *slog.Logger) (*Server, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	jobsCtx, jobsCancel := context.WithCancel(ctx)
	altchaSvc, err := altcha.New(cfg)
	if err != nil {
		jobsCancel()
		return nil, err
	}
	srv := &Server{
		cfg:             cfg,
		store:           store,
		scanner:         scanner,
		metadataMatcher: library.NewMetadataMatcher(store, cfg.CoverDir(), log),
		maintenance:     library.NewMaintenance(store, cfg.CoverDir(), log),
		altcha:          altchaSvc,
		log:             log,
		totpPending:     make(map[string]pendingTOTP),
		jobsCtx:         jobsCtx,
		jobsCancel:      jobsCancel,
	}
	scanner.SetOnComplete(func(ev library.ScanCompleteEvent) {
		srv.emitWebhook(models.WebhookEventLibraryScanComplete, map[string]any{
			"libraryId": ev.LibraryID,
			"indexed":   ev.Indexed,
			"skipped":   ev.Skipped,
			"pruned":    ev.Pruned,
		})
	})
	return srv, nil
}

// Handler builds the root http.Handler with all routes registered.
func (s *Server) Handler() (http.Handler, error) {
	if err := s.loadServerConfig(context.Background()); err != nil {
		return nil, err
	}

	mux := http.NewServeMux()
	s.registerMetricsRoute(mux)
	s.registerAPI(mux)

	spa, err := spaHandler(s.cfg.WebDir)
	if err != nil {
		return nil, err
	}
	mux.Handle("/", spa)

	chain := recoverMiddleware(s.log)(
		telemetry.HTTPMiddleware()(s.withMiddleware(s.withSecurityHeaders(s.withCSRF(s.withAuth(mux))))),
	)
	return chain, nil
}

// withMiddleware applies cross-cutting concerns: request logging and a
// permissive set of security-friendly headers.
func (s *Server) withMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		httpRequestsTotal.Add(1)
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r)
		s.log.Debug("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", sw.status,
			"dur", time.Since(start).String(),
		)
	})
}

// statusWriter captures the response status code for logging.
type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

// Flush implements http.Flusher when the underlying writer supports it.
func (w *statusWriter) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// Run starts the HTTP server and blocks until the context is cancelled,
// then drains connections and waits for background jobs before returning.
func (s *Server) Run(ctx context.Context) error {
	handler, err := s.Handler()
	if err != nil {
		return err
	}

	srv := &http.Server{
		Addr:              announceAddr(s.cfg.Addr),
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		listenArgs := []any{"addr", s.cfg.Addr, "library", s.cfg.LibraryDir}
		if s.cfg.WebDir != "" {
			listenArgs = append(listenArgs, "web_dir", s.cfg.WebDir)
		}
		s.log.Info("athenaeum listening", listenArgs...)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	var runErr error
	select {
	case runErr = <-errCh:
		s.log.Error("http server failed", "err", runErr)
	case <-ctx.Done():
		s.log.Info("shutdown signal received")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), httpShutdownTimeout)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			s.log.Warn("http shutdown incomplete", "err", err)
			runErr = err
		}
	}

	s.cleanup()
	return runErr
}

func (s *Server) cleanup() {
	s.jobsCancel()

	waitCtx, cancel := context.WithTimeout(context.Background(), jobsShutdownTimeout)
	defer cancel()

	if err := s.scanner.Wait(waitCtx); err != nil {
		s.log.Warn("scanner did not finish before shutdown timeout", "err", err)
	}
	if err := s.metadataMatcher.Wait(waitCtx); err != nil {
		s.log.Warn("metadata matcher did not finish before shutdown timeout", "err", err)
	}
	if err := s.maintenance.Wait(waitCtx); err != nil {
		s.log.Warn("maintenance did not finish before shutdown timeout", "err", err)
	}

	library.CloseOutboundHTTP()
	s.log.Info("background cleanup finished")
}

func announceAddr(addr string) string {
	if addr == "" {
		return ":8080"
	}
	return addr
}
