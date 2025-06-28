package integration

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"pvz-cli/internal/domain/vo"
)

func (s *TestSuite) TestListOrders_ErrEmptyUserID() {
	ctx := context.Background()
	pg := vo.Pagination{Page: 1, Limit: 10}

	_, _, err := s.svc.ListOrders(ctx, "", false, 0, pg)
	require.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "empty user id")
}
