package demo

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"athenaeum/internal/brand"
	"athenaeum/internal/library"
	"athenaeum/internal/models"
	"athenaeum/internal/storage"
)

const (
	markerName     = ".athenaeum-demo-seeded"
	catalogVersion = "pd-v2"
)

// Seed writes demo media into libraryDir, covers into coverDir, and indexes them in store.
// It is idempotent when the catalog version marker matches.
func Seed(ctx context.Context, store *storage.Store, libraryDir, coverDir string, log *slog.Logger) error {
	if log == nil {
		log = slog.Default()
	}

	demoRoot := filepath.Join(libraryDir, "demo")
	marker := filepath.Join(demoRoot, markerName)
	if data, err := os.ReadFile(marker); err == nil && strings.Contains(string(data), "catalog="+catalogVersion) { // #nosec G304 -- marker under demo library root
		log.Info("demo library already seeded", "path", demoRoot, "catalog", catalogVersion)
		return nil
	}

	libs, err := store.ListLibraries(ctx)
	if err != nil {
		return err
	}
	var libraryID int64 = 1
	for _, lib := range libs {
		if filepath.Clean(lib.MountPath) == filepath.Clean(libraryDir) {
			libraryID = lib.ID
			break
		}
	}

	if err := os.MkdirAll(demoRoot, 0o750); err != nil {
		return err
	}
	if err := os.MkdirAll(coverDir, 0o750); err != nil {
		return err
	}

	now := time.Now()
	shelfIDs := map[string]int64{}
	var seededIDs []int64
	coversFetched := 0

	for i, e := range Catalog() {
		rel, abs, err := writeMediaFile(libraryDir, e)
		if err != nil {
			return fmt.Errorf("write media %s: %w", e.Slug, err)
		}
		fileSize := e.FileSize
		if e.Format == "epub" && e.GutenbergID > 0 {
			if err := FetchGutenbergEPUB(ctx, e.GutenbergID, abs); err != nil {
				log.Debug("gutenberg epub unavailable, keeping stub", "title", e.Title, "err", err)
			} else if info, err := os.Stat(abs); err == nil {
				fileSize = info.Size()
			}
		}

		var id int64
		if e.Format == "audiobook" {
			tracks := []models.AudiobookTrack{
				{Index: 0, RelPath: filepath.Join(rel, "01 - Opening.mp3"), Title: "Stave I", Format: "mp3", FileSize: 32000},
				{Index: 1, RelPath: filepath.Join(rel, "02 - Kiln.mp3"), Title: "Stave II", Format: "mp3", FileSize: 32000},
				{Index: 2, RelPath: filepath.Join(rel, "03 - Closing.mp3"), Title: "Stave III", Format: "mp3", FileSize: 32000},
			}
			book := &models.Book{
				LibraryID: libraryID, Title: e.Title, Author: e.Author, Series: e.Series,
				SeriesIndex: e.SeriesIndex, Format: models.FormatAudiobook, RelPath: rel, AbsPath: abs,
				FileSize: fileSize, Language: e.Language, Description: e.Description,
				HasCover: true, MetaEdited: true, CoverEdited: true,
				AddedAt: now.Add(-time.Duration(i) * time.Hour), ModifiedAt: now,
			}
			id, err = store.UpsertAudiobookSet(ctx, book, tracks)
		} else {
			book := &models.Book{
				LibraryID: libraryID, Title: e.Title, Author: e.Author, Series: e.Series,
				SeriesIndex: e.SeriesIndex, Format: e.Format, RelPath: rel, AbsPath: abs,
				FileSize: fileSize, Language: e.Language, Description: e.Description,
				HasCover: true, MetaEdited: true, CoverEdited: true,
				AddedAt: now.Add(-time.Duration(i) * time.Hour), ModifiedAt: now,
			}
			id, err = store.UpsertBook(ctx, book, now.Unix()-int64(i*3600))
		}
		if err != nil {
			return fmt.Errorf("upsert %s: %w", e.Slug, err)
		}
		seededIDs = append(seededIDs, id)

		coverPath := library.CoverPath(coverDir, id)
		coverData, fromRemote := resolveCover(ctx, e, log)
		if fromRemote {
			coversFetched++
		}
		if err := os.WriteFile(coverPath, coverData, 0o600); err != nil {
			return err
		}
		if err := store.SetBookCover(ctx, id, true); err != nil {
			return err
		}

		if e.Progress > 0 {
			_ = store.SaveProgress(ctx, models.AnonymousUserID, models.Progress{
				BookID: id, Location: fmt.Sprintf("demo:%.2f", e.Progress), Percent: e.Progress,
			})
		}
		if e.Favorite {
			_ = store.SetFavorite(ctx, models.AnonymousUserID, id, true)
		}
		if e.Shelf != "" {
			shelfID, ok := shelfIDs[e.Shelf]
			if !ok {
				c, err := store.CreateCollection(ctx, models.AnonymousUserID, e.Shelf, "Public-domain demo shelf")
				if err != nil {
					return err
				}
				shelfID = c.ID
				shelfIDs[e.Shelf] = shelfID
			}
			_ = store.AddToCollection(ctx, models.AnonymousUserID, shelfID, id)
		}
	}

	if _, err := store.CreateSmartCollection(ctx, models.AnonymousUserID, "Audiobooks", "Public-domain audio titles", models.SmartQuery{
		Format: "audio",
	}); err != nil {
		log.Warn("demo smart collection", "err", err)
	}

	if err := store.EnsureAutoCollections(ctx); err != nil {
		log.Warn("demo auto collections", "err", err)
	}

	meta := fmt.Sprintf(
		"seeded_at=%s\ncatalog=%s\nbooks=%d\ncovers_fetched=%d\n",
		now.UTC().Format(time.RFC3339), catalogVersion, len(seededIDs), coversFetched,
	)
	if err := os.WriteFile(marker, []byte(meta), 0o640); err != nil { // #nosec G306 -- demo marker readable for operators
		return err
	}

	log.Info("demo library seeded",
		"books", len(seededIDs),
		"covers_fetched", coversFetched,
		"catalog", catalogVersion,
		"library", libraryDir,
	)
	return nil
}

func resolveCover(ctx context.Context, e Entry, log *slog.Logger) (data []byte, fromRemote bool) {
	if e.CoverURL != "" {
		if img, err := library.FetchCoverImage(ctx, e.CoverURL); err == nil && len(img) > 0 {
			return img, true
		} else if err != nil {
			log.Debug("demo cover fetch failed", "title", e.Title, "err", err)
		}
	}
	png, err := EncodeCoverBuffer(e.Title, e.Author)
	if err != nil {
		return []byte("demo"), false
	}
	return png, false
}

// FetchGutenbergEPUB downloads a Project Gutenberg EPUB when online (optional enrichment).
func FetchGutenbergEPUB(ctx context.Context, gutenbergID int, dest string) error {
	if gutenbergID <= 0 {
		return fmt.Errorf("invalid gutenberg id")
	}
	urls := []string{
		fmt.Sprintf("https://www.gutenberg.org/ebooks/%d.epub3.images", gutenbergID),
		fmt.Sprintf("https://www.gutenberg.org/ebooks/%d.epub.images", gutenbergID),
		fmt.Sprintf("https://www.gutenberg.org/ebooks/%d.epub.noimages", gutenbergID),
	}
	client := &http.Client{Timeout: 45 * time.Second}
	var last error
	for _, u := range urls {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
		if err != nil {
			last = err
			continue
		}
		req.Header.Set("User-Agent", brand.UserAgent("demo"))
		res, err := client.Do(req)
		if err != nil {
			last = err
			continue
		}
		body, err := io.ReadAll(io.LimitReader(res.Body, 40<<20))
		_ = res.Body.Close()
		if err != nil {
			last = err
			continue
		}
		if res.StatusCode != http.StatusOK || len(body) < 1000 {
			last = fmt.Errorf("%s: %s", u, res.Status)
			continue
		}
		if err := os.MkdirAll(filepath.Dir(dest), 0o750); err != nil {
			return err
		}
		return os.WriteFile(dest, body, 0o640) // #nosec G306 -- demo cover under library tree
	}
	return last
}
