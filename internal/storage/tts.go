package storage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"athenaeum/internal/models"
)

// GetTTSSettings returns the singleton Kokoro TTS sidecar configuration.
func (s *Store) GetTTSSettings(ctx context.Context) (models.TTSSettings, error) {
	var c models.TTSSettings
	var enabled int
	err := s.queryRowContext(ctx, `
SELECT enabled, base_url, api_key, model, default_voice, response_format, timeout_sec
FROM tts_settings WHERE id=1`).
		Scan(&enabled, &c.BaseURL, &c.APIKey, &c.Model, &c.DefaultVoice, &c.ResponseFormat, &c.TimeoutSec)
	if err != nil {
		return c, err
	}
	c.Enabled = enabled != 0
	applyTTSDefaults(&c)
	return c, nil
}

// SaveTTSSettings updates the singleton Kokoro TTS sidecar configuration.
func (s *Store) SaveTTSSettings(ctx context.Context, c models.TTSSettings) error {
	applyTTSDefaults(&c)
	_, err := s.execContext(ctx, `
UPDATE tts_settings SET enabled=?, base_url=?, api_key=?, model=?, default_voice=?,
	response_format=?, timeout_sec=?, updated_at=?
WHERE id=1`,
		boolToInt(c.Enabled), c.BaseURL, c.APIKey, c.Model, c.DefaultVoice,
		c.ResponseFormat, c.TimeoutSec, time.Now().Unix())
	return err
}

func applyTTSDefaults(c *models.TTSSettings) {
	if c.TimeoutSec <= 0 {
		c.TimeoutSec = 60
	}
	if c.DefaultVoice == "" {
		c.DefaultVoice = "af_heart"
	}
	if c.Model == "" {
		c.Model = "kokoro"
	}
	if c.ResponseFormat == "" {
		c.ResponseFormat = "mp3"
	}
}

// GetTTSUserPrefs returns the user's narration preferences, or defaults.
func (s *Store) GetTTSUserPrefs(ctx context.Context, userID int64) (models.TTSUserPrefs, error) {
	p := models.TTSUserPrefs{UserID: userID, Speed: 1}
	var sched int
	err := s.queryRowContext(ctx, `
SELECT voice, speed, sched_enabled, sched_start, sched_end
FROM tts_user_prefs WHERE user_id=?`, userID).
		Scan(&p.Voice, &p.Speed, &sched, &p.SchedStart, &p.SchedEnd)
	if errors.Is(err, sql.ErrNoRows) {
		return p, nil
	}
	if err != nil {
		return p, err
	}
	p.SchedEnabled = sched != 0
	if p.Speed <= 0 {
		p.Speed = 1
	}
	return p, nil
}

// SaveTTSUserPrefs upserts the user's narration preferences.
func (s *Store) SaveTTSUserPrefs(ctx context.Context, p models.TTSUserPrefs) error {
	if p.Speed <= 0 {
		p.Speed = 1
	}
	_, err := s.execContext(ctx, `
INSERT INTO tts_user_prefs (user_id, voice, speed, sched_enabled, sched_start, sched_end, updated_at)
VALUES (?,?,?,?,?,?,?)
ON CONFLICT(user_id) DO UPDATE SET
	voice=excluded.voice,
	speed=excluded.speed,
	sched_enabled=excluded.sched_enabled,
	sched_start=excluded.sched_start,
	sched_end=excluded.sched_end,
	updated_at=excluded.updated_at`,
		p.UserID, p.Voice, p.Speed, boolToInt(p.SchedEnabled), p.SchedStart, p.SchedEnd,
		time.Now().Unix())
	return err
}

const ttsJobColumns = `
SELECT j.id, j.user_id, j.book_id, b.title, j.voice, j.speed, j.status,
	j.total_chapters, j.done_chapters, j.output_dir, j.error, j.run_at,
	j.created_at, j.started_at, j.finished_at
FROM tts_jobs j LEFT JOIN books b ON b.id = j.book_id`

func scanTTSJob(row interface {
	Scan(dest ...any) error
}) (models.TTSJob, error) {
	var j models.TTSJob
	var title sql.NullString
	var created int64
	var started, finished sql.NullInt64
	err := row.Scan(&j.ID, &j.UserID, &j.BookID, &title, &j.Voice, &j.Speed, &j.Status,
		&j.TotalChapters, &j.DoneChapters, &j.OutputDir, &j.Error, &j.RunAt,
		&created, &started, &finished)
	if err != nil {
		return j, err
	}
	j.BookTitle = title.String
	j.CreatedAt = time.Unix(created, 0)
	if started.Valid && started.Int64 > 0 {
		t := time.Unix(started.Int64, 0)
		j.StartedAt = &t
	}
	if finished.Valid && finished.Int64 > 0 {
		t := time.Unix(finished.Int64, 0)
		j.FinishedAt = &t
	}
	return j, nil
}

// CreateTTSJob queues a narration job.
func (s *Store) CreateTTSJob(ctx context.Context, j models.TTSJob) (int64, error) {
	if j.Speed <= 0 {
		j.Speed = 1
	}
	return s.insertID(ctx, `
INSERT INTO tts_jobs (user_id, book_id, voice, speed, status, run_at, created_at)
VALUES (?,?,?,?,?,?,?)
RETURNING id`,
		j.UserID, j.BookID, j.Voice, j.Speed, models.TTSJobQueued, j.RunAt, time.Now().Unix())
}

// GetTTSJob returns one job by id.
func (s *Store) GetTTSJob(ctx context.Context, id int64) (models.TTSJob, error) {
	j, err := scanTTSJob(s.queryRowContext(ctx, ttsJobColumns+` WHERE j.id=?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return j, ErrNotFound
	}
	return j, err
}

// ListTTSJobs returns a user's jobs, or every job when all is true.
func (s *Store) ListTTSJobs(ctx context.Context, userID int64, all bool) ([]models.TTSJob, error) {
	q := ttsJobColumns
	args := []any{}
	if !all {
		q += ` WHERE j.user_id=?`
		args = append(args, userID)
	}
	q += ` ORDER BY j.created_at DESC, j.id DESC LIMIT 200`
	rows, err := s.queryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.TTSJob
	for rows.Next() {
		j, err := scanTTSJob(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, j)
	}
	return out, rows.Err()
}

// CountActiveTTSJobsForBook returns queued or running jobs for a book.
func (s *Store) CountActiveTTSJobsForBook(ctx context.Context, bookID int64) (int, error) {
	var n int
	err := s.queryRowContext(ctx, `
SELECT COUNT(*) FROM tts_jobs WHERE book_id=? AND status IN ('queued','running')`, bookID).Scan(&n)
	return n, err
}

// CountPendingTTSJobs returns a user's queued or running jobs.
func (s *Store) CountPendingTTSJobs(ctx context.Context, userID int64) (int, error) {
	var n int
	err := s.queryRowContext(ctx, `
SELECT COUNT(*) FROM tts_jobs WHERE user_id=? AND status IN ('queued','running')`, userID).Scan(&n)
	return n, err
}

// ClaimNextTTSJob marks the oldest due queued job as running.
func (s *Store) ClaimNextTTSJob(ctx context.Context, now int64) (models.TTSJob, bool, error) {
	var id int64
	err := s.queryRowContext(ctx, `
UPDATE tts_jobs SET status='running', started_at=?, error=''
WHERE id = (
	SELECT id FROM tts_jobs WHERE status='queued' AND run_at<=? ORDER BY created_at, id LIMIT 1
)
RETURNING id`, now, now).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return models.TTSJob{}, false, nil
	}
	if err != nil {
		return models.TTSJob{}, false, err
	}
	j, err := s.GetTTSJob(ctx, id)
	if err != nil {
		return models.TTSJob{}, false, err
	}
	return j, true, nil
}

// RequeueTTSJob parks a running job until the next allowed run time.
func (s *Store) RequeueTTSJob(ctx context.Context, id int64, runAt int64) error {
	_, err := s.execContext(ctx, `
UPDATE tts_jobs SET status='queued', run_at=? WHERE id=? AND status='running'`, runAt, id)
	return err
}

// RetryTTSJob requeues a finished, failed, or cancelled job. Staged
// chapter files on disk let it resume where it stopped.
func (s *Store) RetryTTSJob(ctx context.Context, id int64, runAt int64) error {
	_, err := s.execContext(ctx, `
UPDATE tts_jobs SET status='queued', run_at=?, error='', started_at=NULL, finished_at=NULL
WHERE id=? AND status NOT IN ('queued','running')`, runAt, id)
	return err
}

// UpdateTTSJobProgress stores chapter totals and progress.
func (s *Store) UpdateTTSJobProgress(ctx context.Context, id int64, total, done int) error {
	_, err := s.execContext(ctx, `
UPDATE tts_jobs SET total_chapters=?, done_chapters=? WHERE id=?`, total, done, id)
	return err
}

// SetTTSJobOutput records the library-relative output directory.
func (s *Store) SetTTSJobOutput(ctx context.Context, id int64, dir string) error {
	_, err := s.execContext(ctx, `UPDATE tts_jobs SET output_dir=? WHERE id=?`, dir, id)
	return err
}

// FinishTTSJob marks a job done, failed, or cancelled.
func (s *Store) FinishTTSJob(ctx context.Context, id int64, status, errMsg string) error {
	_, err := s.execContext(ctx, `
UPDATE tts_jobs SET status=?, error=?, finished_at=? WHERE id=?`,
		status, errMsg, time.Now().Unix(), id)
	return err
}

// TTSJobIsActive reports whether a job is still queued or running.
func (s *Store) TTSJobIsActive(ctx context.Context, id int64) (bool, error) {
	var status string
	err := s.queryRowContext(ctx, `SELECT status FROM tts_jobs WHERE id=?`, id).Scan(&status)
	if errors.Is(err, sql.ErrNoRows) {
		return false, ErrNotFound
	}
	if err != nil {
		return false, err
	}
	return status == models.TTSJobQueued || status == models.TTSJobRunning, nil
}

// ResetRunningTTSJobs requeues jobs orphaned by a restart.
func (s *Store) ResetRunningTTSJobs(ctx context.Context) error {
	_, err := s.execContext(ctx, `
UPDATE tts_jobs SET status='queued', run_at=0 WHERE status='running'`)
	return err
}

// DeleteTTSJob removes a finished, failed, or cancelled job row.
func (s *Store) DeleteTTSJob(ctx context.Context, id int64) error {
	res, err := s.execContext(ctx, `
DELETE FROM tts_jobs WHERE id=? AND status NOT IN ('queued','running')`, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
