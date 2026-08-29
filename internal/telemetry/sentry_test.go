package telemetry

import (
	"testing"
	"time"

	"athenaeum/internal/config"
)

func TestConnectHost(t *testing.T) {
	got := ConnectHost("https://abc@glitchtip.example.com/42")
	if got != "https://glitchtip.example.com" {
		t.Fatalf("ConnectHost = %q", got)
	}
	if ConnectHost("not-a-dsn") != "" {
		t.Fatal("expected empty for invalid dsn")
	}
}

func TestPublicFromConfigEmpty(t *testing.T) {
	if got := PublicFromConfig(config.Config{}); got.SentryDSN != "" {
		t.Fatalf("expected empty public config, got %+v", got)
	}
}

func TestPublicFromConfigUsesPublicDSN(t *testing.T) {
	cfg := config.Config{
		SentryDSN:       "https://secret@backend.example.com/1",
		SentryDSNPublic: "https://public@frontend.example.com/2",
	}
	got := PublicFromConfig(cfg)
	if got.SentryDSN != cfg.SentryDSNPublic {
		t.Fatalf("public dsn = %q", got.SentryDSN)
	}
}

func TestInitAndFlush(t *testing.T) {
	if err := Init(config.Config{}); err != nil {
		t.Fatal(err)
	}
	Flush(10 * time.Millisecond)

	err := Init(config.Config{
		SentryDSN:              "https://ffffffffffffffffffffffffffffffff@127.0.0.1/1",
		SentryEnvironment:      "test",
		SentryRelease:          "test-release",
		SentryTracesSampleRate: 0,
	})
	if err != nil {
		t.Fatalf("Init fake dsn: %v", err)
	}
	if !Enabled() {
		t.Fatal("expected sentry enabled after Init")
	}
	Flush(50 * time.Millisecond)
	CaptureException(nil)
	mw := HTTPMiddleware()
	if mw == nil {
		t.Fatal("middleware nil")
	}
}
