package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		databaseURL = "postgres://postgres:postgres@localhost:5432/dentaldesk?sslmode=disable"
	}
	migrationsDir := strings.TrimSpace(os.Getenv("MIGRATIONS_DIR"))
	if migrationsDir == "" {
		migrationsDir = "infra/migrations"
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ctx := context.Background()
	if err := db.PingContext(ctx); err != nil {
		log.Fatal(err)
	}

	if err := ensureMigrationsTable(ctx, db); err != nil {
		log.Fatal(err)
	}

	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.sql"))
	if err != nil {
		log.Fatal(err)
	}
	sort.Strings(files)

	for _, file := range files {
		name := filepath.Base(file)
		applied, err := migrationApplied(ctx, db, name)
		if err != nil {
			log.Fatal(err)
		}
		if applied {
			log.Printf("skip %s", name)
			continue
		}
		if err := applyMigration(ctx, db, file, name); err != nil {
			log.Fatal(err)
		}
		log.Printf("applied %s", name)
	}
}

func ensureMigrationsTable(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			name TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)
	`)
	return err
}

func migrationApplied(ctx context.Context, db *sql.DB, name string) (bool, error) {
	var exists bool
	err := db.QueryRowContext(ctx, `
		SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE name = $1)
	`, name).Scan(&exists)
	return exists, err
}

func applyMigration(ctx context.Context, db *sql.DB, file, name string) error {
	sqlBytes, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	if len(strings.TrimSpace(string(sqlBytes))) == 0 {
		return fmt.Errorf("%s is empty", file)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if _, err := tx.ExecContext(ctx, string(sqlBytes)); err != nil {
		return errors.Join(fmt.Errorf("apply %s", name), err)
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO schema_migrations (name) VALUES ($1)
	`, name); err != nil {
		return err
	}
	return tx.Commit()
}
