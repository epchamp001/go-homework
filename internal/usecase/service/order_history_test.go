package service

import (
	"context"
	"errors"
	"pvz-cli/internal/domain/models"
	"pvz-cli/internal/domain/vo"
	repoMock "pvz-cli/internal/usecase/mock"
	txMock "pvz-cli/pkg/txmanager/mock"
	"testing"
	"time"

	"github.com/gojuno/minimock/v3"
	"github.com/stretchr/testify/assert"
)

func TestServiceImpl_OrderHistory(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	type fields struct {
		tx     *txMock.TxManagerMock
		hrRepo *repoMock.HistoryAndReturnsRepositoryMock
	}
	type args struct {
		ctx context.Context
		pg  vo.Pagination
	}

	now := time.Now()
	e1 := &models.HistoryEvent{OrderID: "1", Status: models.StatusAccepted, Time: now.Add(-time.Minute)}
	e2 := &models.HistoryEvent{OrderID: "2", Status: models.StatusIssued, Time: now}
	events := []*models.HistoryEvent{e1, e2}

	tests := []struct {
		name       string
		prepare    func(f *fields, a args)
		args       args
		wantErr    assert.ErrorAssertionFunc
		wantEvents []*models.HistoryEvent
		wantCount  int
	}{
		{
			name: "RepoError",
			prepare: func(f *fields, a args) {
				f.tx.WithReadOnlyMock.Set(func(in context.Context) context.Context {
					return in
				})
				f.hrRepo.HistoryMock.Set(func(roCtx context.Context, pg vo.Pagination) ([]*models.HistoryEvent, error) {
					assert.Equal(t, a.pg, pg)
					return nil, errors.New("db error")
				})
			},
			args:       args{ctx: ctx, pg: vo.Pagination{Page: 1, Limit: 10}},
			wantErr:    assert.Error,
			wantEvents: nil,
			wantCount:  0,
		},
		{
			name: "Success",
			prepare: func(f *fields, a args) {
				f.tx.WithReadOnlyMock.Set(func(in context.Context) context.Context {
					return in
				})
				f.hrRepo.HistoryMock.Set(func(roCtx context.Context, pg vo.Pagination) ([]*models.HistoryEvent, error) {
					assert.Equal(t, a.pg, pg)
					return events, nil
				})
			},
			args:       args{ctx: ctx, pg: vo.Pagination{Page: 2, Limit: 5}},
			wantErr:    assert.NoError,
			wantEvents: events,
			wantCount:  len(events),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := minimock.NewController(t)
			f := &fields{
				tx:     txMock.NewTxManagerMock(ctrl),
				hrRepo: repoMock.NewHistoryAndReturnsRepositoryMock(ctrl),
			}
			service := NewService(f.tx, nil, f.hrRepo, nil, nil)

			if tt.prepare != nil {
				tt.prepare(f, tt.args)
			}

			gotEvents, gotCount, err := service.OrderHistory(tt.args.ctx, tt.args.pg)
			tt.wantErr(t, err)
			assert.Equal(t, tt.wantEvents, gotEvents)
			assert.Equal(t, tt.wantCount, gotCount)
		})
	}
}
