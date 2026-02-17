package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"sftrails/models"
)

// statusColumns is the shared SELECT expression for computing trail status.
// It uses a single vote subquery (v) filtered to 12h, with conditional SUMs
// to derive both the 4h and 12h windows from one table scan.
const statusColumns = `
	COALESCE(
		CASE
			WHEN COALESCE(v.total_4h, 0) >= 3 THEN
				CASE
					WHEN v.open_4h > v.closed_4h THEN 'open'
					WHEN v.closed_4h > v.open_4h THEN 'closed'
					ELSE 'unknown'
				END
			WHEN COALESCE(v.total_12h, 0) >= 3 THEN
				CASE
					WHEN v.open_12h > v.closed_12h THEN 'open'
					WHEN v.closed_12h > v.open_12h THEN 'closed'
					ELSE 'unknown'
				END
			ELSE 'unknown'
		END,
		'unknown'
	) AS status,
	CASE
		WHEN COALESCE(v.total_4h, 0) >= 3 THEN COALESCE(v.open_4h, 0)
		ELSE COALESCE(v.open_12h, 0)
	END AS open_votes,
	CASE
		WHEN COALESCE(v.total_4h, 0) >= 3 THEN COALESCE(v.closed_4h, 0)
		ELSE COALESCE(v.closed_12h, 0)
	END AS closed_votes,
	CASE
		WHEN COALESCE(v.total_4h, 0) >= 3 THEN COALESCE(v.total_4h, 0)
		ELSE COALESCE(v.total_12h, 0)
	END AS total_votes`

// voteSubquery returns a single LEFT JOIN that computes both 4h and 12h
// vote windows in one pass over the votes table (filtered to 12h).
// When filtered is true, the subquery includes a WHERE trail_id placeholder.
func voteSubquery(filtered bool) string {
	where := ""
	if filtered {
		where = "trail_id = " + ph(1) + " AND "
	}
	return fmt.Sprintf(`
	LEFT JOIN (
		SELECT
			trail_id,
			SUM(CASE WHEN vote = 'open' AND created_at >= %s THEN 1 ELSE 0 END) AS open_4h,
			SUM(CASE WHEN vote = 'closed' AND created_at >= %s THEN 1 ELSE 0 END) AS closed_4h,
			SUM(CASE WHEN created_at >= %s THEN 1 ELSE 0 END) AS total_4h,
			SUM(CASE WHEN vote = 'open' THEN 1 ELSE 0 END) AS open_12h,
			SUM(CASE WHEN vote = 'closed' THEN 1 ELSE 0 END) AS closed_12h,
			COUNT(*) AS total_12h
		FROM votes
		WHERE %screated_at >= %s
		GROUP BY trail_id
	) v ON t.id = v.trail_id`, datetimeAge(4), datetimeAge(4), datetimeAge(4), where, datetimeAge(12))
}

func scanTrail(scanner interface{ Scan(dest ...any) error }) (models.TrailWithStatus, error) {
	var t models.TrailWithStatus
	var status string
	if err := scanner.Scan(
		&t.ID, &t.Name, &t.Location, &t.City, &t.Description,
		&t.Latitude, &t.Longitude, &t.CreatedAt, &t.UpdatedAt,
		&status, &t.OpenVotes, &t.ClosedVotes, &t.TotalVotes,
	); err != nil {
		return t, err
	}
	t.Status = models.TrailStatus(status)
	return t, nil
}

func GetTrailsWithStatus(ctx context.Context, db *sql.DB) ([]models.TrailWithStatus, error) {
	query := `
	SELECT
		t.id, t.name, t.location, t.city, t.description,
		t.latitude, t.longitude, t.created_at, t.updated_at,` +
		statusColumns + `
	FROM trails t` +
		voteSubquery(false) + `
	ORDER BY t.name`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query trails: %w", err)
	}
	defer rows.Close()

	trails := make([]models.TrailWithStatus, 0, 12)
	for rows.Next() {
		t, err := scanTrail(rows)
		if err != nil {
			return nil, fmt.Errorf("scan trail: %w", err)
		}
		trails = append(trails, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate trails: %w", err)
	}
	return trails, nil
}

func GetTrailWithStatus(ctx context.Context, db *sql.DB, trailID int64) (*models.TrailWithStatus, error) {
	query := `
	SELECT
		t.id, t.name, t.location, t.city, t.description,
		t.latitude, t.longitude, t.created_at, t.updated_at,` +
		statusColumns + `
	FROM trails t` +
		voteSubquery(true) + `
	WHERE t.id = ` + ph(2)

	t, err := scanTrail(db.QueryRowContext(ctx, query, trailID, trailID))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query trail %d: %w", trailID, err)
	}
	return &t, nil
}

func CastVote(ctx context.Context, db *sql.DB, trailID int64, vote models.VoteType, ip string, fingerprint string) error {
	dup, err := HasRecentVote(ctx, db, trailID, ip, fingerprint)
	if err != nil {
		return fmt.Errorf("check recent vote: %w", err)
	}
	if dup {
		return nil
	}
	_, err = db.ExecContext(ctx,
		`INSERT INTO votes (trail_id, vote, ip_address, fingerprint) VALUES (`+placeholders(4)+`)`,
		trailID, string(vote), ip, fingerprint,
	)
	if err != nil {
		return fmt.Errorf("insert vote: %w", err)
	}
	return nil
}

func ResetVotes(ctx context.Context, db *sql.DB) (int64, error) {
	result, err := db.ExecContext(ctx, `DELETE FROM votes`)
	if err != nil {
		return 0, fmt.Errorf("delete votes: %w", err)
	}
	return result.RowsAffected()
}

func HasRecentVote(ctx context.Context, db *sql.DB, trailID int64, ip string, fingerprint string) (bool, error) {
	var exists bool
	err := db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM votes WHERE trail_id = `+ph(1)+` AND ip_address = `+ph(2)+` AND fingerprint = `+ph(3)+` AND created_at >= `+datetimeAge(1)+` LIMIT 1)`,
		trailID, ip, fingerprint,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("query recent vote: %w", err)
	}
	return exists, nil
}
