package simple

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"pvz-cli/internal/domain/models"
	"pvz-cli/internal/usecase/packaging"
	"testing"
	"time"
)

const (
	orderID uint64              = 100
	userID  uint64              = 200
	priceK  models.PriceKopecks = 500_00
)

func TestAcceptOrder_Happy(t *testing.T) {
	// каждый тест работает с чистой базой
	cleanDB(t)

	expiresAt := time.Now().Add(72 * time.Hour)

	total, err := svc.AcceptOrder(
		ctx,
		fmt.Sprint(orderID),
		fmt.Sprint(userID),
		expiresAt,
		1.2,
		priceK, // 50000
		models.PackageBag,
	)
	require.NoError(t, err)

	strat := packaging.NewBagStrategy()
	expected := priceK + strat.Surcharge()
	assert.Equal(t, expected, total)

	var status string
	var totalDb int64
	err = masterPool.QueryRow(ctx,
		`SELECT status, total_price FROM orders WHERE id = $1`, orderID,
	).Scan(&status, &totalDb)
	require.NoError(t, err)
	assert.Equal(t, string(models.StatusAccepted), status)
	assert.Equal(t, int64(expected), totalDb)

	var count int
	err = masterPool.QueryRow(ctx,
		`SELECT COUNT(*) FROM order_history WHERE order_id = $1`, orderID,
	).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}
