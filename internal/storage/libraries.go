package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"athenaeum/internal/libfs"
	"athenaeum/internal/models"
)

// EnsureDefaultLibrary creates or updates the primary library mount from startup config.
func (s *Store) EnsureDefaultLibrary(ctx context.Context, mountPath string) error {
	mountPath, err := filepath.Abs(mountPath)
	if err != nil {
		return err
	}
	var id int64
	var existing string
	err = s.queryRowContext(ctx, `SELECT id, mount_path FROM libraries WHERE id=1`).Scan(&id, &existing)
	if errors.Is(err, sql.ErrNoRows) {
		_, err = s.execContext(ctx, `
INSERT INTO libraries (id, name, mount_path, sort_order, created_at, backend, backend_config)
VALUES (1, 'Main Library', ?, 0, ?, 'local', '{}')`, mountPath, time.Now().Unix())
		return err
	}
	if err != nil {
		return err
	}
	if existing == "" || existing != mountPath {
		_, err = s.execContext(ctx, `
UPDATE libraries SET mount_path=?, backend='local', backend_config='{}' WHERE id=1`, mountPath)
	}
	return err
}

// ListLibraries returns all mounted libraries ordered for display.
func (s *Store) ListLibraries(ctx context.Context) ([]models.Library, error) {
	rows, err := s.queryContext(ctx, `
SELECT id, name, mount_path, sort_order, created_at, backend, backend_config
FROM libraries
ORDER BY sort_order ASC, LOWER(name) ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Library
	for rows.Next() {
		lib, err := scanLibraryRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, lib)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for i := range out {
		if err := s.fillLibraryBookCount(ctx, &out[i]); err != nil {
			return nil, err
		}
	}
	return out, nil
}

// GetLibrary returns one library by id.
func (s *Store) GetLibrary(ctx context.Context, id int64) (models.Library, error) {
	row := s.queryRowContext(ctx, `
SELECT id, name, mount_path, sort_order, created_at, backend, backend_config
FROM libraries WHERE id=?`, id)
	lib, err := scanLibraryRow(row)
	if errors.Is(err, sql.ErrNoRows) {
		return models.Library{}, ErrNotFound
	}
	if err != nil {
		return models.Library{}, err
	}
	if err := s.fillLibraryBookCount(ctx, &lib); err != nil {
		return models.Library{}, err
	}
	return lib, nil
}

// LibraryMountPath returns the absolute mount path for a library.
func (s *Store) LibraryMountPath(ctx context.Context, id int64) (string, error) {
	var path string
	err := s.queryRowContext(ctx, `SELECT mount_path FROM libraries WHERE id=?`, id).Scan(&path)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	return path, err
}

// LibraryBackend returns the storage backend for a library.
func (s *Store) LibraryBackend(ctx context.Context, id int64) (string, error) {
	var backend string
	err := s.queryRowContext(ctx, `SELECT COALESCE(backend,'local') FROM libraries WHERE id=?`, id).Scan(&backend)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	if backend == "" {
		backend = models.LibraryBackendLocal
	}
	return backend, err
}

type libraryBackendRow struct {
	Backend string
	Config  string
	Path    string
}

func (s *Store) libraryBackendRow(ctx context.Context, id int64) (libraryBackendRow, error) {
	var row libraryBackendRow
	err := s.queryRowContext(ctx, `
SELECT mount_path, COALESCE(backend,'local'), COALESCE(backend_config,'{}')
FROM libraries WHERE id=?`, id).Scan(&row.Path, &row.Backend, &row.Config)
	if errors.Is(err, sql.ErrNoRows) {
		return libraryBackendRow{}, ErrNotFound
	}
	return row, err
}

// OpenLibraryFS opens the filesystem backend for a library mount.
func (s *Store) OpenLibraryFS(ctx context.Context, id int64) (libfs.LibraryFS, error) {
	row, err := s.libraryBackendRow(ctx, id)
	if err != nil {
		return nil, err
	}
	cfg, err := libfsConfigFromRow(row)
	if err != nil {
		return nil, err
	}
	return libfs.Open(cfg)
}

func libfsConfigFromRow(row libraryBackendRow) (libfs.Config, error) {
	backend := strings.TrimSpace(strings.ToLower(row.Backend))
	if backend == "" {
		backend = libfs.BackendLocal
	}
	cfg := libfs.Config{Backend: backend, Path: row.Path}
	if backend == libfs.BackendS3 {
		var s3 libfs.S3Config
		if err := json.Unmarshal([]byte(row.Config), &s3); err != nil {
			return libfs.Config{}, errors.New("invalid s3 backend config")
		}
		cfg.S3 = s3
	}
	return cfg, nil
}

// CreateLibrary adds a new mounted library root.
func (s *Store) CreateLibrary(ctx context.Context, name, mountPath string) (models.Library, error) {
	return s.CreateLibraryFull(ctx, models.LibraryCreate{
		Name:      name,
		MountPath: mountPath,
		Backend:   models.LibraryBackendLocal,
	})
}

// CreateLibraryFull adds a local or S3 library mount.
func (s *Store) CreateLibraryFull(ctx context.Context, in models.LibraryCreate) (models.Library, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return models.Library{}, errors.New("name is required")
	}
	backend := strings.TrimSpace(strings.ToLower(in.Backend))
	if backend == "" {
		backend = models.LibraryBackendLocal
	}

	var mountPath string
	var configJSON string
	switch backend {
	case models.LibraryBackendLocal:
		var err error
		mountPath, err = filepath.Abs(strings.TrimSpace(in.MountPath))
		if err != nil {
			return models.Library{}, err
		}
		if err := validateMountPath(mountPath); err != nil {
			return models.Library{}, err
		}
		configJSON = "{}"
	case models.LibraryBackendS3:
		if in.S3 == nil {
			return models.Library{}, errors.New("s3 config is required")
		}
		s3cfg := s3InputToLibfs(*in.S3)
		if err := libfs.ValidateS3Config(s3cfg, true); err != nil {
			return models.Library{}, err
		}
		if err := libfs.TestS3(ctx, s3cfg); err != nil {
			return models.Library{}, err
		}
		mountPath = libfs.MountLabel(s3cfg.Bucket, strings.Trim(s3cfg.Prefix, "/"))
		raw, err := json.Marshal(s3cfg) // #nosec G117 -- AccessKey stored encrypted-at-rest in backend_config
		if err != nil {
			return models.Library{}, err
		}
		configJSON = string(raw)
	default:
		return models.Library{}, errors.New("unsupported library backend")
	}

	var maxOrder int
	_ = s.queryRowContext(ctx, `SELECT COALESCE(MAX(sort_order),0) FROM libraries`).Scan(&maxOrder)
	now := time.Now().Unix()
	id, err := s.insertID(ctx, `
INSERT INTO libraries (name, mount_path, sort_order, created_at, backend, backend_config)
VALUES (?,?,?,?,?,?) RETURNING id`,
		name, mountPath, maxOrder+1, now, backend, configJSON)
	if err != nil {
		return models.Library{}, err
	}
	return s.GetLibrary(ctx, id)
}

// UpdateLibrary renames or changes the mount path of a library.
func (s *Store) UpdateLibrary(ctx context.Context, id int64, name, mountPath string) (models.Library, error) {
	return s.UpdateLibraryFull(ctx, id, models.LibraryCreate{
		Name:      name,
		MountPath: mountPath,
		Backend:   models.LibraryBackendLocal,
	})
}

// UpdateLibraryFull updates a local or S3 library mount.
func (s *Store) UpdateLibraryFull(ctx context.Context, id int64, in models.LibraryCreate) (models.Library, error) {
	existing, err := s.GetLibrary(ctx, id)
	if err != nil {
		return models.Library{}, err
	}
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return models.Library{}, errors.New("name is required")
	}
	backend := strings.TrimSpace(strings.ToLower(in.Backend))
	if backend == "" {
		backend = existing.Backend
	}
	if backend == "" {
		backend = models.LibraryBackendLocal
	}

	var mountPath string
	var configJSON string
	switch backend {
	case models.LibraryBackendLocal:
		mountPath, err = filepath.Abs(strings.TrimSpace(in.MountPath))
		if err != nil {
			return models.Library{}, err
		}
		if err := validateMountPath(mountPath); err != nil {
			return models.Library{}, err
		}
		configJSON = "{}"
	case models.LibraryBackendS3:
		if in.S3 == nil {
			return models.Library{}, errors.New("s3 config is required")
		}
		s3cfg := s3InputToLibfs(*in.S3)
		row, err := s.libraryBackendRow(ctx, id)
		if err != nil {
			return models.Library{}, err
		}
		if s3cfg.SecretKey == "" && row.Backend == libfs.BackendS3 {
			var prev libfs.S3Config
			if err := json.Unmarshal([]byte(row.Config), &prev); err == nil {
				s3cfg.SecretKey = prev.SecretKey
			}
		}
		if err := libfs.ValidateS3Config(s3cfg, true); err != nil {
			return models.Library{}, err
		}
		if err := libfs.TestS3(ctx, s3cfg); err != nil {
			return models.Library{}, err
		}
		mountPath = libfs.MountLabel(s3cfg.Bucket, strings.Trim(s3cfg.Prefix, "/"))
		raw, err := json.Marshal(s3cfg) // #nosec G117 -- AccessKey stored encrypted-at-rest in backend_config
		if err != nil {
			return models.Library{}, err
		}
		configJSON = string(raw)
	default:
		return models.Library{}, errors.New("unsupported library backend")
	}

	res, err := s.execContext(ctx, `
UPDATE libraries SET name=?, mount_path=?, backend=?, backend_config=? WHERE id=?`,
		name, mountPath, backend, configJSON, id)
	if err != nil {
		return models.Library{}, err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return models.Library{}, ErrNotFound
	}
	return s.GetLibrary(ctx, id)
}

// ReorderLibraries updates sort_order for each library id in order.
func (s *Store) ReorderLibraries(ctx context.Context, ids []int64) error {
	for i, id := range ids {
		res, err := s.execContext(ctx, `UPDATE libraries SET sort_order=? WHERE id=?`, i, id)
		if err != nil {
			return err
		}
		n, _ := res.RowsAffected()
		if n == 0 {
			return ErrNotFound
		}
	}
	return nil
}

// DeleteLibrary removes a library mount and all indexed books in it.
func (s *Store) DeleteLibrary(ctx context.Context, id int64) error {
	if _, err := s.GetLibrary(ctx, id); err != nil {
		return err
	}
	if _, err := s.execContext(ctx, `DELETE FROM books WHERE library_id=?`, id); err != nil {
		return err
	}
	res, err := s.execContext(ctx, `DELETE FROM libraries WHERE id=?`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func scanLibraryRow(row scanner) (models.Library, error) {
	var lib models.Library
	var created int64
	var backend string
	var configJSON string
	err := row.Scan(&lib.ID, &lib.Name, &lib.MountPath, &lib.SortOrder, &created, &backend, &configJSON)
	if err != nil {
		return models.Library{}, err
	}
	lib.CreatedAt = time.Unix(created, 0)
	lib.Backend = backend
	if lib.Backend == "" {
		lib.Backend = models.LibraryBackendLocal
	}
	if lib.Backend == models.LibraryBackendS3 {
		var cfg libfs.S3Config
		if err := json.Unmarshal([]byte(configJSON), &cfg); err == nil {
			lib.S3 = &models.LibraryS3Public{
				Endpoint:     cfg.Endpoint,
				Region:       cfg.Region,
				Bucket:       cfg.Bucket,
				Prefix:       strings.Trim(cfg.Prefix, "/"),
				AccessKey:    cfg.AccessKey,
				UsePathStyle: cfg.UsePathStyle,
				TLS:          cfg.TLS,
				HasSecretKey: cfg.SecretKey != "",
			}
		}
	}
	return lib, nil
}

func (s *Store) fillLibraryBookCount(ctx context.Context, lib *models.Library) error {
	return s.queryRowContext(ctx, `SELECT COUNT(*) FROM books WHERE library_id=?`, lib.ID).Scan(&lib.BookCount)
}

func validateMountPath(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return errors.New("mount path must exist and be readable")
	}
	if !info.IsDir() {
		return errors.New("mount path must be a directory")
	}
	return nil
}

func s3InputToLibfs(in models.LibraryS3Input) libfs.S3Config {
	return libfs.S3Config{
		Endpoint:     in.Endpoint,
		Region:       in.Region,
		Bucket:       in.Bucket,
		Prefix:       in.Prefix,
		AccessKey:    in.AccessKey,
		SecretKey:    in.SecretKey,
		UsePathStyle: in.UsePathStyle,
		TLS:          in.TLS,
	}
}
