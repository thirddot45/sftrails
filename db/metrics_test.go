//go:build !postgres

package db

import (
	"context"
	"testing"
)

func TestRecordAndGetSiteMetrics(t *testing.T) {
	d := setupTestDB(t)
	defer d.Close()
	ctx := context.Background()

	// Two distinct visitors; one of them views two pages.
	if err := RecordPageView(ctx, d, "/", "visitorA"); err != nil {
		t.Fatalf("RecordPageView error: %v", err)
	}
	if err := RecordPageView(ctx, d, "/", "visitorA"); err != nil {
		t.Fatalf("RecordPageView error: %v", err)
	}
	if err := RecordPageView(ctx, d, "/trail/markham-park", "visitorB"); err != nil {
		t.Fatalf("RecordPageView error: %v", err)
	}

	m, err := GetSiteMetrics(ctx, d)
	if err != nil {
		t.Fatalf("GetSiteMetrics error: %v", err)
	}

	if m.TotalViews != 3 {
		t.Errorf("Expected 3 total views, got %d", m.TotalViews)
	}
	if m.UniqueVisits != 2 {
		t.Errorf("Expected 2 unique visitors, got %d", m.UniqueVisits)
	}
	if m.ViewsToday != 3 {
		t.Errorf("Expected 3 views today, got %d", m.ViewsToday)
	}
	if m.UniqueToday != 2 {
		t.Errorf("Expected 2 unique today, got %d", m.UniqueToday)
	}

	// 7-day window is always zero-filled to exactly 7 entries.
	if len(m.Days) != 7 {
		t.Fatalf("Expected 7 day buckets, got %d", len(m.Days))
	}
	today := m.Days[len(m.Days)-1]
	if today.Views != 3 || today.Unique != 2 {
		t.Errorf("Expected today bucket 3 views / 2 unique, got %d / %d", today.Views, today.Unique)
	}

	// Top paths: "/" should lead with 2 views.
	if len(m.TopPaths) != 2 {
		t.Fatalf("Expected 2 top paths, got %d", len(m.TopPaths))
	}
	if m.TopPaths[0].Path != "/" || m.TopPaths[0].Views != 2 {
		t.Errorf("Expected top path '/' with 2 views, got %q with %d", m.TopPaths[0].Path, m.TopPaths[0].Views)
	}
}

func TestGetSiteMetricsEmpty(t *testing.T) {
	d := setupTestDB(t)
	defer d.Close()

	m, err := GetSiteMetrics(context.Background(), d)
	if err != nil {
		t.Fatalf("GetSiteMetrics error: %v", err)
	}
	if m.TotalViews != 0 || m.UniqueVisits != 0 {
		t.Errorf("Expected zero metrics, got %d views / %d unique", m.TotalViews, m.UniqueVisits)
	}
	if len(m.Days) != 7 {
		t.Errorf("Expected 7 zero-filled day buckets, got %d", len(m.Days))
	}
	if len(m.TopPaths) != 0 {
		t.Errorf("Expected no top paths, got %d", len(m.TopPaths))
	}
}
