package db

import (
	"database/sql"

	"sftrails/models"
)

func GetTrailsWithStatus(db *sql.DB) ([]models.TrailWithStatus, error) {
	query := `
	SELECT
		t.id, t.name, t.location, t.city, t.description,
		t.latitude, t.longitude, t.created_at, t.updated_at,
		COALESCE(
			CASE
				WHEN COALESCE(v4.total, 0) >= 3 THEN
					CASE
						WHEN v4.open_votes > v4.closed_votes THEN 'open'
						WHEN v4.closed_votes > v4.open_votes THEN 'closed'
						ELSE 'unknown'
					END
				WHEN COALESCE(v12.total, 0) >= 3 THEN
					CASE
						WHEN v12.open_votes > v12.closed_votes THEN 'open'
						WHEN v12.closed_votes > v12.open_votes THEN 'closed'
						ELSE 'unknown'
					END
				ELSE 'unknown'
			END,
			'unknown'
		) AS status,
		CASE
			WHEN COALESCE(v4.total, 0) >= 3 THEN COALESCE(v4.open_votes, 0)
			ELSE COALESCE(v12.open_votes, 0)
		END AS open_votes,
		CASE
			WHEN COALESCE(v4.total, 0) >= 3 THEN COALESCE(v4.closed_votes, 0)
			ELSE COALESCE(v12.closed_votes, 0)
		END AS closed_votes,
		CASE
			WHEN COALESCE(v4.total, 0) >= 3 THEN COALESCE(v4.total, 0)
			ELSE COALESCE(v12.total, 0)
		END AS total_votes
	FROM trails t
	LEFT JOIN (
		SELECT
			trail_id,
			SUM(CASE WHEN vote = 'open' THEN 1 ELSE 0 END) AS open_votes,
			SUM(CASE WHEN vote = 'closed' THEN 1 ELSE 0 END) AS closed_votes,
			COUNT(*) AS total
		FROM votes
		WHERE created_at >= datetime('now', '-4 hours')
		GROUP BY trail_id
	) v4 ON t.id = v4.trail_id
	LEFT JOIN (
		SELECT
			trail_id,
			SUM(CASE WHEN vote = 'open' THEN 1 ELSE 0 END) AS open_votes,
			SUM(CASE WHEN vote = 'closed' THEN 1 ELSE 0 END) AS closed_votes,
			COUNT(*) AS total
		FROM votes
		WHERE created_at >= datetime('now', '-12 hours')
		GROUP BY trail_id
	) v12 ON t.id = v12.trail_id
	ORDER BY t.name
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trails []models.TrailWithStatus
	for rows.Next() {
		var t models.TrailWithStatus
		var status string
		if err := rows.Scan(
			&t.ID, &t.Name, &t.Location, &t.City, &t.Description,
			&t.Latitude, &t.Longitude, &t.CreatedAt, &t.UpdatedAt,
			&status, &t.OpenVotes, &t.ClosedVotes, &t.TotalVotes,
		); err != nil {
			return nil, err
		}
		t.Status = models.TrailStatus(status)
		trails = append(trails, t)
	}
	return trails, rows.Err()
}

func GetTrailWithStatus(db *sql.DB, trailID int64) (*models.TrailWithStatus, error) {
	query := `
	SELECT
		t.id, t.name, t.location, t.city, t.description,
		t.latitude, t.longitude, t.created_at, t.updated_at,
		COALESCE(
			CASE
				WHEN COALESCE(v4.total, 0) >= 3 THEN
					CASE
						WHEN v4.open_votes > v4.closed_votes THEN 'open'
						WHEN v4.closed_votes > v4.open_votes THEN 'closed'
						ELSE 'unknown'
					END
				WHEN COALESCE(v12.total, 0) >= 3 THEN
					CASE
						WHEN v12.open_votes > v12.closed_votes THEN 'open'
						WHEN v12.closed_votes > v12.open_votes THEN 'closed'
						ELSE 'unknown'
					END
				ELSE 'unknown'
			END,
			'unknown'
		) AS status,
		CASE
			WHEN COALESCE(v4.total, 0) >= 3 THEN COALESCE(v4.open_votes, 0)
			ELSE COALESCE(v12.open_votes, 0)
		END AS open_votes,
		CASE
			WHEN COALESCE(v4.total, 0) >= 3 THEN COALESCE(v4.closed_votes, 0)
			ELSE COALESCE(v12.closed_votes, 0)
		END AS closed_votes,
		CASE
			WHEN COALESCE(v4.total, 0) >= 3 THEN COALESCE(v4.total, 0)
			ELSE COALESCE(v12.total, 0)
		END AS total_votes
	FROM trails t
	LEFT JOIN (
		SELECT
			trail_id,
			SUM(CASE WHEN vote = 'open' THEN 1 ELSE 0 END) AS open_votes,
			SUM(CASE WHEN vote = 'closed' THEN 1 ELSE 0 END) AS closed_votes,
			COUNT(*) AS total
		FROM votes
		WHERE trail_id = ? AND created_at >= datetime('now', '-4 hours')
		GROUP BY trail_id
	) v4 ON t.id = v4.trail_id
	LEFT JOIN (
		SELECT
			trail_id,
			SUM(CASE WHEN vote = 'open' THEN 1 ELSE 0 END) AS open_votes,
			SUM(CASE WHEN vote = 'closed' THEN 1 ELSE 0 END) AS closed_votes,
			COUNT(*) AS total
		FROM votes
		WHERE trail_id = ? AND created_at >= datetime('now', '-12 hours')
		GROUP BY trail_id
	) v12 ON t.id = v12.trail_id
	WHERE t.id = ?
	`

	var t models.TrailWithStatus
	var status string
	err := db.QueryRow(query, trailID, trailID, trailID).Scan(
		&t.ID, &t.Name, &t.Location, &t.City, &t.Description,
		&t.Latitude, &t.Longitude, &t.CreatedAt, &t.UpdatedAt,
		&status, &t.OpenVotes, &t.ClosedVotes, &t.TotalVotes,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	t.Status = models.TrailStatus(status)
	return &t, nil
}

func CastVote(db *sql.DB, trailID int64, vote models.VoteType, ip string, fingerprint string) error {
	dup, err := HasRecentVote(db, trailID, ip, fingerprint)
	if err != nil {
		return err
	}
	if dup {
		return nil
	}
	_, err = db.Exec(
		`INSERT INTO votes (trail_id, vote, ip_address, fingerprint) VALUES (?, ?, ?, ?)`,
		trailID, string(vote), ip, fingerprint,
	)
	return err
}

func ResetVotes(db *sql.DB) (int64, error) {
	result, err := db.Exec(`DELETE FROM votes`)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func HasRecentVote(db *sql.DB, trailID int64, ip string, fingerprint string) (bool, error) {
	var count int
	err := db.QueryRow(
		`SELECT COUNT(*) FROM votes WHERE trail_id = ? AND ip_address = ? AND fingerprint = ? AND created_at >= datetime('now', '-1 hour')`,
		trailID, ip, fingerprint,
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
