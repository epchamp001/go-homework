package integration

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"pvz-cli/internal/domain/models"
	"pvz-cli/internal/usecase/packaging"
	"time"
)

const (
	orderID uint64              = 100
	userID  uint64              = 200
	priceK  models.PriceKopecks = 500_00
)

func (s *TestSuite) TestAcceptOrder_Happy() {
	ctx := context.Background()

	expiresAt := time.Now().Add(72 * time.Hour)
	total, err := s.svc.AcceptOrder(
		ctx,
		fmt.Sprint(orderID),
		fmt.Sprint(userID),
		expiresAt,
		1.2,
		priceK,
		models.PackageBag,
	)
	require.NoError(s.T(), err)

	strategy := packaging.NewBagStrategy()
	pkgCost := strategy.Surcharge()
	expected := priceK + pkgCost
	assert.Equal(s.T(), expected, total)

	var (
		status     string
		totalDb    int64
		historyCnt int
	)
	err = s.masterPool.QueryRow(ctx,
		`SELECT status, total_price FROM orders WHERE id=$1`, orderID,
	).Scan(&status, &totalDb)
	require.NoError(s.T(), err)

	assert.Equal(s.T(), "ACCEPTED", status)
	assert.Equal(s.T(), int64(expected), totalDb)

	err = s.masterPool.QueryRow(ctx,
		`SELECT COUNT(*) FROM order_history WHERE order_id=$1`, orderID,
	).Scan(&historyCnt)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 1, historyCnt)
}
