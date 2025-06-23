package txmanager

import (
	"context"
	"errors"
	"pvz-cli/pkg/logger"
	"sync/atomic"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Executor interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type TxManager interface {
	GetExecutor(ctx context.Context) Executor
	WithTx(ctx context.Context, isoLevel pgx.TxIsoLevel, accessMode pgx.TxAccessMode, fn func(ctx context.Context) error) error
	WithReadOnly(ctx context.Context) context.Context
}

var (
	IsolationLevelSerializable   = pgx.Serializable
	IsolationLevelReadCommitted  = pgx.ReadCommitted
	IsolationLevelRepeatableRead = pgx.RepeatableRead

	AccessModeReadWrite = pgx.ReadWrite
	AccessModeReadOnly  = pgx.ReadOnly
)

type txKey struct{}

type Transactor struct {
	writePool *pgxpool.Pool
	readPools []*pgxpool.Pool
	rrCounter uint64
	logger    logger.Logger
}

func NewTransactor(writePool *pgxpool.Pool, readPools []*pgxpool.Pool, logger logger.Logger) *Transactor {
	return &Transactor{
		writePool: writePool,
		readPools: readPools,
		rrCounter: 0,
		logger:    logger,
	}
}

func (t *Transactor) poolForMode(mode pgx.TxAccessMode) *pgxpool.Pool {
	if mode == AccessModeReadOnly {
		return t.pickRead()
	}
	return t.writePool
}

func (t *Transactor) pickRead() *pgxpool.Pool {
	n := len(t.readPools)
	if n == 0 {
		return t.writePool
	}
	idx := atomic.AddUint64(&t.rrCounter, 1) - 1
	return t.readPools[int(idx%uint64(n))]
}

// injectTx injects transaction to context
func injectTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

// extractTx extracts transaction from context
func extractTx(ctx context.Context) pgx.Tx {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return tx
	}
	return nil
}

func (t *Transactor) WithTx(ctx context.Context, isoLevel pgx.TxIsoLevel, accessMode pgx.TxAccessMode, fn func(ctx context.Context) error) (err error) {
	opts := pgx.TxOptions{
		IsoLevel:   isoLevel,
		AccessMode: accessMode,
	}

	poolToUse := t.poolForMode(accessMode)

	tx, err := poolToUse.BeginTx(ctx, opts)
	if err != nil {
		t.logger.Errorw("Failed to begin transaction",
			"error", err,
			"isoLevel", isoLevel,
			"accessMode", accessMode,
		)
		return err
	}

	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(ctx); rbErr != nil && !errors.Is(rbErr, pgx.ErrTxClosed) {
				t.logger.Errorw("Failed to rollback transaction",
					"error", rbErr,
				)
			}
		}
	}()

	ctx = injectTx(ctx, tx)

	if err = fn(ctx); err != nil {
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		t.logger.Errorw("Failed to commit transaction",
			"error", err,
		)
	}
	return err
}

func (t *Transactor) GetExecutor(ctx context.Context) Executor {
	if tx := extractTx(ctx); tx != nil {
		return tx
	}

	if mode, ok := extractAccessMode(ctx); ok && mode == AccessModeReadOnly {
		return t.pickRead()
	}

	// default — master
	return t.writePool
}
