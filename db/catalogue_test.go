package db

import (
	"context"
	"database/sql"
	"testing"
)

func catalogueTestDB(t *testing.T) *sql.DB {
	t.Helper()
	d, err := Open(securityDSN(t))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { d.Close() })
	if err := Initialize(context.Background(), d); err != nil {
		t.Fatal(err)
	}
	return d
}

func assertCatalogueValue(t *testing.T, d *sql.DB, query, want string) {
	t.Helper()
	var got string
	if err := d.QueryRow(query).Scan(&got); err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("%s: got %q, want %q", query, got, want)
	}
}

func TestCatalogueFreshSeed(t *testing.T) {
	d := catalogueTestDB(t)
	assertCatalogueValue(t, d, "SELECT city FROM trails WHERE name = 'Virginia Key'", "Miami")
	assertCatalogueValue(t, d, "SELECT location FROM trails WHERE name = 'Virginia Key'", "Miami")
	assertCatalogueValue(t, d, "SELECT description FROM trails WHERE name = 'Tree Tops Park'", "Recreational cycling on designated bike paths; check park maps and signs")
}

func TestCatalogueCorrectsExistingRowsWithoutReplacingTrails(t *testing.T) {
	d := catalogueTestDB(t)
	ctx := context.Background()
	for _, query := range []string{
		"UPDATE trails SET city = 'Key Biscayne', location = 'Key Biscayne' WHERE name = 'Virginia Key'",
		"UPDATE trails SET description = 'Shaded trails through tropical hardwood hammock' WHERE name = 'Tree Tops Park'",
		"INSERT INTO votes (trail_id, vote) SELECT id, 'open' FROM trails WHERE name = 'Virginia Key'",
	} {
		if _, err := d.Exec(query); err != nil {
			t.Fatal(err)
		}
	}
	var originalID int
	if err := d.QueryRow("SELECT id FROM trails WHERE name = 'Virginia Key'").Scan(&originalID); err != nil {
		t.Fatal(err)
	}
	if err := Initialize(ctx, d); err != nil {
		t.Fatal(err)
	}
	assertCatalogueValue(t, d, "SELECT city FROM trails WHERE name = 'Virginia Key'", "Miami")
	assertCatalogueValue(t, d, "SELECT location FROM trails WHERE name = 'Virginia Key'", "Miami")
	assertCatalogueValue(t, d, "SELECT description FROM trails WHERE name = 'Tree Tops Park'", "Recreational cycling on designated bike paths; check park maps and signs")
	var id, votes, trails int
	if err := d.QueryRow("SELECT trail_id FROM votes").Scan(&id); err != nil || id != originalID {
		t.Fatalf("vote association changed: id=%d, error=%v", id, err)
	}
	if err := d.QueryRow("SELECT COUNT(*) FROM votes").Scan(&votes); err != nil || votes != 1 {
		t.Fatalf("votes changed: count=%d, error=%v", votes, err)
	}
	if err := d.QueryRow("SELECT COUNT(*) FROM trails").Scan(&trails); err != nil || trails != 12 {
		t.Fatalf("trails changed: count=%d, error=%v", trails, err)
	}
	// A second initialization must not touch already-corrected timestamps.
	if _, err := d.Exec("UPDATE trails SET updated_at = '2000-01-01 00:00:00'"); err != nil {
		t.Fatal(err)
	}
	if err := Initialize(ctx, d); err != nil {
		t.Fatal(err)
	}
	var changed int
	if err := d.QueryRow("SELECT COUNT(*) FROM trails WHERE updated_at <> '2000-01-01 00:00:00'").Scan(&changed); err != nil || changed != 0 {
		t.Fatalf("repeat initialization changed timestamps: count=%d, error=%v", changed, err)
	}
}

func TestCataloguePreservesEditedFieldsAndUnrelatedTrails(t *testing.T) {
	d := catalogueTestDB(t)
	for _, query := range []string{
		"UPDATE trails SET city = 'Miami (verified)', location = 'Key Biscayne' WHERE name = 'Virginia Key'",
		"UPDATE trails SET description = 'Locally edited cycling details' WHERE name = 'Tree Tops Park'",
		"UPDATE trails SET city = 'Key Biscayne', location = 'Key Biscayne' WHERE name = 'Markham Park'",
	} {
		if _, err := d.Exec(query); err != nil {
			t.Fatal(err)
		}
	}
	if err := Initialize(context.Background(), d); err != nil {
		t.Fatal(err)
	}
	assertCatalogueValue(t, d, "SELECT city FROM trails WHERE name = 'Virginia Key'", "Miami (verified)")
	assertCatalogueValue(t, d, "SELECT location FROM trails WHERE name = 'Virginia Key'", "Miami")
	assertCatalogueValue(t, d, "SELECT description FROM trails WHERE name = 'Tree Tops Park'", "Locally edited cycling details")
	assertCatalogueValue(t, d, "SELECT city FROM trails WHERE name = 'Markham Park'", "Key Biscayne")
	assertCatalogueValue(t, d, "SELECT location FROM trails WHERE name = 'Markham Park'", "Key Biscayne")
}
