//go:build integration

package integration

import (
	"context"
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

func (s *TestSuite) TestReturnOrder_ErrNotExpired() {
	// неистёкший заказ id=300
	s.loadFixtures()

	ctx := context.Background()
	err := s.svc.ReturnOrder(ctx, "300")
	require.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "expired")

	var status string
	err = s.masterPool.QueryRow(ctx,
		`SELECT status FROM orders WHERE id = $1`, 300,
	).Scan(&status)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "ACCEPTED", status)

	var cnt int
	err = s.masterPool.QueryRow(ctx,
		`SELECT COUNT(*) FROM order_returns WHERE order_id = $1`, 300,
	).Scan(&cnt)
	require.NoError(s.T(), err)
	assert.Zero(s.T(), cnt)
}
