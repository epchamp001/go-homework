package service

import (
	"context"
	"errors"
	"pvz-cli/internal/domain/models"
	"pvz-cli/internal/usecase"
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

func TestServiceImpl_IssueOrders(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	type fields struct {
		tx         *txMock.TxManagerMock
		ordRepo    *repoMock.OrdersRepositoryMock
		hrRepo     *repoMock.HistoryAndReturnsRepositoryMock
		outboxRepo *repoMock.OutboxRepositoryMock
		ordCache   *repoMock.OrderCacheMock
		metrics    *repoMock.BussinesMetricsMock
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
			name: "Tx aborted on first ID -> fatal error, map nil",
			prepare: func(f *fields, a args) {
				// первый id "bad" вернёт txErr
				call := 0
				f.ordCache.GetMock.
					Return(nil, false)

				f.ordCache.SetMock.
					Set(func(string, *models.Order) {})

				f.tx.WithTxMock.Set(func(
					_ context.Context, _ pgx.TxIsoLevel, _ pgx.TxAccessMode,
					fn func(context.Context) error,
				) error {
					call++
					if call == 1 {
						return errors.New("deadlock")
					}
					return fn(context.Background())
				})
				f.ordRepo.GetMock.Set(func(_ context.Context, id string) (*models.Order, error) {
					return &models.Order{
						ID:        id,
						UserID:    a.userID,
						Status:    models.StatusAccepted,
						ExpiresAt: validExpiry,
					}, nil
				})
				f.ordRepo.UpdateMock.Return(nil)
				f.hrRepo.AddHistoryMock.Return(nil)
				f.outboxRepo.AddMock.Return(nil)
				f.metrics.IncOrdersIssuedMock.Set(func() {})
			},
			args:    args{ctx, "u", []string{"bad", "other"}},
			wantErr: assert.Error,
		},
		{
			name: "Mixed results: bisErr & ok",
			prepare: func(f *fields, a args) {
				f.ordCache.GetMock.
					Return(nil, false)

				f.ordCache.SetMock.
					Set(func(string, *models.Order) {})

				f.tx.WithTxMock.Set(pass)
				f.ordRepo.GetMock.Set(func(_ context.Context, id string) (*models.Order, error) {
					switch id {
					case "bis":
						return nil, errors.New("not found")
					default:
						return &models.Order{
							ID: id, UserID: a.userID,
							Status: models.StatusAccepted, ExpiresAt: validExpiry,
						}, nil
					}
				})
				f.ordRepo.UpdateMock.Return(nil)
				f.hrRepo.AddHistoryMock.Return(nil)
				f.outboxRepo.AddMock.Return(nil)
				f.metrics.IncOrdersIssuedMock.Set(func() {})
			},
			args:    args{ctx, "u", []string{"ok", "bis"}},
			wantErr: assert.NoError,
			wantResults: map[string]assert.ErrorAssertionFunc{
				"ok":  assert.NoError,
				"bis": assert.Error,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := minimock.NewController(t)
			f := &fields{
				tx:         txMock.NewTxManagerMock(ctrl),
				ordRepo:    repoMock.NewOrdersRepositoryMock(ctrl),
				hrRepo:     repoMock.NewHistoryAndReturnsRepositoryMock(ctrl),
				outboxRepo: repoMock.NewOutboxRepositoryMock(ctrl),
				ordCache:   repoMock.NewOrderCacheMock(ctrl),
				metrics:    repoMock.NewBussinesMetricsMock(ctrl),
			}
			if tt.prepare != nil {
				tt.prepare(f, tt.args)
			}

			log, _ := logger.NewLogger(logger.WithMode("prod"))
			wp := wpool.NewWorkerPool(4, 16, log)
			defer wp.Stop()

			svc := NewService(f.tx, f.ordRepo, f.hrRepo, f.outboxRepo, nil, wp, f.ordCache, f.metrics)

			got, err := svc.IssueOrders(tt.args.ctx, tt.args.userID, tt.args.ids)
			tt.wantErr(t, err)

			if tt.wantResults == nil {
				assert.Nil(t, got)
			} else {
				assert.Len(t, got, len(tt.args.ids))
				for id, want := range tt.wantResults {
					want(t, got[id], "order %s", id)
				}
			}
		})
	}
}

func TestServiceImpl_issueOne(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	type fields struct {
		tx         *txMock.TxManagerMock
		ordRepo    *repoMock.OrdersRepositoryMock
		hrRepo     *repoMock.HistoryAndReturnsRepositoryMock
		outboxRepo *repoMock.OutboxRepositoryMock
		ordCache   *repoMock.OrderCacheMock
		metrics    *repoMock.BussinesMetricsMock
	}
	type args struct {
		ctx     context.Context
		orderID string
		userID  string
	}
	now := time.Now()
	validExpiry := now.Add(time.Hour)

	tests := []struct {
		name       string
		prepare    func(f *fields, a args)
		args       args
		wantbisErr assert.ErrorAssertionFunc
		wantTxErr  assert.ErrorAssertionFunc
	}{
		{
			name: "order not found -> bisErr",
			prepare: func(f *fields, a args) {
				f.ordCache.GetMock.Return(nil, false)
				f.tx.WithTxMock.Set(pass)
				f.ordRepo.GetMock.Return(nil, errors.New("nf"))
				f.metrics.IncOrdersIssuedMock.Set(func() {})
			},
			args:       args{ctx, "1", "u"},
			wantbisErr: assert.Error,
			wantTxErr:  assert.NoError,
		},
		{
			name: "wrong user -> validation bisErr",
			prepare: func(f *fields, a args) {
				f.ordCache.GetMock.Return(nil, false)
				f.tx.WithTxMock.Set(pass)
				f.ordRepo.GetMock.Return(&models.Order{
					ID: "2", UserID: "other",
					Status: models.StatusAccepted, ExpiresAt: validExpiry,
				}, nil)
				f.metrics.IncOrdersIssuedMock.Set(func() {})
			},
			args:       args{ctx, "2", "u"},
			wantbisErr: assert.Error,
			wantTxErr:  assert.NoError,
		},
		{
			name: "tx aborted -> txErr",
			prepare: func(f *fields, a args) {
				f.ordCache.GetMock.Return(nil, false)
				f.tx.WithTxMock.Set(func(context.Context, pgx.TxIsoLevel,
					pgx.TxAccessMode, func(context.Context) error) error {
					return errors.New("deadlock")
				})
				f.metrics.IncOrdersIssuedMock.Set(func() {})
			},
			args:       args{ctx, "3", "u"},
			wantbisErr: assert.NoError,
			wantTxErr:  assert.Error,
		},
		{
			name: "happy path",
			prepare: func(f *fields, a args) {
				f.ordCache.GetMock.Return(nil, false)
				f.ordCache.SetMock.Set(func(k string, o *models.Order) { // ②
					assert.Equal(t, usecase.OrderKey(a.orderID), k)
					assert.Equal(t, a.orderID, o.ID)
					assert.Equal(t, models.StatusIssued, o.Status)
				})
				f.tx.WithTxMock.Set(pass)
				f.ordRepo.GetMock.Return(&models.Order{
					ID: a.orderID, UserID: a.userID,
					Status: models.StatusAccepted, ExpiresAt: validExpiry,
				}, nil)
				f.ordRepo.UpdateMock.Return(nil)
				f.hrRepo.AddHistoryMock.Return(nil)
				f.outboxRepo.AddMock.Return(nil)
				f.metrics.IncOrdersIssuedMock.Set(func() {})
			},
			args:       args{ctx, "4", "u"},
			wantbisErr: assert.NoError,
			wantTxErr:  assert.NoError,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := minimock.NewController(t)
			f := &fields{
				tx:         txMock.NewTxManagerMock(ctrl),
				ordRepo:    repoMock.NewOrdersRepositoryMock(ctrl),
				hrRepo:     repoMock.NewHistoryAndReturnsRepositoryMock(ctrl),
				outboxRepo: repoMock.NewOutboxRepositoryMock(ctrl),
				ordCache:   repoMock.NewOrderCacheMock(ctrl),
				metrics:    repoMock.NewBussinesMetricsMock(ctrl),
			}
			if tt.prepare != nil {
				tt.prepare(f, tt.args)
			}

			log, _ := logger.NewLogger(
				logger.WithMode("prod"),
				logger.WithEncoding("console"),
			)
			wp := wpool.NewWorkerPool(4, 16, log)
			defer wp.Stop()

			svc := NewService(f.tx, f.ordRepo, f.hrRepo, f.outboxRepo, nil, wp, f.ordCache, f.metrics)

			bis, txErr := svc.issueOne(tt.args.ctx, tt.args.orderID, tt.args.userID, now)
			tt.wantbisErr(t, bis)
			tt.wantTxErr(t, txErr)
		})
	}
}

// пасс-сквозной WithTx для краткости
func pass(
	txCtx context.Context,
	_ pgx.TxIsoLevel,
	_ pgx.TxAccessMode,
	fn func(context.Context) error,
) error {
	return fn(txCtx)
}
