package postgres

import (
	"context"
	"pvz-cli/internal/domain/models"
	"pvz-cli/internal/domain/vo"
	"pvz-cli/pkg/errs"
	"pvz-cli/pkg/txmanager"
)

type HistoryAndReturnsPostgresRepo struct {
	conn txmanager.TxManager
}

func NewHistoryAndReturnsPostgresRepo(conn txmanager.TxManager) *HistoryAndReturnsPostgresRepo {
	return &HistoryAndReturnsPostgresRepo{
		conn: conn,
	}
}

const (
	selectReturnsBase = `
		SELECT
			order_id, user_id, returned_at
		FROM order_returns`

	selectHistoryBase = `
		SELECT
			order_id, status, event_time
		FROM order_history`

	insertHistorySQL = `
        INSERT INTO order_history (
            order_id, status, event_time
        ) VALUES (
            $1, $2, $3
        );`

	insertReturnSQL = `
		INSERT INTO order_returns (
			order_id, user_id, returned_at
		) VALUES (
			$1, $2, $3
		);`

	selectReturnsByUserSQL = `
		SELECT
			order_id, user_id, returned_at
		FROM order_returns
		WHERE user_id = $1
		ORDER BY returned_at DESC, order_id DESC;`
)

func (r *HistoryAndReturnsPostgresRepo) ListReturns(
	ctx context.Context,
	pg vo.Pagination,
) ([]*models.ReturnRecord, error) {
	exec := r.conn.GetExecutor(ctx)

	limitClause, args := buildLimitOffset(pg, nil)
	sql := selectReturnsBase + `
		ORDER BY returned_at DESC, order_id DESC
		` + limitClause

	rows, err := exec.Query(ctx, sql, args...)
	if err != nil {
		return nil, errs.Wrap(err,
			errs.CodeDatabaseError, "failed to query returns")
	}
	records, err := scanReturnRecords(rows)
	if err != nil {
		return nil, errs.Wrap(err,
			errs.CodeDatabaseError, "failed to scan returns")
	}
	return records, nil
}

func (r *HistoryAndReturnsPostgresRepo) History(
	ctx context.Context,
	pg vo.Pagination,
) ([]*models.HistoryEvent, error) {
	exec := r.conn.GetExecutor(ctx)

	// page
	limitClause, args := buildLimitOffset(pg, nil)
	sql := selectHistoryBase + `
		ORDER BY event_time DESC, order_id DESC
		` + limitClause

	rows, err := exec.Query(ctx, sql, args...)
	if err != nil {
		return nil, errs.Wrap(err,
			errs.CodeDatabaseError, "failed to query history")
	}
	events, err := scanHistoryEvents(rows)
	if err != nil {
		return nil, errs.Wrap(err,
			errs.CodeDatabaseError, "failed to scan history")
	}

	return events, nil
}

func (r *HistoryAndReturnsPostgresRepo) AddHistory(ctx context.Context, e *models.HistoryEvent) error {
	exec := r.conn.GetExecutor(ctx)

	if _, err := exec.Exec(ctx,
		insertHistorySQL,
		e.OrderID, e.Status, e.Time,
	); err != nil {
		return errs.Wrap(err,
			errs.CodeDatabaseError,
			"failed to insert history event",
			"order_id", e.OrderID,
		)
	}
	return nil
}

func (r *HistoryAndReturnsPostgresRepo) AddReturn(ctx context.Context, rec *models.ReturnRecord) error {
	exec := r.conn.GetExecutor(ctx)

	if _, err := exec.Exec(ctx,
		insertReturnSQL,
		rec.OrderID, rec.UserID, rec.ReturnedAt,
	); err != nil {
		return errs.Wrap(err,
			errs.CodeDatabaseError,
			"failed to insert return record",
			"order_id", rec.OrderID,
		)
	}
	return nil
}

func (r *HistoryAndReturnsPostgresRepo) ListReturnsByUser(
	ctx context.Context,
	userID string,
) ([]*models.ReturnRecord, error) {
	exec := r.conn.GetExecutor(ctx)

	rows, err := exec.Query(ctx, selectReturnsByUserSQL, userID)
	if err != nil {
		return nil, errs.Wrap(err,
			errs.CodeDatabaseError, "failed to query returns by user", "user_id", userID)
	}
	return scanReturnRecords(rows)
}
