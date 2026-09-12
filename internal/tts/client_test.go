package tts

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"athenaeum/internal/models"
)

func TestTruncateErr(t *testing.T) {
	long := make([]byte, 250)
	for i := range long {
		long[i] = 'a'
	}
	if trunc := truncateErr(long); len(trunc) <= 200 || !strings.HasSuffix(trunc, "…") {
		t.Fatalf("truncateErr=%q", trunc)
	}
	if truncateErr([]byte(" short ")) != "short" {
		t.Fatal("truncateErr short")
	}
}

func TestSynthesizeSendsOpenAIShape(t *testing.T) {
	var got map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/audio/speech" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewDecoder(r.Body).Decode(&got)
		w.Header().Set("Content-Type", "audio/mpeg")
		_, _ = w.Write([]byte("ID3fake"))
	}))
	defer srv.Close()

	c := NewClient(models.TTSSettings{
		Enabled: true, BaseURL: srv.URL, Model: "kokoro", DefaultVoice: "af_heart",
		ResponseFormat: "mp3", TimeoutSec: 10,
	})
	audio, ct, err := c.Synthesize(context.Background(), "Hello world", "", 1.25)
	if err != nil {
		t.Fatal(err)
	}
	if string(audio) != "ID3fake" || ct != "audio/mpeg" {
		t.Fatalf("unexpected audio=%q ct=%q", audio, ct)
	}
	if got["input"] != "Hello world" || got["voice"] != "af_heart" ||
		got["response_format"] != "mp3" || got["model"] != "kokoro" || got["speed"] != 1.25 {
		t.Fatalf("unexpected request body %v", got)
	}
}

func TestVoicesFallbackToV1(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/audio/voices" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"data":[{"id":"af_heart","name":"Heart"}]}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	c := NewClient(models.TTSSettings{Enabled: true, BaseURL: srv.URL, TimeoutSec: 10})
	voices, err := c.Voices(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(voices) != 1 || voices[0].ID != "af_heart" || voices[0].Label != "Heart" {
		t.Fatalf("unexpected voices %#v", voices)
	}
}
