//go:build !postgres

package db

import (
	"path/filepath"
	"testing"
)

func securityDSN(t *testing.T) string {
	return filepath.Join(t.TempDir(), "votes.db")
}
