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

func TestIssueOrders_Simple(t *testing.T) {
	cleanDB(t)

	ctx := context.Background()

	orderID := "100"
	userID := "200"
	priceK := models.PriceKopecks(50000)

	expires := time.Now().Add(48 * time.Hour)
	_, err := svc.AcceptOrder(
		ctx,
		orderID,
		userID,
		expires,
		1.0,
		priceK,
		models.PackageBag,
	)
	require.NoError(t, err)

	res, err := svc.IssueOrders(ctx, userID, []string{orderID})
	require.NoError(t, err)

	require.Contains(t, res, orderID)
	assert.NoError(t, res[orderID])

	var status string
	var issuedAt time.Time
	err = masterPool.QueryRow(ctx,
		`SELECT status, issued_at FROM orders WHERE id = $1`, orderID,
	).Scan(&status, &issuedAt)
	require.NoError(t, err)
	assert.Equal(t, models.StatusIssued, models.OrderStatus(status))
	assert.WithinDuration(t, time.Now(), issuedAt, 2*time.Second)

	var histCount int
	err = masterPool.QueryRow(ctx,
		`SELECT COUNT(*) FROM order_history WHERE order_id = $1`, orderID,
	).Scan(&histCount)
	require.NoError(t, err)
	assert.Equal(t, 2, histCount)
}
