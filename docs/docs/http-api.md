---
sidebar_position: 8
title: HTTP API
description: REST, OpenAPI, OPDS, KOSync, and admin endpoints.
---

# HTTP API

Interactive API docs (served by a running Athenaeum instance): `/docs`.
Machine-readable: `GET /api/openapi.json` (OpenAPI 3) and `GET /api/docs` (JSON).
Key management is under Settings -> API.

Mutating browser requests require the `X-CSRF-Token` header (not needed for
Basic Auth or API keys). Auth when users exist: session cookie, HTTP Basic, or
API key (`Authorization: Bearer ath_<token>` or `X-API-Key: ath_<token>`).


## Docs and health

| Method | Path | Description |
| ------ | ---- | ----------- |
| GET | /api/docs | API reference (JSON) |
| GET | /api/health | Health check and optional telemetry config |
| GET | /api/i18n/locales | List UI locales |
| GET | /api/i18n/template | Translation template |
| GET | `/api/i18n/{locale}` | Messages for one locale |
| GET | /api/openapi.json | OpenAPI 3.0 specification |
| GET | /docs | Interactive OpenAPI UI |
| GET | /docs/app.js | Interactive OpenAPI UI |
| GET | /metrics | Prometheus metrics (if enabled) |

## Authentication

| Method | Path | Description |
| ------ | ---- | ----------- |
| GET | /api/altcha/challenge | ALTCHA PoW challenge (builtin) |
| GET | /api/auth/api-keys | List API keys |
| POST | /api/auth/api-keys | Create API key |
| DELETE | `/api/auth/api-keys/{id}` | Revoke API key |
| GET | /api/auth/audit | Audit log |
| GET | /api/auth/csrf | Issue CSRF token cookie |
| GET | /api/auth/kindle-email | Per-user Kindle/email delivery address |
| PUT | /api/auth/kindle-email | Per-user Kindle/email delivery address |
| POST | /api/auth/login | Create session |
| POST | /api/auth/logout | End session |
| GET | /api/auth/me | Current user |
| GET | /api/auth/methods | Available login methods |
| GET | /api/auth/oidc/config | OIDC / SSO configuration |
| PUT | /api/auth/oidc/config | OIDC / SSO configuration |
| POST | /api/auth/oidc/discover | Discover OIDC endpoints from issuer |
| PUT | /api/auth/password | Change own password |
| POST | /api/auth/password/check | Check password against policy |
| PUT | /api/auth/profile | Rename current user |
| GET | /api/auth/reader-prefs | Synced reader preferences |
| PUT | /api/auth/reader-prefs | Synced reader preferences |
| POST | /api/auth/refresh | Rotate access token |
| POST | /api/auth/register | Create user (admin) |
| POST | /api/auth/register-public | Self-register (when enabled) |
| DELETE | /api/auth/sessions | Revoke all other sessions |
| GET | /api/auth/sessions | List your sessions |
| DELETE | `/api/auth/sessions/{id}` | Revoke one session |
| GET | /api/auth/settings | Auth settings (registration, require TOTP) |
| PUT | /api/auth/settings | Auth settings (registration, require TOTP) |
| GET | /api/auth/setup | Setup / auth status |
| POST | /api/auth/setup | Create first admin |
| POST | /api/auth/totp/disable | Disable TOTP |
| POST | /api/auth/totp/enable | Confirm TOTP with code |
| POST | /api/auth/totp/setup | Create first admin |
| POST | /api/auth/totp/verify | Verify TOTP during login |
| GET | /api/auth/users | List users |
| POST | /api/auth/users/guest | Create guest account |
| GET | /api/auth/users/guests | List guests |
| POST | /api/auth/users/guests/bulk-delete | Revoke guest accounts |
| POST | `/api/auth/users/guests/{id}/extend` | Extend guest expiry |
| POST | /api/invites | Create invite |
| GET | /api/invites | List invites |
| DELETE | `/api/invites/{id}` | Revoke invite |
| GET | `/api/invite/{token}` | Invite metadata (public) |
| POST | `/api/invite/{token}/accept` | Accept invite (public) |
| DELETE | `/api/auth/users/{id}` | Delete user |
| PUT | `/api/auth/users/{id}/admin` | Grant/revoke admin |
| GET | `/api/auth/users/{id}/libraries` | Per-user library access |
| PUT | `/api/auth/users/{id}/libraries` | Per-user library access |
| PUT | `/api/auth/users/{id}/password` | Reset password (admin) |
| PUT | `/api/auth/users/{id}/permissions` | Set non-admin permissions |
| GET | `/api/auth/users/{id}/sessions` | List sessions for a user (admin) |
| GET | /api/authors | Authors with counts |
| GET | /auth/oidc/callback | OIDC callback |
| GET | /auth/oidc/login | Create session |

## Library mounts and scan

| Method | Path | Description |
| ------ | ---- | ----------- |
| GET | /api/fs/browse | Browse host filesystem (admin) |
| GET | /api/libraries | List library mounts |
| POST | /api/libraries | Add library mount (local or s3) |
| POST | /api/libraries/test-s3 | Validate S3 credentials |
| PUT | /api/libraries/reorder | Reorder mounts |
| DELETE | `/api/libraries/{id}` | Remove mount |
| GET | `/api/libraries/{id}` | Get one mount |
| PUT | `/api/libraries/{id}` | Update mount |
| POST | `/api/libraries/{id}/scan` | Scan one mount |
| POST | `/api/libraries/{id}/uploads` | Start resumable upload |
| DELETE | `/api/libraries/{id}/uploads/{uploadId}` | Cancel upload |
| GET | `/api/libraries/{id}/uploads/{uploadId}` | Upload status |
| PATCH | `/api/libraries/{id}/uploads/{uploadId}` | Upload chunk |
| POST | /api/library/metadata/match | Start bulk metadata match |
| GET | /api/library/metadata/match/status | Bulk metadata match status |
| POST | /api/library/scan | Background rescan all |
| GET | /api/library/scan/status | Live scan progress |
| POST | /api/library/series/cleanup | Cleanup series names |
| GET | /api/library/stats | Library counts and flags |
| GET | /api/metadata/providers | Metadata providers |

## Books and reading

| Method | Path | Description |
| ------ | ---- | ----------- |
| GET | /api/authors | Authors with counts |
| GET | /api/books | List books |
| DELETE | `/api/books/{id}` | Delete book |
| GET | `/api/books/{id}` | Book detail |
| PUT | `/api/books/{id}` | Edit metadata |
| GET | `/api/books/{id}/bookmarks` | List bookmarks |
| POST | `/api/books/{id}/bookmarks` | Add bookmark |
| DELETE | `/api/books/{id}/bookmarks/{bookmarkId}` | Delete bookmark |
| GET | `/api/books/{id}/chapters` | Audiobook chapters |
| POST | `/api/books/{id}/convert` | Convert book format |
| DELETE | `/api/books/{id}/cover` | Remove cover |
| GET | `/api/books/{id}/cover` | Cover image |
| PUT | `/api/books/{id}/cover` | Upload cover |
| PUT | `/api/books/{id}/cover-from-url` | Set cover from URL |
| GET | `/api/books/{id}/download` | Download attachment |
| GET | `/api/books/{id}/favorite` | Favorite status / toggle |
| PUT | `/api/books/{id}/favorite` | Favorite status / toggle |
| GET | `/api/books/{id}/file` | Stream file (HTTP Range) |
| GET | `/api/books/{id}/highlights` | List highlights |
| POST | `/api/books/{id}/highlights` | Add highlight |
| DELETE | `/api/books/{id}/highlights/{highlightId}` | Delete highlight |
| POST | `/api/books/{id}/metadata/apply` | Apply metadata match |
| POST | `/api/books/{id}/metadata/search` | Search external metadata |
| GET | `/api/books/{id}/mobi-sections` | MOBI/AZW section list |
| GET | `/api/books/{id}/pages` | Comic page list |
| GET | `/api/books/{id}/pages/{page}` | Comic page image |
| GET | `/api/books/{id}/progress` | Reading progress |
| PUT | `/api/books/{id}/progress` | Reading progress |
| GET | `/api/books/{id}/rating` | Star rating (1-5, 0 clears) |
| PUT | `/api/books/{id}/rating` | Star rating (1-5, 0 clears) |
| POST | `/api/books/{id}/reading-time` | Add reading seconds |
| POST | `/api/books/{id}/send` | Send via SMTP/Kindle |
| GET | `/api/books/{id}/share` | List share links |
| POST | `/api/books/{id}/share` | Create share link |
| DELETE | `/api/books/{id}/share/{shareId}` | Revoke share link |
| GET | `/api/books/{id}/tags` | List tags on book |
| POST | `/api/books/{id}/tags` | Add tag to book |
| PUT | `/api/books/{id}/tags` | Replace book tags |
| DELETE | `/api/books/{id}/tags/{tagId}` | Remove tag from book |
| GET | `/api/books/{id}/tracks` | Audiobook tracks |
| GET | /api/favorites | Favorite status / toggle |
| GET | /api/series | Series with counts |
| GET | /api/stats/reading | User reading stats |
| GET | /api/tags | List tags |
| POST | /api/tags | Create tag |

## Collections

| Method | Path | Description |
| ------ | ---- | ----------- |
| GET | /api/collections | List shelves |
| POST | /api/collections | Create shelf |
| DELETE | `/api/collections/{id}` | Delete shelf |
| GET | `/api/collections/{id}` | Get shelf |
| PUT | `/api/collections/{id}` | Update shelf |
| DELETE | `/api/collections/{id}/books/{bookId}` | Remove book from shelf |
| POST | `/api/collections/{id}/books/{bookId}` | Add book to shelf |

## Sharing and delivery

| Method | Path | Description |
| ------ | ---- | ----------- |
| GET | /api/admin/smtp | SMTP delivery settings |
| PUT | /api/admin/smtp | SMTP delivery settings |
| GET | /api/auth/kindle-email | Per-user Kindle/email delivery address |
| PUT | /api/auth/kindle-email | Per-user Kindle/email delivery address |
| POST | `/api/books/{id}/send` | Send via SMTP/Kindle |
| GET | `/api/books/{id}/share` | List share links |
| POST | `/api/books/{id}/share` | Create share link |
| DELETE | `/api/books/{id}/share/{shareId}` | Revoke share link |
| GET | `/api/share/{token}` | Share metadata (public) |
| GET | /opds/kindle | OPDS Kindle-friendly feed |
| GET | `/share/{token}/download` | Download attachment |

## Narration (TTS)

| Method | Path | Description |
| ------ | ---- | ----------- |
| GET | /api/admin/tts | Kokoro TTS settings |
| PUT | /api/admin/tts | Update Kokoro TTS settings |
| POST | /api/admin/tts/test | Ping Kokoro sidecar |
| GET | /api/tts/status | Whether Kokoro TTS is enabled |
| POST | /api/tts/synthesize | Synthesize speech via Kokoro |
| GET | /api/tts/voices | List Kokoro voices |

## Admin and maintenance

| Method | Path | Description |
| ------ | ---- | ----------- |
| GET | /api/admin/backup | Download data zip |
| GET | /api/admin/config/export | Export config JSON |
| POST | /api/admin/config/import | Import config JSON |
| POST | /api/admin/content-index | Index EPUB full text |
| POST | /api/admin/restore | Restore from zip |
| GET | /api/admin/server | Server settings (metrics, proxy, CSP, CORS) |
| PUT | /api/admin/server | Server settings (metrics, proxy, CSP, CORS) |
| GET | /api/admin/smtp | SMTP delivery settings |
| PUT | /api/admin/smtp | SMTP delivery settings |
| GET | /api/admin/pocketid | Pocket ID connector |
| PUT | /api/admin/pocketid | Pocket ID connector |
| POST | /api/admin/pocketid/test | Test Pocket ID API |
| POST | /api/admin/pocketid/apply-oidc | Apply Pocket ID to OIDC |
| GET | /api/admin/pocketid/signup-tokens | List Pocket ID signup tokens |
| POST | /api/admin/pocketid/signup-tokens | Create Pocket ID signup token |
| DELETE | `/api/admin/pocketid/signup-tokens/{id}` | Delete signup token |
| GET | /api/admin/webhooks | List webhooks |
| POST | /api/admin/webhooks | Create webhook |
| GET | `/api/admin/webhooks/{id}` | Get webhook |
| PUT | `/api/admin/webhooks/{id}` | Update webhook |
| DELETE | `/api/admin/webhooks/{id}` | Delete webhook |
| GET | `/api/admin/webhooks/{id}/deliveries` | Delivery log |
| POST | `/api/admin/webhooks/{id}/test` | Send test ping |
| POST | /api/admin/tasks/cleanup-covers | Cleanup orphan covers |
| POST | /api/admin/tasks/cleanup-series | Cleanup series metadata |
| POST | /api/admin/tasks/cleanup-text | Cleanup text fields |
| POST | /api/admin/tasks/prune-missing | Prune missing files from DB |
| POST | /api/admin/tasks/regenerate-covers | Regenerate covers |
| GET | /api/admin/tasks/status | Maintenance task status |
| POST | /api/admin/tasks/verify | Verify library integrity |
| GET | /api/admin/tts | Kokoro TTS settings |
| PUT | /api/admin/tts | Update Kokoro TTS settings |
| POST | /api/admin/tts/test | Ping Kokoro sidecar |
| DELETE | /api/offline | Revoke offline grant |
| GET | /api/offline | List offline grants |
| POST | /api/offline | Grant offline book |
| GET | /api/system/stats | Host CPU/memory/disk and version |

## OPDS and KOSync

| Method | Path | Description |
| ------ | ---- | ----------- |
| PUT | /kosync/syncs/progress | Reading progress |
| GET | `/kosync/syncs/progress/{document}` | Reading progress |
| GET | /kosync/users/auth | KOSync authenticate |
| GET | /opds/ | OPDS 1.2 navigation catalog |
| GET | /opds/comics | OPDS comics feed |
| GET | /opds/kindle | OPDS Kindle-friendly feed |
| GET | /opds/recent | OPDS recent acquisitions |
| GET | /opds/search | OPDS search feed |
| GET | /opds/series | OPDS series navigation |
| GET | `/opds/series/{name}` | OPDS books in series |
| GET | /opds/v2/ | OPDS 2 catalog |
| GET | /opds/v2/recent | OPDS 2 catalog |
