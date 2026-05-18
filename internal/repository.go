package internal

import "context"

type Repository interface {
	InsertViewing(ctx context.Context, v *Viewing) (int64, error)
	GetViewingByID(ctx context.Context, id int64) (*Viewing, error)
	ListViewings(ctx context.Context, filter ListFilter) ([]Viewing, error)
	BulkUpdateStatus(ctx context.Context, ids []int64, status ViewingStatus) error
	BulkUpdateNotes(ctx context.Context, ids []int64, notes string) error
	MarkMissedViewings(ctx context.Context) (int64, error)
}
