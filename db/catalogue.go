package db

import (
	"context"
	"database/sql"
	"fmt"
)

// correctTrailCatalogue updates only the original seeded values. It is safe to
// repeat on startup and preserves edited fields, trail IDs, names and votes.
// Sources and verification limits: docs/catalogue-verification-2026-10-13.md.
func correctTrailCatalogue(ctx context.Context, d *sql.DB) error {
	tx, err := d.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin catalogue corrections: %w", err)
	}
	defer tx.Rollback()

	for _, query := range []string{
		`UPDATE trails SET city = 'Miami', updated_at = CURRENT_TIMESTAMP
		 WHERE name = 'Virginia Key' AND city = 'Key Biscayne'`,
		`UPDATE trails SET location = 'Miami', updated_at = CURRENT_TIMESTAMP
		 WHERE name = 'Virginia Key' AND location = 'Key Biscayne'`,
		`UPDATE trails SET description = 'Recreational cycling on designated bike paths; check park maps and signs', updated_at = CURRENT_TIMESTAMP
		 WHERE name = 'Tree Tops Park' AND description = 'Shaded trails through tropical hardwood hammock'`,
	} {
		if _, err := tx.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("update catalogue: %w", err)
		}
	}
	return tx.Commit()
}
