package repository

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/uptimebeats/heartbeat-api/internal/models"
)

type HeartbeatRepository struct {
	DB *pgxpool.Pool
}

func NewHeartbeatRepository(db *pgxpool.Pool) *HeartbeatRepository {
	return &HeartbeatRepository{DB: db}
}

// GetSiteByUniqueID retrieves a heartbeat site by its unique ID
func (r *HeartbeatRepository) GetSiteByUniqueID(ctx context.Context, uniqueID string) (*models.HeartbeatSite, error) {
	query := `
		SELECT id, name, type, unique_id, created_at, updated_at, check_frequency, tolerance_duration, site_up, in_maintenance, last_incident, failure_count, org_id
		FROM sites_list_heartbeat
		WHERE unique_id = $1
	`
	row := r.DB.QueryRow(ctx, query, uniqueID)

	var site models.HeartbeatSite
	err := row.Scan(
		&site.ID, &site.Name, &site.Type, &site.UniqueID, &site.CreatedAt, &site.UpdatedAt,
		&site.CheckFrequency, &site.ToleranceDuration, &site.SiteUp, &site.InMaintenance,
		&site.LastIncident, &site.FailureCount, &site.OrgID,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get site by unique id: %w", err)
	}
	return &site, nil
}

// RecordHeartbeat inserts a new record into sites_status_heartbeat and handles recovery
func (r *HeartbeatRepository) RecordHeartbeat(ctx context.Context, status *models.HeartbeatStatus) error {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// 1. Insert status record and get ID (needed for incident FK)
	query := `
		INSERT INTO sites_status_heartbeat (heartbeat_site_id, received_at, source_ip, http_method, user_agent, status_code, error)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`
	var statusID int64
	err = tx.QueryRow(ctx, query,
		status.HeartbeatSiteID, status.ReceivedAt, status.SourceIP, status.HTTPMethod,
		status.UserAgent, status.StatusCode, status.Error,
	).Scan(&statusID)

	if err != nil {
		return fmt.Errorf("failed to record heartbeat: %w", err)
	}

	// 2. Mark site as UP (Recovery logic)
	// Only update if it was previously DOWN (site_up = false).
	// This prevents duplicate recovery incidents if the site is already UP.
	updateQuery := `UPDATE sites_list_heartbeat SET site_up = true WHERE id = $1 AND site_up = false`
	cmdTag, err := tx.Exec(ctx, updateQuery, status.HeartbeatSiteID)
	if err != nil {
		return fmt.Errorf("failed to update site status to UP: %w", err)
	}

	// 3. Create Recovery Incident if site was recovered
	if cmdTag.RowsAffected() > 0 {
		incidentQuery := `
			INSERT INTO incidents_heartbeat (heartbeat_site_id, heartbeat_status_id, status, created_at, comments)
			VALUES ($1, $2, true, NOW(), 'Heartbeat recovered')
		`
		_, err = tx.Exec(ctx, incidentQuery, status.HeartbeatSiteID, statusID)
		if err != nil {
			return fmt.Errorf("failed to create recovery incident: %w", err)
		}
	}

	return tx.Commit(ctx)
}

// CheckHeartbeatFailures identifies sites that have missed their heartbeat window
// Logic:
// 1. Get all active heartbeat sites that are NOT in maintenance.
// 2. Join with the *latest* status for each site.
// 3. Check if (latest_received_at + frequency + tolerance) < now
// 4. OR if no status exists and (created_at + frequency + tolerance) < now (edge case for new sites never pinged)
func (r *HeartbeatRepository) CheckHeartbeatFailures(ctx context.Context) error {
	// This function will likely need to be more complex: identifying failures, creating incidents, and updating site status.
	// For efficiency, we can do this in a single query or a transaction.

	// Simplified approach for the MVP implementation:
	// Find sites that are UP but should be DOWN.

	now := time.Now()

	query := `
		SELECT s.id, s.check_frequency, s.tolerance_duration, MAX(st.received_at) as last_received
		FROM sites_list_heartbeat s
		LEFT JOIN sites_status_heartbeat st ON s.id = st.heartbeat_site_id
		WHERE s.site_up = true AND s.in_maintenance = false
		GROUP BY s.id
	`

	rows, err := r.DB.Query(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to query sites for failure check: %w", err)
	}
	defer rows.Close()

	var sitesToDown []string

	for rows.Next() {
		var id string
		var checkFreq, tolerance int
		var lastReceived *time.Time

		if err := rows.Scan(&id, &checkFreq, &tolerance, &lastReceived); err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		thresholdDuration := time.Duration(checkFreq+tolerance) * time.Second

		var lastTime time.Time
		if lastReceived != nil {
			lastTime = *lastReceived
		} else {
			// If never received, we do nothing.
			// Logic: "if never received heartbeat incase of new recorsds do nothing, no incident or no insert into status table"
			continue
		}

		if lastTime.Add(thresholdDuration).Before(now) {
			sitesToDown = append(sitesToDown, id)
		}
	}

	if len(sitesToDown) > 0 {
		return r.markSitesDown(ctx, sitesToDown)
	}

	return nil
}

func (r *HeartbeatRepository) markSitesDown(ctx context.Context, siteIDs []string) error {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, id := range siteIDs {
		// 1. Update site status
		_, err := tx.Exec(ctx, `UPDATE sites_list_heartbeat SET site_up = false, last_incident = NOW() WHERE id = $1`, id)
		if err != nil {
			return err
		}

		// 2. Create incident
		// We need a heartbeat_status_id. If the site is down due to missing heartbeat, we might not have a *new* status ID to link to.
		// The schema requires `heartbeat_status_id` (not null).
		// This implies we might need to insert a "missing" status record or link to the last one.
		// "if latest record... > current time... update status... and add record in incidents_heartbeat"
		// If we don't have a new status (because it's missing), we might need to use the *last* known status or insert a dummy "timed out" status.
		// Let's insert a "Timed Out" status record first.

		var statusID int64
		err = tx.QueryRow(ctx, `
			INSERT INTO sites_status_heartbeat (heartbeat_site_id, error, received_at)
			VALUES ($1, 'Heartbeat missing (Timeout)', NOW())
			RETURNING id
		`, id).Scan(&statusID)

		if err != nil {
			return fmt.Errorf("failed to insert timeout status for site %s: %w", id, err)
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO incidents_heartbeat (heartbeat_site_id, heartbeat_status_id, status, created_at, comments)
			VALUES ($1, $2, true, NOW(), 'Heartbeat missing')
		`, id, statusID)
		if err != nil {
			return fmt.Errorf("failed to create incident for site %s: %w", id, err)
		}
	}

	return tx.Commit(ctx)
}
