//go:build integration

package integration

import (
	"context"
	"pvz-cli/internal/domain/models"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (s *TestSuite) TestReturnOrdersByClient_Happy() {
	ctx := context.Background()

	_, err := s.masterPool.Exec(ctx, `
		TRUNCATE orders, order_history, order_returns RESTART IDENTITY CASCADE`)
	s.Require().NoError(err)

	issuedAt := time.Now().Add(-1 * time.Hour).UTC()

	_, err = s.masterPool.Exec(ctx, `
		INSERT INTO orders (id, user_id, status, issued_at, expires_at,
		                    package, weight, price, total_price)
		VALUES ($1,$2,'ISSUED',$3,$4,'bag',1,50000,50500)`,
		108, 200, issuedAt, issuedAt.Add(240*time.Hour))
	s.Require().NoError(err)

	_, err = s.masterPool.Exec(ctx, `
		INSERT INTO order_history (order_id, status, event_time)
		VALUES ($1,'ISSUED',$2)`,
		108, issuedAt)
	s.Require().NoError(err)

	res, err := s.svc.ReturnOrdersByClient(ctx, "200", []string{"108"})
	require.NoError(s.T(), err)

	require.Contains(s.T(), res, "108")
	assert.NoError(s.T(), res["108"])

	var status string
	var returnedAt time.Time
	err = s.masterPool.QueryRow(ctx,
		`SELECT status, returned_at FROM orders WHERE id = $1`, 108,
	).Scan(&status, &returnedAt)
	require.NoError(s.T(), err)

	assert.Equal(s.T(), models.StatusReturned, models.OrderStatus(status))
	assert.WithinDuration(s.T(), time.Now(), returnedAt, 5*time.Second)

	var cnt int
	err = s.masterPool.QueryRow(ctx,
		`SELECT COUNT(*) FROM order_returns WHERE order_id = $1`, 108,
	).Scan(&cnt)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 1, cnt)

	err = s.masterPool.QueryRow(ctx,
		`SELECT COUNT(*) FROM order_history WHERE order_id = $1 AND status = 'RETURNED'`,
		108,
	).Scan(&cnt)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 1, cnt)
}
