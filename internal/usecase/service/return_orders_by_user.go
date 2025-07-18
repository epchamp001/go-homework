package service

import (
	"context"
	"math/rand"
	"pvz-cli/internal/domain/models"
	"pvz-cli/internal/usecase"
	"pvz-cli/pkg/errs"
	"pvz-cli/pkg/txmanager"
	"pvz-cli/pkg/wpool"
	"time"
)

func (s *ServiceImpl) ReturnOrdersByClient(
	ctx context.Context,
	userID string,
	ids []string,
) (map[string]error, error) {

	type result struct {
		id     string
		bisErr error
		txErr  error
	}
	resCh := make(chan result, len(ids))

	now := time.Now()

	for _, id := range ids {
		orderID := id

		s.wp.Submit(wpool.Job{
			Ctx:    ctx,
			Result: make(chan wpool.Response, 1), // важно!
			Do: func(c context.Context) (any, error) {
				bErr, txErr := s.returnOneByClient(c, orderID, userID, now)
				resCh <- result{id: orderID, bisErr: bErr, txErr: txErr}
				return nil, nil
			},
		})
	}

	agg := make(map[string]error, len(ids))
	var fatalTx error

	for i := 0; i < len(ids); i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()

		case r := <-resCh:
			agg[r.id] = r.bisErr
			if r.txErr != nil && fatalTx == nil {
				fatalTx = errs.Wrap(r.txErr, errs.CodeDBTransactionError,
					"return by client tx failed", "order_id", r.id)
			}
		}
	}

	if fatalTx != nil {
		return nil, fatalTx
	}
	return agg, nil
}

func (s *ServiceImpl) returnOneByClient(
	ctx context.Context, orderID, userID string, now time.Time,
) (bisErr, txErr error) {

	cached, cacheHit := s.cache.Get(usecase.OrderKey(orderID))

	txErr = s.tx.WithTx(
		ctx,
		txmanager.IsolationLevelRepeatableRead,
		txmanager.AccessModeReadWrite,
		func(txCtx context.Context) error {

			var o *models.Order
			if cacheHit {
				o = cached
			} else {
				var err error
				o, err = s.ordRepo.Get(txCtx, orderID)
				if err != nil {
					bisErr = errs.Wrap(err, errs.CodeRecordNotFound,
						"order not found", "order_id", orderID)
					return nil
				}
			}

			if err := validateClientReturn(o, userID, now); err != nil {
				bisErr = err
				return nil
			}

			rec := &models.ReturnRecord{
				OrderID:    o.ID,
				UserID:     o.UserID,
				ReturnedAt: now,
			}
			if err := s.hrRepo.AddReturn(txCtx, rec); err != nil {
				bisErr = errs.Wrap(err, errs.CodeDatabaseError,
					"failed to add return record", "order_id", orderID)
				return nil
			}

			o.Status, o.ReturnedAt = models.StatusReturned, &now
			if err := s.ordRepo.Update(txCtx, o); err != nil {
				bisErr = errs.Wrap(err, errs.CodeDatabaseError,
					"failed to mark order returned", "order_id", orderID)
				return nil
			}

			hEvt := &models.HistoryEvent{OrderID: o.ID, Status: models.StatusReturned, Time: now}
			if err := s.hrRepo.AddHistory(txCtx, hEvt); err != nil {
				bisErr = errs.Wrap(err, errs.CodeDatabaseError,
					"failed to add history event", "order_id", orderID)
				return nil
			}

			clientID := rand.Int63n(1_000_000_000) + 1
			outEvt, err := models.NewOrderEvent(
				models.OrderReturnedByClient,
				o.ID,
				"returned_by_client",
				o.UserID,
				models.Actor{Type: models.ActorClient, ID: clientID},
			)
			if err != nil {
				bisErr = errs.Wrap(err, errs.CodeInternalError,
					"failed to build client-return event", "order_id", orderID)
				return nil
			}
			if err := s.outboxRepo.Add(txCtx, outEvt); err != nil {
				bisErr = errs.Wrap(err, errs.CodeDatabaseError,
					"failed to enqueue client-return event", "order_id", orderID)
				return nil
			}

			cached = o
			return nil
		},
	)

	if txErr == nil && bisErr == nil {
		s.cache.Set(usecase.OrderKey(orderID), cached)
	}
	return
}
