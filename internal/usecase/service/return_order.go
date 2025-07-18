package service

import (
	"context"
	"math/rand"
	"pvz-cli/internal/domain/models"
	"pvz-cli/internal/usecase"
	"pvz-cli/pkg/errs"
	"pvz-cli/pkg/txmanager"
	"time"
)

func (s *ServiceImpl) ReturnOrder(ctx context.Context, orderID string) error {
	if orderID == "" {
		return errs.New(errs.CodeValidationError, "empty order id")
	}

	var cached *models.Order

	if o, ok := s.cache.Get(usecase.OrderKey(orderID)); ok {
		cached = o
		if err := validateReturn(o); err != nil {
			return err
		}
	}

	err := s.tx.WithTx(
		ctx,
		txmanager.IsolationLevelRepeatableRead,
		txmanager.AccessModeReadWrite,
		func(txCtx context.Context) error {
			var o *models.Order
			if cached != nil {
				o = cached
			} else {
				var err error
				o, err = s.ordRepo.Get(txCtx, orderID)
				if err != nil {
					return errs.Wrap(err, errs.CodeRecordNotFound, "order not found", "order_id", orderID)
				}

				if err := validateReturn(o); err != nil {
					return err
				}
			}

			now := time.Now()

			rec := &models.ReturnRecord{
				OrderID:    o.ID,
				UserID:     o.UserID,
				ReturnedAt: now,
			}
			if err := s.hrRepo.AddReturn(txCtx, rec); err != nil {
				return errs.Wrap(err, errs.CodeDatabaseError, "failed to add return", "order_id", orderID)
			}

			evt := &models.HistoryEvent{
				OrderID: o.ID,
				Status:  models.StatusReturned,
				Time:    now,
			}
			if err := s.hrRepo.AddHistory(txCtx, evt); err != nil {
				return errs.Wrap(err, errs.CodeDatabaseError, "failed to add history", "order_id", orderID)
			}

			o.Status = models.StatusReturned
			o.ReturnedAt = &now
			if err := s.ordRepo.Update(txCtx, o); err != nil {
				return errs.Wrap(err, errs.CodeDatabaseError,
					"failed to mark order returned", "order_id", orderID)
			}

			randomCourierID := rand.Int63n(1_000_000_000) + 1
			kafkaEvt, err := models.NewOrderEvent(
				models.OrderReturnedToCourier,
				o.ID,
				"returned_to_courier",
				o.UserID,
				models.Actor{Type: models.ActorCourier, ID: randomCourierID},
			)
			if err != nil {
				return errs.Wrap(err, errs.CodeInternalError,
					"failed to build return event", "order_id", orderID)
			}
			if err := s.outboxRepo.Add(txCtx, kafkaEvt); err != nil {
				return errs.Wrap(err, errs.CodeDatabaseError,
					"failed to enqueue return event", "order_id", orderID)
			}

			cached = o
			return nil
		},
	)
	if err != nil {
		return errs.Wrap(err, errs.CodeDBTransactionError, "return order tx failed", "order_id", orderID)
	}

	if cached != nil {
		s.cache.Set(usecase.OrderKey(orderID), cached)
	}

	return nil
}
