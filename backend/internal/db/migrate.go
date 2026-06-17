package db

import (
	"context"
	"embed"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// RunMigrations applies all pending SQL migrations.
func RunMigrations(pool *pgxpool.Pool) error {
	ctx := context.Background()

	if _, err := pool.Exec(
		ctx,
		`CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
	); err != nil {
		return fmt.Errorf("create schema_migrations table: %w", err)
	}

	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("read migrations directory: %w", err)
	}

	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".up.sql") {
			continue
		}
		files = append(files, name)
	}
	sort.Strings(files)

	for _, fileName := range files {
		if err := applyMigration(pool, fileName); err != nil {
			return err
		}
	}

	return nil
}

func applyMigration(pool *pgxpool.Pool, fileName string) error {
	ctx := context.Background()

	var exists bool
	if err := pool.QueryRow(
		ctx,
		"SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)",
		fileName,
	).Scan(&exists); err != nil {
		return fmt.Errorf("check migration %s: %w", fileName, err)
	}
	if exists {
		return nil
	}

	sqlBytes, err := migrationsFS.ReadFile(filepath.Join("migrations", fileName))
	if err != nil {
		return fmt.Errorf("read migration %s: %w", fileName, err)
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin migration tx %s: %w", fileName, err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, string(sqlBytes)); err != nil {
		return fmt.Errorf("execute migration %s: %w", fileName, err)
	}

	if _, err := tx.Exec(ctx, "INSERT INTO schema_migrations(version) VALUES ($1)", fileName); err != nil {
		return fmt.Errorf("record migration %s: %w", fileName, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit migration %s: %w", fileName, err)
	}

	return nil
}
