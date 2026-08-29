package auth_test

import (
	"testing"

	"athenaeum/internal/auth"
	"athenaeum/internal/brand"
)

// Wire-critical cookie and key prefixes must stay aligned with
// web/src/lib/brand/config.ts (see brand.config.test.ts).
func TestAuthBrandWireConstants(t *testing.T) {
	cases := []struct {
		name string
		got  string
		want string
	}{
		{"CSRFCookie", auth.CSRFCookie, "athenaeum_csrf"},
		{"SessionCookie", auth.SessionCookie, "athenaeum_session"},
		{"RefreshCookie", auth.RefreshCookie, "athenaeum_refresh"},
		{"APIKeyPrefix", brand.APIKeyPrefix, "ath_"},
		{"auth.APIKeyPrefix", auth.APIKeyPrefix, "ath_"},
		{"ConfigExportName", brand.ConfigExportName, "athenaeum-config.json"},
		{"OIDCStateCookie", brand.OIDCStateCookie, "athenaeum_oidc_state"},
	}
	for _, tc := range cases {
		if tc.got != tc.want {
			t.Errorf("%s: got %q want %q", tc.name, tc.got, tc.want)
		}
	}
}
