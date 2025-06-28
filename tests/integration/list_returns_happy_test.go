package integration

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"pvz-cli/internal/domain/vo"
	"time"
)

func (s *TestSuite) TestListReturns_Happy() {
	s.loadFixtures()

	ctx := context.Background()
	pg := vo.Pagination{Page: 1, Limit: 10}

	recs, err := s.svc.ListReturns(ctx, pg)
	require.NoError(s.T(), err)

	require.Len(s.T(), recs, 2)

	assert.WithinDuration(s.T(),
		time.Date(2025, 6, 28, 11, 0, 0, 0, time.UTC),
		recs[0].ReturnedAt,
		time.Second,
	)
	assert.Equal(s.T(), "108", recs[0].OrderID)

	assert.WithinDuration(s.T(),
		time.Date(2025, 6, 27, 11, 0, 0, 0, time.UTC),
		recs[1].ReturnedAt,
		time.Second,
	)
	assert.Equal(s.T(), "109", recs[1].OrderID)
}

func (s *TestSuite) TestListReturns_Pagination() {
	s.loadFixtures()

	ctx := context.Background()

	pg1 := vo.Pagination{Page: 1, Limit: 1}
	first, err := s.svc.ListReturns(ctx, pg1)
	require.NoError(s.T(), err)
	require.Len(s.T(), first, 1)

	assert.Equal(s.T(), "108", first[0].OrderID)

	pg2 := vo.Pagination{Page: 2, Limit: 1}
	second, err := s.svc.ListReturns(ctx, pg2)
	require.NoError(s.T(), err)
	require.Len(s.T(), second, 1)

	assert.Equal(s.T(), "109", second[0].OrderID)
}

func (s *TestSuite) TestListReturns_Empty() {

	ctx := context.Background()
	pg := vo.Pagination{Page: 1, Limit: 10}

	recs, err := s.svc.ListReturns(ctx, pg)
	require.NoError(s.T(), err)
	assert.Empty(s.T(), recs)
}
