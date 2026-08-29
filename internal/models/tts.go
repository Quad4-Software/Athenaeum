package models

// TTSSettings holds optional Kokoro (or compatible) sidecar configuration
// used for server-proxied narration.
type TTSSettings struct {
	Enabled      bool   `json:"enabled"`
	BaseURL      string `json:"baseUrl"`
	APIKey       string `json:"apiKey,omitempty"`
	DefaultVoice string `json:"defaultVoice"`
	TimeoutSec   int    `json:"timeoutSec"`
}

// TTSSettingsPublic masks the API key for admin responses.
type TTSSettingsPublic struct {
	Enabled      bool   `json:"enabled"`
	BaseURL      string `json:"baseUrl"`
	DefaultVoice string `json:"defaultVoice"`
	APIKeySet    bool   `json:"apiKeySet"`
	TimeoutSec   int    `json:"timeoutSec"`
}

// Public strips secrets from TTS settings.
func (c TTSSettings) Public() TTSSettingsPublic {
	return TTSSettingsPublic{
		Enabled:      c.Enabled,
		BaseURL:      c.BaseURL,
		DefaultVoice: c.DefaultVoice,
		APIKeySet:    c.APIKey != "",
		TimeoutSec:   c.TimeoutSec,
	}
}

// TTSStatus is the reduced view exposed to any authenticated reader.
type TTSStatus struct {
	Enabled      bool   `json:"enabled"`
	DefaultVoice string `json:"defaultVoice"`
}
