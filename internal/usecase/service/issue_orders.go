package service

import (
	"context"
	"math/rand"
	"pvz-cli/internal/domain/models"
	"pvz-cli/internal/usecase"
	"pvz-cli/pkg/errs"
	"pvz-cli/pkg/txmanager"
	"pvz-cli/pkg/wpool"
	"strings"
	"time"
)

func (s *ServiceImpl) IssueOrders(ctx context.Context, userID string, ids []string) (map[string]error, error) {
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
			Ctx: ctx,
			Do: func(c context.Context) (any, error) {
				bErr, txErr := s.issueOne(c, orderID, userID, now)
				resCh <- result{id: orderID, bisErr: bErr, txErr: txErr}
				return nil, nil
			},
			Result: make(chan wpool.Response, 1),
		})
	}

	agg := make(map[string]error, len(ids))
	var fatalTxErr error

	for i := 0; i < len(ids); i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()

		case r := <-resCh:
			agg[r.id] = r.bisErr
			if r.txErr != nil && fatalTxErr == nil {
				fatalTxErr = errs.Wrap(r.txErr, errs.CodeDBTransactionError,
					"issue order tx failed", "order_id", r.id)
			}
		}
	}

	if fatalTxErr != nil {
		return nil, fatalTxErr
	}
	return agg, nil
}

func (s *ServiceImpl) issueOne(ctx context.Context, orderID, userID string, now time.Time) (bisErr, txErr error) {

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

			if err := validateIssue(o, userID, now); err != nil {
				bisErr = err
				return nil
			}

			o.Status, o.IssuedAt = models.StatusIssued, &now
			if err := s.ordRepo.Update(txCtx, o); err != nil {
				bisErr = errs.Wrap(err, errs.CodeDatabaseError,
					"failed to update order", "order_id", orderID)
				return nil
			}

			evt := &models.HistoryEvent{OrderID: o.ID, Status: models.StatusIssued, Time: now}
			if err := s.hrRepo.AddHistory(txCtx, evt); err != nil {
				bisErr = errs.Wrap(err, errs.CodeDatabaseError,
					"failed to add history", "order_id", orderID)
				return nil
			}

			courierID := rand.Int63n(1_000_000_000) + 1
			kafkaEvt, err := models.NewOrderEvent(
				models.OrderIssued, o.ID,
				strings.ToLower(string(o.Status)), o.UserID,
				models.Actor{Type: models.ActorCourier, ID: courierID},
			)
			if err != nil {
				bisErr = errs.Wrap(err, errs.CodeInternalError,
					"failed to build order event", "order_id", orderID)
				return nil
			}
			if err := s.outboxRepo.Add(txCtx, kafkaEvt); err != nil {
				bisErr = errs.Wrap(err, errs.CodeDatabaseError,
					"failed to enqueue order event", "order_id", orderID)
				return nil
			}

			cached = o
			return nil
		},
	)

	if txErr == nil && bisErr == nil {
		s.cache.Set(usecase.OrderKey(orderID), cached)
	}

	s.metrics.IncOrdersIssued()

	return
}
