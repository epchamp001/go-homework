package service

import (
	"context"
	"pvz-cli/internal/domain/models"
	"pvz-cli/internal/domain/vo"
	"pvz-cli/pkg/errs"
)

func (s *ServiceImpl) ScrollOrders(ctx context.Context, userID string, cur vo.ScrollCursor) ([]*models.Order, vo.ScrollCursor, error) {
	if userID == "" {
		return nil, vo.ScrollCursor{}, errs.New(
			errs.CodeValidationError, "empty user id",
		)
	}
	roCtx := s.tx.WithReadOnly(ctx)
	orders, next, err := s.ordRepo.NextBatchAfter(roCtx, userID, cur)
	if err != nil {
		return nil, vo.ScrollCursor{}, errs.Wrap(err, errs.CodeDatabaseError,
			"next batch query failed", "user_id", userID)
	}

	return orders, next, nil
}
