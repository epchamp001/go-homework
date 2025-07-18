package service

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"math/rand"
	"pvz-cli/internal/domain/models"
	"pvz-cli/internal/usecase"
	"pvz-cli/pkg/errs"
	"pvz-cli/pkg/txmanager"
	"strings"
	"time"
)

func (s *ServiceImpl) AcceptOrder(
	ctx context.Context,
	orderID, userID string,
	exp time.Time,
	weight float64,
	price models.PriceKopecks,
	pkgType models.PackageType,
) (models.PriceKopecks, error) {

	ctx, span := otel.Tracer("pvz-cli/usecase").Start(ctx, "UseCase.AcceptOrder",
		trace.WithAttributes(
			attribute.String("order.id", orderID),
			attribute.String("user.id", userID),
		))
	defer span.End()

	// валидация входных данных
	if err := validateAccept(orderID, userID, exp, weight); err != nil {
		return 0, errs.Wrap(err, errs.CodeValidationError, "validation failed")
	}

	// расчёт наценки
	strat, err := s.strategies.Strategy(pkgType)
	if err != nil {
		return 0, errs.Wrap(err, errs.CodeValidationError, "invalid package type")
	}
	if err := strat.Validate(weight); err != nil {
		return 0, errs.Wrap(err, errs.CodeValidationError, "weight validation failed")
	}

	total := price + strat.Surcharge()
	now := time.Now()
	o := &models.Order{
		ID:         orderID,
		UserID:     userID,
		Status:     models.StatusAccepted,
		ExpiresAt:  exp,
		CreatedAt:  now,
		Weight:     weight,
		Price:      price,
		TotalPrice: int64(total),
		Package:    pkgType,
	}

	ctxTx, txSpan := otel.Tracer("pvz-cli/usecase").Start(ctx, "DB.TX.AcceptOrder")
	err = s.tx.WithTx(
		ctxTx,
		txmanager.IsolationLevelReadCommitted,
		txmanager.AccessModeReadWrite,
		func(txCtx context.Context) error {
			if err := s.ordRepo.Create(txCtx, o); err != nil {
				return errs.Wrap(err, errs.CodeDatabaseError, "failed to create order", "order_id", orderID)
			}
			evt := &models.HistoryEvent{
				OrderID: o.ID,
				Status:  o.Status,
				Time:    now,
			}
			if err := s.hrRepo.AddHistory(txCtx, evt); err != nil {
				return errs.Wrap(err, errs.CodeDatabaseError, "failed to add history", "order_id", orderID)
			}

			randomCourierID := rand.Int63n(1_000_000_000) + 1
			kafkaEvt, err := models.NewOrderEvent(
				models.OrderAccepted,
				o.ID,
				strings.ToLower(string(o.Status)),
				o.UserID,
				models.Actor{
					Type: models.ActorCourier,
					ID:   randomCourierID,
				},
			)
			if err != nil {
				return errs.Wrap(err, errs.CodeInternalError, "failed to build order event", "order_id", orderID)
			}
			if err := s.outboxRepo.Add(txCtx, kafkaEvt); err != nil {
				return errs.Wrap(err, errs.CodeDatabaseError, "failed to enqueue order event", "order_id", orderID)
			}

			return nil
		},
	)
	txSpan.End()
	if err != nil {
		return 0, errs.Wrap(err, errs.CodeDBTransactionError, "transaction failed", "order_id", orderID)
	}

	s.metrics.IncOrdersAccepted()
	s.cache.Set(usecase.OrderKey(orderID), o)
	return total, nil
}
