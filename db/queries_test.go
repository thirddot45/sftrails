package db

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"sftrails/models"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	d, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Failed to open test DB: %v", err)
	}
	if err := Initialize(context.Background(), d); err != nil {
		t.Fatalf("Failed to initialize test DB: %v", err)
	}
	return d
}

func TestGetTrailsWithStatus(t *testing.T) {
	d := setupTestDB(t)
	defer d.Close()

	trails, err := GetTrailsWithStatus(context.Background(), d)
	if err != nil {
		t.Fatalf("GetTrailsWithStatus error: %v", err)
	}
	if len(trails) != 12 {
		t.Fatalf("Expected 12 trails, got %d", len(trails))
	}
	for _, trail := range trails {
		if trail.Status != models.StatusUnknown {
			t.Errorf("Expected unknown status for %s, got %s", trail.Name, trail.Status)
		}
	}
}

func TestCastVote(t *testing.T) {
	d := setupTestDB(t)
	defer d.Close()
	ctx := context.Background()

	err := CastVote(ctx, d, 1, models.VoteOpen, "127.0.0.1", "abc123")
	if err != nil {
		t.Fatalf("CastVote error: %v", err)
	}

	trail, err := GetTrailWithStatus(ctx, d, 1)
	if err != nil {
		t.Fatalf("GetTrailWithStatus error: %v", err)
	}
	if trail.OpenVotes != 1 {
		t.Errorf("Expected 1 open vote, got %d", trail.OpenVotes)
	}
	if trail.TotalVotes != 1 {
		t.Errorf("Expected 1 total vote, got %d", trail.TotalVotes)
	}
}

func TestDuplicateVoteRejection(t *testing.T) {
	d := setupTestDB(t)
	defer d.Close()
	ctx := context.Background()

	// First vote should succeed
	if err := CastVote(ctx, d, 1, models.VoteOpen, "127.0.0.1", "abc123"); err != nil {
		t.Fatalf("First CastVote error: %v", err)
	}

	// Second vote from same IP+fingerprint within 1 hour should be silently rejected
	if err := CastVote(ctx, d, 1, models.VoteOpen, "127.0.0.1", "abc123"); err != nil {
		t.Fatalf("Second CastVote error: %v", err)
	}

	trail, err := GetTrailWithStatus(ctx, d, 1)
	if err != nil {
		t.Fatalf("GetTrailWithStatus error: %v", err)
	}
	if trail.OpenVotes != 1 {
		t.Errorf("Expected 1 open vote (duplicate rejected), got %d", trail.OpenVotes)
	}
}

func TestDifferentIPCanVote(t *testing.T) {
	d := setupTestDB(t)
	defer d.Close()
	ctx := context.Background()

	if err := CastVote(ctx, d, 1, models.VoteOpen, "127.0.0.1", "abc123"); err != nil {
		t.Fatalf("First CastVote error: %v", err)
	}
	if err := CastVote(ctx, d, 1, models.VoteOpen, "192.168.1.1", "def456"); err != nil {
		t.Fatalf("Second CastVote error: %v", err)
	}

	trail, err := GetTrailWithStatus(ctx, d, 1)
	if err != nil {
		t.Fatalf("GetTrailWithStatus error: %v", err)
	}
	if trail.OpenVotes != 2 {
		t.Errorf("Expected 2 open votes, got %d", trail.OpenVotes)
	}
}

func TestStatusComputation(t *testing.T) {
	d := setupTestDB(t)
	defer d.Close()
	ctx := context.Background()

	// Cast 3 open votes (different IPs to avoid dedup)
	for i := range 3 {
		ip := "10.0.0." + string(rune('1'+i))
		if err := CastVote(ctx, d, 1, models.VoteOpen, ip, "fp"+string(rune('1'+i))); err != nil {
			t.Fatalf("CastVote error: %v", err)
		}
	}

	trail, err := GetTrailWithStatus(ctx, d, 1)
	if err != nil {
		t.Fatalf("GetTrailWithStatus error: %v", err)
	}
	if trail.Status != models.StatusOpen {
		t.Errorf("Expected open status with 3 open votes, got %s", trail.Status)
	}
}

func TestStatusMajorityClosed(t *testing.T) {
	d := setupTestDB(t)
	defer d.Close()
	ctx := context.Background()

	// 2 closed, 1 open = closed majority (with 3+ votes)
	CastVote(ctx, d, 1, models.VoteClosed, "10.0.0.1", "fp1")
	CastVote(ctx, d, 1, models.VoteClosed, "10.0.0.2", "fp2")
	CastVote(ctx, d, 1, models.VoteOpen, "10.0.0.3", "fp3")

	trail, err := GetTrailWithStatus(ctx, d, 1)
	if err != nil {
		t.Fatalf("GetTrailWithStatus error: %v", err)
	}
	if trail.Status != models.StatusClosed {
		t.Errorf("Expected closed status, got %s", trail.Status)
	}
}

func TestHasRecentVote(t *testing.T) {
	d := setupTestDB(t)
	defer d.Close()
	ctx := context.Background()

	has, err := HasRecentVote(ctx, d, 1, "127.0.0.1", "abc123")
	if err != nil {
		t.Fatalf("HasRecentVote error: %v", err)
	}
	if has {
		t.Error("Expected no recent vote")
	}

	CastVote(ctx, d, 1, models.VoteOpen, "127.0.0.1", "abc123")

	has, err = HasRecentVote(ctx, d, 1, "127.0.0.1", "abc123")
	if err != nil {
		t.Fatalf("HasRecentVote error: %v", err)
	}
	if !has {
		t.Error("Expected recent vote")
	}
}

func TestGetTrailWithStatusNotFound(t *testing.T) {
	d := setupTestDB(t)
	defer d.Close()

	trail, err := GetTrailWithStatus(context.Background(), d, 9999)
	if err != nil {
		t.Fatalf("GetTrailWithStatus error: %v", err)
	}
	if trail != nil {
		t.Error("Expected nil for non-existent trail")
	}
}

func TestStatusTimeWindow(t *testing.T) {
	d := setupTestDB(t)
	defer d.Close()
	ctx := context.Background()

	// Insert votes manually with timestamps older than 4 hours but within 12 hours
	fiveHoursAgo := time.Now().UTC().Add(-5 * time.Hour).Format("2006-01-02 15:04:05")
	for i := range 3 {
		_, err := d.Exec(
			`INSERT INTO votes (trail_id, vote, ip_address, fingerprint, created_at) VALUES (?, ?, ?, ?, ?)`,
			1, "open", "10.0.0."+string(rune('1'+i)), "fp"+string(rune('1'+i)), fiveHoursAgo,
		)
		if err != nil {
			t.Fatalf("Insert vote error: %v", err)
		}
	}

	trail, err := GetTrailWithStatus(ctx, d, 1)
	if err != nil {
		t.Fatalf("GetTrailWithStatus error: %v", err)
	}
	// Votes are outside 4h window but inside 12h window, and >= 3, so should compute status
	if trail.Status != models.StatusOpen {
		t.Errorf("Expected open status from 12h window, got %s", trail.Status)
	}
	if trail.TotalVotes != 3 {
		t.Errorf("Expected 3 total votes from 12h window, got %d", trail.TotalVotes)
	}
}
