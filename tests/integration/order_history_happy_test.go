//go:build integration

package integration

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"pvz-cli/internal/domain/models"
	"pvz-cli/internal/domain/vo"
	"time"
)

func (s *TestSuite) TestOrderHistory_Happy() {
	s.loadFixtures()

	ctx := context.Background()
	pg := vo.Pagination{Page: 1, Limit: 10}

	events, total, err := s.svc.OrderHistory(ctx, pg)
	require.NoError(s.T(), err)

	require.Len(s.T(), events, 7)
	assert.Equal(s.T(), 7, total)

	expected := []struct {
		OrderID string
		Status  models.OrderStatus
		Time    time.Time
	}{
		{"800", models.StatusAccepted, time.Date(2025, 6, 28, 10, 0, 0, 0, time.UTC)},
		{"300", models.StatusAccepted, time.Date(2025, 6, 28, 10, 0, 0, 0, time.UTC)},
		{"108", models.StatusIssued, time.Date(2025, 6, 27, 10, 0, 0, 0, time.UTC)},
		{"108", models.StatusAccepted, time.Date(2025, 6, 27, 9, 0, 0, 0, time.UTC)},
		{"109", models.StatusIssued, time.Date(2025, 6, 25, 10, 0, 0, 0, time.UTC)},
		{"109", models.StatusAccepted, time.Date(2025, 6, 25, 9, 0, 0, 0, time.UTC)},
		{"110", models.StatusAccepted, time.Date(2025, 6, 25, 0, 0, 0, 0, time.UTC)},
	}

	for i, exp := range expected {
		ev := events[i]
		assert.Equal(s.T(), exp.OrderID, ev.OrderID, "event #%d OrderID", i)
		assert.Equal(s.T(), exp.Status, ev.Status, "event #%d Status", i)
		assert.WithinDuration(s.T(), exp.Time, ev.Time, time.Second, "event #%d Time", i)
	}
}
