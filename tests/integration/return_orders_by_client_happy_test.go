//go:build integration

package integration

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"pvz-cli/internal/domain/models"
	"time"
)

func (s *TestSuite) TestReturnOrdersByClient_Happy() {
	s.loadFixtures()

	ctx := context.Background()
	_, err := s.masterPool.Exec(ctx, `TRUNCATE order_returns`)
	s.Require().NoError(err)

	res, err := s.svc.ReturnOrdersByClient(ctx, "200", []string{"108"})
	require.NoError(s.T(), err)

	require.Contains(s.T(), res, "108")
	assert.NoError(s.T(), res["108"])

	var (
		status     string
		returnedAt time.Time
	)
	err = s.masterPool.QueryRow(ctx,
		`SELECT status, returned_at FROM orders WHERE id = $1`, 108,
	).Scan(&status, &returnedAt)
	require.NoError(s.T(), err)

	assert.Equal(s.T(), models.StatusReturned, models.OrderStatus(status))
	assert.WithinDuration(s.T(), time.Now(), returnedAt, 2*time.Second)

	var cnt int
	err = s.masterPool.QueryRow(ctx,
		`SELECT COUNT(*) FROM order_returns WHERE order_id = $1`, 108,
	).Scan(&cnt)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 1, cnt)
}
