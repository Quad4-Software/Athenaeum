package library

import (
	"context"
	"path/filepath"
	"sort"
	"strings"

	"athenaeum/internal/models"
)

type audioTrack struct {
	id      int64
	relPath string
	title   string
	format  string
	size    int64
	author  string
	series  string
}

// MergeAudiobookFolders groups multi-file audio directories into audiobook sets.
func (s *Scanner) MergeAudiobookFolders(ctx context.Context, libraryID int64) error {
	rows, err := s.store.ListAudioBooksByLibrary(ctx, libraryID)
	if err != nil {
		return err
	}
	byDir := make(map[string][]audioTrack)
	for _, b := range rows {
		if b.Hidden {
			continue
		}
		dir := filepath.Dir(b.RelPath)
		if dir == "." {
			continue
		}
		byDir[dir] = append(byDir[dir], audioTrack{
			id: b.ID, relPath: b.RelPath, title: b.Title, author: b.Author,
			series: b.Series, format: b.Format, size: b.FileSize,
		})
	}
	for dir, tracks := range byDir {
		if len(tracks) < 2 {
			continue
		}
		skipDir := false
		for _, t := range tracks {
			if t.format == models.FormatM4B {
				skipDir = true
				break
			}
		}
		if skipDir {
			continue
		}
		sort.Slice(tracks, func(i, j int) bool {
			return naturalLess(tracks[i].relPath, tracks[j].relPath)
		})
		title := tracks[0].title
		if tracks[0].series != "" {
			title = tracks[0].series
		}
		var totalSize int64
		at := make([]models.AudiobookTrack, len(tracks))
		for i, t := range tracks {
			totalSize += t.size
			at[i] = models.AudiobookTrack{
				Index: i, Title: t.title, RelPath: t.relPath, Format: t.format, FileSize: t.size,
			}
		}
		mount, err := s.store.LibraryMountPath(ctx, libraryID)
		if err != nil {
			return err
		}
		setBook := &models.Book{
			LibraryID: libraryID,
			Title:     title,
			Author:    tracks[0].author,
			Series:    tracks[0].series,
			Format:    models.FormatAudiobook,
			RelPath:   dir + "/",
			AbsPath:   mount + "/" + dir,
			FileSize:  totalSize,
		}
		if !strings.HasPrefix(mount, "s3://") {
			setBook.AbsPath = filepath.Join(mount, dir)
		}
		setID, err := s.store.UpsertAudiobookSet(ctx, setBook, at)
		if err != nil {
			s.log.Warn("audiobook set failed", "dir", dir, "err", err)
			continue
		}
		for _, t := range tracks {
			if err := s.store.HideBook(ctx, t.id, setID); err != nil {
				s.log.Warn("hide track failed", "id", t.id, "err", err)
			}
		}
	}
	return s.store.PruneOrphanAudiobookSets(ctx, libraryID)
}
