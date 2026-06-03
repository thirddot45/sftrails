//go:build !postgres

package db

import "fmt"

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
