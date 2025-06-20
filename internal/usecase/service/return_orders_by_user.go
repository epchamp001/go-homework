package service

import (
	"context"
	"pvz-cli/internal/domain/models"
	"pvz-cli/pkg/errs"
	"pvz-cli/pkg/txmanager"
	"time"
)

func (s *ServiceImpl) ReturnOrdersByClient(ctx context.Context, userID string, ids []string) (map[string]error, error) {
	result := make(map[string]error, len(ids))
	now := time.Now()

	for _, orderID := range ids {
		// отдельная транзакция на каждый заказ
		errTx := s.tx.WithTx(
			ctx,
			txmanager.IsolationLevelRepeatableRead,
			txmanager.AccessModeReadWrite,
			func(txCtx context.Context) error {
				o, err := s.ordRepo.Get(txCtx, orderID)
				if err != nil {
					result[orderID] = errs.Wrap(
						err, errs.CodeRecordNotFound,
						"order not found", "order_id", orderID,
					)
					return nil
				}

				if err := validateClientReturn(o, userID, now); err != nil {
					result[orderID] = err
					return nil
				}

				rec := &models.ReturnRecord{
					OrderID:    o.ID,
					UserID:     o.UserID,
					ReturnedAt: now,
				}
				if err := s.hrRepo.AddReturn(txCtx, rec); err != nil {
					result[orderID] = errs.Wrap(
						err, errs.CodeDatabaseError,
						"failed to add return record", "order_id", orderID,
					)
					return nil
				}

				o.Status = models.StatusReturned
				o.ReturnedAt = &now
				if err := s.ordRepo.Update(txCtx, o); err != nil {
					result[orderID] = errs.Wrap(
						err, errs.CodeDatabaseError,
						"failed to mark order returned", "order_id", orderID,
					)
					return nil
				}

				evt := &models.HistoryEvent{
					OrderID: o.ID,
					Status:  models.StatusReturned,
					Time:    now,
				}
				if err := s.hrRepo.AddHistory(txCtx, evt); err != nil {
					result[orderID] = errs.Wrap(
						err, errs.CodeDatabaseError,
						"failed to add history event", "order_id", orderID,
					)
					return nil
				}

				result[orderID] = nil
				return nil
			},
		)

		if errTx != nil {
			return nil, errs.Wrap(
				errTx, errs.CodeDBTransactionError,
				"return by client tx failed", "order_id", orderID,
			)
		}
	}

	return result, nil
}
