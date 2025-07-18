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

func TestServiceImpl_ListReturns(t *testing.T) {
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
	r1 := &models.ReturnRecord{OrderID: "a", UserID: "user1", ReturnedAt: now.Add(-time.Hour)}
	r2 := &models.ReturnRecord{OrderID: "b", UserID: "user2", ReturnedAt: now}
	records := []*models.ReturnRecord{r1, r2}

	tests := []struct {
		name        string
		prepare     func(f *fields, a args)
		args        args
		wantErr     assert.ErrorAssertionFunc
		wantRecords []*models.ReturnRecord
	}{
		{
			name: "RepoError",
			prepare: func(f *fields, a args) {
				f.tx.WithReadOnlyMock.Set(func(in context.Context) context.Context {
					return in
				})
				f.hrRepo.ListReturnsMock.Set(func(roCtx context.Context, pg vo.Pagination) ([]*models.ReturnRecord, error) {
					assert.Equal(t, a.pg, pg)
					return nil, errors.New("db error")
				})
			},
			args:        args{ctx: ctx, pg: vo.Pagination{Page: 1, Limit: 10}},
			wantErr:     assert.Error,
			wantRecords: nil,
		},
		{
			name: "Success",
			prepare: func(f *fields, a args) {
				f.tx.WithReadOnlyMock.Set(func(in context.Context) context.Context {
					return in
				})
				f.hrRepo.ListReturnsMock.Set(func(roCtx context.Context, pg vo.Pagination) ([]*models.ReturnRecord, error) {
					assert.Equal(t, a.pg, pg)
					return records, nil
				})
			},
			args:        args{ctx: ctx, pg: vo.Pagination{Page: 2, Limit: 5}},
			wantErr:     assert.NoError,
			wantRecords: records,
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
			service := NewService(f.tx, nil, f.hrRepo, nil, nil, nil, nil)

			if tt.prepare != nil {
				tt.prepare(f, tt.args)
			}

			got, err := service.ListReturns(tt.args.ctx, tt.args.pg)
			tt.wantErr(t, err)
			assert.Equal(t, tt.wantRecords, got)
		})
	}
}
