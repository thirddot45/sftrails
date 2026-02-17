package db

import (
	"context"
	"database/sql"
	"fmt"
)

func Initialize(ctx context.Context, d *sql.DB) error {
	if err := RunMigrations(ctx, d); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}
	if err := SeedTrails(ctx, d); err != nil {
		return fmt.Errorf("seed trails: %w", err)
	}
	return nil
}
