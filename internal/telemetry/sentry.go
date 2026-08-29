package telemetry

import (
	"net/http"
	"net/url"
	"time"

	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"

	"athenaeum/internal/config"
	"athenaeum/internal/version"
)

// PublicConfig exposes client-side Sentry/GlitchTip settings via /api/health.
type PublicConfig struct {
	SentryDSN        string  `json:"sentryDsn,omitempty"`
	Environment      string  `json:"environment,omitempty"`
	Release          string  `json:"release,omitempty"`
	TracesSampleRate float64 `json:"tracesSampleRate,omitempty"`
}

// Init configures the Sentry SDK when a DSN is set. GlitchTip uses the same protocol.
func Init(cfg config.Config) error {
	dsn := cfg.SentryDSN
	if dsn == "" {
		return nil
	}
	release := cfg.SentryRelease
	if release == "" {
		release = version.Version
	}
	return sentry.Init(sentry.ClientOptions{
		Dsn:              dsn,
		Environment:      cfg.SentryEnvironment,
		Release:          release,
		TracesSampleRate: cfg.SentryTracesSampleRate,
		AttachStacktrace: true,
	})
}

// Flush waits for pending Sentry events before shutdown.
func Flush(timeout time.Duration) {
	if !Enabled() {
		return
	}
	sentry.Flush(timeout)
}

// Enabled reports whether the Sentry client is active.
func Enabled() bool {
	return sentry.CurrentHub().Client() != nil
}

// PublicFromConfig returns frontend telemetry settings derived from server config.
func PublicFromConfig(cfg config.Config) PublicConfig {
	dsn := cfg.SentryPublicDSN()
	if dsn == "" {
		return PublicConfig{}
	}
	release := cfg.SentryRelease
	if release == "" {
		release = version.Version
	}
	return PublicConfig{
		SentryDSN:        dsn,
		Environment:      cfg.SentryEnvironment,
		Release:          release,
		TracesSampleRate: cfg.SentryTracesSampleRate,
	}
}

// ConnectHost returns the origin to allow in connect-src CSP for the given DSN.
func ConnectHost(dsn string) string {
	u, err := url.Parse(dsn)
	if err != nil || u.Host == "" {
		return ""
	}
	return u.Scheme + "://" + u.Host
}

// HTTPMiddleware wraps handlers with panic recovery and request context for Sentry.
func HTTPMiddleware() func(http.Handler) http.Handler {
	if !Enabled() {
		return func(next http.Handler) http.Handler { return next }
	}
	return sentryhttp.New(sentryhttp.Options{Repanic: true}).Handle
}

// CaptureException reports an error when Sentry is enabled.
func CaptureException(err error) {
	if err == nil || !Enabled() {
		return
	}
	sentry.CaptureException(err)
}
