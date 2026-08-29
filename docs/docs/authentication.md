---
sidebar_position: 6
title: Authentication
description: Bootstrap, guests, permissions, TOTP, OIDC, API keys, and ALTCHA.
---

# Authentication

## First admin

Create the first admin on startup (only when no users exist yet):

```sh
./bin/athenaeum --admin-user admin --admin-pass 'your-secure-password' ...
```

Or complete the browser setup wizard at first visit. After that, sign in at
`/login`. OPDS and script clients can use HTTP Basic Auth with the same
credentials, or an API key (see below).

Offline account management without the HTTP server is covered in
[CLI users](./cli-users).

## Sessions and CSRF

Browser logins create a session cookie. `POST /api/auth/refresh` rotates the
access token using the refresh cookie. List or revoke sessions under
Settings -> Profile (or `GET/DELETE /api/auth/sessions`).

Mutating browser API calls require the `X-CSRF-Token` header from
`GET /api/auth/csrf`. Basic Auth and API keys skip CSRF.

When auth is enabled, unauthenticated API calls return JSON for XHR/fetch
clients. Browser navigations to `/api/*` redirect to `/login` with a reason
query (`required`, `session_expired`, `logged_out`). Repeated 403 responses
redirect to `/error/forbidden`. SPA routes are served without a session. The
frontend enforces login.

## Permissions

Non-admin users carry a permission mask. Admins always have every permission.

| Permission | Meaning |
| ---------- | ------- |
| `read` | Browse and open books |
| `edit_metadata` | Edit book metadata and covers |
| `delete_books` | Delete books from the library |
| `manage_library` | Manage mounts, scans, uploads |
| `manage_users` | Manage users (admin-adjacent) |

Default for new local users: `read`, `edit_metadata`, `delete_books`. Set
permissions in Settings -> Administration -> Users or
`PUT /api/auth/users/{id}/permissions`. Guests can receive a subset via the
`permissions` array on create.

Admins can also restrict non-admin users to specific library mounts
(`GET/PUT /api/auth/users/{id}/libraries`).

## Guest accounts

Admins create temporary guests under Settings -> Administration -> Users.
Each guest receives a one-time password and an expiry (default 24 hours).
Expired guests are purged automatically. The panel supports bulk revoke,
extending expiry, copying an invite (login) link, optional permissions, and
listing guests expiring within 72 hours.

## Invites

Admins can send tokenized invites (Settings -> Administration -> Users) for
permanent accounts or guests. Create with `POST /api/invites` (`kind`:
`permanent` or `guest`). Recipients open `/invite/{token}` and complete
acceptance (`GET/POST /api/invite/{token}`).

When SMTP is configured and an email is provided, Athenaeum emails the invite
link. The admin UI always shows a copyable link as well.

Permanent invites may set `provisionPocketId` when the Pocket ID connector is
enabled. That creates the user in Pocket ID, returns a passkey setup URL
(`{pocketId}/lc/{token}`), and accept continues through SSO instead of a local
password.

## Public registration

Enable self-registration under Settings -> Administration -> Auth settings
(`allowRegistration`), or `PUT /api/auth/settings`. New users call
`POST /api/auth/register-public`. Password strength still follows the
configured policy (see [Configuration](./configuration)).

## TOTP (2FA)

Users enroll under Settings -> Profile:

1. `POST /api/auth/totp/setup` returns a secret / otpauth URI
2. Confirm with `POST /api/auth/totp/enable` and a code
3. On login, when the response indicates `needsTotp`, finish with
   `POST /api/auth/totp/verify`

Disable with `POST /api/auth/totp/disable` (password required). Admins can set
`requireTotp` via auth settings (API) even if the UI only exposes registration
today.

## OIDC / SSO

Configure under Settings -> Administration -> SSO:

1. Enable OIDC and set issuer URL, client ID, and client secret
2. Optionally use Discover to fill authorize / token / userinfo URLs
3. Choose how to match existing users (`username`, `email`, or `sub`)
4. Optionally list `adminGroups` so matching IdP groups become Athenaeum admins

Browser flow: `/auth/oidc/login` -> IdP -> `/auth/oidc/callback`. Admin APIs:
`GET/PUT /api/auth/oidc/config`, `POST /api/auth/oidc/discover`.

### Pocket ID

[Pocket ID](https://pocket-id.org/docs/api) is supported as an optional Admin
API connector (Settings -> Administration -> Pocket ID):

1. Set the Pocket ID base URL and an API key (`X-API-KEY`)
2. Use **Apply to OIDC** to fill issuer discovery (`matchBy: email`,
   `autoRegister` on). Paste the OIDC client id/secret from the Pocket ID app
   registration into the SSO section
3. Permanent invites can provision users in Pocket ID and return a
   `/lc/{token}` passkey setup link

Signup-token helpers: `GET/POST/DELETE /api/admin/pocketid/signup-tokens`.

## API keys

Create per-user API keys under Settings -> API. Authenticate with
`Authorization: Bearer ath_<token>` or `X-API-Key: ath_<token>`. The full
secret is shown only once at creation.

Interactive API reference lives in the web UI at Settings -> API and at
`/docs` on a running instance. Machine-readable docs: `GET /api/docs`.

## ALTCHA (optional)

[ALTCHA](https://altcha.org/docs/) is a privacy-first proof-of-work captcha
([widget docs](https://altcha.org/docs/integration/widget/)). It is off by
default.

### Builtin mode

Recommended for most installs. Athenaeum generates challenges at
`GET /api/altcha/challenge` and verifies payloads on login/setup. HMAC secrets
are auto-persisted under `{data}/altcha_hmac_secret` when unset.

```sh
ATHENAEUM_ALTCHA_ENABLED=true
ATHENAEUM_ALTCHA_MODE=builtin
```

Optional widget customization:

```sh
# ATHENAEUM_ALTCHA_WIDGET_THEME=auto
# ATHENAEUM_ALTCHA_WIDGET_DISPLAY=standard
# ATHENAEUM_ALTCHA_WIDGET_TYPE=checkbox
# ATHENAEUM_ALTCHA_WIDGET_AUTO=onsubmit
# ATHENAEUM_ALTCHA_PROTECT=login,setup
```

### Sentinel mode

Run ALTCHA Sentinel (compose profile `altcha`), create an API key in its UI,
then point Athenaeum at it:

```sh
ATHENAEUM_ALTCHA_ENABLED=true
ATHENAEUM_ALTCHA_MODE=sentinel
ATHENAEUM_ALTCHA_CHALLENGE_URL=https://sentinel.example.com/v1/challenge?apiKey=...
ATHENAEUM_ALTCHA_VERIFY_URL=https://sentinel.example.com/v1/verify/signature
ATHENAEUM_ALTCHA_API_KEY_SECRET=...
# optional: ATHENAEUM_ALTCHA_SENTINEL_URL=...
```

Sentinel requires a license after its trial. See
[ALTCHA Sentinel Docker](https://altcha.org/docs/v2/sentinel/install/docker/).
