package service

import (
	"context"
	"errors"
	"pvz-cli/internal/domain/models"
	repoMock "pvz-cli/internal/usecase/mock"
	"pvz-cli/pkg/logger"
	txMock "pvz-cli/pkg/txmanager/mock"
	"pvz-cli/pkg/wpool"
	"testing"
	"time"

	"github.com/gojuno/minimock/v3"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
)

func TestServiceImpl_ReturnOrdersByClient(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	type fields struct {
		tx         *txMock.TxManagerMock
		ordRepo    *repoMock.OrdersRepositoryMock
		hrRepo     *repoMock.HistoryAndReturnsRepositoryMock
		outboxRepo *repoMock.OutboxRepositoryMock
	}
	type args struct {
		ctx    context.Context
		userID string
		ids    []string
	}

	now := time.Now()

	tests := []struct {
		name        string
		prepare     func(f *fields, a args)
		args        args
		wantErr     assert.ErrorAssertionFunc
		wantResults map[string]assert.ErrorAssertionFunc
	}{
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
			args:    args{ctx, "user1", []string{"1"}},
			wantErr: assert.NoError,
			wantResults: map[string]assert.ErrorAssertionFunc{
				"1": assert.Error,
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
				f.ordRepo.GetMock.Set(func(_ context.Context, id string) (*models.Order, error) {
					return &models.Order{
						ID:       id,
						UserID:   "otherUser",
						Status:   models.StatusIssued,
						IssuedAt: &now,
					}, nil
				})
			},
			args:    args{ctx, "user1", []string{"2"}},
			wantErr: assert.NoError,
			wantResults: map[string]assert.ErrorAssertionFunc{
				"2": assert.Error,
			},
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
						ID:       id,
						UserID:   a.userID,
						Status:   models.StatusIssued,
						IssuedAt: &now,
					}, nil
				})
				f.hrRepo.AddReturnMock.Set(func(_ context.Context, rec *models.ReturnRecord) error {
					assert.Equal(t, "3", rec.OrderID)
					return errors.New("add return failed")
				})
			},
			args:    args{ctx, "user1", []string{"3"}},
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
						ID:       id,
						UserID:   a.userID,
						Status:   models.StatusIssued,
						IssuedAt: &now,
					}, nil
				})
				f.hrRepo.AddReturnMock.Set(func(_ context.Context, rec *models.ReturnRecord) error {
					return nil
				})
				f.ordRepo.UpdateMock.Set(func(_ context.Context, o *models.Order) error {
					assert.Equal(t, models.StatusReturned, o.Status)
					return errors.New("update failed")
				})
			},
			args:    args{ctx, "user1", []string{"4"}},
			wantErr: assert.NoError,
			wantResults: map[string]assert.ErrorAssertionFunc{
				"4": assert.Error,
			},
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
						ID:       id,
						UserID:   a.userID,
						Status:   models.StatusIssued,
						IssuedAt: &now,
					}, nil
				})
				f.hrRepo.AddReturnMock.Set(func(_ context.Context, rec *models.ReturnRecord) error {
					return nil
				})
				f.ordRepo.UpdateMock.Set(func(_ context.Context, o *models.Order) error {
					return nil
				})
				f.hrRepo.AddHistoryMock.Set(func(_ context.Context, evt *models.HistoryEvent) error {
					assert.Equal(t, models.StatusReturned, evt.Status)
					return errors.New("add history failed")
				})
			},
			args:    args{ctx, "user1", []string{"5"}},
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
			args:    args{ctx, "user1", []string{"6"}},
			wantErr: assert.Error,
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
						ID:       id,
						UserID:   a.userID,
						Status:   models.StatusIssued,
						IssuedAt: &now,
					}, nil
				})
				f.hrRepo.AddReturnMock.Set(func(_ context.Context, rec *models.ReturnRecord) error {
					return nil
				})
				f.ordRepo.UpdateMock.Set(func(_ context.Context, o *models.Order) error {
					return nil
				})
				f.hrRepo.AddHistoryMock.Set(func(_ context.Context, evt *models.HistoryEvent) error {
					assert.Equal(t, models.StatusReturned, evt.Status)
					return nil
				})
				f.outboxRepo.AddMock.Set(func(_ context.Context, evt *models.OrderEvent) error {
					assert.Equal(t, models.OrderReturnedByClient, evt.EventType)
					assert.Contains(t, []string{"7", "8"}, evt.Order.ID)
					return nil
				})
			},
			args:    args{ctx, "user1", []string{"7", "8"}},
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
			defer ctrl.Finish()

			f := &fields{
				tx:         txMock.NewTxManagerMock(ctrl),
				ordRepo:    repoMock.NewOrdersRepositoryMock(ctrl),
				hrRepo:     repoMock.NewHistoryAndReturnsRepositoryMock(ctrl),
				outboxRepo: repoMock.NewOutboxRepositoryMock(ctrl),
			}

			if tt.prepare != nil {
				tt.prepare(f, tt.args)
			}

			log, _ := logger.NewLogger(logger.WithMode("prod"))
			wp := wpool.NewWorkerPool(4, 16, log)
			defer wp.Stop()
			svc := NewService(f.tx, f.ordRepo, f.hrRepo, f.outboxRepo, nil, wp)

			res, err := svc.ReturnOrdersByClient(tt.args.ctx, tt.args.userID, tt.args.ids)
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

func TestServiceImpl_returnOneByClient(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Now()

	type fields struct {
		tx         *txMock.TxManagerMock
		ordRepo    *repoMock.OrdersRepositoryMock
		hrRepo     *repoMock.HistoryAndReturnsRepositoryMock
		outboxRepo *repoMock.OutboxRepositoryMock
	}
	type args struct {
		ctx     context.Context
		orderID string
		userID  string
		now     time.Time
	}

	tests := []struct {
		name       string
		prepare    func(f *fields, a args)
		args       args
		wantBisErr assert.ErrorAssertionFunc
		wantTxErr  assert.ErrorAssertionFunc
	}{
		{
			name: "OrderNotFound",
			prepare: func(f *fields, a args) {
				f.tx.WithTxMock.Set(pass)
				f.ordRepo.GetMock.Set(func(_ context.Context, id string) (*models.Order, error) {
					return nil, errors.New("not found")
				})
			},
			args:       args{ctx, "1", "user1", now},
			wantBisErr: assert.Error,
			wantTxErr:  assert.NoError,
		},
		{
			name: "WrongUser",
			prepare: func(f *fields, a args) {
				f.tx.WithTxMock.Set(pass)
				f.ordRepo.GetMock.Set(func(_ context.Context, id string) (*models.Order, error) {
					return &models.Order{
						ID:       id,
						UserID:   "other",
						Status:   models.StatusIssued,
						IssuedAt: &now,
					}, nil
				})
			},
			args:       args{ctx, "2", "user1", now},
			wantBisErr: assert.Error,
			wantTxErr:  assert.NoError,
		},
		{
			name: "AddReturnFailed",
			prepare: func(f *fields, a args) {
				f.tx.WithTxMock.Set(pass)
				f.ordRepo.GetMock.Set(func(_ context.Context, id string) (*models.Order, error) {
					return &models.Order{
						ID:       id,
						UserID:   a.userID,
						Status:   models.StatusIssued,
						IssuedAt: &now,
					}, nil
				})

				f.hrRepo.AddReturnMock.Set(func(_ context.Context, rec *models.ReturnRecord) error {
					assert.Equal(t, a.orderID, rec.OrderID)
					return errors.New("return failed")
				})
			},
			args:       args{ctx, "3", "user1", now},
			wantBisErr: assert.Error,
			wantTxErr:  assert.NoError,
		},
		{
			name: "UpdateFailed",
			prepare: func(f *fields, a args) {
				f.tx.WithTxMock.Set(pass)
				f.ordRepo.GetMock.Set(func(_ context.Context, id string) (*models.Order, error) {
					return &models.Order{
						ID:       id,
						UserID:   a.userID,
						Status:   models.StatusIssued,
						IssuedAt: &now,
					}, nil
				})
				f.hrRepo.AddReturnMock.Set(func(_ context.Context, rec *models.ReturnRecord) error {
					return nil
				})

				f.ordRepo.UpdateMock.Set(func(_ context.Context, o *models.Order) error {
					assert.Equal(t, models.StatusReturned, o.Status)
					return errors.New("update failed")
				})
			},
			args:       args{ctx, "4", "user1", now},
			wantBisErr: assert.Error,
			wantTxErr:  assert.NoError,
		},
		{
			name: "AddHistoryFailed",
			prepare: func(f *fields, a args) {
				f.tx.WithTxMock.Set(pass)
				f.ordRepo.GetMock.Set(func(_ context.Context, id string) (*models.Order, error) {
					return &models.Order{
						ID:       id,
						UserID:   a.userID,
						Status:   models.StatusIssued,
						IssuedAt: &now,
					}, nil
				})
				f.hrRepo.AddReturnMock.Set(func(_ context.Context, rec *models.ReturnRecord) error {
					return nil
				})
				f.ordRepo.UpdateMock.Set(func(_ context.Context, o *models.Order) error {
					return nil
				})

				f.hrRepo.AddHistoryMock.Set(func(_ context.Context, evt *models.HistoryEvent) error {
					assert.Equal(t, models.StatusReturned, evt.Status)
					return errors.New("history failed")
				})
			},
			args:       args{ctx, "5", "user1", now},
			wantBisErr: assert.Error,
			wantTxErr:  assert.NoError,
		},
		{
			name: "OutboxFailed",
			prepare: func(f *fields, a args) {
				f.tx.WithTxMock.Set(pass)
				f.ordRepo.GetMock.Set(func(_ context.Context, id string) (*models.Order, error) {
					return &models.Order{
						ID:       id,
						UserID:   a.userID,
						Status:   models.StatusIssued,
						IssuedAt: &now,
					}, nil
				})
				f.hrRepo.AddReturnMock.Set(func(_ context.Context, rec *models.ReturnRecord) error {
					return nil
				})
				f.ordRepo.UpdateMock.Set(func(_ context.Context, o *models.Order) error {
					return nil
				})
				f.hrRepo.AddHistoryMock.Set(func(_ context.Context, evt *models.HistoryEvent) error {
					return nil
				})

				f.outboxRepo.AddMock.Set(func(_ context.Context, evt *models.OrderEvent) error {
					assert.Equal(t, models.OrderReturnedByClient, evt.EventType)
					return errors.New("outbox failed")
				})
			},
			args:       args{ctx, "6", "user1", now},
			wantBisErr: assert.Error,
			wantTxErr:  assert.NoError,
		},
		{
			name: "TxAborted",
			prepare: func(f *fields, a args) {

				f.tx.WithTxMock.Set(func(
					_ context.Context,
					_ pgx.TxIsoLevel,
					_ pgx.TxAccessMode,
					_ func(context.Context) error,
				) error {
					return errors.New("deadlock")
				})
			},
			args:       args{ctx, "7", "user1", now},
			wantBisErr: assert.NoError,
			wantTxErr:  assert.Error,
		},
		{
			name: "Success",
			prepare: func(f *fields, a args) {
				f.tx.WithTxMock.Set(pass)
				f.ordRepo.GetMock.Set(func(_ context.Context, id string) (*models.Order, error) {
					return &models.Order{
						ID:       id,
						UserID:   a.userID,
						Status:   models.StatusIssued,
						IssuedAt: &now,
					}, nil
				})
				f.hrRepo.AddReturnMock.Set(func(_ context.Context, rec *models.ReturnRecord) error {
					return nil
				})
				f.ordRepo.UpdateMock.Set(func(_ context.Context, o *models.Order) error {
					assert.Equal(t, models.StatusReturned, o.Status)
					return nil
				})
				f.hrRepo.AddHistoryMock.Set(func(_ context.Context, evt *models.HistoryEvent) error {
					assert.Equal(t, models.StatusReturned, evt.Status)
					return nil
				})
				f.outboxRepo.AddMock.Set(func(_ context.Context, evt *models.OrderEvent) error {
					assert.Equal(t, models.OrderReturnedByClient, evt.EventType)
					assert.Equal(t, a.orderID, evt.Order.ID)
					return nil
				})
			},
			args:       args{ctx, "8", "user1", now},
			wantBisErr: assert.NoError,
			wantTxErr:  assert.NoError,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := minimock.NewController(t)
			defer ctrl.Finish()

			f := &fields{
				tx:         txMock.NewTxManagerMock(ctrl),
				ordRepo:    repoMock.NewOrdersRepositoryMock(ctrl),
				hrRepo:     repoMock.NewHistoryAndReturnsRepositoryMock(ctrl),
				outboxRepo: repoMock.NewOutboxRepositoryMock(ctrl),
			}
			if tt.prepare != nil {
				tt.prepare(f, tt.args)
			}

			log, _ := logger.NewLogger(logger.WithMode("prod"))
			wp := wpool.NewWorkerPool(1, 1, log)
			defer wp.Stop()

			svc := NewService(f.tx, f.ordRepo, f.hrRepo, f.outboxRepo, nil, wp)

			bisErr, txErr := svc.returnOneByClient(
				tt.args.ctx, tt.args.orderID, tt.args.userID, tt.args.now,
			)
			tt.wantBisErr(t, bisErr)
			tt.wantTxErr(t, txErr)
		})
	}
}
