package db

import "database/sql"

func RunMigrations(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS trails (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		location TEXT NOT NULL DEFAULT '',
		city TEXT NOT NULL DEFAULT '',
		description TEXT NOT NULL DEFAULT '',
		latitude REAL NOT NULL DEFAULT 0,
		longitude REAL NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS votes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		trail_id INTEGER NOT NULL REFERENCES trails(id),
		vote TEXT NOT NULL CHECK(vote IN ('open', 'closed')),
		ip_address TEXT NOT NULL DEFAULT '',
		fingerprint TEXT NOT NULL DEFAULT '',
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_votes_trail_created ON votes(trail_id, created_at DESC);
	CREATE INDEX IF NOT EXISTS idx_votes_dedup ON votes(trail_id, ip_address, fingerprint, created_at DESC);
	`
	_, err := db.Exec(schema)
	return err
}
