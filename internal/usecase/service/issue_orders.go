package service

import (
	"context"
	"pvz-cli/internal/domain/models"
	"pvz-cli/pkg/errs"
	"pvz-cli/pkg/txmanager"
	"time"
)

func (s *ServiceImpl) IssueOrders(ctx context.Context, userID string, ids []string) (map[string]error, error) {
	result := make(map[string]error, len(ids))
	now := time.Now()

	for _, orderID := range ids {
		// каждая запись — своя транзакция
		errTx := s.tx.WithTx(
			ctx,
			txmanager.IsolationLevelRepeatableRead,
			txmanager.AccessModeReadWrite,
			func(txCtx context.Context) error {
				o, err := s.ordRepo.Get(txCtx, orderID)
				if err != nil {
					result[orderID] = errs.Wrap(err, errs.CodeRecordNotFound,
						"order not found", "order_id", orderID)
					return nil // не откатываем всю пачку. Норм? Или лучше "всё или ничего"?
				}

				if err := validateIssue(o, userID, now); err != nil {
					result[orderID] = err
					return nil
				}

				o.Status = models.StatusIssued
				o.IssuedAt = &now
				if err := s.ordRepo.Update(txCtx, o); err != nil {
					result[orderID] = errs.Wrap(err, errs.CodeDatabaseError,
						"failed to update order", "order_id", orderID)
					return nil
				}

				evt := &models.HistoryEvent{
					OrderID: o.ID,
					Status:  models.StatusIssued,
					Time:    now,
				}
				if err := s.hrRepo.AddHistory(txCtx, evt); err != nil {
					result[orderID] = errs.Wrap(err, errs.CodeDatabaseError,
						"failed to add history", "order_id", orderID)
					return nil
				}

				result[orderID] = nil
				return nil
			},
		)

		// системная ошибка транзакции
		if errTx != nil {
			return nil, errs.Wrap(errTx, errs.CodeDBTransactionError,
				"issue order tx failed", "order_id", orderID)
		}
	}
	return result, nil
}
