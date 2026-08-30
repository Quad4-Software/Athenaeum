// Package brand holds product identity constants for forks.
// Change values here to rename the app across server surfaces.
package brand

import "fmt"

const (
	Name        = "Athenaeum"
	ShortName   = "Athenaeum"
	Description = "Self-hosted EPUB, PDF, and audiobook library with search, collections, and reading progress."

	DBFilename          = "athenaeum.db"
	BackupPrefix        = "athenaeum-backup"
	ConfigExportName    = "athenaeum-config.json"
	APIKeyPrefix        = "ath_"
	OIDCStateCookie     = "athenaeum_oidc_state"
	InvitePendingCookie = "athenaeum_invite_pending"

	MetricsPrefix = "athenaeum_"
)

// UserAgent is sent on outbound HTTP requests (cover fetch, metadata).
func UserAgent(version string) string {
	if version == "" {
		return Name + "/1.0"
	}
	return fmt.Sprintf("%s/%s", Name, version)
}
