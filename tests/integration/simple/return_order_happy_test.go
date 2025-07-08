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

func TestReturnOrder_Simple(t *testing.T) {
	cleanDB(t)

	ctx := context.Background()

	expiredAt := time.Now().Add(-1 * time.Hour)
	createdAt := time.Now().Add(-24 * time.Hour)

	_, err := masterPool.Exec(ctx, `
        INSERT INTO orders
            (id, user_id, status, expires_at, created_at, package, weight, price, total_price)
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
    `,
		110,
		200,
		models.StatusAccepted,
		expiredAt,
		createdAt,
		models.PackageBag,
		1.0,
		50000,
		50500,
	)
	require.NoError(t, err)

	_, err = masterPool.Exec(ctx, `
        INSERT INTO order_history (order_id, status, event_time)
        VALUES ($1,$2,$3)
    `, 110, models.StatusAccepted, createdAt)
	require.NoError(t, err)

	err = svc.ReturnOrder(ctx, "110")
	require.NoError(t, err)

	var (
		status     string
		returnedAt time.Time
	)
	err = masterPool.QueryRow(ctx,
		`SELECT status, returned_at FROM orders WHERE id = $1`, 110,
	).Scan(&status, &returnedAt)
	require.NoError(t, err)

	assert.Equal(t, models.StatusReturned, models.OrderStatus(status))
	assert.WithinDuration(t, time.Now(), returnedAt, 2*time.Second)

	var cnt int
	err = masterPool.QueryRow(ctx,
		`SELECT COUNT(*) FROM order_returns WHERE order_id = $1`, 110,
	).Scan(&cnt)
	require.NoError(t, err)
	assert.Equal(t, 1, cnt)

	err = masterPool.QueryRow(ctx,
		`SELECT COUNT(*) FROM order_history WHERE order_id = $1 AND status = $2`,
		110, models.StatusReturned,
	).Scan(&cnt)
	require.NoError(t, err)
	assert.Equal(t, 1, cnt)
}
