package service

import (
	"context"
	"errors"
	"github.com/gojuno/minimock/v3"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"pvz-cli/internal/domain/models"
	repoMock "pvz-cli/internal/usecase/mock"
	txMock "pvz-cli/pkg/txmanager/mock"
	"testing"
	"time"
)

func TestServiceImpl_IssueOrders(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	type fields struct {
		tx      *txMock.TxManagerMock
		ordRepo *repoMock.OrdersRepositoryMock
		hrRepo  *repoMock.HistoryAndReturnsRepositoryMock
	}
	type args struct {
		ctx    context.Context
		userID string
		ids    []string
	}

	now := time.Now()
	validExpiry := now.Add(time.Hour)

	tests := []struct {
		name        string
		prepare     func(f *fields, a args)
		args        args
		wantErr     assert.ErrorAssertionFunc
		wantResults map[string]assert.ErrorAssertionFunc
	}{
		{
			name: "GetOrderNotFound",
			prepare: func(f *fields, a args) {
				f.tx.WithTxMock.Set(func(
					txCtx context.Context,
					_ pgx.TxIsoLevel,
					_ pgx.TxAccessMode,
					fn func(context.Context) error,
				) error {
					return fn(txCtx)
				})
				f.ordRepo.GetMock.Set(func(_ context.Context, orderID string) (*models.Order, error) {
					return nil, errors.New("not found")
				})
			},
			args:    args{ctx: ctx, userID: "user1", ids: []string{"1"}},
			wantErr: assert.NoError,
			wantResults: map[string]assert.ErrorAssertionFunc{
				"1": assert.Error, // wrapped "order not found"
			},
		},
		{
			name: "ValidationFailed_WrongUser",
			prepare: func(f *fields, a args) {
				f.tx.WithTxMock.Set(func(
					txCtx context.Context,
					_ pgx.TxIsoLevel,
					_ pgx.TxAccessMode,
					fn func(context.Context) error,
				) error {
					return fn(txCtx)
				})
				f.ordRepo.GetMock.Set(func(_ context.Context, orderID string) (*models.Order, error) {
					return &models.Order{
						ID:        orderID,
						UserID:    "otherUser",
						Status:    models.StatusAccepted,
						ExpiresAt: validExpiry,
					}, nil
				})
			},
			args:    args{ctx: ctx, userID: "user1", ids: []string{"2"}},
			wantErr: assert.NoError,
			wantResults: map[string]assert.ErrorAssertionFunc{
				"2": assert.Error,
			},
		},
		{
			name: "ValidationFailed_Expired",
			prepare: func(f *fields, a args) {
				f.tx.WithTxMock.Set(func(
					txCtx context.Context,
					_ pgx.TxIsoLevel,
					_ pgx.TxAccessMode,
					fn func(context.Context) error,
				) error {
					return fn(txCtx)
				})
				f.ordRepo.GetMock.Set(func(_ context.Context, orderID string) (*models.Order, error) {
					return &models.Order{
						ID:        orderID,
						UserID:    a.userID,
						Status:    models.StatusAccepted,
						ExpiresAt: now.Add(-time.Minute),
					}, nil
				})
			},
			args:    args{ctx: ctx, userID: "user1", ids: []string{"3"}},
			wantErr: assert.NoError,
			wantResults: map[string]assert.ErrorAssertionFunc{
				"3": assert.Error,
			},
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
						UserID:    a.userID,
						Status:    models.StatusAccepted,
						ExpiresAt: validExpiry,
					}, nil
				})
				f.ordRepo.UpdateMock.Set(func(_ context.Context, _ *models.Order) error {
					return errors.New("update failed")
				})
			},
			args:    args{ctx: ctx, userID: "user1", ids: []string{"4"}},
			wantErr: assert.NoError,
			wantResults: map[string]assert.ErrorAssertionFunc{
				"4": assert.Error,
			},
		},
		{
			name: "HistoryFailed",
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
						UserID:    a.userID,
						Status:    models.StatusAccepted,
						ExpiresAt: validExpiry,
					}, nil
				})
				f.ordRepo.UpdateMock.Set(func(_ context.Context, _ *models.Order) error {
					return nil
				})
				f.hrRepo.AddHistoryMock.Set(func(_ context.Context, evt *models.HistoryEvent) error {
					return errors.New("history error")
				})
			},
			args:    args{ctx: ctx, userID: "user1", ids: []string{"5"}},
			wantErr: assert.NoError,
			wantResults: map[string]assert.ErrorAssertionFunc{
				"5": assert.Error,
			},
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
			args:        args{ctx: ctx, userID: "user1", ids: []string{"6"}},
			wantErr:     assert.Error,
			wantResults: nil,
		},
		{
			name: "SuccessMultiple",
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
						UserID:    a.userID,
						Status:    models.StatusAccepted,
						ExpiresAt: validExpiry,
					}, nil
				})
				f.ordRepo.UpdateMock.Set(func(_ context.Context, _ *models.Order) error {
					return nil
				})
				f.hrRepo.AddHistoryMock.Set(func(_ context.Context, evt *models.HistoryEvent) error {
					assert.Equal(t, models.StatusIssued, evt.Status)
					assert.WithinDuration(t, now, evt.Time, time.Second)
					return nil
				})
			},
			args:    args{ctx: ctx, userID: "user1", ids: []string{"7", "8"}},
			wantErr: assert.NoError,
			wantResults: map[string]assert.ErrorAssertionFunc{
				"7": assert.NoError,
				"8": assert.NoError,
			},
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

			res, err := service.IssueOrders(tt.args.ctx, tt.args.userID, tt.args.ids)
			tt.wantErr(t, err)

			if tt.wantResults == nil {
				assert.Nil(t, res)
			} else {
				assert.Len(t, res, len(tt.args.ids))
				for id, assertFn := range tt.wantResults {
					assertFn(t, res[id], "order %s", id)
				}
			}
		})
	}
}
