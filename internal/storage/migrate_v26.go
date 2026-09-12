package storage

import (
	"context"
)

// ttsV26DDL is shared by the SQLite and Postgres migrations. Placeholders are
// dialect-neutral here (both accept ?).
const ttsV26DDL = `
CREATE TABLE IF NOT EXISTS tts_user_prefs (
	user_id       INTEGER PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
	voice         TEXT    NOT NULL DEFAULT '',
	speed         REAL    NOT NULL DEFAULT 1,
	sched_enabled INTEGER NOT NULL DEFAULT 0,
	sched_start   TEXT    NOT NULL DEFAULT '',
	sched_end     TEXT    NOT NULL DEFAULT '',
	updated_at    INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS tts_jobs (
	id             INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id        INTEGER NOT NULL,
	book_id        INTEGER NOT NULL REFERENCES books(id) ON DELETE CASCADE,
	voice          TEXT    NOT NULL DEFAULT '',
	speed          REAL    NOT NULL DEFAULT 1,
	status         TEXT    NOT NULL DEFAULT 'queued',
	total_chapters INTEGER NOT NULL DEFAULT 0,
	done_chapters  INTEGER NOT NULL DEFAULT 0,
	output_dir     TEXT    NOT NULL DEFAULT '',
	error          TEXT    NOT NULL DEFAULT '',
	run_at         INTEGER NOT NULL DEFAULT 0,
	created_at     INTEGER NOT NULL DEFAULT 0,
	started_at     INTEGER,
	finished_at    INTEGER
);
CREATE INDEX IF NOT EXISTS idx_tts_jobs_user   ON tts_jobs(user_id);
CREATE INDEX IF NOT EXISTS idx_tts_jobs_status ON tts_jobs(status, run_at);
CREATE INDEX IF NOT EXISTS idx_tts_jobs_book   ON tts_jobs(book_id);
`

// migrateV26 adds TTS model/format columns and the narration job tables.
func (s *Store) migrateV26(ctx context.Context) error {
	for _, c := range []struct {
		name string
		ddl  string
	}{
		{"model", `ALTER TABLE tts_settings ADD COLUMN model TEXT NOT NULL DEFAULT 'kokoro'`},
		{"response_format", `ALTER TABLE tts_settings ADD COLUMN response_format TEXT NOT NULL DEFAULT 'mp3'`},
	} {
		has, err := s.tableHasColumn(ctx, "tts_settings", c.name)
		if err != nil {
			return err
		}
		if !has {
			if err := s.exec(ctx, c.ddl); err != nil {
				return err
			}
		}
	}
	return s.exec(ctx, ttsV26DDL)
}

// migratePostgresV26 applies the same changes on Postgres dialect.
func (s *Store) migratePostgresV26(ctx context.Context) error {
	for _, c := range []struct {
		name string
		ddl  string
	}{
		{"model", `ALTER TABLE tts_settings ADD COLUMN model TEXT NOT NULL DEFAULT 'kokoro'`},
		{"response_format", `ALTER TABLE tts_settings ADD COLUMN response_format TEXT NOT NULL DEFAULT 'mp3'`},
	} {
		has, err := s.tableHasColumn(ctx, "tts_settings", c.name)
		if err != nil {
			return err
		}
		if !has {
			if err := s.exec(ctx, c.ddl); err != nil {
				return err
			}
		}
	}
	return s.exec(ctx, `
CREATE TABLE IF NOT EXISTS tts_user_prefs (
	user_id       BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
	voice         TEXT    NOT NULL DEFAULT '',
	speed         DOUBLE PRECISION NOT NULL DEFAULT 1,
	sched_enabled INTEGER NOT NULL DEFAULT 0,
	sched_start   TEXT    NOT NULL DEFAULT '',
	sched_end     TEXT    NOT NULL DEFAULT '',
	updated_at    BIGINT  NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS tts_jobs (
	id             BIGSERIAL PRIMARY KEY,
	user_id        BIGINT NOT NULL,
	book_id        BIGINT NOT NULL REFERENCES books(id) ON DELETE CASCADE,
	voice          TEXT   NOT NULL DEFAULT '',
	speed          DOUBLE PRECISION NOT NULL DEFAULT 1,
	status         TEXT   NOT NULL DEFAULT 'queued',
	total_chapters INTEGER NOT NULL DEFAULT 0,
	done_chapters  INTEGER NOT NULL DEFAULT 0,
	output_dir     TEXT   NOT NULL DEFAULT '',
	error          TEXT   NOT NULL DEFAULT '',
	run_at         BIGINT NOT NULL DEFAULT 0,
	created_at     BIGINT NOT NULL DEFAULT 0,
	started_at     BIGINT,
	finished_at    BIGINT
);
CREATE INDEX IF NOT EXISTS idx_tts_jobs_user   ON tts_jobs(user_id);
CREATE INDEX IF NOT EXISTS idx_tts_jobs_status ON tts_jobs(status, run_at);
CREATE INDEX IF NOT EXISTS idx_tts_jobs_book   ON tts_jobs(book_id);
`)
}
