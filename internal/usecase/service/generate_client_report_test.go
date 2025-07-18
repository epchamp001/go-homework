package service

import (
	"bytes"
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
	"github.com/xuri/excelize/v2"
)

func TestServiceImpl_generateClientReport(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	type fields struct {
		tx      *txMock.TxManagerMock
		ordRepo *repoMock.OrdersRepositoryMock
	}
	type args struct {
		ctx    context.Context
		sortBy string
	}

	now := time.Now()
	orders := []*models.Order{
		{UserID: "u1", Price: models.PriceKopecks(100), ReturnedAt: &now},
		{UserID: "u1", Price: models.PriceKopecks(50)},
		{UserID: "u1", Price: models.PriceKopecks(25)},
		{UserID: "u2", Price: models.PriceKopecks(200)},
	}

	tests := []struct {
		name        string
		prepare     func(f *fields)
		args        args
		wantErr     assert.ErrorAssertionFunc
		checkReport func(t *testing.T, reps []*models.ClientReport)
	}{
		{
			name: "ListAllOrdersError",
			prepare: func(f *fields) {
				f.tx.WithTxMock.Set(func(txCtx context.Context, _ pgx.TxIsoLevel, _ pgx.TxAccessMode, fn func(context.Context) error) error {
					return fn(txCtx)
				})
				f.ordRepo.ListAllOrdersMock.Set(func(_ context.Context) ([]*models.Order, error) {
					return nil, errors.New("db fail")
				})
			},
			args:    args{ctx: ctx, sortBy: "orders"},
			wantErr: assert.Error,
		},
		{
			name: "InvalidSortParameter",
			prepare: func(f *fields) {
				f.tx.WithTxMock.Set(func(txCtx context.Context, _ pgx.TxIsoLevel, _ pgx.TxAccessMode, fn func(context.Context) error) error {
					return fn(txCtx)
				})
				f.ordRepo.ListAllOrdersMock.Set(func(_ context.Context) ([]*models.Order, error) {
					return orders, nil
				})
			},
			args:    args{ctx: ctx, sortBy: "bad_key"},
			wantErr: assert.Error,
		},
		{
			name: "SortByOrders",
			prepare: func(f *fields) {
				f.tx.WithTxMock.Set(func(txCtx context.Context, _ pgx.TxIsoLevel, _ pgx.TxAccessMode, fn func(context.Context) error) error {
					return fn(txCtx)
				})
				f.ordRepo.ListAllOrdersMock.Set(func(_ context.Context) ([]*models.Order, error) {
					return orders, nil
				})
			},
			args:    args{ctx: ctx, sortBy: "orders"},
			wantErr: assert.NoError,
			checkReport: func(t *testing.T, reps []*models.ClientReport) {
				assert.Len(t, reps, 2)
				// reports sorted descending by TotalOrders
				// u1 has 3, u2 has 1
				assert.Equal(t, "u1", reps[0].UserID)
				assert.Equal(t, 3, reps[0].TotalOrders)
				assert.Equal(t, "u2", reps[1].UserID)
				assert.Equal(t, 1, reps[1].TotalOrders)
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
			}

			service := NewService(f.tx, f.ordRepo, nil, nil, nil, nil, nil, nil)

			if tt.prepare != nil {
				tt.prepare(f)
			}

			reps, err := service.generateClientReport(tt.args.ctx, tt.args.sortBy)
			tt.wantErr(t, err)
			if err == nil && tt.checkReport != nil {
				tt.checkReport(t, reps)
			}
		})
	}
}

func TestAggregateOrders(t *testing.T) {
	t.Parallel()

	now := time.Now()
	orders := []*models.Order{
		{UserID: "u1", Status: models.StatusAccepted, Price: models.PriceKopecks(100)},
		{UserID: "u1", Status: models.StatusReturned, ReturnedAt: &now, Price: models.PriceKopecks(100)},
		{UserID: "u2", Status: models.StatusAccepted, Price: models.PriceKopecks(200)},
		{UserID: "u1", Status: models.StatusAccepted, Price: models.PriceKopecks(50)},
		{UserID: "u2", Status: models.StatusReturned, ReturnedAt: &now, Price: models.PriceKopecks(200)},
	}

	clientsMap := make(map[string]*models.ClientReport)
	aggregateOrders(clientsMap, orders)

	cr1, ok1 := clientsMap["u1"]
	assert.True(t, ok1, "client u1 should exist")
	assert.Equal(t, 3, cr1.TotalOrders)
	assert.Equal(t, 1, cr1.ReturnedOrders)
	assert.Equal(t, models.PriceKopecks(150), cr1.TotalPurchaseSum)

	cr2, ok2 := clientsMap["u2"]
	assert.True(t, ok2, "client u2 should exist")
	assert.Equal(t, 2, cr2.TotalOrders)
	assert.Equal(t, 1, cr2.ReturnedOrders)
	assert.Equal(t, models.PriceKopecks(200), cr2.TotalPurchaseSum)
}

func TestServiceImpl_GenerateClientReportByte(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	now := time.Now()
	orders := []*models.Order{
		{UserID: "u1", Status: models.StatusReturned, Price: models.PriceKopecks(100), ReturnedAt: &now},
		{UserID: "u1", Status: models.StatusAccepted, Price: models.PriceKopecks(50)},
		{UserID: "u1", Status: models.StatusAccepted, Price: models.PriceKopecks(25)},
		{UserID: "u2", Status: models.StatusAccepted, Price: models.PriceKopecks(200)},
	}

	ctrl := minimock.NewController(t)
	tx := txMock.NewTxManagerMock(ctrl)
	ordRepo := repoMock.NewOrdersRepositoryMock(ctrl)

	tx.WithTxMock.Set(func(
		txCtx context.Context,
		_ pgx.TxIsoLevel,
		_ pgx.TxAccessMode,
		fn func(context.Context) error,
	) error {
		return fn(txCtx)
	})
	ordRepo.ListAllOrdersMock.Set(func(_ context.Context) ([]*models.Order, error) {
		return orders, nil
	})

	service := NewService(tx, ordRepo, nil, nil, nil, nil, nil, nil)

	data, err := service.GenerateClientReportByte(ctx, "orders")
	assert.NoError(t, err)

	f, err := excelize.OpenReader(bytes.NewReader(data))
	assert.NoError(t, err)
	sheet := "ClientsReport"

	expectedHeaders := []struct{ cell, want string }{
		{"A1", "UserID"},
		{"B1", "Total Orders"},
		{"C1", "Returned Orders"},
		{"D1", "Total Purchase Sum (₽)"},
	}
	for _, h := range expectedHeaders {
		v, err := f.GetCellValue(sheet, h.cell)
		assert.NoError(t, err)
		assert.Equal(t, h.want, v)
	}

	rows, err := f.GetRows(sheet)
	assert.NoError(t, err)

	results := map[string][]string{}
	for i, row := range rows {
		if i == 0 {
			continue
		}
		if len(row) < 4 {
			continue
		}
		results[row[0]] = []string{row[1], row[2], row[3]}
	}

	r1, ok := results["u1"]
	assert.True(t, ok)
	assert.Equal(t, "3", r1[0])
	assert.Equal(t, "1", r1[1])
	assert.Equal(t, "0.75", r1[2])

	r2, ok := results["u2"]
	assert.True(t, ok)
	assert.Equal(t, "1", r2[0])
	assert.Equal(t, "0", r2[1])
	assert.Equal(t, "2", r2[2])
}
