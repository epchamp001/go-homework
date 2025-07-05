//go:build integration

package integration

import (
	"context"
	"pvz-cli/internal/domain/models"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (s *TestSuite) TestReturnOrder_Happy() {
	s.loadFixtures()

	ctx := context.Background()

	err := s.svc.ReturnOrder(ctx, "110")
	require.NoError(s.T(), err)

	var (
		status     string
		returnedAt time.Time
	)
	err = s.masterPool.QueryRow(ctx,
		`SELECT status, returned_at FROM orders WHERE id = $1`, 110,
	).Scan(&status, &returnedAt)
	require.NoError(s.T(), err)

	assert.Equal(s.T(), models.StatusReturned, models.OrderStatus(status))
	assert.WithinDuration(s.T(), time.Now(), returnedAt, 2*time.Second)

	var cnt int
	err = s.masterPool.QueryRow(ctx,
		`SELECT COUNT(*) FROM order_returns WHERE order_id = $1`, 110,
	).Scan(&cnt)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 1, cnt)

	err = s.masterPool.QueryRow(ctx,
		`SELECT COUNT(*) FROM order_history WHERE order_id = $1 AND status = $2`,
		110, string(models.StatusReturned),
	).Scan(&cnt)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 1, cnt)
}
