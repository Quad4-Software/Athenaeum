package storage

import (
	"context"
	"database/sql"
	"errors"
	"slices"

	"athenaeum/internal/models"
)

// LibraryAccess describes which libraries a user may access.
type LibraryAccess struct {
	Restricted bool
	LibraryIDs []int64
}

// AccessibleLibraries resolves library access for a user.
// Admins and anonymous users have unrestricted access.
// Non-admins with rows in user_libraries are restricted to those mounts.
// Non-admins without rows may access all libraries.
func (s *Store) AccessibleLibraries(ctx context.Context, user models.User) (LibraryAccess, error) {
	if user.ID == 0 || user.IsAdmin {
		return LibraryAccess{}, nil
	}
	ids, err := s.ListUserLibraryIDs(ctx, user.ID)
	if err != nil {
		return LibraryAccess{}, err
	}
	if len(ids) == 0 {
		return LibraryAccess{}, nil
	}
	return LibraryAccess{Restricted: true, LibraryIDs: ids}, nil
}

// UserCanAccessLibrary reports whether user may use the library mount.
func (s *Store) UserCanAccessLibrary(ctx context.Context, user models.User, libraryID int64) (bool, error) {
	acc, err := s.AccessibleLibraries(ctx, user)
	if err != nil {
		return false, err
	}
	if !acc.Restricted {
		return true, nil
	}
	if slices.Contains(acc.LibraryIDs, libraryID) {
		return true, nil
	}
	return false, nil
}

// ListUserLibraryIDs returns explicit library grants for a user.
func (s *Store) ListUserLibraryIDs(ctx context.Context, userID int64) ([]int64, error) {
	rows, err := s.queryContext(ctx,
		`SELECT library_id FROM user_libraries WHERE user_id=? ORDER BY library_id`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// SetUserLibraries replaces library grants for a user.
func (s *Store) SetUserLibraries(ctx context.Context, userID int64, libraryIDs []int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `DELETE FROM user_libraries WHERE user_id=?`, userID); err != nil {
		return err
	}
	for _, libID := range libraryIDs {
		var exists int
		if err := tx.QueryRowContext(ctx, `SELECT 1 FROM libraries WHERE id=?`, libID).Scan(&exists); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return ErrNotFound
			}
			return err
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO user_libraries (user_id, library_id) VALUES (?,?)`, userID, libID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// ListLibrariesForUser returns libraries visible to the user.
func (s *Store) ListLibrariesForUser(ctx context.Context, user models.User) ([]models.Library, error) {
	acc, err := s.AccessibleLibraries(ctx, user)
	if err != nil {
		return nil, err
	}
	if !acc.Restricted {
		return s.ListLibraries(ctx)
	}
	if len(acc.LibraryIDs) == 0 {
		return []models.Library{}, nil
	}
	all, err := s.ListLibraries(ctx)
	if err != nil {
		return nil, err
	}
	allowed := make(map[int64]struct{}, len(acc.LibraryIDs))
	for _, id := range acc.LibraryIDs {
		allowed[id] = struct{}{}
	}
	var out []models.Library
	for _, lib := range all {
		if _, ok := allowed[lib.ID]; ok {
			out = append(out, lib)
		}
	}
	return out, nil
}
