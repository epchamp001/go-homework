package integration

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"pvz-cli/internal/domain/models"
	"pvz-cli/internal/domain/vo"
)

func (s *TestSuite) TestListOrders_OnlyInPVZ() {
	s.loadFixtures()

	ctx := context.Background()
	pg := vo.Pagination{Page: 1, Limit: 10}
	orders, total, err := s.svc.ListOrders(ctx, "400", true, 0, pg)
	require.NoError(s.T(), err)

	// только один заказ в ПВЗ у user=400
	assert.Equal(s.T(), 1, total)
	require.Len(s.T(), orders, 1)
	assert.Equal(s.T(), "300", orders[0].ID)
	assert.Equal(s.T(), models.StatusAccepted, orders[0].Status)
}

func (s *TestSuite) TestListOrders_LastN() {
	s.loadFixtures()

	ctx := context.Background()
	pg := vo.Pagination{Page: 1, Limit: 10}

	orders, total, err := s.svc.ListOrders(ctx, "200", false, 2, pg)
	require.NoError(s.T(), err)

	assert.Equal(s.T(), 3, total)
	require.Len(s.T(), orders, 2)

	assert.Equal(s.T(), "109", orders[0].ID)
	assert.Equal(s.T(), "108", orders[1].ID)
}
