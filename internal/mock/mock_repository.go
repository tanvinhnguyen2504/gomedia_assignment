package mock

import (
	context "context"
	reflect "reflect"

	internal "property-viewings-service/internal"

	gomock "go.uber.org/mock/gomock"
)

type MockRepository struct {
	ctrl     *gomock.Controller
	recorder *MockRepositoryMockRecorder
}

type MockRepositoryMockRecorder struct {
	mock *MockRepository
}

func NewMockRepository(ctrl *gomock.Controller) *MockRepository {
	m := &MockRepository{ctrl: ctrl}
	m.recorder = &MockRepositoryMockRecorder{m}
	return m
}

func (m *MockRepository) EXPECT() *MockRepositoryMockRecorder {
	return m.recorder
}

func (m *MockRepository) InsertViewing(ctx context.Context, v *internal.Viewing) (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InsertViewing", ctx, v)
	return ret[0].(int64), errOrNil(ret[1])
}

func (mr *MockRepositoryMockRecorder) InsertViewing(ctx, v any) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InsertViewing", reflect.TypeOf((*MockRepository)(nil).InsertViewing), ctx, v)
}

func (m *MockRepository) GetViewingByID(ctx context.Context, id int64) (*internal.Viewing, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetViewingByID", ctx, id)
	res, _ := ret[0].(*internal.Viewing)
	return res, errOrNil(ret[1])
}

func (mr *MockRepositoryMockRecorder) GetViewingByID(ctx, id any) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetViewingByID", reflect.TypeOf((*MockRepository)(nil).GetViewingByID), ctx, id)
}

func (m *MockRepository) ListViewings(ctx context.Context, filter internal.ListFilter) ([]internal.Viewing, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListViewings", ctx, filter)
	res, _ := ret[0].([]internal.Viewing)
	return res, errOrNil(ret[1])
}

func (mr *MockRepositoryMockRecorder) ListViewings(ctx, filter any) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListViewings", reflect.TypeOf((*MockRepository)(nil).ListViewings), ctx, filter)
}

func (m *MockRepository) BulkUpdateStatus(ctx context.Context, ids []int64, status internal.ViewingStatus) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BulkUpdateStatus", ctx, ids, status)
	return errOrNil(ret[0])
}

func (mr *MockRepositoryMockRecorder) BulkUpdateStatus(ctx, ids, status any) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BulkUpdateStatus", reflect.TypeOf((*MockRepository)(nil).BulkUpdateStatus), ctx, ids, status)
}

func (m *MockRepository) BulkUpdateNotes(ctx context.Context, ids []int64, notes string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BulkUpdateNotes", ctx, ids, notes)
	return errOrNil(ret[0])
}

func (mr *MockRepositoryMockRecorder) BulkUpdateNotes(ctx, ids, notes any) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BulkUpdateNotes", reflect.TypeOf((*MockRepository)(nil).BulkUpdateNotes), ctx, ids, notes)
}

func (m *MockRepository) MarkMissedViewings(ctx context.Context) (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MarkMissedViewings", ctx)
	return ret[0].(int64), errOrNil(ret[1])
}

func (mr *MockRepositoryMockRecorder) MarkMissedViewings(ctx any) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MarkMissedViewings", reflect.TypeOf((*MockRepository)(nil).MarkMissedViewings), ctx)
}

func errOrNil(v any) error {
	if v == nil {
		return nil
	}
	return v.(error)
}
