package internal

import "time"

type Viewing struct {
	ID              int64         `db:"id"               json:"id"`
	AgentID         int64         `db:"agent_id"         json:"agent_id"`
	LeadID          int64         `db:"lead_id"          json:"lead_id"`
	PropertyAddress string        `db:"property_address" json:"property_address"`
	ScheduledAt     time.Time     `db:"scheduled_at"     json:"scheduled_at"`
	Status          ViewingStatus `db:"status"           json:"status"`
	Notes           *string       `db:"notes"            json:"notes"`
	CreatedAt       time.Time     `db:"created_at"       json:"created_at"`
	UpdatedAt       time.Time     `db:"updated_at"       json:"updated_at"`
}

type SortDirection string

const (
	SortAsc  SortDirection = "ASC"
	SortDesc SortDirection = "DESC"
)

type OrderClause struct {
	Field string
	Dir   SortDirection
}

type ListFilter struct {
	AgentID       *int64
	Status        *string
	ScheduledFrom *time.Time
	ScheduledTo   *time.Time
	StartingAfter *Cursor
	Limit         int
	OrderBy       []OrderClause
}
