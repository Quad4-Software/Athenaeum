package tts

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode"

	"athenaeum/internal/config"
	"athenaeum/internal/libfs"
	"athenaeum/internal/library"
	"athenaeum/internal/models"
	"athenaeum/internal/storage"
)

const (
	// PollInterval is how often the worker checks for due jobs.
	PollInterval = 20 * time.Second
	// chunkRetryPause waits between a failed chunk and its single retry.
	chunkRetryPause = 3 * time.Second
	// outputRoot is the library-relative directory for generated books.
	outputRoot = "audiobooks"
)

var spaceRuns = regexp.MustCompile(`\s+`)

// nameSafeRune allows letters and digits from any script plus a few
// punctuation marks that are safe in file names and S3 keys.
func nameSafeRune(r rune) bool {
	if unicode.IsLetter(r) || unicode.IsDigit(r) {
		return true
	}
	switch r {
	case ' ', '.', '_', '-', '\'', '(', ')', '[', ']':
		return true
	}
	return false
}

// Worker runs queued whole-book narration jobs one at a time.
type Worker struct {
	store   *storage.Store
	scanner *library.Scanner
	cfg     config.Config
	log     *slog.Logger
	kick    chan struct{}
}

// NewWorker builds a job worker. Start must be called to run it.
func NewWorker(store *storage.Store, scanner *library.Scanner, cfg config.Config, log *slog.Logger) *Worker {
	return &Worker{store: store, scanner: scanner, cfg: cfg, log: log, kick: make(chan struct{}, 1)}
}

// Kick nudges the loop to check for work immediately.
func (w *Worker) Kick() {
	select {
	case w.kick <- struct{}{}:
	default:
	}
}

// DropStaging removes a job's staged chapter files.
func (w *Worker) DropStaging(jobID int64) {
	_ = os.RemoveAll(w.stagingDir(jobID))
}

// Start requeues jobs orphaned by a restart and loops until ctx ends.
func (w *Worker) Start(ctx context.Context) {
	if err := w.store.ResetRunningTTSJobs(ctx); err != nil {
		w.log.Warn("tts job reset failed", "err", err)
	}
	w.cleanOrphanStaging(ctx)
	tick := time.NewTicker(PollInterval)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
		case <-w.kick:
		}
		for {
			job, ok, err := w.store.ClaimNextTTSJob(ctx, time.Now().Unix())
			if err != nil {
				w.log.Warn("tts claim failed", "err", err)
				break
			}
			if !ok {
				break
			}
			w.runJob(ctx, job)
			if ctx.Err() != nil {
				return
			}
		}
	}
}

// stagingDir holds per-chapter audio while a job runs so the library only
// ever sees a complete book.
func (w *Worker) stagingDir(jobID int64) string {
	return filepath.Join(w.cfg.DataDir, "tts-jobs", fmt.Sprint(jobID))
}

func (w *Worker) cleanOrphanStaging(ctx context.Context) {
	root := filepath.Join(w.cfg.DataDir, "tts-jobs")
	entries, err := os.ReadDir(root)
	if err != nil {
		return
	}
	jobs, err := w.store.ListTTSJobs(ctx, 0, true)
	if err != nil {
		return
	}
	active := make(map[string]bool, len(jobs))
	for _, j := range jobs {
		if j.Status == models.TTSJobQueued || j.Status == models.TTSJobRunning {
			active[fmt.Sprint(j.ID)] = true
		}
	}
	for _, e := range entries {
		if e.IsDir() && !active[e.Name()] {
			_ = os.RemoveAll(filepath.Join(root, e.Name()))
		}
	}
}

func (w *Worker) runJob(ctx context.Context, job models.TTSJob) {
	err := w.process(ctx, job)
	switch {
	case err == nil:
		// completed inside process
	case errors.Is(err, errJobAborted):
		// cancelled or deleted while running; leave status alone
	case errors.Is(err, context.Canceled):
		// Server is shutting down; park the job so it resumes on boot.
		bg := context.WithoutCancel(ctx)
		_ = w.store.RequeueTTSJob(bg, job.ID, 0)
	default:
		w.log.Warn("tts job failed", "job", job.ID, "err", err)
		_ = w.store.FinishTTSJob(ctx, job.ID, models.TTSJobFailed, err.Error())
	}
}

var errJobAborted = errors.New("job aborted")

func (w *Worker) process(ctx context.Context, job models.TTSJob) error {
	prefs, err := w.store.GetTTSUserPrefs(ctx, job.UserID)
	if err != nil {
		return err
	}
	now := time.Now()
	if !InWindow(now, prefs) {
		return w.store.RequeueTTSJob(ctx, job.ID, NextWindowStart(now, prefs).Unix())
	}

	book, err := w.store.GetBook(ctx, job.BookID)
	if err != nil {
		return fmt.Errorf("book lookup: %w", err)
	}
	if book.Format != models.FormatEPUB {
		return errors.New("only EPUB sources are supported")
	}

	fs, err := w.store.OpenLibraryFS(ctx, book.LibraryID)
	if err != nil {
		return err
	}
	srcPath, cleanup, err := w.materialize(ctx, fs, book.RelPath)
	if err != nil {
		return fmt.Errorf("open book file: %w", err)
	}
	defer cleanup()

	chapters, err := library.ExtractEPUBChapters(srcPath)
	if err != nil {
		return fmt.Errorf("epub parse: %w", err)
	}
	if len(chapters) == 0 {
		return errors.New("no readable text found in book")
	}

	settings, err := w.store.GetTTSSettings(ctx)
	if err != nil {
		return err
	}
	if !settings.Enabled || settings.BaseURL == "" {
		return errors.New("TTS sidecar is not configured")
	}
	client := NewClient(settings)
	timeout := time.Duration(settings.TimeoutSec) * time.Second
	voice := job.Voice
	if voice == "" {
		voice = settings.DefaultVoice
	}

	staging := w.stagingDir(job.ID)
	if err := os.MkdirAll(staging, 0o750); err != nil {
		return err
	}
	total := len(chapters)
	done := w.countStaged(staging, total)
	if err := w.store.UpdateTTSJobProgress(ctx, job.ID, total, done); err != nil {
		return err
	}

	for i := done; i < total; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if active, err := w.store.TTSJobIsActive(ctx, job.ID); err != nil {
			return err
		} else if !active {
			return errJobAborted
		}
		if p, err := w.store.GetTTSUserPrefs(ctx, job.UserID); err == nil && !InWindow(time.Now(), p) {
			return w.store.RequeueTTSJob(ctx, job.ID, NextWindowStart(time.Now(), p).Unix())
		}
		if err := w.synthChapter(ctx, client, staging, i, chapters[i], voice, job.Speed, timeout); err != nil {
			return err
		}
		if err := w.store.UpdateTTSJobProgress(ctx, job.ID, total, i+1); err != nil {
			return err
		}
	}

	outDir := outputDirName(book.Title, voice)
	if err := w.publish(ctx, fs, staging, outDir, chapters); err != nil {
		return fmt.Errorf("publish: %w", err)
	}
	if err := w.store.SetTTSJobOutput(ctx, job.ID, outDir+"/"); err != nil {
		return err
	}
	w.finishBook(ctx, book, outDir)
	_ = os.RemoveAll(staging)
	return w.store.FinishTTSJob(ctx, job.ID, models.TTSJobDone, "")
}

// synthChapter renders one chapter's chunks into a .part file, renamed to
// the final .mp3 only when the chapter is complete. A crash mid-chapter
// leaves a partial .part that a later run discards and re-renders, so the
// library never sees truncated audio.
func (w *Worker) synthChapter(ctx context.Context, client *Client, staging string, idx int, ch library.EPUBChapter, voice string, speed float64, timeout time.Duration) error {
	final := filepath.Join(staging, chapterStagingName(idx))
	part := final + ".part"
	_ = os.Remove(part)
	f, err := os.OpenFile(part, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o640)
	if err != nil {
		return err
	}
	ok := false
	defer func() {
		_ = f.Close()
		if !ok {
			_ = os.Remove(part)
		}
	}()

	for _, chunk := range ChunkParagraphs(ch.Paragraphs, JobChunkChars) {
		audio, _, err := w.synthOnce(ctx, client, chunk, voice, speed, timeout)
		if err != nil {
			// One retry after a short pause handles transient sidecar hiccups.
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(chunkRetryPause):
			}
			audio, _, err = w.synthOnce(ctx, client, chunk, voice, speed, timeout)
			if err != nil {
				return err
			}
		}
		if _, err := f.Write(audio); err != nil {
			return err
		}
	}
	if err := f.Close(); err != nil {
		return err
	}
	if err := os.Rename(part, final); err != nil {
		return err
	}
	ok = true
	return nil
}

func (w *Worker) synthOnce(ctx context.Context, client *Client, text, voice string, speed float64, timeout time.Duration) ([]byte, string, error) {
	cctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	audio, ct, err := client.SynthesizeFormat(cctx, text, voice, speed, JobAudioFormat)
	if err != nil {
		return nil, "", err
	}
	// Guard against endpoints that ignore response_format and hand back
	// another container; writing WAV bytes into a .mp3 would be silent corruption.
	ct = strings.ToLower(ct)
	if strings.HasPrefix(ct, "audio/") && !strings.Contains(ct, "mpeg") && !strings.Contains(ct, "mp3") {
		return nil, "", fmt.Errorf("sidecar returned %s, expected mp3", ct)
	}
	return audio, ct, nil
}

// countStaged counts completed chapter files already in the staging dir.
func (w *Worker) countStaged(staging string, total int) int {
	done := 0
	for i := 0; i < total; i++ {
		info, err := os.Stat(filepath.Join(staging, chapterStagingName(i)))
		if err != nil || info.Size() == 0 {
			break
		}
		done++
	}
	return done
}

func chapterStagingName(idx int) string {
	return fmt.Sprintf("chapter-%03d.mp3", idx+1)
}

// materialize returns a local path for the book file.
func (w *Worker) materialize(ctx context.Context, fs libfs.LibraryFS, relPath string) (string, func(), error) {
	if fs.Backend() == libfs.BackendLocal {
		p, err := fs.LocalAbsPath(relPath)
		return p, func() {}, err
	}
	tmp, err := libfs.Materialize(ctx, fs, relPath, w.cfg.TempDir())
	if err != nil {
		return "", nil, err
	}
	return tmp, func() { _ = os.Remove(tmp) }, nil
}

// publish copies staged chapters into the library under audiobooks/.
func (w *Worker) publish(ctx context.Context, fs libfs.LibraryFS, staging, outDir string, chapters []library.EPUBChapter) error {
	if err := fs.MkdirAll(ctx, outDir); err != nil {
		return err
	}
	for i, ch := range chapters {
		src := filepath.Join(staging, chapterStagingName(i))
		f, err := os.Open(src)
		if err != nil {
			return err
		}
		info, err := f.Stat()
		if err != nil {
			_ = f.Close()
			return err
		}
		rel := outDir + "/" + chapterFileName(i, ch)
		err = fs.Write(ctx, rel, f, info.Size())
		_ = f.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// finishBook rescans the library then copies the source book's metadata
// and cover onto the generated audiobook entry.
func (w *Worker) finishBook(ctx context.Context, book models.Book, outDir string) {
	if err := w.scanner.ScanLibrary(ctx, book.LibraryID); err != nil {
		w.log.Warn("post-tts scan failed", "err", err)
		return
	}
	// The merge collapses the folder into one "audiobook" set row keyed by
	// dir/; a single-chapter book stays a plain mp3 row instead.
	gen, err := w.store.GetBookByPath(ctx, book.LibraryID, outDir+"/")
	if err != nil {
		entries, lerr := w.generatedFileRows(ctx, book.LibraryID, outDir)
		if lerr != nil || len(entries) == 0 {
			return
		}
		gen = entries[0]
	}
	_, err = w.store.UpdateBookMetadata(ctx, gen.ID, models.BookUpdate{
		Title:       book.Title,
		Author:      book.Author,
		Series:      book.Series,
		SeriesIndex: book.SeriesIndex,
		Language:    book.Language,
		Description: book.Description,
	})
	if err != nil {
		w.log.Warn("tts metadata copy failed", "err", err)
	}
	if book.HasCover {
		src := library.CoverPath(w.cfg.CoverDir(), book.ID)
		if data, err := os.ReadFile(src); err == nil && len(data) > 0 {
			dst := library.CoverPath(w.cfg.CoverDir(), gen.ID)
			if err := os.WriteFile(dst, data, 0o600); err == nil {
				_ = w.store.SetBookHasCover(ctx, gen.ID, true)
			}
		}
	}
}

func (w *Worker) generatedFileRows(ctx context.Context, libraryID int64, outDir string) ([]models.Book, error) {
	page, err := w.store.ListBooks(ctx, models.BookQuery{
		LibraryID: libraryID,
		Format:    models.FormatMP3,
		Limit:     5000,
	})
	if err != nil {
		return nil, err
	}
	var out []models.Book
	for _, b := range page.Items {
		if strings.HasPrefix(b.RelPath, outDir+"/") {
			out = append(out, b)
		}
	}
	return out, nil
}

// outputDirName builds audiobooks/<Title> (<voice>) with unsafe chars removed.
func outputDirName(title, voice string) string {
	base := sanitizeName(title)
	if base == "" {
		base = "untitled"
	}
	if v := sanitizeName(voice); v != "" {
		base += " (" + v + ")"
	}
	if len(base) > 96 {
		base = strings.TrimRight(base[:96], " -_(")
	}
	return outputRoot + "/" + base
}

// chapterFileName produces NN - Title.mp3 for library output.
func chapterFileName(idx int, ch library.EPUBChapter) string {
	title := sanitizeName(ch.Title)
	if title == "" {
		title = fmt.Sprintf("Chapter %d", idx+1)
	}
	name := fmt.Sprintf("%02d - %s", idx+1, title)
	if len([]rune(name)) > 90 {
		name = string([]rune(name)[:90])
	}
	return name + ".mp3"
}

func sanitizeName(s string) string {
	var b strings.Builder
	for _, r := range s {
		if nameSafeRune(r) {
			b.WriteRune(r)
		} else {
			b.WriteByte(' ')
		}
	}
	s = spaceRuns.ReplaceAllString(b.String(), " ")
	return strings.TrimSpace(strings.Trim(s, "-."))
}
