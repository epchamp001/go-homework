package service

import (
	"context"
	"errors"
	"pvz-cli/internal/domain/models"
	"pvz-cli/internal/usecase"
	repoMock "pvz-cli/internal/usecase/mock"
	pkgMock "pvz-cli/internal/usecase/packaging/mock"
	txMock "pvz-cli/pkg/txmanager/mock"
	"testing"
	"time"

	"github.com/gojuno/minimock/v3"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
)

func TestServiceImpl_AcceptOrder(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	type fields struct {
		tx         *txMock.TxManagerMock
		ordRepo    *repoMock.OrdersRepositoryMock
		hrRepo     *repoMock.HistoryAndReturnsRepositoryMock
		outboxRepo *repoMock.OutboxRepositoryMock
		ordCache   *repoMock.OrderCacheMock
		pkgSvc     *pkgMock.PackagingStrategyMock
		strategy   *pkgMock.ProviderMock
	}

	type args struct {
		ctx     context.Context
		orderID string
		userID  string
		exp     time.Time
		weight  float64
		price   models.PriceKopecks
		pkgType models.PackageType
	}

	nowPlus := time.Now().Add(24 * time.Hour)
	tests := []struct {
		name    string
		prepare func(f *fields, a args)
		args    args
		want    models.PriceKopecks
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "SuccessAcceptOrder",
			prepare: func(f *fields, a args) {
				f.strategy.StrategyMock.
					Expect(a.pkgType).
					Return(f.pkgSvc, nil)

				f.pkgSvc.ValidateMock.
					Expect(a.weight).
					Return(nil)

				f.pkgSvc.SurchargeMock.
					Expect().
					Return(models.PriceKopecks(20))

				f.tx.WithTxMock.Set(func(
					_ context.Context,
					_ pgx.TxIsoLevel,
					_ pgx.TxAccessMode,
					fn func(context.Context) error,
				) error {
					return fn(ctx)
				})

				f.ordRepo.CreateMock.
					Set(func(_ context.Context, o *models.Order) error {
						assert.Equal(t, a.orderID, o.ID) // t — из родительской области
						return nil
					})

				f.hrRepo.AddHistoryMock.
					Set(func(_ context.Context, _ *models.HistoryEvent) error {
						return nil
					})

				f.outboxRepo.AddMock.
					Set(func(_ context.Context, evt *models.OrderEvent) error {
						assert.Equal(t, models.OrderAccepted, evt.EventType)
						assert.Equal(t, a.orderID, evt.Order.ID)
						return nil
					})

				f.ordCache.SetMock.
					Set(func(key string, val *models.Order) {
						assert.Equal(t, usecase.OrderKey(a.orderID), key)
						assert.Equal(t, a.orderID, val.ID)
					})
			}, args: args{
				ctx:     ctx,
				orderID: "123",
				userID:  "100",
				exp:     time.Now().Add(24 * time.Hour),
				weight:  5.0,
				price:   models.PriceKopecks(100),
				pkgType: models.PackageBag,
			},
			want:    models.PriceKopecks(120),
			wantErr: assert.NoError,
		},
		{
			name: "ValidationError_EmptyOrderID",
			args: args{
				ctx:     ctx,
				orderID: "",
				userID:  "100",
				exp:     nowPlus,
				weight:  5.0,
				price:   models.PriceKopecks(100),
				pkgType: models.PackageBag,
			},
			want:    0,
			wantErr: assert.Error,
		},
		{
			name: "ValidationError_ExpInPast",
			args: args{
				ctx:     ctx,
				orderID: "123",
				userID:  "100",
				exp:     time.Now().Add(-time.Hour),
				weight:  5.0,
				price:   models.PriceKopecks(100),
				pkgType: models.PackageBag,
			},
			want:    0,
			wantErr: assert.Error,
		},
		{
			name: "ValidationError_NonPositiveWeight",
			args: args{
				ctx:     ctx,
				orderID: "123",
				userID:  "100",
				exp:     nowPlus,
				weight:  0,
				price:   models.PriceKopecks(100),
				pkgType: models.PackageBag,
			},
			want:    0,
			wantErr: assert.Error,
		},
		{
			name: "InvalidPackageType",
			prepare: func(f *fields, a args) {
				f.strategy.StrategyMock.
					Expect(a.pkgType).
					Return(nil, errors.New("no strategy"))
			},
			args: args{
				ctx:     ctx,
				orderID: "123",
				userID:  "100",
				exp:     nowPlus,
				weight:  5.0,
				price:   models.PriceKopecks(100),
				pkgType: models.PackageType("invalid"),
			},
			want:    0,
			wantErr: assert.Error,
		},
		{
			name: "WeightValidationFailed",
			prepare: func(f *fields, a args) {
				f.strategy.StrategyMock.
					Expect(a.pkgType).
					Return(f.pkgSvc, nil)
				f.pkgSvc.ValidateMock.
					Expect(a.weight).
					Return(errors.New("too heavy"))
			},
			args: args{
				ctx:     ctx,
				orderID: "123",
				userID:  "100",
				exp:     nowPlus,
				weight:  1000.0,
				price:   models.PriceKopecks(100),
				pkgType: models.PackageBag,
			},
			want:    0,
			wantErr: assert.Error,
		},
		{
			name: "CreateOrderError",
			prepare: func(f *fields, a args) {
				f.strategy.StrategyMock.
					Expect(a.pkgType).
					Return(f.pkgSvc, nil)
				f.pkgSvc.ValidateMock.
					Expect(a.weight).
					Return(nil)
				f.pkgSvc.SurchargeMock.
					Expect().
					Return(models.PriceKopecks(20))
				f.tx.WithTxMock.Set(func(
					_ context.Context,
					_ pgx.TxIsoLevel,
					_ pgx.TxAccessMode,
					fn func(context.Context) error,
				) error {
					return fn(ctx)
				})
				f.ordRepo.CreateMock.
					Set(func(_ context.Context, _ *models.Order) error {
						return errors.New("db failure")
					})
			},
			args: args{
				ctx:     ctx,
				orderID: "123",
				userID:  "100",
				exp:     nowPlus,
				weight:  5.0,
				price:   models.PriceKopecks(100),
				pkgType: models.PackageBag,
			},
			want:    0,
			wantErr: assert.Error,
		},
		{
			name: "AddHistoryError",
			prepare: func(f *fields, a args) {
				f.strategy.StrategyMock.
					Expect(a.pkgType).
					Return(f.pkgSvc, nil)
				f.pkgSvc.ValidateMock.
					Expect(a.weight).
					Return(nil)
				f.pkgSvc.SurchargeMock.
					Expect().
					Return(models.PriceKopecks(20))
				f.tx.WithTxMock.Set(func(
					_ context.Context,
					_ pgx.TxIsoLevel,
					_ pgx.TxAccessMode,
					fn func(context.Context) error,
				) error {
					return fn(ctx)
				})
				f.ordRepo.CreateMock.
					Set(func(_ context.Context, _ *models.Order) error {
						return nil
					})
				f.hrRepo.AddHistoryMock.
					Set(func(_ context.Context, _ *models.HistoryEvent) error {
						return errors.New("history failure")
					})
			},
			args: args{
				ctx:     ctx,
				orderID: "123",
				userID:  "100",
				exp:     nowPlus,
				weight:  5.0,
				price:   models.PriceKopecks(100),
				pkgType: models.PackageBag,
			},
			want:    0,
			wantErr: assert.Error,
		},
		{
			name: "TransactionError",
			prepare: func(f *fields, a args) {
				f.strategy.StrategyMock.
					Expect(a.pkgType).
					Return(f.pkgSvc, nil)
				f.pkgSvc.ValidateMock.
					Expect(a.weight).
					Return(nil)
				f.pkgSvc.SurchargeMock.
					Expect().
					Return(models.PriceKopecks(20))
				// эмулируем провал самого WithTx
				f.tx.WithTxMock.Set(func(
					_ context.Context,
					_ pgx.TxIsoLevel,
					_ pgx.TxAccessMode,
					_ func(context.Context) error,
				) error {
					return errors.New("tx aborted")
				})
			},
			args: args{
				ctx:     ctx,
				orderID: "123",
				userID:  "100",
				exp:     nowPlus,
				weight:  5.0,
				price:   models.PriceKopecks(100),
				pkgType: models.PackageBag,
			},
			want:    0,
			wantErr: assert.Error,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := minimock.NewController(t)
			fieldsForTests := &fields{
				tx:         txMock.NewTxManagerMock(ctrl),
				ordRepo:    repoMock.NewOrdersRepositoryMock(ctrl),
				hrRepo:     repoMock.NewHistoryAndReturnsRepositoryMock(ctrl),
				outboxRepo: repoMock.NewOutboxRepositoryMock(ctrl),
				pkgSvc:     pkgMock.NewPackagingStrategyMock(ctrl),
				strategy:   pkgMock.NewProviderMock(ctrl),
				ordCache:   repoMock.NewOrderCacheMock(ctrl),
			}

			s := NewService(fieldsForTests.tx, fieldsForTests.ordRepo, fieldsForTests.hrRepo, fieldsForTests.outboxRepo, fieldsForTests.strategy, nil, fieldsForTests.ordCache)

			if tt.prepare != nil {
				tt.prepare(fieldsForTests, tt.args)
			}
			got, err := s.AcceptOrder(
				tt.args.ctx, tt.args.orderID, tt.args.userID,
				tt.args.exp, tt.args.weight, tt.args.price, tt.args.pkgType,
			)
			tt.wantErr(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
