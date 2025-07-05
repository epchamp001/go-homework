//go:build integration

package simple

import (
	"context"
	"pvz-cli/internal/domain/models"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImportOrders_Happy(t *testing.T) {
	cleanDB(t)

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

	count, err := svc.ImportOrders(ctx, orders)
	require.NoError(t, err)
	assert.Equal(t, 2, count)

	var total int
	err = masterPool.QueryRow(ctx,
		`SELECT COUNT(*) FROM orders`,
	).Scan(&total)
	require.NoError(t, err)
	assert.Equal(t, 2, total)

	for _, id := range []string{"500", "501"} {
		var cnt int
		err = masterPool.QueryRow(ctx,
			`SELECT COUNT(*) FROM order_history WHERE order_id = $1`, id,
		).Scan(&cnt)
		require.NoError(t, err)
		assert.Equal(t, 1, cnt, "history events for %s", id)
	}

	var status string
	err = masterPool.QueryRow(ctx,
		`SELECT status FROM orders WHERE id = $1`, "500",
	).Scan(&status)
	require.NoError(t, err)
	assert.Equal(t, models.StatusAccepted, models.OrderStatus(status))
}
