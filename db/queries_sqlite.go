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
