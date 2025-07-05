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

func TestServiceImpl_ReturnOrder(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	type fields struct {
		tx      *txMock.TxManagerMock
		ordRepo *repoMock.OrdersRepositoryMock
		hrRepo  *repoMock.HistoryAndReturnsRepositoryMock
	}
	type args struct {
		ctx     context.Context
		orderID string
	}

	now := time.Now()

	tests := []struct {
		name    string
		prepare func(f *fields, a args)
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:    "EmptyOrderID",
			args:    args{ctx: ctx, orderID: ""},
			wantErr: assert.Error,
		},
		{
			name: "GetNotFound",
			prepare: func(f *fields, a args) {
				f.tx.WithTxMock.Set(func(
					txCtx context.Context,
					_ pgx.TxIsoLevel,
					_ pgx.TxAccessMode,
					fn func(context.Context) error,
				) error {
					return fn(txCtx)
				})
				f.ordRepo.GetMock.Set(func(_ context.Context, id string) (*models.Order, error) {
					return nil, errors.New("not found")
				})
			},
			args:    args{ctx: ctx, orderID: "1"},
			wantErr: assert.Error,
		},
		{
			name: "ValidateReturnFailed",
			prepare: func(f *fields, a args) {
				f.tx.WithTxMock.Set(func(
					txCtx context.Context,
					_ pgx.TxIsoLevel,
					_ pgx.TxAccessMode,
					fn func(context.Context) error,
				) error {
					return fn(txCtx)
				})

				f.ordRepo.GetMock.Set(func(_ context.Context, id string) (*models.Order, error) {
					return &models.Order{
						ID:        id,
						UserID:    "user1",
						Status:    models.StatusIssued,
						CreatedAt: now,
					}, nil
				})
			},
			args:    args{ctx: ctx, orderID: "2"},
			wantErr: assert.Error,
		},
		{
			name: "AddReturnFailed",
			prepare: func(f *fields, a args) {
				f.tx.WithTxMock.Set(func(
					txCtx context.Context,
					_ pgx.TxIsoLevel,
					_ pgx.TxAccessMode,
					fn func(context.Context) error,
				) error {
					return fn(txCtx)
				})

				f.ordRepo.GetMock.Set(func(_ context.Context, id string) (*models.Order, error) {
					return &models.Order{
						ID:        id,
						UserID:    "user1",
						Status:    models.StatusAccepted,
						CreatedAt: now,
					}, nil
				})
				f.hrRepo.AddReturnMock.Set(func(_ context.Context, rec *models.ReturnRecord) error {
					assert.Equal(t, "3", rec.OrderID)
					assert.Equal(t, "user1", rec.UserID)
					return errors.New("add return failed")
				})
			},
			args:    args{ctx: ctx, orderID: "3"},
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
				f.ordRepo.GetMock.Set(func(_ context.Context, id string) (*models.Order, error) {
					return &models.Order{
						ID:        id,
						UserID:    "user1",
						Status:    models.StatusAccepted,
						CreatedAt: now,
					}, nil
				})
				f.hrRepo.AddReturnMock.Set(func(_ context.Context, rec *models.ReturnRecord) error {
					return nil
				})
				f.hrRepo.AddHistoryMock.Set(func(_ context.Context, evt *models.HistoryEvent) error {
					assert.Equal(t, models.StatusReturned, evt.Status)
					return errors.New("add history failed")
				})
			},
			args:    args{ctx: ctx, orderID: "4"},
			wantErr: assert.Error,
		},
		{
			name: "UpdateFailed",
			prepare: func(f *fields, a args) {
				f.tx.WithTxMock.Set(func(
					txCtx context.Context,
					_ pgx.TxIsoLevel,
					_ pgx.TxAccessMode,
					fn func(context.Context) error,
				) error {
					return fn(txCtx)
				})
				f.ordRepo.GetMock.Set(func(_ context.Context, id string) (*models.Order, error) {
					return &models.Order{
						ID:        id,
						UserID:    "user1",
						Status:    models.StatusAccepted,
						CreatedAt: now,
					}, nil
				})
				f.hrRepo.AddReturnMock.Set(func(_ context.Context, rec *models.ReturnRecord) error {
					return nil
				})
				f.hrRepo.AddHistoryMock.Set(func(_ context.Context, evt *models.HistoryEvent) error {
					return nil
				})
				f.ordRepo.UpdateMock.Set(func(_ context.Context, o *models.Order) error {
					assert.Equal(t, models.StatusReturned, o.Status)
					assert.NotNil(t, o.ReturnedAt)
					return errors.New("update failed")
				})
			},
			args:    args{ctx: ctx, orderID: "5"},
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
			args:    args{ctx: ctx, orderID: "6"},
			wantErr: assert.Error,
		},
		{
			name: "Success",
			prepare: func(f *fields, a args) {
				f.tx.WithTxMock.Set(func(
					txCtx context.Context,
					_ pgx.TxIsoLevel,
					_ pgx.TxAccessMode,
					fn func(context.Context) error,
				) error {
					return fn(txCtx)
				})
				f.ordRepo.GetMock.Set(func(_ context.Context, id string) (*models.Order, error) {
					return &models.Order{
						ID:        id,
						UserID:    "user1",
						Status:    models.StatusAccepted,
						CreatedAt: now,
					}, nil
				})
				f.hrRepo.AddReturnMock.Set(func(_ context.Context, rec *models.ReturnRecord) error {
					assert.Equal(t, "7", rec.OrderID)
					return nil
				})
				f.hrRepo.AddHistoryMock.Set(func(_ context.Context, evt *models.HistoryEvent) error {
					assert.Equal(t, models.StatusReturned, evt.Status)
					return nil
				})
				f.ordRepo.UpdateMock.Set(func(_ context.Context, o *models.Order) error {
					assert.Equal(t, models.StatusReturned, o.Status)
					assert.NotNil(t, o.ReturnedAt)
					return nil
				})
			},
			args:    args{ctx: ctx, orderID: "7"},
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

			err := service.ReturnOrder(tt.args.ctx, tt.args.orderID)
			tt.wantErr(t, err)
		})
	}
}
