// cmd/initdb/main.go
// 用于初始化/迁移数据库表结构。
//
// 用法：
//   go run ./cmd/initdb        # 执行 up 迁移
//   go run ./cmd/initdb down   # 执行 down 迁移
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/t2-travel-terminal/t2-travel-terminal/internal/config"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/datastore"
	"go.uber.org/zap"
)

const schemaMigrationsTable = `
CREATE TABLE IF NOT EXISTS schema_migrations (
    version TEXT PRIMARY KEY,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
`

type migration struct {
	version string
	up      string
	down    string
}

func main() {
	logger, _ := zap.NewProduction()
	defer func() { _ = logger.Sync() }()

	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("failed to load config", zap.Error(err))
	}

	pool, err := datastore.NewPool(ctx, cfg)
	if err != nil {
		logger.Fatal("failed to create database pool", zap.Error(err))
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		logger.Fatal("failed to ping database", zap.Error(err))
	}
	logger.Info("database connected")

	if _, err := pool.Exec(ctx, schemaMigrationsTable); err != nil {
		logger.Fatal("failed to create schema_migrations table", zap.Error(err))
	}

	migrationsDir, err := migrationsDir()
	if err != nil {
		logger.Fatal("failed to locate migrations directory", zap.Error(err))
	}

	migrations, err := loadMigrations(migrationsDir)
	if err != nil {
		logger.Fatal("failed to load migrations", zap.Error(err))
	}

	action := "up"
	if len(os.Args) > 1 {
		action = os.Args[1]
	}

	switch action {
	case "up":
		if err := migrateUp(ctx, pool, migrations, logger); err != nil {
			logger.Fatal("migration up failed", zap.Error(err))
		}
	case "down":
		if err := migrateDown(ctx, pool, migrations, logger); err != nil {
			logger.Fatal("migration down failed", zap.Error(err))
		}
	default:
		fmt.Fprintf(os.Stderr, "usage: %s [up|down]\n", os.Args[0])
		os.Exit(1)
	}
}

func migrationsDir() (string, error) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("failed to get current file path")
	}
	projectRoot := filepath.Join(filepath.Dir(currentFile), "..", "..")
	return filepath.Abs(filepath.Join(projectRoot, "migrations"))
}

func loadMigrations(dir string) ([]migration, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read migrations dir: %w", err)
	}

	byVersion := make(map[string]*migration)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".sql") {
			continue
		}

		parts := strings.SplitN(name, "_", 2)
		if len(parts) != 2 {
			continue
		}
		version := parts[0]

		m, ok := byVersion[version]
		if !ok {
			m = &migration{version: version}
			byVersion[version] = m
		}

		path := filepath.Join(dir, name)
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read migration file %s: %w", name, err)
		}

		if strings.HasSuffix(name, ".up.sql") {
			m.up = string(content)
		} else if strings.HasSuffix(name, ".down.sql") {
			m.down = string(content)
		}
	}

	var versions []string
	for v := range byVersion {
		versions = append(versions, v)
	}
	sort.Strings(versions)

	var migrations []migration
	for _, v := range versions {
		m := byVersion[v]
		if m.up == "" {
			return nil, fmt.Errorf("migration %s is missing .up.sql file", v)
		}
		migrations = append(migrations, *m)
	}

	return migrations, nil
}

func appliedVersions(ctx context.Context, pool *datastore.Pool) (map[string]bool, error) {
	rows, err := pool.Query(ctx, "SELECT version FROM schema_migrations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = true
	}
	return applied, rows.Err()
}

func migrateUp(ctx context.Context, pool *datastore.Pool, migrations []migration, logger *zap.Logger) error {
	applied, err := appliedVersions(ctx, pool)
	if err != nil {
		return fmt.Errorf("fetch applied versions: %w", err)
	}

	for _, m := range migrations {
		if applied[m.version] {
			logger.Info("migration already applied, skipping", zap.String("version", m.version))
			continue
		}

		tx, err := pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("begin transaction for %s: %w", m.version, err)
		}

		if _, err := tx.Exec(ctx, m.up); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("execute up migration %s: %w", m.version, err)
		}

		if _, err := tx.Exec(ctx,
			"INSERT INTO schema_migrations (version) VALUES ($1)", m.version,
		); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("record migration %s: %w", m.version, err)
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit migration %s: %w", m.version, err)
		}

		logger.Info("migration applied", zap.String("version", m.version))
	}

	logger.Info("all up migrations completed")
	return nil
}

func migrateDown(ctx context.Context, pool *datastore.Pool, migrations []migration, logger *zap.Logger) error {
	rows, err := pool.Query(ctx,
		"SELECT version FROM schema_migrations ORDER BY version DESC",
	)
	if err != nil {
		return fmt.Errorf("fetch applied versions: %w", err)
	}
	defer rows.Close()

	var applied []string
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return err
		}
		applied = append(applied, version)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	migrationMap := make(map[string]migration)
	for _, m := range migrations {
		migrationMap[m.version] = m
	}

	for _, version := range applied {
		m, ok := migrationMap[version]
		if !ok {
			logger.Warn("no down migration found, skipping", zap.String("version", version))
			continue
		}
		if m.down == "" {
			logger.Warn("down migration is empty, skipping", zap.String("version", version))
			continue
		}

		tx, err := pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("begin transaction for %s: %w", version, err)
		}

		if _, err := tx.Exec(ctx,
			"DELETE FROM schema_migrations WHERE version = $1", version,
		); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("delete migration record %s: %w", version, err)
		}

		if _, err := tx.Exec(ctx, m.down); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("execute down migration %s: %w", version, err)
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit down migration %s: %w", version, err)
		}

		logger.Info("migration rolled back", zap.String("version", version))
	}

	logger.Info("all down migrations completed")
	return nil
}
