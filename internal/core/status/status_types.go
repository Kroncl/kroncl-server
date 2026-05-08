package corestatus

import "time"

type Status string

const (
	StatusOperational   Status = "operational"
	StatusDegraded      Status = "degraded"
	StatusPartialOutage Status = "partial_outage"
	StatusMajorOutage   Status = "major_outage"
)

type IncidentSeverity string

const (
	SeverityMinor IncidentSeverity = "minor"
	SeverityMajor IncidentSeverity = "major"
)

type ComponentType string

const (
	ComponentAll     ComponentType = "all"
	ComponentServer  ComponentType = "server"
	ComponentStorage ComponentType = "storage"
)

type ComponentStatus struct {
	Type  ComponentType `json:"type"`
	Daily []DailyStatus `json:"daily"`
}

type Incident struct {
	ID          string           `json:"id"`
	StartTime   time.Time        `json:"start_time"`
	EndTime     *time.Time       `json:"end_time,omitempty"`
	Severity    IncidentSeverity `json:"severity"`
	Title       string           `json:"title"`
	Description string           `json:"description"`
	MetricsType string           `json:"metrics_type"` // db/server
}

type DailyStatus struct {
	Date      string     `json:"date"`
	Status    Status     `json:"status"`
	Incidents []Incident `json:"incidents,omitempty"`
}

type SystemStatusResponse struct {
	CurrentStatus   Status                          `json:"current_status"`
	Daily           []DailyStatus                   `json:"daily"`
	ActiveIncidents []Incident                      `json:"active_incidents"`
	Components      map[ComponentType][]DailyStatus `json:"components"`
}
