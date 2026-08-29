package models

// ServerConfig holds admin-configurable server behavior.
type ServerConfig struct {
	MetricsEnabled   bool   `json:"metricsEnabled"`
	MetricsAuth      bool   `json:"metricsAuth"`
	MetricsUsername  string `json:"metricsUsername"`
	MetricsPassword  string `json:"metricsPassword,omitempty"`
	PasswordSet      bool   `json:"metricsPasswordSet"`
	TrustedProxies   string `json:"trustedProxies"`
	CORSEnabled      bool   `json:"corsEnabled"`
	CORSOrigins      string `json:"corsOrigins"`
	CSPEnabled       bool   `json:"cspEnabled"`
	CSPPolicy        string `json:"cspPolicy"`
	AutoScanEnabled  bool   `json:"autoScanEnabled"`
	AutoScanInterval int    `json:"autoScanIntervalSec"`
	ScanWorkers      int    `json:"scanWorkers"`
}

// ServerConfigPublic is exposed to admins without secrets.
type ServerConfigPublic struct {
	MetricsEnabled   bool   `json:"metricsEnabled"`
	MetricsAuth      bool   `json:"metricsAuth"`
	MetricsUsername  string `json:"metricsUsername"`
	PasswordSet      bool   `json:"metricsPasswordSet"`
	TrustedProxies   string `json:"trustedProxies"`
	CORSEnabled      bool   `json:"corsEnabled"`
	CORSOrigins      string `json:"corsOrigins"`
	CSPEnabled       bool   `json:"cspEnabled"`
	CSPPolicy        string `json:"cspPolicy"`
	AutoScanEnabled  bool   `json:"autoScanEnabled"`
	AutoScanInterval int    `json:"autoScanIntervalSec"`
	ScanWorkers      int    `json:"scanWorkers"`
}

func (c ServerConfig) Public() ServerConfigPublic {
	return ServerConfigPublic{
		MetricsEnabled:   c.MetricsEnabled,
		MetricsAuth:      c.MetricsAuth,
		MetricsUsername:  c.MetricsUsername,
		PasswordSet:      c.PasswordSet,
		TrustedProxies:   c.TrustedProxies,
		CORSEnabled:      c.CORSEnabled,
		CORSOrigins:      c.CORSOrigins,
		CSPEnabled:       c.CSPEnabled,
		CSPPolicy:        c.CSPPolicy,
		AutoScanEnabled:  c.AutoScanEnabled,
		AutoScanInterval: c.AutoScanInterval,
		ScanWorkers:      c.ScanWorkers,
	}
}
