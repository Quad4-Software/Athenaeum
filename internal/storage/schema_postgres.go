package storage

// schemaPostgres is the full Athenaeum schema for a fresh PostgreSQL database
// (equivalent to SQLite user_version 24 plus tsvector FTS).
const schemaPostgres = `
CREATE TABLE IF NOT EXISTS schema_version (
	version INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS libraries (
	id             BIGSERIAL PRIMARY KEY,
	name           TEXT    NOT NULL,
	mount_path     TEXT    NOT NULL UNIQUE,
	sort_order     INTEGER NOT NULL DEFAULT 0,
	created_at     BIGINT  NOT NULL,
	backend        TEXT    NOT NULL DEFAULT 'local',
	backend_config TEXT    NOT NULL DEFAULT '{}'
);

CREATE TABLE IF NOT EXISTS books (
	id               BIGSERIAL PRIMARY KEY,
	library_id       BIGINT  NOT NULL DEFAULT 1 REFERENCES libraries(id) ON DELETE CASCADE,
	title            TEXT    NOT NULL,
	author           TEXT    NOT NULL DEFAULT '',
	series           TEXT    NOT NULL DEFAULT '',
	series_index     DOUBLE PRECISION NOT NULL DEFAULT 0,
	format           TEXT    NOT NULL,
	rel_path         TEXT    NOT NULL,
	abs_path         TEXT    NOT NULL,
	file_size        BIGINT  NOT NULL DEFAULT 0,
	has_cover        INTEGER NOT NULL DEFAULT 0,
	language         TEXT    NOT NULL DEFAULT '',
	description      TEXT    NOT NULL DEFAULT '',
	mtime            BIGINT  NOT NULL DEFAULT 0,
	added_at         BIGINT  NOT NULL,
	modified_at      BIGINT  NOT NULL,
	meta_edited      INTEGER NOT NULL DEFAULT 0,
	cover_edited     INTEGER NOT NULL DEFAULT 0,
	content_hash     TEXT    NOT NULL DEFAULT '',
	duplicate_of     BIGINT  NOT NULL DEFAULT 0,
	hidden           INTEGER NOT NULL DEFAULT 0,
	audiobook_set_id BIGINT  NOT NULL DEFAULT 0,
	search_tsv       tsvector GENERATED ALWAYS AS (
		setweight(to_tsvector('simple', coalesce(title, '')), 'A') ||
		setweight(to_tsvector('simple', coalesce(author, '')), 'B') ||
		setweight(to_tsvector('simple', coalesce(series, '')), 'C') ||
		setweight(to_tsvector('simple', coalesce(description, '')), 'D')
	) STORED,
	UNIQUE (library_id, rel_path)
);
CREATE INDEX IF NOT EXISTS idx_books_library ON books(library_id);
CREATE INDEX IF NOT EXISTS idx_books_title ON books(title);
CREATE INDEX IF NOT EXISTS idx_books_author ON books(author);
CREATE INDEX IF NOT EXISTS idx_books_format ON books(format);
CREATE INDEX IF NOT EXISTS idx_books_added ON books(added_at);
CREATE INDEX IF NOT EXISTS idx_books_series ON books(series);
CREATE INDEX IF NOT EXISTS idx_books_content_hash ON books(content_hash);
CREATE INDEX IF NOT EXISTS idx_books_visible_library_added ON books(library_id, added_at DESC) WHERE hidden = 0;
CREATE INDEX IF NOT EXISTS idx_books_search_tsv ON books USING GIN (search_tsv);

CREATE TABLE IF NOT EXISTS users (
	id            BIGSERIAL PRIMARY KEY,
	username      TEXT    NOT NULL,
	password_hash TEXT    NOT NULL,
	is_admin      INTEGER NOT NULL DEFAULT 0,
	created_at    BIGINT  NOT NULL,
	email         TEXT    NOT NULL DEFAULT '',
	oidc_sub      TEXT    NOT NULL DEFAULT '',
	permissions   INTEGER NOT NULL DEFAULT 0,
	is_guest      INTEGER NOT NULL DEFAULT 0,
	expires_at    BIGINT  NOT NULL DEFAULT 0,
	totp_secret   TEXT,
	totp_enabled  INTEGER NOT NULL DEFAULT 0
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username_lower ON users (LOWER(username));
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_oidc_sub ON users(oidc_sub) WHERE oidc_sub <> '';
CREATE INDEX IF NOT EXISTS idx_users_expires ON users(expires_at) WHERE expires_at > 0;

CREATE TABLE IF NOT EXISTS sessions (
	token        TEXT PRIMARY KEY,
	user_id      BIGINT  NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	expires_at   BIGINT  NOT NULL,
	session_id   TEXT    NOT NULL DEFAULT '',
	ip           TEXT    NOT NULL DEFAULT '',
	user_agent   TEXT    NOT NULL DEFAULT '',
	device       TEXT    NOT NULL DEFAULT '',
	auth_method  TEXT    NOT NULL DEFAULT 'local',
	created_at   BIGINT  NOT NULL DEFAULT 0,
	last_seen_at BIGINT  NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_sessions_user ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires ON sessions(expires_at);
CREATE INDEX IF NOT EXISTS idx_sessions_session_id ON sessions(session_id);

CREATE TABLE IF NOT EXISTS refresh_tokens (
	token      TEXT PRIMARY KEY,
	user_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	expires_at BIGINT NOT NULL,
	session_id TEXT   NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user ON refresh_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires ON refresh_tokens(expires_at);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_session_id ON refresh_tokens(session_id);

CREATE TABLE IF NOT EXISTS collections (
	id          BIGSERIAL PRIMARY KEY,
	user_id     BIGINT  NOT NULL DEFAULT 0,
	name        TEXT    NOT NULL,
	description TEXT    NOT NULL DEFAULT '',
	created_at  BIGINT  NOT NULL,
	kind        TEXT    NOT NULL DEFAULT 'manual',
	query_json  TEXT    NOT NULL DEFAULT '',
	UNIQUE (user_id, name)
);
CREATE INDEX IF NOT EXISTS idx_collections_user ON collections(user_id);

CREATE TABLE IF NOT EXISTS collection_items (
	collection_id BIGINT NOT NULL REFERENCES collections(id) ON DELETE CASCADE,
	book_id       BIGINT NOT NULL REFERENCES books(id) ON DELETE CASCADE,
	sort_order    INTEGER NOT NULL DEFAULT 0,
	PRIMARY KEY (collection_id, book_id)
);
CREATE INDEX IF NOT EXISTS idx_collection_items_book ON collection_items(book_id);
CREATE INDEX IF NOT EXISTS idx_collection_items_collection ON collection_items(collection_id);

CREATE TABLE IF NOT EXISTS progress (
	user_id      BIGINT NOT NULL DEFAULT 0,
	book_id      BIGINT NOT NULL REFERENCES books(id) ON DELETE CASCADE,
	location     TEXT   NOT NULL DEFAULT '',
	percent      DOUBLE PRECISION NOT NULL DEFAULT 0,
	updated_at   BIGINT NOT NULL,
	read_seconds BIGINT NOT NULL DEFAULT 0,
	PRIMARY KEY (user_id, book_id)
);
CREATE INDEX IF NOT EXISTS idx_progress_user_updated ON progress(user_id, updated_at DESC);

CREATE TABLE IF NOT EXISTS book_chapters (
	book_id  BIGINT  NOT NULL REFERENCES books(id) ON DELETE CASCADE,
	idx      INTEGER NOT NULL,
	title    TEXT    NOT NULL DEFAULT '',
	start_ms BIGINT  NOT NULL DEFAULT 0,
	PRIMARY KEY (book_id, idx)
);
CREATE INDEX IF NOT EXISTS idx_book_chapters_book ON book_chapters(book_id);

CREATE TABLE IF NOT EXISTS user_favorites (
	user_id    BIGINT NOT NULL,
	book_id    BIGINT NOT NULL REFERENCES books(id) ON DELETE CASCADE,
	created_at BIGINT NOT NULL,
	PRIMARY KEY (user_id, book_id)
);
CREATE INDEX IF NOT EXISTS idx_user_favorites_user ON user_favorites(user_id);

CREATE TABLE IF NOT EXISTS audit_log (
	id             BIGSERIAL PRIMARY KEY,
	actor_id       BIGINT  NOT NULL,
	actor_name     TEXT    NOT NULL DEFAULT '',
	target_user_id BIGINT  NOT NULL DEFAULT 0,
	target_name    TEXT    NOT NULL DEFAULT '',
	action         TEXT    NOT NULL,
	details        TEXT    NOT NULL DEFAULT '',
	ip             TEXT    NOT NULL DEFAULT '',
	created_at     BIGINT  NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_audit_log_created ON audit_log(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_log_actor ON audit_log(actor_id);

CREATE TABLE IF NOT EXISTS user_libraries (
	user_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	library_id BIGINT NOT NULL REFERENCES libraries(id) ON DELETE CASCADE,
	PRIMARY KEY (user_id, library_id)
);
CREATE INDEX IF NOT EXISTS idx_user_libraries_user ON user_libraries(user_id);

CREATE TABLE IF NOT EXISTS upload_sessions (
	id         TEXT PRIMARY KEY,
	library_id BIGINT  NOT NULL REFERENCES libraries(id) ON DELETE CASCADE,
	user_id    BIGINT  NOT NULL,
	rel_path   TEXT    NOT NULL,
	total_size BIGINT  NOT NULL DEFAULT 0,
	"offset"   BIGINT  NOT NULL DEFAULT 0,
	done       INTEGER NOT NULL DEFAULT 0,
	book_id    BIGINT  NOT NULL DEFAULT 0,
	created_at BIGINT  NOT NULL,
	updated_at BIGINT  NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_upload_sessions_user ON upload_sessions(user_id);

CREATE TABLE IF NOT EXISTS api_keys (
	id           BIGSERIAL PRIMARY KEY,
	user_id      BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	name         TEXT   NOT NULL,
	prefix       TEXT   NOT NULL,
	key_hash     TEXT   NOT NULL,
	created_at   BIGINT NOT NULL,
	last_used_at BIGINT NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_api_keys_prefix ON api_keys(prefix);
CREATE INDEX IF NOT EXISTS idx_api_keys_user ON api_keys(user_id);

CREATE TABLE IF NOT EXISTS oidc_config (
	id                INTEGER PRIMARY KEY CHECK (id = 1),
	enabled           INTEGER NOT NULL DEFAULT 0,
	login_local       INTEGER NOT NULL DEFAULT 1,
	issuer_url        TEXT    NOT NULL DEFAULT '',
	authorize_url     TEXT    NOT NULL DEFAULT '',
	token_url         TEXT    NOT NULL DEFAULT '',
	userinfo_url      TEXT    NOT NULL DEFAULT '',
	jwks_url          TEXT    NOT NULL DEFAULT '',
	logout_url        TEXT    NOT NULL DEFAULT '',
	client_id         TEXT    NOT NULL DEFAULT '',
	client_secret     TEXT    NOT NULL DEFAULT '',
	signing_algorithm TEXT    NOT NULL DEFAULT 'RS256',
	button_text       TEXT    NOT NULL DEFAULT 'Sign in with SSO',
	match_by          TEXT    NOT NULL DEFAULT 'username',
	auto_register     INTEGER NOT NULL DEFAULT 0,
	auto_launch       INTEGER NOT NULL DEFAULT 0,
	admin_groups      TEXT    NOT NULL DEFAULT '',
	group_claim       TEXT    NOT NULL DEFAULT 'groups'
);

CREATE TABLE IF NOT EXISTS server_config (
	id                     INTEGER PRIMARY KEY CHECK (id = 1),
	metrics_enabled        INTEGER NOT NULL DEFAULT 0,
	metrics_auth           INTEGER NOT NULL DEFAULT 1,
	metrics_username       TEXT    NOT NULL DEFAULT '',
	metrics_password_hash  TEXT    NOT NULL DEFAULT '',
	trusted_proxies        TEXT    NOT NULL DEFAULT '',
	cors_enabled           INTEGER NOT NULL DEFAULT 0,
	cors_origins           TEXT    NOT NULL DEFAULT '',
	csp_enabled            INTEGER NOT NULL DEFAULT 1,
	csp_policy             TEXT    NOT NULL DEFAULT '',
	auto_scan_enabled      INTEGER NOT NULL DEFAULT 0,
	auto_scan_interval_sec INTEGER NOT NULL DEFAULT 300,
	scan_workers           INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS bookmarks (
	id         BIGSERIAL PRIMARY KEY,
	user_id    BIGINT NOT NULL,
	book_id    BIGINT NOT NULL REFERENCES books(id) ON DELETE CASCADE,
	location   TEXT   NOT NULL,
	label      TEXT   NOT NULL DEFAULT '',
	created_at BIGINT NOT NULL,
	UNIQUE (user_id, book_id, location)
);
CREATE INDEX IF NOT EXISTS idx_bookmarks_user_book ON bookmarks(user_id, book_id);

CREATE TABLE IF NOT EXISTS highlights (
	id         BIGSERIAL PRIMARY KEY,
	user_id    BIGINT NOT NULL,
	book_id    BIGINT NOT NULL REFERENCES books(id) ON DELETE CASCADE,
	location   TEXT   NOT NULL,
	excerpt    TEXT   NOT NULL DEFAULT '',
	note       TEXT   NOT NULL DEFAULT '',
	color      TEXT   NOT NULL DEFAULT 'yellow',
	created_at BIGINT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_highlights_user_book ON highlights(user_id, book_id);

CREATE TABLE IF NOT EXISTS audiobook_tracks (
	set_book_id BIGINT  NOT NULL REFERENCES books(id) ON DELETE CASCADE,
	track_index INTEGER NOT NULL,
	rel_path    TEXT    NOT NULL,
	title       TEXT    NOT NULL DEFAULT '',
	format      TEXT    NOT NULL DEFAULT '',
	file_size   BIGINT  NOT NULL DEFAULT 0,
	PRIMARY KEY (set_book_id, track_index)
);
CREATE INDEX IF NOT EXISTS idx_audiobook_tracks_set ON audiobook_tracks(set_book_id);

CREATE TABLE IF NOT EXISTS tags (
	id   BIGSERIAL PRIMARY KEY,
	name TEXT NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_tags_name_lower ON tags (LOWER(name));

CREATE TABLE IF NOT EXISTS book_tags (
	book_id BIGINT NOT NULL REFERENCES books(id) ON DELETE CASCADE,
	tag_id  BIGINT NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
	PRIMARY KEY (book_id, tag_id)
);
CREATE INDEX IF NOT EXISTS idx_book_tags_tag ON book_tags(tag_id);

CREATE TABLE IF NOT EXISTS book_ratings (
	user_id    BIGINT  NOT NULL,
	book_id    BIGINT  NOT NULL REFERENCES books(id) ON DELETE CASCADE,
	rating     INTEGER NOT NULL CHECK (rating BETWEEN 1 AND 5),
	updated_at BIGINT  NOT NULL,
	PRIMARY KEY (user_id, book_id)
);

CREATE TABLE IF NOT EXISTS reader_prefs (
	user_id    BIGINT NOT NULL PRIMARY KEY,
	prefs_json TEXT   NOT NULL DEFAULT '{}',
	updated_at BIGINT NOT NULL
);

CREATE TABLE IF NOT EXISTS auth_settings (
	id                 INTEGER PRIMARY KEY CHECK (id = 1),
	allow_registration INTEGER NOT NULL DEFAULT 0,
	require_totp       INTEGER NOT NULL DEFAULT 0,
	updated_at         BIGINT  NOT NULL
);

CREATE TABLE IF NOT EXISTS share_links (
	id             BIGSERIAL PRIMARY KEY,
	token          TEXT   NOT NULL UNIQUE,
	book_id        BIGINT NOT NULL REFERENCES books(id) ON DELETE CASCADE,
	created_by     BIGINT NOT NULL,
	expires_at     BIGINT,
	created_at     BIGINT NOT NULL,
	download_count INTEGER NOT NULL DEFAULT 0,
	max_downloads  INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_share_links_book ON share_links(book_id);

CREATE TABLE IF NOT EXISTS smtp_settings (
	id         INTEGER PRIMARY KEY CHECK (id = 1),
	enabled    INTEGER NOT NULL DEFAULT 0,
	host       TEXT    NOT NULL DEFAULT '',
	port       INTEGER NOT NULL DEFAULT 587,
	username   TEXT    NOT NULL DEFAULT '',
	password   TEXT    NOT NULL DEFAULT '',
	from_addr  TEXT    NOT NULL DEFAULT '',
	use_tls    INTEGER NOT NULL DEFAULT 1,
	updated_at BIGINT  NOT NULL
);

CREATE TABLE IF NOT EXISTS user_kindle_email (
	user_id    BIGINT NOT NULL PRIMARY KEY,
	email      TEXT   NOT NULL DEFAULT '',
	updated_at BIGINT NOT NULL
);

CREATE TABLE IF NOT EXISTS book_content (
	book_id     BIGINT  NOT NULL,
	chunk_index INTEGER NOT NULL,
	content     TEXT    NOT NULL,
	content_tsv tsvector GENERATED ALWAYS AS (
		to_tsvector('simple', coalesce(content, ''))
	) STORED,
	PRIMARY KEY (book_id, chunk_index)
);
CREATE INDEX IF NOT EXISTS idx_book_content_tsv ON book_content USING GIN (content_tsv);

CREATE TABLE IF NOT EXISTS kosync_documents (
	user_id    BIGINT NOT NULL,
	document   TEXT   NOT NULL,
	progress   TEXT   NOT NULL DEFAULT '',
	percentage DOUBLE PRECISION NOT NULL DEFAULT 0,
	device     TEXT   NOT NULL DEFAULT '',
	device_id  TEXT   NOT NULL DEFAULT '',
	timestamp  BIGINT NOT NULL DEFAULT 0,
	PRIMARY KEY (user_id, document)
);

CREATE TABLE IF NOT EXISTS offline_grants (
	user_id    BIGINT NOT NULL,
	book_id    BIGINT NOT NULL REFERENCES books(id) ON DELETE CASCADE,
	created_at BIGINT NOT NULL,
	PRIMARY KEY (user_id, book_id)
);

CREATE TABLE IF NOT EXISTS tts_settings (
	id            INTEGER PRIMARY KEY CHECK (id = 1),
	enabled       INTEGER NOT NULL DEFAULT 0,
	base_url      TEXT    NOT NULL DEFAULT '',
	api_key       TEXT    NOT NULL DEFAULT '',
	default_voice TEXT    NOT NULL DEFAULT 'af_heart',
	timeout_sec   INTEGER NOT NULL DEFAULT 60,
	updated_at    BIGINT  NOT NULL
);

CREATE TABLE IF NOT EXISTS invites (
	id                BIGSERIAL PRIMARY KEY,
	token             TEXT   NOT NULL UNIQUE,
	kind              TEXT   NOT NULL,
	email             TEXT   NOT NULL DEFAULT '',
	permissions       BIGINT NOT NULL DEFAULT 0,
	created_by        BIGINT NOT NULL,
	expires_at        BIGINT,
	guest_expires_at  BIGINT,
	pocket_id_user_id TEXT   NOT NULL DEFAULT '',
	accepted_at       BIGINT,
	accepted_user_id  BIGINT,
	revoked_at        BIGINT,
	created_at        BIGINT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_invites_token ON invites(token);

CREATE TABLE IF NOT EXISTS webhooks (
	id         BIGSERIAL PRIMARY KEY,
	url        TEXT   NOT NULL,
	secret     TEXT   NOT NULL DEFAULT '',
	events     TEXT   NOT NULL DEFAULT '[]',
	enabled    INTEGER NOT NULL DEFAULT 1,
	created_at BIGINT NOT NULL,
	updated_at BIGINT NOT NULL
);

CREATE TABLE IF NOT EXISTS webhook_deliveries (
	id           BIGSERIAL PRIMARY KEY,
	webhook_id   BIGINT NOT NULL,
	event        TEXT   NOT NULL,
	payload      TEXT   NOT NULL,
	status_code  INTEGER NOT NULL DEFAULT 0,
	success      INTEGER NOT NULL DEFAULT 0,
	attempts     INTEGER NOT NULL DEFAULT 0,
	last_error   TEXT    NOT NULL DEFAULT '',
	created_at   BIGINT  NOT NULL,
	delivered_at BIGINT
);
CREATE INDEX IF NOT EXISTS idx_webhook_deliveries_webhook ON webhook_deliveries(webhook_id, created_at DESC);

CREATE TABLE IF NOT EXISTS pocketid_settings (
	id                INTEGER PRIMARY KEY CHECK (id = 1),
	enabled           INTEGER NOT NULL DEFAULT 0,
	base_url          TEXT    NOT NULL DEFAULT '',
	api_key           TEXT    NOT NULL DEFAULT '',
	default_group_ids TEXT    NOT NULL DEFAULT '[]',
	updated_at        BIGINT  NOT NULL
);
`

const schemaPostgresSeeds = `
INSERT INTO libraries (id, name, mount_path, sort_order, created_at)
VALUES (1, 'Main Library', '', 0, EXTRACT(EPOCH FROM NOW())::BIGINT)
ON CONFLICT (id) DO NOTHING;
SELECT setval(pg_get_serial_sequence('libraries', 'id'), GREATEST((SELECT MAX(id) FROM libraries), 1));

INSERT INTO oidc_config (id) VALUES (1) ON CONFLICT (id) DO NOTHING;
INSERT INTO server_config (id) VALUES (1) ON CONFLICT (id) DO NOTHING;
INSERT INTO auth_settings (id, allow_registration, require_totp, updated_at) VALUES (1, 0, 0, 0) ON CONFLICT (id) DO NOTHING;
INSERT INTO smtp_settings (id, enabled, host, port, username, password, from_addr, use_tls, updated_at)
VALUES (1, 0, '', 587, '', '', '', 1, 0) ON CONFLICT (id) DO NOTHING;
INSERT INTO tts_settings (id, enabled, base_url, api_key, default_voice, timeout_sec, updated_at)
VALUES (1, 0, '', '', 'af_heart', 60, 0) ON CONFLICT (id) DO NOTHING;
INSERT INTO pocketid_settings (id, enabled, base_url, api_key, default_group_ids, updated_at)
VALUES (1, 0, '', '', '[]', 0) ON CONFLICT (id) DO NOTHING;
`
