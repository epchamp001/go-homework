package service

import (
	"context"
	"math/rand"
	"pvz-cli/internal/domain/models"
	"pvz-cli/pkg/errs"
	"pvz-cli/pkg/txmanager"
	"pvz-cli/pkg/wpool"
	"time"
)

func (s *ServiceImpl) ReturnOrdersByClient(ctx context.Context, userID string, ids []string) (map[string]error, error) {
	result := make(map[string]error, len(ids))

	type res struct {
		id    string
		err   error
		txErr error
	}
	resCh := make(chan res, len(ids))

	now := time.Now()

	for _, orderID := range ids {
		oid := orderID

		s.wp.Submit(wpool.Job{
			Ctx:    ctx,
			Result: make(chan wpool.Response, 1),
			Do: func(c context.Context) (any, error) {

				var bisErr error

				errTx := s.tx.WithTx(
					c,
					txmanager.IsolationLevelRepeatableRead,
					txmanager.AccessModeReadWrite,
					func(txCtx context.Context) error {

						o, err := s.ordRepo.Get(txCtx, oid)
						if err != nil {
							bisErr = errs.Wrap(
								err, errs.CodeRecordNotFound,
								"order not found", "order_id", oid,
							)
							return nil
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
							bisErr = errs.Wrap(
								err, errs.CodeDatabaseError,
								"failed to add return record", "order_id", oid,
							)
							return nil
						}

						o.Status = models.StatusReturned
						o.ReturnedAt = &now
						if err := s.ordRepo.Update(txCtx, o); err != nil {
							bisErr = errs.Wrap(
								err, errs.CodeDatabaseError,
								"failed to mark order returned", "order_id", oid,
							)
							return nil
						}

						evt := &models.HistoryEvent{
							OrderID: o.ID,
							Status:  models.StatusReturned,
							Time:    now,
						}
						if err := s.hrRepo.AddHistory(txCtx, evt); err != nil {
							bisErr = errs.Wrap(
								err, errs.CodeDatabaseError,
								"failed to add history event", "order_id", oid,
							)
							return nil
						}

						randomClientID := rand.Int63n(1_000_000_000) + 1
						kafkaEvt, err := models.NewOrderEvent(
							models.OrderReturnedByClient,
							o.ID,
							"returned_by_client",
							o.UserID,
							models.Actor{Type: models.ActorClient, ID: randomClientID},
						)
						if err != nil {
							bisErr = errs.Wrap(err, errs.CodeInternalError,
								"failed to build client-return event", "order_id", oid)
							return nil
						}
						if err := s.outboxRepo.Add(txCtx, kafkaEvt); err != nil {
							bisErr = errs.Wrap(err, errs.CodeDatabaseError,
								"failed to enqueue return-by-client event", "order_id", oid)
							return nil
						}

						bisErr = nil
						return nil
					},
				)

				if errTx != nil {
					resCh <- res{
						id: oid,
						txErr: errs.Wrap(
							errTx, errs.CodeDBTransactionError,
							"return by client tx failed", "order_id", oid,
						),
					}
				} else {
					resCh <- res{id: oid, err: bisErr}
				}
				return nil, nil
			},
		})
	}

	var fatalTxErr error

	for i := 0; i < len(ids); i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()

		case r := <-resCh:
			result[r.id] = r.err
			if r.txErr != nil && fatalTxErr == nil {
				fatalTxErr = r.txErr
			}
		}
	}

	if fatalTxErr != nil {
		return nil, fatalTxErr
	}
	return result, nil
}
