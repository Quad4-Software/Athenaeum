package storage

import (
	"context"
	"path/filepath"
	"testing"

	"athenaeum/internal/models"
)

func TestTTSSettingsRoundTrip(t *testing.T) {
	dir := t.TempDir()
	store, err := Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { store.Close() })

	ctx := context.Background()
	got, err := store.GetTTSSettings(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got.Enabled {
		t.Fatal("expected disabled default")
	}
	if got.DefaultVoice != "af_heart" {
		t.Fatalf("default voice=%q", got.DefaultVoice)
	}

	in := models.TTSSettings{
		Enabled:      true,
		BaseURL:      "http://kokoro:8880",
		APIKey:       "secret",
		DefaultVoice: "am_adam",
		TimeoutSec:   30,
	}
	if err := store.SaveTTSSettings(ctx, in); err != nil {
		t.Fatal(err)
	}
	got, err = store.GetTTSSettings(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Enabled || got.BaseURL != in.BaseURL || got.APIKey != in.APIKey || got.DefaultVoice != in.DefaultVoice || got.TimeoutSec != 30 {
		t.Fatalf("round trip mismatch: %+v", got)
	}
	pub := got.Public()
	if pub.APIKeySet != true || pub.BaseURL != in.BaseURL {
		t.Fatalf("public mismatch: %+v", pub)
	}
}
