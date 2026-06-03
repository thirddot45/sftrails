//go:build postgres

package db

import "fmt"

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
func dateOnly(col string) string { return "to_char(" + col + ", 'YYYY-MM-DD')" }

// todayPredicate returns a boolean expression true when col falls on today.
func todayPredicate(col string) string { return col + "::date = CURRENT_DATE" }
