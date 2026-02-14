package models

import (
	"time"
)

// HeartbeatSite represents a row in the sites_list_heartbeat table
type HeartbeatSite struct {
	ID                string     `db:"id"`
	Name              *string    `db:"name"`
	Type              string     `db:"type"`
	UniqueID          string     `db:"unique_id"`
	CreatedAt         time.Time  `db:"created_at"`
	UpdatedAt         time.Time  `db:"updated_at"`
	CheckFrequency    int        `db:"check_frequency"`
	ToleranceDuration int        `db:"tolerance_duration"`
	SiteUp            bool       `db:"site_up"`
	InMaintenance     bool       `db:"in_maintenance"`
	LastIncident      *time.Time `db:"last_incident"`
	FailureCount      int        `db:"failure_count"`
	OrgID             *string    `db:"org_id"`
}

// HeartbeatStatus represents a row in the sites_status_heartbeat table
type HeartbeatStatus struct {
	ID              int64      `db:"id"`
	HeartbeatSiteID string     `db:"heartbeat_site_id"`
	ReceivedAt      *time.Time `db:"received_at"`
	SourceIP        *string    `db:"source_ip"`
	HTTPMethod      *string    `db:"http_method"`
	UserAgent       *string    `db:"user_agent"`
	StatusCode      *int64     `db:"status_code"`
	Error           *string    `db:"error"`
}

// HeartbeatIncident represents a row in the incidents_heartbeat table
type HeartbeatIncident struct {
	ID                int64     `db:"id"`
	CreatedAt         time.Time `db:"created_at"`
	Status            *bool     `db:"status"`
	HeartbeatStatusID int64     `db:"heartbeat_status_id"`
	HeartbeatSiteID   string    `db:"heartbeat_site_id"`
	Comments          *string   `db:"comments"`
}
