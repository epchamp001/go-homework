package service

import (
	"context"
	"errors"
	"pvz-cli/internal/domain/models"
	repoMock "pvz-cli/internal/usecase/mock"
	txMock "pvz-cli/pkg/txmanager/mock"
	"testing"
	"time"

	"github.com/gojuno/minimock/v3"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
)

func TestServiceImpl_ImportOrders(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	type fields struct {
		tx      *txMock.TxManagerMock
		ordRepo *repoMock.OrdersRepositoryMock
		hrRepo  *repoMock.HistoryAndReturnsRepositoryMock
	}
	type args struct {
		ctx    context.Context
		orders []*models.Order
	}

	now := time.Now()

	tests := []struct {
		name    string
		prepare func(f *fields, a args)
		args    args
		want    int
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:    "EmptyOrdersSlice",
			args:    args{ctx: ctx, orders: []*models.Order{}},
			want:    0,
			wantErr: assert.Error,
		},
		{
			name: "ImportManyFailed",
			prepare: func(f *fields, a args) {
				f.tx.WithTxMock.Set(func(
					txCtx context.Context,
					_ pgx.TxIsoLevel,
					_ pgx.TxAccessMode,
					fn func(context.Context) error,
				) error {
					return fn(txCtx)
				})
				f.ordRepo.ImportManyMock.Set(func(_ context.Context, _ []*models.Order) error {
					return errors.New("import failed")
				})
			},
			args:    args{ctx: ctx, orders: []*models.Order{{ID: "1"}}},
			want:    0,
			wantErr: assert.Error,
		},
		{
			name: "AddHistoryFailed",
			prepare: func(f *fields, a args) {
				f.tx.WithTxMock.Set(func(
					txCtx context.Context,
					_ pgx.TxIsoLevel,
					_ pgx.TxAccessMode,
					fn func(context.Context) error,
				) error {
					return fn(txCtx)
				})
				f.ordRepo.ImportManyMock.Set(func(_ context.Context, _ []*models.Order) error {
					return nil
				})
				f.hrRepo.AddHistoryMock.Set(func(_ context.Context, evt *models.HistoryEvent) error {
					assert.Equal(t, models.StatusAccepted, evt.Status)
					return errors.New("history error")
				})
			},
			args:    args{ctx: ctx, orders: []*models.Order{{ID: "1"}, {ID: "2"}}},
			want:    0,
			wantErr: assert.Error,
		},
		{
			name: "TransactionFailed",
			prepare: func(f *fields, a args) {
				f.tx.WithTxMock.Set(func(
					_ context.Context,
					_ pgx.TxIsoLevel,
					_ pgx.TxAccessMode,
					_ func(context.Context) error,
				) error {
					return errors.New("tx aborted")
				})
			},
			args:    args{ctx: ctx, orders: []*models.Order{{ID: "1"}}},
			want:    0,
			wantErr: assert.Error,
		},
		{
			name: "Success",
			prepare: func(f *fields, a args) {
				f.tx.WithTxMock.Set(func(
					txCtx context.Context,
					level pgx.TxIsoLevel,
					mode pgx.TxAccessMode,
					fn func(context.Context) error,
				) error {
					return fn(txCtx)
				})
				f.ordRepo.ImportManyMock.Set(func(_ context.Context, _ []*models.Order) error {
					return nil
				})
				f.hrRepo.AddHistoryMock.Set(func(_ context.Context, evt *models.HistoryEvent) error {
					assert.Contains(t, []string{"1", "2"}, evt.OrderID)
					assert.Equal(t, models.StatusAccepted, evt.Status)
					assert.WithinDuration(t, now, evt.Time, time.Second)
					return nil
				})
			},
			args:    args{ctx: ctx, orders: []*models.Order{{ID: "1"}, {ID: "2"}}},
			want:    2,
			wantErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := minimock.NewController(t)

			f := &fields{
				tx:      txMock.NewTxManagerMock(ctrl),
				ordRepo: repoMock.NewOrdersRepositoryMock(ctrl),
				hrRepo:  repoMock.NewHistoryAndReturnsRepositoryMock(ctrl),
			}
			service := NewService(f.tx, f.ordRepo, f.hrRepo, nil)

			if tt.prepare != nil {
				tt.prepare(f, tt.args)
			}
			got, err := service.ImportOrders(tt.args.ctx, tt.args.orders)
			tt.wantErr(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
