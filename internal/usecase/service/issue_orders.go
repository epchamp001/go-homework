package service

import (
	"context"
	"math/rand"
	"pvz-cli/internal/domain/models"
	"pvz-cli/pkg/errs"
	"pvz-cli/pkg/txmanager"
	"pvz-cli/pkg/wpool"
	"strings"
	"time"
)

func (s *ServiceImpl) IssueOrders(ctx context.Context, userID string, ids []string) (map[string]error, error) {

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
			Result: make(chan wpool.Response, 1), // заглушка
			Do: func(c context.Context) (any, error) {

				var bisErr error

				errTx := s.tx.WithTx(
					c,
					txmanager.IsolationLevelRepeatableRead,
					txmanager.AccessModeReadWrite,
					func(txCtx context.Context) error {

						o, err := s.ordRepo.Get(txCtx, oid)
						if err != nil {
							bisErr = errs.Wrap(err, errs.CodeRecordNotFound,
								"order not found", "order_id", oid)
							return nil
						}

						if err := validateIssue(o, userID, now); err != nil {
							bisErr = err
							return nil
						}

						o.Status = models.StatusIssued
						o.IssuedAt = &now
						if err := s.ordRepo.Update(txCtx, o); err != nil {
							bisErr = errs.Wrap(err, errs.CodeDatabaseError,
								"failed to update order", "order_id", oid)
							return nil
						}

						evt := &models.HistoryEvent{
							OrderID: o.ID,
							Status:  models.StatusIssued,
							Time:    now,
						}
						if err := s.hrRepo.AddHistory(txCtx, evt); err != nil {
							bisErr = errs.Wrap(err, errs.CodeDatabaseError,
								"failed to add history", "order_id", oid)
							return nil
						}

						randomCourierID := rand.Int63n(1_000_000_000) + 1
						kafkaEvt, err := models.NewOrderEvent(
							models.OrderIssued,
							o.ID,
							strings.ToLower(string(o.Status)),
							o.UserID,
							models.Actor{Type: models.ActorCourier, ID: randomCourierID},
						)
						if err != nil {
							bisErr = errs.Wrap(err, errs.CodeInternalError,
								"failed to build order event", "order_id", oid)
							return nil
						}
						if err := s.outboxRepo.Add(txCtx, kafkaEvt); err != nil {
							bisErr = errs.Wrap(err, errs.CodeDatabaseError,
								"failed to enqueue order event", "order_id", oid)
							return nil
						}

						bisErr = nil // успех
						return nil
					},
				)

				if errTx != nil {
					resCh <- res{
						id: oid,
						txErr: errs.Wrap(errTx, errs.CodeDBTransactionError,
							"issue order tx failed", "order_id", oid),
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
