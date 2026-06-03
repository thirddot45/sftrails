package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"sftrails/models"
)

// RecordPageView stores a single page view. visitorHash is a salted, one-way
// hash of the visitor used only for approximate unique counts; no IP address is
// ever stored.
func RecordPageView(ctx context.Context, db *sql.DB, path, visitorHash string) error {
	_, err := db.ExecContext(ctx,
		`INSERT INTO page_views (path, visitor_hash) VALUES (`+placeholders(2)+`)`,
		path, visitorHash,
	)
	if err != nil {
		return fmt.Errorf("insert page view: %w", err)
	}
	return nil
}

// GetSiteMetrics returns aggregate traffic metrics for the dashboard. It
// includes all-time totals, today's counts, a per-day breakdown for the last 7
// days, and the most-visited paths. No IP information is returned.
func GetSiteMetrics(ctx context.Context, db *sql.DB) (models.SiteMetrics, error) {
	var m models.SiteMetrics

	// All-time totals and today's counts in a single scan.
	row := db.QueryRowContext(ctx, fmt.Sprintf(`
		SELECT
			COUNT(*),
			COUNT(DISTINCT visitor_hash),
			COALESCE(SUM(CASE WHEN %s THEN 1 ELSE 0 END), 0),
			COUNT(DISTINCT CASE WHEN %s THEN visitor_hash END)
		FROM page_views`,
		todayPredicate("created_at"), todayPredicate("created_at")))
	if err := row.Scan(&m.TotalViews, &m.UniqueVisits, &m.ViewsToday, &m.UniqueToday); err != nil {
		return m, fmt.Errorf("query metric totals: %w", err)
	}

	days, err := getDailyCounts(ctx, db)
	if err != nil {
		return m, err
	}
	m.Days = days

	paths, err := getTopPaths(ctx, db, 10)
	if err != nil {
		return m, err
	}
	m.TopPaths = paths

	return m, nil
}

// getDailyCounts returns per-day view and unique-visitor counts for the last 7
// days (UTC), oldest first, with zero-filled gaps for days that had no traffic.
func getDailyCounts(ctx context.Context, db *sql.DB) ([]models.DayCount, error) {
	query := fmt.Sprintf(`
		SELECT %s AS day, COUNT(*), COUNT(DISTINCT visitor_hash)
		FROM page_views
		WHERE created_at >= %s
		GROUP BY day`,
		dateOnly("created_at"), datetimeAge(24*7))

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query daily counts: %w", err)
	}
	defer rows.Close()

	seen := make(map[string]models.DayCount, 8)
	for rows.Next() {
		var d models.DayCount
		if err := rows.Scan(&d.Date, &d.Views, &d.Unique); err != nil {
			return nil, fmt.Errorf("scan daily count: %w", err)
		}
		seen[d.Date] = d
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate daily counts: %w", err)
	}

	// Zero-fill the 7-day window so the chart always shows a continuous range.
	days := make([]models.DayCount, 0, 7)
	now := time.Now().UTC()
	for i := 6; i >= 0; i-- {
		key := now.AddDate(0, 0, -i).Format("2006-01-02")
		if d, ok := seen[key]; ok {
			days = append(days, d)
		} else {
			days = append(days, models.DayCount{Date: key})
		}
	}
	return days, nil
}

// getTopPaths returns the most-visited paths, highest first.
func getTopPaths(ctx context.Context, db *sql.DB, limit int) ([]models.PathCount, error) {
	query := fmt.Sprintf(`
		SELECT path, COUNT(*) AS views
		FROM page_views
		GROUP BY path
		ORDER BY views DESC, path ASC
		LIMIT %d`, limit)

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query top paths: %w", err)
	}
	defer rows.Close()

	paths := make([]models.PathCount, 0, limit)
	for rows.Next() {
		var p models.PathCount
		if err := rows.Scan(&p.Path, &p.Views); err != nil {
			return nil, fmt.Errorf("scan top path: %w", err)
		}
		paths = append(paths, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate top paths: %w", err)
	}
	return paths, nil
}
