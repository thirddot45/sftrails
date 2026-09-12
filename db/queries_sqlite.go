//go:build !postgres

package db

import (
	"context"
	"database/sql"
	"fmt"
)

func lockVoteTrail(ctx context.Context, tx *sql.Tx, trailID int64) error {
	// Obtain SQLite's write lock before taking the duplicate-check snapshot.
	result, err := tx.ExecContext(ctx, `UPDATE trails SET id = id WHERE id = ?`, trailID)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err == nil && n == 0 {
		return ErrTrailNotFound
	}
	return err
}

func ph(n int) string { return "?" }

func placeholders(n int) string {
	s := "?"
	for i := 1; i < n; i++ {
		s += ", ?"
	}
	return s
}

func datetimeAge(hours int) string {
	return fmt.Sprintf("datetime('now', '-%d hours')", hours)
}

// dateOnly returns an expression yielding the YYYY-MM-DD date for a column.
func dateOnly(col string) string { return "date(" + col + ")" }

// todayPredicate returns a boolean expression true when col falls on today.
func todayPredicate(col string) string { return "date(" + col + ") = date('now')" }
