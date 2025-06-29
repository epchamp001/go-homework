//go:build integration

package integration

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"pvz-cli/internal/domain/codes"
	"pvz-cli/internal/domain/models"
	"pvz-cli/pkg/errs"
	"time"
)

func (s *TestSuite) TestReturnOrdersByClient_ErrNotFound() {
	ctx := context.Background()

	res, err := s.svc.ReturnOrdersByClient(ctx, "200", []string{"999"})
	require.NoError(s.T(), err)

	perOrderErr, ok := res["999"]
	require.True(s.T(), ok)
	require.ErrorIs(s.T(), perOrderErr, codes.ErrOrderNotFound)

	var cnt int
	err = s.masterPool.QueryRow(ctx, `SELECT COUNT(*) FROM orders`).Scan(&cnt)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 0, cnt)
}

func (s *TestSuite) TestReturnOrdersByClient_ErrTooLate() {
	s.loadFixtures()

	ctx := context.Background()
	_, err := s.masterPool.Exec(ctx, `TRUNCATE order_returns`)
	s.Require().NoError(err)

	res, err := s.svc.ReturnOrdersByClient(ctx, "200", []string{"109"})
	require.NoError(s.T(), err)

	perErr := res["109"]
	require.NotNil(s.T(), perErr)

	errStr := perErr.Error()
	assert.Contains(s.T(), errStr, errs.CodeValidationError)
	assert.Contains(s.T(), errStr, "return window expired")

	var status string
	err = s.masterPool.QueryRow(ctx,
		`SELECT status FROM orders WHERE id = $1`, 109,
	).Scan(&status)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), string(models.StatusIssued), status)

	var cnt int
	err = s.masterPool.QueryRow(ctx,
		`SELECT COUNT(*) FROM order_returns WHERE order_id = $1`, 109,
	).Scan(&cnt)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 0, cnt)
}

func (s *TestSuite) TestReturnOrdersByClient_ErrAlreadyReturned() {
	ctx := context.Background()

	now := time.Now()
	issuedAt := now.Add(-72 * time.Hour)   // выдан 72ч назад
	returnedAt := now.Add(-48 * time.Hour) // возвращён 48ч назад

	_, err := s.masterPool.Exec(ctx, `
        INSERT INTO orders 
           (id, user_id, status, expires_at, created_at, issued_at, returned_at,
            package, weight, price, total_price)
        VALUES
           ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
    `,
		500,
		200,
		models.StatusReturned,
		now.Add(24*time.Hour), // ещё не истёк срок хранения
		issuedAt.Add(-time.Hour),
		issuedAt,
		returnedAt,
		models.PackageBag,
		1.0,
		50000,
		50500,
	)
	require.NoError(s.T(), err)

	events := []struct {
		status string
		ts     time.Time
	}{
		{string(models.StatusAccepted), issuedAt.Add(-2 * time.Hour)},
		{string(models.StatusIssued), issuedAt},
		{string(models.StatusReturned), returnedAt},
	}
	for _, e := range events {
		_, err := s.masterPool.Exec(ctx, `
            INSERT INTO order_history (order_id, status, event_time)
            VALUES ($1, $2, $3)
        `, 500, e.status, e.ts)
		require.NoError(s.T(), err)
	}

	_, err = s.masterPool.Exec(ctx, `
        INSERT INTO order_returns (order_id, user_id, returned_at)
        VALUES ($1, $2, $3)
    `, 500, 200, returnedAt)
	require.NoError(s.T(), err)

	res, err := s.svc.ReturnOrdersByClient(ctx, "200", []string{"500"})
	require.NoError(s.T(), err)

	perErr := res["500"]
	require.NotNil(s.T(), perErr)

	errMsg := perErr.Error()
	assert.Contains(s.T(), errMsg, "order not in issued status")

	var status string
	err = s.masterPool.QueryRow(ctx,
		`SELECT status FROM orders WHERE id = $1`, 500,
	).Scan(&status)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), string(models.StatusReturned), status)

	var cnt int
	err = s.masterPool.QueryRow(ctx,
		`SELECT COUNT(*) FROM order_returns WHERE order_id = $1`, 500,
	).Scan(&cnt)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 1, cnt)
}
