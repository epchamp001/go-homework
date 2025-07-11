//go:build integration

package integration

import (
	"context"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (s *TestSuite) TestReturnOrder_ErrEmptyID() {
	ctx := context.Background()

	err := s.svc.ReturnOrder(ctx, "")
	require.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "empty order id")
}

func (s *TestSuite) TestReturnOrder_ErrNotFound() {
	ctx := context.Background()

	err := s.svc.ReturnOrder(ctx, "999") // такого заказа нет в fixtures
	require.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "order not found")
}

func (s *TestSuite) TestReturnOrder_ErrNotExpired_NoFixtures() {
	ctx := context.Background()

	future := time.Now().UTC().Add(24 * time.Hour)
	_, err := s.masterPool.Exec(ctx, `
        INSERT INTO orders (
            id, user_id, status, expires_at, created_at,
            package, weight, price, total_price
        ) VALUES (
            $1, $2, $3, $4, $5,
            'bag', 1.0, 100, 120
        )
    `, 300, 400, "ACCEPTED", future, time.Now().UTC())
	s.Require().NoError(err)

	err = s.svc.ReturnOrder(ctx, "300")
	s.Require().Error(err)
	s.Assert().Contains(err.Error(), "expired")

	var status string
	err = s.masterPool.QueryRow(ctx,
		`SELECT status FROM orders WHERE id = $1`, 300,
	).Scan(&status)
	s.Require().NoError(err)
	s.Assert().Equal("ACCEPTED", status)

	var cnt int
	err = s.masterPool.QueryRow(ctx,
		`SELECT COUNT(*) FROM order_returns WHERE order_id = $1`, 300,
	).Scan(&cnt)
	s.Require().NoError(err)
	s.Assert().Zero(cnt)
}
