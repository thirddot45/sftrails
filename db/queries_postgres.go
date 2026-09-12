//go:build postgres

package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

func lockVoteTrail(ctx context.Context, tx *sql.Tx, trailID int64) error {
	var id int64
	err := tx.QueryRowContext(ctx, `SELECT id FROM trails WHERE id = $1 FOR UPDATE`, trailID).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrTrailNotFound
	}
	return err
}

func ph(n int) string { return fmt.Sprintf("$%d", n) }

func placeholders(n int) string {
	s := "$1"
	for i := 2; i <= n; i++ {
		s += fmt.Sprintf(", $%d", i)
	}
	return s
}

func datetimeAge(hours int) string {
	return fmt.Sprintf("NOW() - INTERVAL '%d hours'", hours)
}

// dateOnly returns an expression yielding the YYYY-MM-DD date for a column.
func dateOnly(col string) string { return "to_char(" + col + " AT TIME ZONE 'UTC', 'YYYY-MM-DD')" }

// todayPredicate returns a boolean expression true when col falls on today.
func todayPredicate(col string) string {
	return "(" + col + " AT TIME ZONE 'UTC')::date = (NOW() AT TIME ZONE 'UTC')::date"
}
