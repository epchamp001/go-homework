//go:build integration

package integration

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"pvz-cli/internal/domain/models"
	"time"
)

func (s *TestSuite) TestIssueOrders_Happy() {
	ctx := context.Background()

	_, err := s.svc.AcceptOrder(
		ctx,
		fmt.Sprint(orderID),
		fmt.Sprint(userID),
		time.Now().Add(48*time.Hour),
		1.0,
		priceK,
		models.PackageBag,
	)
	require.NoError(s.T(), err)

	res, err := s.svc.IssueOrders(ctx, fmt.Sprint(userID), []string{fmt.Sprint(orderID)})
	require.NoError(s.T(), err)

	require.Contains(s.T(), res, fmt.Sprint(orderID))
	assert.NoError(s.T(), res[fmt.Sprint(orderID)])

	var (
		status    string
		issuedAt  *time.Time
		histCount int
	)

	err = s.masterPool.QueryRow(ctx,
		`SELECT status, issued_at FROM orders WHERE id = $1`, orderID,
	).Scan(&status, &issuedAt)
	require.NoError(s.T(), err)

	assert.Equal(s.T(), models.StatusIssued, models.OrderStatus(status))
	require.NotNil(s.T(), issuedAt)
	assert.WithinDuration(s.T(), time.Now(), *issuedAt, 2*time.Second)

	err = s.masterPool.QueryRow(ctx,
		`SELECT COUNT(*) FROM order_history WHERE order_id = $1`, orderID,
	).Scan(&histCount)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 2, histCount)
}
