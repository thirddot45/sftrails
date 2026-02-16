package db

import (
	"context"
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

func Open(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	db.SetMaxOpenConns(1)
	pragmas := []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA foreign_keys=ON",
		"PRAGMA busy_timeout=5000",
	}
	for _, p := range pragmas {
		if _, err := db.Exec(p); err != nil {
			db.Close()
			return nil, fmt.Errorf("exec pragma %q: %w", p, err)
		}
	}
	return db, nil
}

func Initialize(ctx context.Context, d *sql.DB) error {
	if err := RunMigrations(ctx, d); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}
	if err := SeedTrails(ctx, d); err != nil {
		return fmt.Errorf("seed trails: %w", err)
	}
	return nil
}
