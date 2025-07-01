package postgres

import (
	"context"
	"errors"
	"pvz-cli/internal/domain/codes"
	"pvz-cli/internal/domain/models"
	"pvz-cli/internal/domain/vo"
	"pvz-cli/pkg/errs"
	"pvz-cli/pkg/txmanager"
	"strings"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type OrdersPostgresRepo struct {
	conn txmanager.TxManager
}

func NewOrdersPostgresRepo(conn txmanager.TxManager) *OrdersPostgresRepo {
	return &OrdersPostgresRepo{
		conn: conn,
	}
}

const (
	insertOrderSQL = `
		INSERT INTO orders (
			id, user_id, status, expires_at,
			issued_at, returned_at, created_at,
			package, weight, price, total_price
		) VALUES (
			$1,$2,$3,$4,
			$5,$6,$7,
			$8,$9,$10,$11
		);`

	updateOrderSQL = `
		UPDATE orders SET
			user_id      = $2,
			status       = $3,
			expires_at   = $4,
			issued_at    = $5,
			returned_at  = $6,
			package      = $7,
			weight       = $8,
			price        = $9,
			total_price  = $10
		WHERE id = $1;`

	selectOrderSQL = `
		SELECT
			id, user_id, status, expires_at,
			issued_at, returned_at, created_at,
			package, weight, price, total_price
		FROM orders
		WHERE id = $1;`

	deleteOrderSQL = `DELETE FROM orders WHERE id = $1;`

	listAllSQL = `
		SELECT
			id, user_id, status, expires_at,
			issued_at, returned_at, created_at,
			package, weight, price, total_price
		FROM orders
		ORDER BY created_at ASC, id ASC;`
)

func (r *OrdersPostgresRepo) Create(ctx context.Context, o *models.Order) error {
	exec := r.conn.GetExecutor(ctx)

	_, err := exec.Exec(ctx,
		insertOrderSQL,
		o.ID, o.UserID, o.Status, o.ExpiresAt,
		o.IssuedAt, o.ReturnedAt, o.CreatedAt,
		o.Package, o.Weight, o.Price, o.TotalPrice,
	)
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
		return codes.ErrOrderAlreadyExists
	}

	return errs.Wrap(
		err,
		errs.CodeDatabaseError,
		"failed to insert order",
		"order_id", o.ID,
	)
}

func (r *OrdersPostgresRepo) Update(ctx context.Context, o *models.Order) error {
	exec := r.conn.GetExecutor(ctx)

	tag, err := exec.Exec(ctx, updateOrderSQL,
		o.ID, o.UserID, o.Status, o.ExpiresAt,
		o.IssuedAt, o.ReturnedAt,
		o.Package, o.Weight, o.Price, o.TotalPrice,
	)
	if err != nil {
		return errs.Wrap(err, errs.CodeDatabaseError,
			"failed to update order", "order_id", o.ID)
	}
	if tag.RowsAffected() == 0 {
		return codes.ErrOrderNotFound
	}
	return nil
}

func (r *OrdersPostgresRepo) Get(ctx context.Context, id string) (*models.Order, error) {
	exec := r.conn.GetExecutor(ctx)

	var o models.Order
	if err := exec.QueryRow(ctx, selectOrderSQL, id).Scan(
		&o.ID, &o.UserID, &o.Status, &o.ExpiresAt,
		&o.IssuedAt, &o.ReturnedAt, &o.CreatedAt,
		&o.Package, &o.Weight, &o.Price, &o.TotalPrice,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, codes.ErrOrderNotFound
		}
		return nil, errs.Wrap(err, errs.CodeDatabaseError,
			"failed to select order", "order_id", id)
	}
	return &o, nil
}

func (r *OrdersPostgresRepo) Delete(ctx context.Context, id string) error {
	exec := r.conn.GetExecutor(ctx)

	tag, err := exec.Exec(ctx, deleteOrderSQL, id)
	if err != nil {
		return errs.Wrap(err, errs.CodeDatabaseError,
			"failed to delete order", "order_id", id)
	}
	if tag.RowsAffected() == 0 {
		return codes.ErrOrderNotFound
	}
	return nil
}

func (r *OrdersPostgresRepo) ListByUser(
	ctx context.Context,
	userID string,
	onlyInPVZ bool,
	lastN int,
	pg *vo.Pagination,
) ([]*models.Order, error) {
	exec := r.conn.GetExecutor(ctx)

	where, fArgs := filterClause(userID, onlyInPVZ)

	sel := selectClause()
	ord := orderClause(false)
	pag, args := paginationClause(lastN, pg, fArgs)

	fullQ := strings.Join([]string{sel, where, ord, pag}, " ")
	rows, err := exec.Query(ctx, fullQ, args...)
	if err != nil {
		return nil, errs.Wrap(err, errs.CodeDatabaseError, "query failed")
	}

	orders, err := scanOrders(rows)
	if err != nil {
		return nil, errs.Wrap(err, errs.CodeDatabaseError, "scan failed")
	}
	return orders, nil
}

func (r *OrdersPostgresRepo) ImportMany(ctx context.Context, list []*models.Order) error {
	exec := r.conn.GetExecutor(ctx)

	for _, o := range list {
		if _, err := exec.Exec(ctx, insertOrderSQL,
			o.ID, o.UserID, o.Status, o.ExpiresAt,
			o.IssuedAt, o.ReturnedAt, o.CreatedAt,
			o.Package, o.Weight, o.Price, o.TotalPrice,
		); err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				return errs.New(errs.CodeRecordAlreadyExists,
					"order already exists", "orderID", o.ID)
			}
			return errs.Wrap(err, errs.CodeDatabaseError,
				"failed to import order", "orderID", o.ID)
		}
	}
	return nil
}

func (r *OrdersPostgresRepo) ListAllOrders(ctx context.Context) ([]*models.Order, error) {

	exec := r.conn.GetExecutor(ctx)

	rows, err := exec.Query(ctx, listAllSQL)
	if err != nil {
		return nil, errs.Wrap(err,
			errs.CodeDatabaseError, "failed to list all orders")
	}
	defer rows.Close()

	orders, err := scanOrders(rows)
	if err != nil {
		return nil, errs.Wrap(err,
			errs.CodeDatabaseError, "failed to scan orders")
	}
	return orders, nil
}
