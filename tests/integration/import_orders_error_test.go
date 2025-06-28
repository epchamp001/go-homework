package integration

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"pvz-cli/internal/domain/models"
	"time"
)

func (s *TestSuite) TestImportOrders_ErrEmptySlice() {
	ctx := context.Background()

	count, err := s.svc.ImportOrders(ctx, []*models.Order{})
	require.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "empty orders slice")
	assert.Equal(s.T(), 0, count)
}

func (s *TestSuite) TestImportOrders_ErrDuplicateID() {
	ctx := context.Background()

	s.loadFixtures()

	exp := time.Now().Add(24 * time.Hour)
	now := time.Now()
	orders := []*models.Order{
		{
			ID:         "800",
			UserID:     "300",
			Status:     models.StatusAccepted,
			ExpiresAt:  exp,
			CreatedAt:  now,
			Package:    models.PackageBag,
			Weight:     1.0,
			Price:      50000,
			TotalPrice: 50500,
		},
		{
			ID:         "801",
			UserID:     "301",
			Status:     models.StatusAccepted,
			ExpiresAt:  exp,
			CreatedAt:  now,
			Package:    models.PackageBox,
			Weight:     2.0,
			Price:      150000,
			TotalPrice: 152000,
		},
	}

	count, err := s.svc.ImportOrders(ctx, orders)

	require.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "import many failed")
	assert.Equal(s.T(), 0, count)

	var cnt int

	err = s.masterPool.QueryRow(ctx,
		`SELECT COUNT(*) FROM orders WHERE id = $1`, "800",
	).Scan(&cnt)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 1, cnt)

	err = s.masterPool.QueryRow(ctx,
		`SELECT COUNT(*) FROM orders WHERE id = $1`, "801",
	).Scan(&cnt)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 0, cnt)
}
