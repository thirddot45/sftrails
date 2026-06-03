//go:build postgres

package db

import (
	"context"
	"database/sql"
	"fmt"
)

func RunMigrations(ctx context.Context, db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS trails (
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL,
		location TEXT NOT NULL DEFAULT '',
		city TEXT NOT NULL DEFAULT '',
		description TEXT NOT NULL DEFAULT '',
		latitude DOUBLE PRECISION NOT NULL DEFAULT 0,
		longitude DOUBLE PRECISION NOT NULL DEFAULT 0,
		created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS votes (
		id SERIAL PRIMARY KEY,
		trail_id INTEGER NOT NULL REFERENCES trails(id),
		vote TEXT NOT NULL CHECK(vote IN ('open', 'closed')),
		ip_address TEXT NOT NULL DEFAULT '',
		fingerprint TEXT NOT NULL DEFAULT '',
		created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_votes_trail_created ON votes(trail_id, created_at DESC);
	CREATE INDEX IF NOT EXISTS idx_votes_dedup ON votes(trail_id, ip_address, fingerprint, created_at DESC);

	-- page_views records site traffic for the metrics dashboard. We deliberately
	-- do NOT store IP addresses: visitor_hash is a salted, one-way hash used only
	-- to approximate unique-visitor counts, and cannot be reversed to an IP.
	CREATE TABLE IF NOT EXISTS page_views (
		id SERIAL PRIMARY KEY,
		path TEXT NOT NULL DEFAULT '',
		visitor_hash TEXT NOT NULL DEFAULT '',
		created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_page_views_created ON page_views(created_at DESC);
	CREATE INDEX IF NOT EXISTS idx_page_views_visitor ON page_views(visitor_hash);
	`
	if _, err := db.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("exec schema: %w", err)
	}
	return nil
}
