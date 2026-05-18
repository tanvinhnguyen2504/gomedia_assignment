package internal

import "time"

type Cursor struct {
	ScheduledAt time.Time `json:"scheduled_at"`
	ID          int64     `json:"id"`
}

type CreateViewingRequest struct {
	AgentID         int64     `json:"agent_id"`
	LeadID          int64     `json:"lead_id"`
	PropertyAddress string    `json:"property_address"`
	ScheduledAt     time.Time `json:"scheduled_at"`
	Notes           *string   `json:"notes"`
}

type ListViewingsRequest struct {
	AgentID       *int64        `json:"agent_id"`
	Status        *string       `json:"status"`
	ScheduledFrom *time.Time    `json:"scheduled_from"`
	ScheduledTo   *time.Time    `json:"scheduled_to"`
	StartingAfter *Cursor       `json:"starting_after"`
	Limit         *int          `json:"limit"`
	OrderBy       []OrderClause `json:"order_by"`
}

type BulkUpdateRequest struct {
	IDs    []int64       `json:"ids"`
	Action ViewingAction `json:"action"`
	Notes  *string       `json:"notes"`
}

type CreateViewingResponse struct {
	Data struct {
		ID int64 `json:"id"`
	} `json:"data"`
}

type ListViewingsResponse struct {
	Data       []Viewing `json:"data"`
	HasMore    bool      `json:"has_more"`
	NextCursor *Cursor   `json:"next_cursor"`
}
