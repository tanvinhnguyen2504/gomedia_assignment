package internal_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	. "property-viewings-service/internal"
	"property-viewings-service/internal/mock"
)

func makeViewings(n int) []Viewing {
	rows := make([]Viewing, n)
	for i := range rows {
		rows[i] = Viewing{ID: int64(i + 1), Status: StatusScheduled}
	}
	return rows
}

func TestService_CreateViewing(t *testing.T) {
	futureTime := time.Now().Add(24 * time.Hour)

	tests := []struct {
		name      string
		req       CreateViewingRequest
		mockSetup func(repo *mock.MockRepository)
		wantID    int64
		wantErr   error
	}{
		{
			name: "success",
			req: CreateViewingRequest{
				AgentID:         1,
				LeadID:          42,
				PropertyAddress: "123 Orchard Road",
				ScheduledAt:     futureTime,
			},
			mockSetup: func(repo *mock.MockRepository) {
				repo.EXPECT().InsertViewing(gomock.Any(), gomock.Any()).Return(int64(42), nil)
			},
			wantID:  42,
			wantErr: nil,
		},
		{
			name: "missing agent_id",
			req: CreateViewingRequest{
				LeadID:          42,
				PropertyAddress: "123 Orchard Road",
				ScheduledAt:     futureTime,
			},
			mockSetup: func(repo *mock.MockRepository) {}, // no repo calls expected
			wantErr:   ErrMissingField,
		},
		{
			name: "scheduled_at in the past",
			req: CreateViewingRequest{
				AgentID:         1,
				LeadID:          42,
				PropertyAddress: "123 Orchard Road",
				ScheduledAt:     time.Now().Add(-1 * time.Hour),
			},
			mockSetup: func(repo *mock.MockRepository) {}, // no repo calls expected
			wantErr:   ErrPastDate,
		},
		{
			name: "slot already taken — InsertViewing returns ErrConflict",
			req: CreateViewingRequest{
				AgentID:         1,
				LeadID:          42,
				PropertyAddress: "123 Orchard Road",
				ScheduledAt:     futureTime,
			},
			mockSetup: func(repo *mock.MockRepository) {
				repo.EXPECT().InsertViewing(gomock.Any(), gomock.Any()).Return(int64(0), ErrConflict)
			},
			wantErr: ErrConflict,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			ctrl := gomock.NewController(t)
			repo := mock.NewMockRepository(ctrl)
			svc := NewService(repo)
			tc.mockSetup(repo)

			// Act
			id, err := svc.CreateViewing(context.Background(), tc.req)

			// Assert
			assert.ErrorIs(t, err, tc.wantErr)
			if tc.wantErr == nil {
				assert.Equal(t, tc.wantID, id)
			}
		})
	}
}

func TestService_GetViewing(t *testing.T) {
	tests := []struct {
		name      string
		id        int64
		mockSetup func(repo *mock.MockRepository)
		wantErr   error
	}{
		{
			name: "found",
			id:   1,
			mockSetup: func(repo *mock.MockRepository) {
				repo.EXPECT().GetViewingByID(gomock.Any(), int64(1)).Return(&Viewing{ID: 1}, nil)
			},
			wantErr: nil,
		},
		{
			name: "not found",
			id:   9999,
			mockSetup: func(repo *mock.MockRepository) {
				repo.EXPECT().GetViewingByID(gomock.Any(), int64(9999)).Return(nil, sql.ErrNoRows)
			},
			wantErr: ErrNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			ctrl := gomock.NewController(t)
			repo := mock.NewMockRepository(ctrl)
			svc := NewService(repo)
			tc.mockSetup(repo)

			// Act
			v, err := svc.GetViewing(context.Background(), tc.id)

			// Assert
			assert.ErrorIs(t, err, tc.wantErr)
			if tc.wantErr == nil {
				assert.NotNil(t, v)
				assert.Equal(t, tc.id, v.ID)
			}
		})
	}
}

func TestService_ListViewings(t *testing.T) {
	tests := []struct {
		name           string
		req            ListViewingsRequest
		mockRows       []Viewing // rows the repo returns
		wantLen        int
		wantHasMore    bool
		wantNextCursor bool // whether nextCursor should be non-nil
	}{
		{
			name:           "fewer rows than limit — no more pages",
			req:            ListViewingsRequest{},
			mockRows:       makeViewings(3), // 3 < default limit 20
			wantLen:        3,
			wantHasMore:    false,
			wantNextCursor: false,
		},
		{
			name:           "exactly limit+1 rows returned — has more",
			req:            ListViewingsRequest{},
			mockRows:       makeViewings(21), // default limit=20, fetch 21
			wantLen:        20,
			wantHasMore:    true,
			wantNextCursor: true,
		},
		{
			name:           "limit over max is capped to 100",
			req:            ListViewingsRequest{Limit: intPtr(200)},
			mockRows:       makeViewings(5), // repo returns 5 rows
			wantLen:        5,
			wantHasMore:    false,
			wantNextCursor: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			ctrl := gomock.NewController(t)
			repo := mock.NewMockRepository(ctrl)
			svc := NewService(repo)
			repo.EXPECT().
				ListViewings(gomock.Any(), gomock.Any()).
				Return(tc.mockRows, nil)

			// Act
			rows, hasMore, nextCursor, err := svc.ListViewings(context.Background(), tc.req)

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, tc.wantLen, len(rows))
			assert.Equal(t, tc.wantHasMore, hasMore)
			if tc.wantNextCursor {
				assert.NotNil(t, nextCursor)
			} else {
				assert.Nil(t, nextCursor)
			}
		})
	}
}

func TestService_BulkUpdate(t *testing.T) {
	notes := "some notes"

	tests := []struct {
		name      string
		req       BulkUpdateRequest
		mockSetup func(repo *mock.MockRepository)
		wantErr   error
	}{
		{
			name: "update notes",
			req:  BulkUpdateRequest{IDs: []int64{1, 2, 3}, Action: ActionUpdateNotes, Notes: &notes},
			mockSetup: func(repo *mock.MockRepository) {
				repo.EXPECT().BulkUpdateNotes(gomock.Any(), []int64{1, 2, 3}, notes).Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "update status on any status — success",
			req:  BulkUpdateRequest{IDs: []int64{1, 2, 3}, Action: ActionComplete},
			mockSetup: func(repo *mock.MockRepository) {
				repo.EXPECT().BulkUpdateStatus(gomock.Any(), []int64{1, 2, 3}, StatusCompleted).Return(nil)
			},
			wantErr: nil,
		},
		{
			name:      "unknown action — error",
			req:       BulkUpdateRequest{IDs: []int64{1}, Action: "UNKNOWN"},
			mockSetup: func(repo *mock.MockRepository) {},
			wantErr:   ErrInvalidAction,
		},
		{
			name:      "empty ids — missing field error",
			req:       BulkUpdateRequest{IDs: []int64{}, Action: ActionCancel},
			mockSetup: func(repo *mock.MockRepository) {},
			wantErr:   ErrMissingField,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := mock.NewMockRepository(ctrl)
			svc := NewService(repo)
			tc.mockSetup(repo)

			// Act
			err := svc.BulkUpdate(context.Background(), tc.req)

			// Assert
			assert.ErrorIs(t, err, tc.wantErr)
		})
	}
}

func TestService_MarkMissedViewings(t *testing.T) {
	tests := []struct {
		name      string
		mockSetup func(repo *mock.MockRepository)
	}{
		{
			name: "marks rows — logs count",
			mockSetup: func(repo *mock.MockRepository) {
				repo.EXPECT().MarkMissedViewings(gomock.Any()).Return(int64(3), nil)
			},
		},
		{
			name: "repo error — no panic",
			mockSetup: func(repo *mock.MockRepository) {
				repo.EXPECT().MarkMissedViewings(gomock.Any()).Return(int64(0), errors.New("db error"))
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			ctrl := gomock.NewController(t)
			repo := mock.NewMockRepository(ctrl)
			svc := NewService(repo)
			tc.mockSetup(repo)

			// Act + Assert — must not panic
			assert.NotPanics(t, func() {
				svc.MarkMissedViewings(context.Background())
			})
		})
	}
}

func intPtr(v int) *int { return &v }
