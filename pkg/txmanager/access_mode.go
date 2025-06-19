package txmanager

import (
	"context"
	"github.com/jackc/pgx/v5"
)

type accessModeKey struct{}

func (t *Transactor) WithReadOnly(ctx context.Context) context.Context {
	return injectAccessMode(ctx, AccessModeReadOnly)
}

// injectAccessMode returns a new Context that carries the given access mode.
func injectAccessMode(ctx context.Context, mode pgx.TxAccessMode) context.Context {
	return context.WithValue(ctx, accessModeKey{}, mode)
}

// extractAccessMode retrieves the access mode from the Context.
// The second return value is false if no access mode was set.
func extractAccessMode(ctx context.Context) (pgx.TxAccessMode, bool) {
	mode, ok := ctx.Value(accessModeKey{}).(pgx.TxAccessMode)
	return mode, ok
}
