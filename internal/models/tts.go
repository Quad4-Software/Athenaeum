package models

import "time"

// TTS job lifecycle states stored in tts_jobs.status.
const (
	TTSJobQueued    = "queued"
	TTSJobRunning   = "running"
	TTSJobDone      = "done"
	TTSJobFailed    = "failed"
	TTSJobCancelled = "cancelled"
)

// TTSSettings holds optional Kokoro (or compatible) sidecar configuration
// used for server-proxied narration and audiobook generation.
type TTSSettings struct {
	Enabled        bool   `json:"enabled"`
	BaseURL        string `json:"baseUrl"`
	APIKey         string `json:"apiKey,omitempty"`
	Model          string `json:"model"`
	DefaultVoice   string `json:"defaultVoice"`
	ResponseFormat string `json:"responseFormat"`
	TimeoutSec     int    `json:"timeoutSec"`
}

// TTSSettingsPublic masks the API key for admin responses.
type TTSSettingsPublic struct {
	Enabled        bool   `json:"enabled"`
	BaseURL        string `json:"baseUrl"`
	Model          string `json:"model"`
	DefaultVoice   string `json:"defaultVoice"`
	ResponseFormat string `json:"responseFormat"`
	APIKeySet      bool   `json:"apiKeySet"`
	TimeoutSec     int    `json:"timeoutSec"`
}

// Public strips secrets from TTS settings.
func (c TTSSettings) Public() TTSSettingsPublic {
	return TTSSettingsPublic{
		Enabled:        c.Enabled,
		BaseURL:        c.BaseURL,
		Model:          c.Model,
		DefaultVoice:   c.DefaultVoice,
		ResponseFormat: c.ResponseFormat,
		APIKeySet:      c.APIKey != "",
		TimeoutSec:     c.TimeoutSec,
	}
}

// TTSStatus is the reduced view exposed to any authenticated reader.
type TTSStatus struct {
	Enabled      bool   `json:"enabled"`
	DefaultVoice string `json:"defaultVoice"`
}

// TTSUserPrefs holds a user's narration defaults and optional daily
// generation window for queued audiobook jobs.
type TTSUserPrefs struct {
	UserID       int64   `json:"-"`
	Voice        string  `json:"voice"`
	Speed        float64 `json:"speed"`
	SchedEnabled bool    `json:"schedEnabled"`
	SchedStart   string  `json:"schedStart"` // "HH:MM" server-local
	SchedEnd     string  `json:"schedEnd"`   // "HH:MM" server-local
}

// TTSJob is a queued whole-book narration job.
type TTSJob struct {
	ID            int64      `json:"id"`
	UserID        int64      `json:"userId"`
	BookID        int64      `json:"bookId"`
	BookTitle     string     `json:"bookTitle,omitempty"`
	Voice         string     `json:"voice"`
	Speed         float64    `json:"speed"`
	Status        string     `json:"status"`
	TotalChapters int        `json:"totalChapters"`
	DoneChapters  int        `json:"doneChapters"`
	OutputDir     string     `json:"outputDir,omitempty"`
	Error         string     `json:"error,omitempty"`
	RunAt         int64      `json:"runAt,omitempty"`
	CreatedAt     time.Time  `json:"createdAt"`
	StartedAt     *time.Time `json:"startedAt,omitempty"`
	FinishedAt    *time.Time `json:"finishedAt,omitempty"`
}
