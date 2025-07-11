package service

import (
	"context"
	"errors"
	"pvz-cli/internal/domain/models"
	"pvz-cli/internal/domain/vo"
	repoMock "pvz-cli/internal/usecase/mock"
	txMock "pvz-cli/pkg/txmanager/mock"
	"sort"
	"testing"
	"time"

	"github.com/gojuno/minimock/v3"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
)

func TestServiceImpl_ListOrders(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	type fields struct {
		tx      *txMock.TxManagerMock
		ordRepo *repoMock.OrdersRepositoryMock
		hrRepo  *repoMock.HistoryAndReturnsRepositoryMock
	}
	type args struct {
		ctx       context.Context
		userID    string
		onlyInPVZ bool
		lastN     int
		pg        vo.Pagination
	}

	now := time.Now()
	o1 := &models.Order{ID: "1", UserID: "user1", CreatedAt: now.Add(-3 * time.Hour)}
	o2 := &models.Order{ID: "2", UserID: "user1", CreatedAt: now.Add(-1 * time.Hour)}
	o3 := &models.Order{ID: "3", UserID: "user1", CreatedAt: now.Add(-2 * time.Hour)}
	activeOrders := []*models.Order{o1, o2, o3}

	tests := []struct {
		name         string
		prepare      func(f *fields, a args)
		args         args
		wantErr      assert.ErrorAssertionFunc
		wantSliceNil bool
		wantTotal    int
		active       []*models.Order
	}{
		{
			name:         "EmptyUserID",
			args:         args{ctx: ctx, userID: "", onlyInPVZ: false, lastN: 0, pg: vo.Pagination{}},
			wantErr:      assert.Error,
			wantSliceNil: true,
			wantTotal:    0,
		},
		{
			name: "ListByUserError",
			prepare: func(f *fields, a args) {
				f.tx.WithTxMock.Set(func(
					txCtx context.Context,
					_ pgx.TxIsoLevel,
					_ pgx.TxAccessMode,
					fn func(context.Context) error,
				) error {
					return fn(txCtx)
				})
				f.ordRepo.ListByUserMock.Set(func(
					_ context.Context,
					userID string,
					onlyInPVZ bool,
					lastN int,
					pg *vo.Pagination,
				) ([]*models.Order, error) {
					assert.Equal(t, a.userID, userID)
					assert.Equal(t, a.onlyInPVZ, onlyInPVZ)
					assert.Zero(t, lastN)
					assert.Nil(t, pg)
					return nil, errors.New("db failure")
				})
			},
			args:         args{ctx: ctx, userID: "user1", onlyInPVZ: true, lastN: 0, pg: vo.Pagination{}},
			wantErr:      assert.Error,
			wantSliceNil: true,
			wantTotal:    0,
		},
		{
			name: "TransactionError",
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
			args:         args{ctx: ctx, userID: "user1", onlyInPVZ: false, lastN: 0, pg: vo.Pagination{}},
			wantErr:      assert.Error,
			wantSliceNil: true,
			wantTotal:    0,
		},
		{
			name: "SuccessSortedAndPaginated",
			prepare: func(f *fields, a args) {
				f.tx.WithTxMock.Set(func(
					txCtx context.Context,
					_ pgx.TxIsoLevel,
					_ pgx.TxAccessMode,
					fn func(context.Context) error,
				) error {
					return fn(txCtx)
				})
				f.ordRepo.ListByUserMock.Set(func(
					_ context.Context,
					_userID string,
					_onlyInPVZ bool,
					_lastN int,
					pg *vo.Pagination,
				) ([]*models.Order, error) {
					assert.Nil(t, pg)
					return activeOrders, nil
				})
			},
			args:         args{ctx: ctx, userID: "user1", onlyInPVZ: false, lastN: 5, pg: vo.Pagination{}},
			wantErr:      assert.NoError,
			wantSliceNil: false,
			wantTotal:    len(activeOrders),
			active:       activeOrders,
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
			service := NewService(f.tx, f.ordRepo, f.hrRepo, nil, nil, nil)

			if tt.prepare != nil {
				tt.prepare(f, tt.args)
			}

			got, total, err := service.ListOrders(
				tt.args.ctx, tt.args.userID, tt.args.onlyInPVZ, tt.args.lastN, tt.args.pg,
			)
			tt.wantErr(t, err)

			if tt.wantSliceNil {
				assert.Nil(t, got)
				assert.Zero(t, total)
			} else {
				assert.Equal(t, tt.wantTotal, total)

				expected := make([]*models.Order, len(tt.active))
				copy(expected, tt.active)
				sort.Slice(expected, func(i, j int) bool {
					return expected[i].CreatedAt.Before(expected[j].CreatedAt)
				})
				assert.Equal(t, expected, got)
			}
		})
	}
}
