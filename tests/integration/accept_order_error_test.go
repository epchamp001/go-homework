//go:build integration

package integration

import (
	"context"
	"pvz-cli/internal/domain/codes"
	"pvz-cli/internal/domain/models"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (s *TestSuite) TestAcceptOrder_ErrDuplicateID() {
	ctx := context.Background()

	_, err := s.svc.AcceptOrder(
		ctx, "100", "200",
		time.Now().Add(24*time.Hour),
		1.0, 500_00, models.PackageBag,
	)
	require.NoError(s.T(), err)

	_, err = s.svc.AcceptOrder(
		ctx, "100", "201",
		time.Now().Add(24*time.Hour),
		1.0, 500_00, models.PackageBag,
	)

	require.ErrorIs(s.T(), err, codes.ErrOrderAlreadyExists)

	var cnt int
	err = s.masterPool.QueryRow(ctx,
		`SELECT COUNT(*) FROM orders WHERE id = 100`,
	).Scan(&cnt)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 1, cnt)

	err = s.masterPool.QueryRow(ctx,
		`SELECT COUNT(*) FROM order_history WHERE order_id = 100`,
	).Scan(&cnt)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 1, cnt)
}

func (s *TestSuite) TestAcceptOrder_ErrWeightTooHeavy() {
	ctx := context.Background()

	_, err := s.svc.AcceptOrder(
		ctx, "101", "300",
		time.Now().Add(48*time.Hour),
		25.0,
		1_000_00,
		models.PackageBag,
	)

	require.ErrorIs(s.T(), err, codes.ErrWeightTooHeavy)

	var cnt int
	err = s.masterPool.
		QueryRow(ctx, `SELECT COUNT(*) FROM orders WHERE id = 101`).
		Scan(&cnt)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 0, cnt)

	err = s.masterPool.
		QueryRow(ctx, `SELECT COUNT(*) FROM order_history WHERE order_id = 101`).
		Scan(&cnt)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 0, cnt)
}
