package service

import (
	"context"
	"pvz-cli/internal/domain/models"
	"pvz-cli/internal/domain/vo"
	"pvz-cli/pkg/errs"
	"pvz-cli/pkg/txmanager"
)

func (s *ServiceImpl) ListOrders(ctx context.Context, userID string, onlyInPVZ bool, lastN int, pg vo.Pagination) ([]*models.Order, int, error) {

	if userID == "" {
		return nil, 0, errs.New(errs.CodeValidationError, "empty user id")
	}

	var (
		paged []*models.Order
		total int
	)

	errTx := s.tx.WithTx(
		ctx,
		txmanager.IsolationLevelRepeatableRead,
		txmanager.AccessModeReadOnly,
		func(txCtx context.Context) error {

			// активные заказы
			active, err := s.ordRepo.ListByUser(
				txCtx, userID, onlyInPVZ, 0, nil, // nil => без лимита
			)
			if err != nil {
				return errs.Wrap(err, errs.CodeDatabaseError,
					"listByUser failed", "user_id", userID)
			}

			sortOrders(active)
			paged, total = paginate(active, lastN, pg)

			return nil
		},
	)

	if errTx != nil {
		return nil, 0, errs.Wrap(errTx, errs.CodeDBTransactionError,
			"list orders tx failed", "user_id", userID)
	}

	return paged, total, nil
}
