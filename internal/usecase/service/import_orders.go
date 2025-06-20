package service

import (
	"context"
	"pvz-cli/internal/domain/models"
	"pvz-cli/pkg/errs"
	"pvz-cli/pkg/txmanager"
	"time"
)

func (s *ServiceImpl) ImportOrders(ctx context.Context, orders []*models.Order) (int, error) {

	if len(orders) == 0 {
		return 0, errs.New(errs.CodeValidationError, "empty orders slice")
	}

	errTx := s.tx.WithTx(
		ctx,
		txmanager.IsolationLevelReadCommitted,
		txmanager.AccessModeReadWrite,
		func(txCtx context.Context) error {
			if err := s.ordRepo.ImportMany(txCtx, orders); err != nil {
				return errs.Wrap(err,
					errs.CodeDatabaseError, "import many failed")
			}
			now := time.Now()
			for _, o := range orders {
				evt := &models.HistoryEvent{
					OrderID: o.ID,
					Status:  models.StatusAccepted,
					Time:    now,
				}
				if err := s.hrRepo.AddHistory(txCtx, evt); err != nil {
					return errs.Wrap(err, errs.CodeDatabaseError,
						"failed to add history for imported order", "order_id", o.ID)
				}
			}
			return nil
		},
	)
	if errTx != nil {
		return 0, errs.Wrap(errTx,
			errs.CodeDBTransactionError, "import orders tx failed")
	}

	return len(orders), nil
}
