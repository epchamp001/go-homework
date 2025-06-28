package integration

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"pvz-cli/internal/domain/models"
	"time"
)

func (s *TestSuite) TestImportOrders_Happy() {
	ctx := context.Background()

	exp1 := time.Now().Add(24 * time.Hour)
	exp2 := time.Now().Add(48 * time.Hour)

	orders := []*models.Order{
		{
			ID:         "500",
			UserID:     "300",
			Status:     models.StatusAccepted,
			ExpiresAt:  exp1,
			CreatedAt:  time.Now(),
			Package:    models.PackageBag,
			Weight:     1.0,
			Price:      50000,
			TotalPrice: 50500,
		},
		{
			ID:         "501",
			UserID:     "301",
			Status:     models.StatusAccepted,
			ExpiresAt:  exp2,
			CreatedAt:  time.Now(),
			Package:    models.PackageBox,
			Weight:     2.0,
			Price:      150000,
			TotalPrice: 152000,
		},
	}

	count, err := s.svc.ImportOrders(ctx, orders)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 2, count)

	var got int
	err = s.masterPool.QueryRow(ctx,
		`SELECT COUNT(*) FROM orders`,
	).Scan(&got)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 2, got)

	err = s.masterPool.QueryRow(ctx,
		`SELECT COUNT(*) FROM order_history WHERE order_id = '500'`,
	).Scan(&got)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 1, got)

	err = s.masterPool.QueryRow(ctx,
		`SELECT COUNT(*) FROM order_history WHERE order_id = '501'`,
	).Scan(&got)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 1, got)

	var status string
	err = s.masterPool.QueryRow(ctx,
		`SELECT status FROM orders WHERE id = '500'`,
	).Scan(&status)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), models.StatusAccepted, models.OrderStatus(status))
}
