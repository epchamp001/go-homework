package service

import (
	"context"
	"pvz-cli/internal/domain/models"
	"pvz-cli/pkg/errs"
	"pvz-cli/pkg/txmanager"
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

	err = s.tx.WithTx(
		ctx,
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
			return nil
		},
	)
	if err != nil {
		return 0, errs.Wrap(err, errs.CodeDBTransactionError, "transaction failed", "order_id", orderID)
	}

	return total, nil
}
