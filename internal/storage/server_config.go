package storage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"athenaeum/internal/models"
)

// GetServerConfig loads admin server settings. Password hash is included when includeSecret is true.
func (s *Store) GetServerConfig(ctx context.Context, includeSecret bool) (models.ServerConfig, error) {
	var cfg models.ServerConfig
	var metricsEnabled, metricsAuth, corsEnabled, cspEnabled, autoScan int
	var passwordHash string
	var autoScanInterval, scanWorkers int
	err := s.queryRowContext(ctx, `
SELECT metrics_enabled, metrics_auth, metrics_username, metrics_password_hash,
       trusted_proxies, cors_enabled, cors_origins, csp_enabled, csp_policy,
       COALESCE(auto_scan_enabled,0), COALESCE(auto_scan_interval_sec,300),
       COALESCE(scan_workers,0)
FROM server_config WHERE id=1`).
		Scan(&metricsEnabled, &metricsAuth, &cfg.MetricsUsername, &passwordHash,
			&cfg.TrustedProxies, &corsEnabled, &cfg.CORSOrigins, &cspEnabled, &cfg.CSPPolicy,
			&autoScan, &autoScanInterval, &scanWorkers)
	if errors.Is(err, sql.ErrNoRows) {
		return models.ServerConfig{MetricsAuth: true, CSPEnabled: true}, nil
	}
	if err != nil {
		return cfg, err
	}
	cfg.MetricsEnabled = metricsEnabled != 0
	cfg.MetricsAuth = metricsAuth != 0
	cfg.CORSEnabled = corsEnabled != 0
	cfg.CSPEnabled = cspEnabled != 0
	cfg.AutoScanEnabled = autoScan != 0
	if autoScanInterval > 0 {
		cfg.AutoScanInterval = autoScanInterval
	} else {
		cfg.AutoScanInterval = 300
	}
	cfg.ScanWorkers = scanWorkers
	cfg.PasswordSet = passwordHash != ""
	if includeSecret {
		cfg.MetricsPassword = passwordHash
	}
	return cfg, nil
}

// SaveServerConfig persists admin server settings. An empty metrics password keeps the existing hash.
func (s *Store) SaveServerConfig(ctx context.Context, cfg models.ServerConfig) error {
	existing, err := s.GetServerConfig(ctx, true)
	if err != nil {
		return err
	}
	passwordHash := cfg.MetricsPassword
	if passwordHash == "" {
		passwordHash = existing.MetricsPassword
	}
	_, err = s.execContext(ctx, `
UPDATE server_config SET
	metrics_enabled=?, metrics_auth=?, metrics_username=?, metrics_password_hash=?,
	trusted_proxies=?, cors_enabled=?, cors_origins=?, csp_enabled=?, csp_policy=?,
	auto_scan_enabled=?, auto_scan_interval_sec=?, scan_workers=?
WHERE id=1`,
		boolToInt(cfg.MetricsEnabled), boolToInt(cfg.MetricsAuth), cfg.MetricsUsername, passwordHash,
		cfg.TrustedProxies, boolToInt(cfg.CORSEnabled), cfg.CORSOrigins, boolToInt(cfg.CSPEnabled), cfg.CSPPolicy,
		boolToInt(cfg.AutoScanEnabled), cfg.AutoScanInterval, cfg.ScanWorkers)
	return err
}

// CreateGuestUser inserts a temporary account with an expiry timestamp.
func (s *Store) CreateGuestUser(ctx context.Context, username, passwordHash string, expiresAt time.Time, perms int64) (int64, error) {
	return s.insertID(ctx, `
INSERT INTO users (username, password_hash, is_admin, permissions, is_guest, expires_at, created_at)
VALUES (?,?,0,?,1,?,?) RETURNING id`,
		username, passwordHash, perms, expiresAt.Unix(), time.Now().Unix())
}

// PurgeExpiredGuests deletes guest accounts past their expiry.
func (s *Store) PurgeExpiredGuests(ctx context.Context) (int64, error) {
	now := time.Now().Unix()
	rows, err := s.queryContext(ctx,
		`SELECT id FROM users WHERE is_guest=1 AND expires_at > 0 AND expires_at < ?`, now)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return 0, err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}
	for _, id := range ids {
		if err := s.DeleteUser(ctx, id); err != nil {
			return 0, err
		}
	}
	return int64(len(ids)), nil
}
