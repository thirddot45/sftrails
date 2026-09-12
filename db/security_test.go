package db

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"testing"
	"time"

	"sftrails/models"
)

func TestConcurrentVoteDeduplicationAcrossConnections(t *testing.T) {
	dsn := securityDSN(t)
	first, err := Open(dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer first.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := Initialize(ctx, first); err != nil {
		t.Fatal(err)
	}
	second, err := Open(dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
	var wg sync.WaitGroup
	start := make(chan struct{})
	failures := make(chan error, 24)
	for i := range 24 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			d := []*sql.DB{first, second}[i%2]
			failures <- CastVote(ctx, d, 1, models.VoteOpen, "203.0.113.1", "same-browser")
		}(i)
	}
	close(start)
	wg.Wait()
	close(failures)
	for err := range failures {
		if err != nil {
			t.Fatal(err)
		}
	}
	var count int
	if err := first.QueryRowContext(ctx, "SELECT COUNT(*) FROM votes").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("concurrent duplicate votes accepted: %d", count)
	}
	if err := CastVote(ctx, second, 1, models.VoteClosed, "203.0.113.2", "different-browser"); err != nil {
		t.Fatal(err)
	}
	trail, err := GetTrailWithStatus(ctx, first, 1)
	if err != nil || trail == nil || trail.TotalVotes != 2 {
		t.Fatalf("independent voter rejected: %+v, %v", trail, err)
	}
	if err := CastVote(ctx, second, 999999, models.VoteOpen, "203.0.113.1", "same-browser"); !errors.Is(err, ErrTrailNotFound) {
		t.Errorf("missing trail error: %v", err)
	}
	if err := RecordPageView(ctx, first, "/", "test-pseudonym"); err != nil {
		t.Fatal(err)
	}
	metrics, err := GetSiteMetrics(ctx, first)
	if err != nil || metrics.TotalViews != 1 || metrics.ViewsToday != 1 {
		t.Fatalf("metrics: %+v, %v", metrics, err)
	}
}
