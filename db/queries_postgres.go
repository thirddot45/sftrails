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
