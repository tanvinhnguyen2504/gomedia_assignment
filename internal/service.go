package internal

import (
	"context"
	"database/sql"
	"errors"
	"log"
)

type Service interface {
	CreateViewing(ctx context.Context, req CreateViewingRequest) (int64, error)
	GetViewing(ctx context.Context, id int64) (*Viewing, error)
	ListViewings(ctx context.Context, req ListViewingsRequest) ([]Viewing, bool, *int64, error)
	BulkUpdate(ctx context.Context, req BulkUpdateRequest) error
	MarkMissedViewings(ctx context.Context)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) CreateViewing(ctx context.Context, req CreateViewingRequest) (int64, error) {
	if err := req.Validate(); err != nil {
		return 0, err
	}

	v := &Viewing{
		AgentID:         req.AgentID,
		LeadID:          req.LeadID,
		PropertyAddress: req.PropertyAddress,
		ScheduledAt:     req.ScheduledAt,
		Notes:           req.Notes,
	}
	return s.repo.InsertViewing(ctx, v)
}

func (s *service) GetViewing(ctx context.Context, id int64) (*Viewing, error) {
	v, err := s.repo.GetViewingByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return v, nil
}

func (s *service) ListViewings(ctx context.Context, req ListViewingsRequest) ([]Viewing, bool, *int64, error) {
	limit := DEFAULT_LIMIT
	if req.Limit != nil {
		limit = *req.Limit
	}
	if limit > MAX_LIMIT {
		limit = MAX_LIMIT
	}

	filter := ListFilter{
		AgentID:       req.AgentID,
		Status:        req.Status,
		ScheduledFrom: req.ScheduledFrom,
		ScheduledTo:   req.ScheduledTo,
		StartingAfter: req.StartingAfter,
		Limit:         limit + 1,
		OrderBy:       req.OrderBy,
	}

	rows, err := s.repo.ListViewings(ctx, filter)
	if err != nil {
		return nil, false, nil, err
	}

	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}

	var nextCursor *int64
	if hasMore && len(rows) > 0 {
		last := rows[len(rows)-1].ID
		nextCursor = &last
	}

	return rows, hasMore, nextCursor, nil
}

func (s *service) BulkUpdate(ctx context.Context, req BulkUpdateRequest) error {
	if err := req.Validate(); err != nil {
		return err
	}

	switch req.Action {
	case ActionCancel:
		return s.repo.BulkUpdateStatus(ctx, req.IDs, StatusCancelled)
	case ActionComplete:
		return s.repo.BulkUpdateStatus(ctx, req.IDs, StatusCompleted)
	case ActionUpdateNotes:
		return s.repo.BulkUpdateNotes(ctx, req.IDs, *req.Notes)
	default:
		return nil
	}
}

func (s *service) MarkMissedViewings(ctx context.Context) {
	count, err := s.repo.MarkMissedViewings(ctx)
	if err != nil {
		log.Printf("MarkMissedViewings error: %v", err)
		return
	}
	log.Printf("MarkMissedViewings: marked %d viewings as MISSED", count)
}
