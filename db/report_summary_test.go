package db

import (
	"context"
	"testing"
	"time"

	"sftrails/models"
)

// Runs against SQLite and, with TEST_DATABASE_URL, PostgreSQL.
func TestReportSummaryWindowsAndTimestamp(t *testing.T) {
	cases := []struct {
		name                 string
		ages                 []int
		votes                []string
		open, closed, window int
		status               models.TrailStatus
	}{
		{"never reported", nil, nil, 0, 0, 12, models.StatusUnknown},
		{"stale report", []int{13}, []string{"open"}, 0, 0, 12, models.StatusUnknown},
		{"insufficient reports", []int{1, 6}, []string{"open", "closed"}, 1, 1, 12, models.StatusUnknown},
		{"four hour window excludes older reports", []int{1, 2, 3, 6}, []string{"open", "open", "closed", "closed"}, 2, 1, 4, models.StatusOpen},
		{"twelve hour fallback", []int{1, 5, 6}, []string{"closed", "closed", "open"}, 1, 2, 12, models.StatusClosed},
		{"tied reports", []int{1, 2, 3, 3}, []string{"open", "closed", "open", "closed"}, 2, 2, 4, models.StatusUnknown},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d, err := Open(securityDSN(t))
			if err != nil {
				t.Fatal(err)
			}
			defer d.Close()
			ctx := context.Background()
			if err := Initialize(ctx, d); err != nil {
				t.Fatal(err)
			}
			var latest time.Time
			for i, age := range tc.ages {
				stamp := time.Now().UTC().Truncate(time.Second).Add(-time.Duration(age) * time.Hour)
				if latest.IsZero() || stamp.After(latest) {
					latest = stamp
				}
				_, err := d.ExecContext(ctx, `INSERT INTO votes (trail_id, vote, created_at) VALUES (`+placeholders(3)+`)`, 1, tc.votes[i], stamp.Format("2006-01-02 15:04:05-07:00"))
				if err != nil {
					t.Fatal(err)
				}
			}
			check := func(trail models.TrailWithStatus) {
				t.Helper()
				if trail.OpenVotes != tc.open || trail.ClosedVotes != tc.closed || trail.TotalVotes != tc.open+tc.closed || trail.ReportWindowHours != tc.window || trail.Status != tc.status || !trail.LastReportAt.Equal(latest) {
					t.Fatalf("unexpected report summary: %+v; want counts %d/%d, window %d, status %s, last %s", trail, tc.open, tc.closed, tc.window, tc.status, latest)
				}
			}
			// Repeated reads must retain the report time rather than the load time.
			for range 2 {
				trail, err := GetTrailWithStatus(ctx, d, 1)
				if err != nil || trail == nil {
					t.Fatalf("single trail: %v, %v", trail, err)
				}
				check(*trail)
				trails, err := GetTrailsWithStatus(ctx, d)
				if err != nil {
					t.Fatal(err)
				}
				found := false
				for _, trail := range trails {
					if trail.ID == 1 {
						check(trail)
						found = true
					}
				}
				if !found {
					t.Fatal("trail missing from list")
				}
			}
		})
	}
}
