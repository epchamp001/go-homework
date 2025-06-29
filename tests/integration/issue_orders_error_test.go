//go:build integration

package integration

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"pvz-cli/internal/domain/codes"
	"pvz-cli/internal/domain/models"
	"time"
)

func (s *TestSuite) TestIssueOrders_ErrNotFound() {
	ctx := context.Background()

	res, err := s.svc.IssueOrders(
		ctx,
		"200",
		[]string{"999"},
	)
	require.NoError(s.T(), err)

	perOrderErr, ok := res["999"]
	require.True(s.T(), ok)
	require.ErrorIs(s.T(), perOrderErr, codes.ErrOrderNotFound)

	var cnt int
	require.NoError(s.T(),
		s.masterPool.QueryRow(ctx,
			`SELECT COUNT(*) FROM orders`).Scan(&cnt))
	assert.Equal(s.T(), 0, cnt)
}

func (s *TestSuite) TestIssueOrders_ErrWrongUser() {
	s.loadFixtures()

	ctx := context.Background()
	res, err := s.svc.IssueOrders(ctx, "200", []string{"300"})
	require.NoError(s.T(), err)

	perOrderErr := res["300"]
	require.NotNil(s.T(), perOrderErr)

	assert.Contains(s.T(), perOrderErr.Error(), "order belongs to another user")

	var st string
	var issuedAt *time.Time
	require.NoError(s.T(), s.masterPool.
		QueryRow(ctx, `SELECT status, issued_at FROM orders WHERE id='300'`).
		Scan(&st, &issuedAt))
	assert.Equal(s.T(), string(models.StatusAccepted), st)
	assert.Nil(s.T(), issuedAt)

	var hc int
	require.NoError(s.T(), s.masterPool.
		QueryRow(ctx, `SELECT COUNT(*) FROM order_history WHERE order_id='300'`).
		Scan(&hc))
	assert.Equal(s.T(), 1, hc)
}
