package models

// PocketIDSettings holds the Pocket ID Admin API connector configuration.
type PocketIDSettings struct {
	Enabled         bool     `json:"enabled"`
	BaseURL         string   `json:"baseUrl"`
	APIKey          string   `json:"apiKey,omitempty"`
	DefaultGroupIDs []string `json:"defaultGroupIds"`
}

// PocketIDSettingsPublic masks the API key.
type PocketIDSettingsPublic struct {
	Enabled         bool     `json:"enabled"`
	BaseURL         string   `json:"baseUrl"`
	APIKeySet       bool     `json:"apiKeySet"`
	DefaultGroupIDs []string `json:"defaultGroupIds"`
}

// Public strips the API key for client responses.
func (c PocketIDSettings) Public() PocketIDSettingsPublic {
	groups := c.DefaultGroupIDs
	if groups == nil {
		groups = []string{}
	}
	return PocketIDSettingsPublic{
		Enabled:         c.Enabled,
		BaseURL:         c.BaseURL,
		APIKeySet:       c.APIKey != "",
		DefaultGroupIDs: groups,
	}
}
