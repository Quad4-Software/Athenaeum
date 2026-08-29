package models

// KosyncDocument is one user's stored reading position for a KOReader
// sync document hash, matching the fields KOReader's sync plugin sends.
type KosyncDocument struct {
	Document   string  `json:"document"`
	Progress   string  `json:"progress"`
	Percentage float64 `json:"percentage"`
	Device     string  `json:"device"`
	DeviceID   string  `json:"device_id"`
	Timestamp  int64   `json:"timestamp"`
}
