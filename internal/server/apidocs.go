package server

import "athenaeum/internal/brand"

// APIDoc describes the HTTP API for clients and the settings documentation page.
type APIDoc struct {
	Title        string          `json:"title"`
	Version      string          `json:"version"`
	Auth         []string        `json:"auth"`
	BaseURL      string          `json:"baseUrl"`
	ContentTypes []string        `json:"contentTypes"`
	Sections     []APIDocSection `json:"sections"`
}

type APIDocSection struct {
	Title     string           `json:"title"`
	Endpoints []APIDocEndpoint `json:"endpoints"`
}

type APIDocEndpoint struct {
	Method   string `json:"method"`
	Path     string `json:"path"`
	Summary  string `json:"summary"`
	Auth     string `json:"auth,omitempty"`
	Query    string `json:"query,omitempty"`
	Body     string `json:"body,omitempty"`
	Response string `json:"response,omitempty"`
}

func apiDocumentation() APIDoc {
	return APIDoc{
		Title:   brand.Name + " HTTP API",
		Version: "1",
		Auth: []string{
			"Session cookie (browser): sign in via POST /api/auth/login; cookies are sent automatically.",
			"HTTP Basic Auth: Authorization: Basic base64(username:password) — useful for OPDS clients and scripts.",
			"API key: Authorization: Bearer " + brand.APIKeyPrefix + "<token> or X-API-Key: " + brand.APIKeyPrefix + "<token> — create keys in Settings -> API.",
		},
		BaseURL: "/api",
		ContentTypes: []string{
			"JSON endpoints accept and return application/json unless noted.",
			"File endpoints stream binary content with appropriate Content-Type and Accept-Ranges.",
			"Mutating browser requests require the X-CSRF-Token header (not needed for Basic Auth or API keys).",
		},
		Sections: []APIDocSection{
			{
				Title: "Health and auth",
				Endpoints: []APIDocEndpoint{
					{Method: "GET", Path: "/api/health", Summary: "Health check and optional Sentry/GlitchTip client config", Auth: "public"},
					{Method: "GET", Path: "/api/i18n/locales", Summary: "List available UI locales", Auth: "public"},
					{Method: "GET", Path: "/api/i18n/template", Summary: "Translation template (keys with empty values)", Auth: "public"},
					{Method: "GET", Path: "/api/i18n/{locale}", Summary: "Messages for one locale", Auth: "public"},
					{Method: "GET", Path: "/api/docs", Summary: "This API reference (JSON)", Auth: "public"},
					{Method: "GET", Path: "/api/openapi.json", Summary: "OpenAPI 3.0 specification", Auth: "public"},
					{Method: "GET", Path: "/docs", Summary: "Interactive OpenAPI documentation UI", Auth: "public"},
					{Method: "GET", Path: "/api/auth/setup", Summary: "Whether initial setup is needed (includes optional altcha config)", Auth: "public"},
					{Method: "POST", Path: "/api/auth/setup", Summary: "Create first admin", Auth: "public", Body: `{"username","password","altcha?"}`},
					{Method: "POST", Path: "/api/auth/login", Summary: "Create session", Auth: "public", Body: `{"username","password","altcha?"}`},
					{Method: "GET", Path: "/api/altcha/challenge", Summary: "Fresh ALTCHA PoW challenge (builtin mode)", Auth: "public"},
					{Method: "POST", Path: "/api/auth/logout", Summary: "End session", Auth: "session"},
					{Method: "POST", Path: "/api/auth/refresh", Summary: "Rotate access token", Auth: "refresh cookie"},
					{Method: "GET", Path: "/api/auth/me", Summary: "Current user", Auth: "required"},
					{Method: "GET", Path: "/api/auth/csrf", Summary: "Issue CSRF token cookie", Auth: "public"},
					{Method: "GET", Path: "/api/auth/methods", Summary: "Available login methods", Auth: "public"},
					{Method: "GET", Path: "/api/auth/reader-prefs", Summary: "Your synced reader preferences", Auth: "required"},
					{Method: "PUT", Path: "/api/auth/reader-prefs", Summary: "Save reader preferences", Auth: "required", Body: `{"prefs":{}}`},
					{Method: "GET", Path: "/api/auth/api-keys", Summary: "List your API keys", Auth: "required"},
					{Method: "POST", Path: "/api/auth/api-keys", Summary: "Create API key (secret shown once)", Auth: "required", Body: `{"name"}`, Response: `{"id","userId","name","prefix","createdAt","key"}`},
					{Method: "DELETE", Path: "/api/auth/api-keys/{id}", Summary: "Revoke API key", Auth: "required"},
				},
			},
			{
				Title: "Library mounts",
				Endpoints: []APIDocEndpoint{
					{Method: "GET", Path: "/api/libraries", Summary: "List mounted library roots", Auth: "required"},
					{Method: "POST", Path: "/api/libraries", Summary: "Add mount (local or s3)", Auth: "admin", Body: `{"name","backend","mountPath","s3"}`},
					{Method: "POST", Path: "/api/libraries/test-s3", Summary: "Validate S3 credentials", Auth: "admin", Body: `{"endpoint","bucket","accessKey","secretKey",...}`},
					{Method: "PUT", Path: "/api/libraries/reorder", Summary: "Reorder mounts", Auth: "admin", Body: `{"ids":[1,2]}`},
					{Method: "GET", Path: "/api/libraries/{id}", Summary: "Get one mount", Auth: "required"},
					{Method: "PUT", Path: "/api/libraries/{id}", Summary: "Update mount", Auth: "admin"},
					{Method: "DELETE", Path: "/api/libraries/{id}", Summary: "Remove mount", Auth: "admin"},
					{Method: "POST", Path: "/api/libraries/{id}/scan", Summary: "Scan one mount", Auth: "admin"},
					{Method: "GET", Path: "/api/library/stats", Summary: "Counts and flags", Auth: "required", Query: "library"},
					{Method: "GET", Path: "/api/library/scan/status", Summary: "Live scan progress", Auth: "required"},
					{Method: "POST", Path: "/api/library/scan", Summary: "Background rescan all", Auth: "admin"},
				},
			},
			{
				Title: "Books",
				Endpoints: []APIDocEndpoint{
					{Method: "GET", Path: "/api/books", Summary: "List books (paginated)", Auth: "required", Query: "search, sort, format, series, author, library, collection, inProgress, favorites, tag, limit, offset"},
					{Method: "GET", Path: "/api/books/{id}", Summary: "Book detail", Auth: "required"},
					{Method: "PUT", Path: "/api/books/{id}", Summary: "Edit metadata", Auth: "admin"},
					{Method: "GET", Path: "/api/books/{id}/cover", Summary: "Cover image", Auth: "required"},
					{Method: "GET", Path: "/api/books/{id}/file", Summary: "Stream file (HTTP Range)", Auth: "required"},
					{Method: "GET", Path: "/api/books/{id}/download", Summary: "Download attachment", Auth: "required"},
					{Method: "GET", Path: "/api/books/{id}/chapters", Summary: "Audiobook chapters", Auth: "required"},
					{Method: "GET", Path: "/api/books/{id}/progress", Summary: "Reading progress", Auth: "required"},
					{Method: "PUT", Path: "/api/books/{id}/progress", Summary: "Save progress", Auth: "required", Body: `{"location","percent"}`},
					{Method: "PUT", Path: "/api/books/{id}/favorite", Summary: "Toggle favorite", Auth: "required", Body: `{"favorite":true}`},
					{Method: "GET", Path: "/api/series", Summary: "Series with counts", Auth: "required", Query: "library"},
					{Method: "GET", Path: "/api/authors", Summary: "Authors with counts", Auth: "required", Query: "library"},
					{Method: "GET", Path: "/api/favorites", Summary: "Favorite book IDs", Auth: "required"},
					{Method: "GET", Path: "/api/tags", Summary: "List all tags", Auth: "required"},
					{Method: "POST", Path: "/api/tags", Summary: "Create a tag", Auth: "required", Body: `{"name"}`},
					{Method: "GET", Path: "/api/books/{id}/tags", Summary: "List tags on a book", Auth: "required"},
					{Method: "PUT", Path: "/api/books/{id}/tags", Summary: "Replace a book's tags", Auth: "required", Body: `{"tags":["fiction","favorite"]}`},
					{Method: "POST", Path: "/api/books/{id}/tags", Summary: "Add one tag to a book", Auth: "required", Body: `{"name"}`},
					{Method: "DELETE", Path: "/api/books/{id}/tags/{tagId}", Summary: "Remove one tag from a book", Auth: "required"},
					{Method: "GET", Path: "/api/books/{id}/rating", Summary: "Your rating for a book", Auth: "required"},
					{Method: "PUT", Path: "/api/books/{id}/rating", Summary: "Set your rating (1-5, or 0 to clear)", Auth: "required", Body: `{"rating"}`},
				},
			},
			{
				Title: "Collections",
				Endpoints: []APIDocEndpoint{
					{Method: "GET", Path: "/api/collections", Summary: "List shelves", Auth: "required"},
					{Method: "POST", Path: "/api/collections", Summary: "Create shelf", Auth: "required", Body: `{"name","description","kind","query"}`},
					{Method: "GET", Path: "/api/collections/{id}", Summary: "Get shelf", Auth: "required"},
					{Method: "PUT", Path: "/api/collections/{id}", Summary: "Update shelf", Auth: "required"},
					{Method: "DELETE", Path: "/api/collections/{id}", Summary: "Delete shelf", Auth: "required"},
					{Method: "POST", Path: "/api/collections/{id}/books/{bookId}", Summary: "Add book to manual shelf", Auth: "required"},
					{Method: "DELETE", Path: "/api/collections/{id}/books/{bookId}", Summary: "Remove book from shelf", Auth: "required"},
				},
			},
			{
				Title: "Uploads",
				Endpoints: []APIDocEndpoint{
					{Method: "POST", Path: "/api/libraries/{id}/uploads", Summary: "Start resumable upload", Auth: "required", Body: `{"relPath","totalSize"}`},
					{Method: "GET", Path: "/api/libraries/{id}/uploads/{uploadId}", Summary: "Upload status", Auth: "required"},
					{Method: "PATCH", Path: "/api/libraries/{id}/uploads/{uploadId}", Summary: "Upload chunk", Auth: "required", Body: "application/octet-stream", Query: "Content-Range header"},
					{Method: "DELETE", Path: "/api/libraries/{id}/uploads/{uploadId}", Summary: "Cancel upload", Auth: "required"},
				},
			},
			{
				Title: "Admin",
				Endpoints: []APIDocEndpoint{
					{Method: "POST", Path: "/api/auth/register", Summary: "Create user", Auth: "admin"},
					{Method: "POST", Path: "/api/auth/users/guest", Summary: "Create temporary guest account", Auth: "admin", Body: `{"username","expiresInHours","permissions"}`, Response: `{"user","password"}`},
					{Method: "POST", Path: "/api/invites", Summary: "Create invite (permanent or guest)", Auth: "admin", Body: `{"kind","email","permissions","expiresInHours","guestExpiresInHours","provisionPocketId"}`},
					{Method: "GET", Path: "/api/invites", Summary: "List invites", Auth: "admin", Query: "status"},
					{Method: "DELETE", Path: "/api/invites/{id}", Summary: "Revoke invite", Auth: "admin"},
					{Method: "GET", Path: "/api/invite/{token}", Summary: "Invite metadata for accept page", Auth: "public"},
					{Method: "POST", Path: "/api/invite/{token}/accept", Summary: "Accept invite", Auth: "public"},
					{Method: "GET", Path: "/api/admin/webhooks", Summary: "List webhooks", Auth: "admin"},
					{Method: "POST", Path: "/api/admin/webhooks", Summary: "Create webhook", Auth: "admin"},
					{Method: "PUT", Path: "/api/admin/webhooks/{id}", Summary: "Update webhook", Auth: "admin"},
					{Method: "DELETE", Path: "/api/admin/webhooks/{id}", Summary: "Delete webhook", Auth: "admin"},
					{Method: "GET", Path: "/api/admin/webhooks/{id}/deliveries", Summary: "Webhook delivery log", Auth: "admin"},
					{Method: "POST", Path: "/api/admin/webhooks/{id}/test", Summary: "Send test ping", Auth: "admin"},
					{Method: "GET", Path: "/api/admin/pocketid", Summary: "Pocket ID connector settings", Auth: "admin"},
					{Method: "PUT", Path: "/api/admin/pocketid", Summary: "Update Pocket ID connector", Auth: "admin"},
					{Method: "POST", Path: "/api/admin/pocketid/test", Summary: "Test Pocket ID API key", Auth: "admin"},
					{Method: "POST", Path: "/api/admin/pocketid/apply-oidc", Summary: "Apply Pocket ID issuer to OIDC config", Auth: "admin"},
					{Method: "GET", Path: "/api/auth/users", Summary: "List users", Auth: "admin"},
					{Method: "PUT", Path: "/api/auth/users/{id}/password", Summary: "Reset password", Auth: "admin"},
					{Method: "PUT", Path: "/api/auth/users/{id}/admin", Summary: "Grant/revoke admin", Auth: "admin"},
					{Method: "DELETE", Path: "/api/auth/users/{id}", Summary: "Delete user", Auth: "admin"},
					{Method: "GET", Path: "/api/auth/users/{id}/libraries", Summary: "Per-user library access", Auth: "admin"},
					{Method: "PUT", Path: "/api/auth/users/{id}/libraries", Summary: "Set library access", Auth: "admin"},
					{Method: "GET", Path: "/api/auth/audit", Summary: "Audit log", Auth: "admin", Query: "limit, offset, action, q"},
					{Method: "GET", Path: "/api/admin/server", Summary: "Server settings (metrics, proxy, CSP, CORS)", Auth: "admin"},
					{Method: "PUT", Path: "/api/admin/server", Summary: "Update server settings", Auth: "admin"},
					{Method: "GET", Path: "/api/admin/tts", Summary: "Kokoro TTS sidecar settings", Auth: "admin"},
					{Method: "PUT", Path: "/api/admin/tts", Summary: "Update Kokoro TTS settings", Auth: "admin"},
					{Method: "POST", Path: "/api/admin/tts/test", Summary: "Ping Kokoro TTS sidecar", Auth: "admin"},
					{Method: "GET", Path: "/api/tts/status", Summary: "Whether Kokoro TTS is enabled", Auth: "optional"},
					{Method: "GET", Path: "/api/tts/voices", Summary: "List Kokoro voices", Auth: "optional"},
					{Method: "POST", Path: "/api/tts/synthesize", Summary: "Synthesize speech via Kokoro sidecar", Auth: "optional"},
					{Method: "GET", Path: "/api/system/stats", Summary: "Host CPU/memory/disk and version", Auth: "admin"},
					{Method: "GET", Path: "/metrics", Summary: "Prometheus metrics", Auth: "optional Basic Auth when enabled"},
					{Method: "GET", Path: "/api/auth/sessions", Summary: "Your active sessions", Auth: "required"},
					{Method: "DELETE", Path: "/api/auth/sessions/{id}", Summary: "Revoke session", Auth: "required"},
				},
			},
			{
				Title: "OPDS catalog",
				Endpoints: []APIDocEndpoint{
					{Method: "GET", Path: "/opds/", Summary: "Navigation catalog", Auth: "required when users exist"},
					{Method: "GET", Path: "/opds/recent", Summary: "Recent acquisitions", Auth: "required when users exist"},
					{Method: "GET", Path: "/opds/series", Summary: "Series navigation", Auth: "required when users exist"},
					{Method: "GET", Path: "/opds/series/{name}", Summary: "Books in series", Auth: "required when users exist"},
					{Method: "GET", Path: "/opds/search", Summary: "Search feed", Auth: "required when users exist", Query: "q"},
				},
			},
		},
	}
}
