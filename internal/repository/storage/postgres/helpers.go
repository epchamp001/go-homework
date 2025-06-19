package postgres

import (
	"fmt"
	"github.com/jackc/pgx/v5"
	"pvz-cli/internal/domain/models"
	"pvz-cli/internal/domain/vo"
	"strings"
)

// selectClause возвращает основу SELECT-части без WHERE
func selectClause() string {
	return `
SELECT
	id, user_id, status, expires_at,
	issued_at, returned_at, created_at,
	package, weight, price, total_price
FROM orders`
}

// filterClause собирает WHERE-часть и параметры
func filterClause(userID string, onlyInPVZ bool) (string, []any) {
	conds := []string{"user_id = $1"}
	args := []any{userID}

	if onlyInPVZ {
		conds = append(conds, "status = 'ACCEPTED'")
	}
	return "WHERE " + strings.Join(conds, " AND "), args
}

// orderClause строит ORDER BY.
func orderClause(asc bool) string {
	dir := "DESC"
	if asc {
		dir = "ASC"
	}
	return fmt.Sprintf("ORDER BY created_at %s, id %s", dir, dir)
}

func isEmptyPaging(p *vo.Pagination) bool {
	if p == nil {
		return true
	}
	return p.Page == 0 && p.Limit == 0
}

// paginationClause строит LIMIT/OFFSET, возвращая новую args
func paginationClause(lastN int, pg *vo.Pagination, args []any) (string, []any) {
	if lastN > 0 {
		args = append(args, lastN)
		return fmt.Sprintf("LIMIT $%d", len(args)), args
	}
	if isEmptyPaging(pg) {
		return "", args
	}
	// sane defaults
	if pg.Page <= 0 {
		pg.Page = 1
	}
	if pg.Limit <= 0 {
		pg.Limit = 20
	}
	offset := (pg.Page - 1) * pg.Limit
	args = append(args, pg.Limit, offset)
	return fmt.Sprintf("LIMIT $%d OFFSET $%d", len(args)-1, len(args)), args
}

// scanOrders конвертирует pgx.Rows -> []*models.Order
func scanOrders(rows pgx.Rows) ([]*models.Order, error) {
	defer rows.Close()

	var out []*models.Order
	for rows.Next() {
		var o models.Order
		if err := rows.Scan(
			&o.ID, &o.UserID, &o.Status, &o.ExpiresAt,
			&o.IssuedAt, &o.ReturnedAt, &o.CreatedAt,
			&o.Package, &o.Weight, &o.Price, &o.TotalPrice,
		); err != nil {
			return nil, err
		}
		out = append(out, &o)
	}
	return out, nil
}

// cursorWhereClause возвращает WHERE и срез args
func cursorWhereClause(userID, lastID string) (string, []any) {
	where := "WHERE user_id = $1"
	args := []any{userID}

	if lastID != "" {
		where += `
		  AND (created_at, id) > (
		        SELECT created_at, id FROM orders WHERE id = $2
		      )`
		args = append(args, lastID)
	}
	return where, args
}

// cursorQuery собирает финальную строку запроса
func cursorQuery(where string, args []any) string {
	return fmt.Sprintf(`
		%s
		%s
		ORDER BY created_at ASC, id ASC
		LIMIT $%d;`, selectClause(), where, len(args))
}

// scanReturnRecords конвертирует pgx.Rows → []*models.ReturnRecord
func scanReturnRecords(rows pgx.Rows) ([]*models.ReturnRecord, error) {
	defer rows.Close()

	var out []*models.ReturnRecord
	for rows.Next() {
		var rr models.ReturnRecord
		if err := rows.Scan(&rr.OrderID, &rr.UserID, &rr.ReturnedAt); err != nil {
			return nil, err
		}
		out = append(out, &rr)
	}
	return out, nil
}

// scanHistoryEvents конвертирует pgx.Rows → []*models.HistoryEvent.
func scanHistoryEvents(rows pgx.Rows) ([]*models.HistoryEvent, error) {
	defer rows.Close()

	var out []*models.HistoryEvent
	for rows.Next() {
		var he models.HistoryEvent
		if err := rows.Scan(&he.OrderID, &he.Status, &he.Time); err != nil {
			return nil, err
		}
		out = append(out, &he)
	}
	return out, nil
}

// buildLimitOffset возвращает LIMIT/OFFSET и обновляет args
func buildLimitOffset(pg vo.Pagination, args []any) (string, []any) {
	if pg.Page <= 0 {
		pg.Page = 1
	}
	if pg.Limit <= 0 {
		pg.Limit = 20
	}
	offset := (pg.Page - 1) * pg.Limit
	args = append(args, pg.Limit, offset)
	clause := fmt.Sprintf("LIMIT $%d OFFSET $%d", len(args)-1, len(args))
	return clause, args
}
