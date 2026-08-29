package telemetry

import (
	"testing"

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
