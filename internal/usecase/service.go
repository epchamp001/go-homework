// Package usecase содержит реализацию бизнес-логики приложения.
package usecase

import (
	"context"
	"fmt"
	"pvz-cli/pkg/errs"
	"pvz-cli/pkg/txmanager"
	"time"

	"github.com/xuri/excelize/v2"
	"pvz-cli/internal/domain/models"
	"pvz-cli/internal/domain/vo"
	"pvz-cli/internal/usecase/packaging"
)

// Service определяет бизнес-логику работы Пункта Выдачи Заказов.
type Service interface {
	// AcceptOrder регистрирует новый заказ и рассчитывает итоговую стоимость с учётом упаковки.
	AcceptOrder(ctx context.Context, orderID, userID string, expires time.Time, weight float64, price models.PriceKopecks, pkgType models.PackageType) (models.PriceKopecks, error)

	// ReturnOrder выполняет возврат заказа по его ID (если срок хранения истёк и не был выдан).
	ReturnOrder(ctx context.Context, orderID string) error

	// IssueOrders выполняет массовую выдачу заказов клиенту.
	IssueOrders(ctx context.Context, userID string, ids []string) (map[string]error, error)

	// ReturnOrdersByClient обрабатывает массовый возврат заказов клиентом в течение 48 часов после выдачи.
	ReturnOrdersByClient(ctx context.Context, userID string, ids []string) (map[string]error, error)

	// ListOrders возвращает заказы клиента: активные и (опционально) возвращённые,
	// с фильтрацией onlyInPVZ, lastN, или обычной пагинацией.
	ListOrders(ctx context.Context, userID string, onlyInPVZ bool, lastN int, pg vo.Pagination) ([]*models.Order, int, error)

	// ScrollOrders возвращает порцию заказов по курсору (key-set пагинация).
	ScrollOrders(ctx context.Context, userID string, cursor vo.ScrollCursor) ([]*models.Order, vo.ScrollCursor, error)

	// ListReturns возвращает список возвратов (с пагинацией).
	ListReturns(ctx context.Context, pg vo.Pagination) ([]*models.ReturnRecord, error)

	// OrderHistory возвращает историю событий по заказам (с пагинацией).
	OrderHistory(ctx context.Context, pg vo.Pagination) ([]*models.HistoryEvent, int, error)

	// ImportOrders импортирует пачку заказов в одну транзакцию.
	ImportOrders(ctx context.Context, orders []*models.Order) (int, error)

	// GenerateClientReportByte генерирует .xlsx-отчёт по клиентам.
	GenerateClientReportByte(ctx context.Context, sortBy string) ([]byte, error)
}

type ServiceImpl struct {
	tx      txmanager.TxManager
	ordRepo OrdersRepository
	hrRepo  HistoryAndReturnsRepository
}

func NewService(tx txmanager.TxManager, ordRepo OrdersRepository, hrRepo HistoryAndReturnsRepository) *ServiceImpl {
	return &ServiceImpl{
		tx:      tx,
		ordRepo: ordRepo,
		hrRepo:  hrRepo,
	}
}

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
	strat, err := packaging.GetStrategy(pkgType)
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

func (s *ServiceImpl) ReturnOrder(ctx context.Context, orderID string) error {
	if orderID == "" {
		return errs.New(errs.CodeValidationError, "empty order id")
	}

	err := s.tx.WithTx(
		ctx,
		txmanager.IsolationLevelRepeatableRead,
		txmanager.AccessModeReadWrite,
		func(txCtx context.Context) error {
			o, err := s.ordRepo.Get(txCtx, orderID)
			if err != nil {
				return errs.Wrap(err, errs.CodeRecordNotFound, "order not found", "order_id", orderID)
			}

			if err := validateReturn(o); err != nil {
				return err
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
			return nil
		},
	)
	if err != nil {
		return errs.Wrap(err, errs.CodeDBTransactionError, "return order tx failed", "order_id", orderID)
	}

	return nil
}

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

func (s *ServiceImpl) ListOrders(ctx context.Context, userID string, onlyInPVZ bool, lastN int, pg vo.Pagination) ([]*models.Order, int, error) {

	if userID == "" {
		return nil, 0, errs.New(errs.CodeValidationError, "empty user id")
	}

	var (
		paged []*models.Order
		total int
	)

	errTx := s.tx.WithTx(
		ctx,
		txmanager.IsolationLevelRepeatableRead,
		txmanager.AccessModeReadOnly,
		func(txCtx context.Context) error {

			// активные заказы
			active, err := s.ordRepo.ListByUser(
				txCtx, userID, onlyInPVZ, 0, nil, // nil => без лимита
			)
			if err != nil {
				return errs.Wrap(err, errs.CodeDatabaseError,
					"listByUser failed", "user_id", userID)
			}

			sortOrders(active)
			paged, total = paginate(active, lastN, pg)

			return nil
		},
	)

	if errTx != nil {
		return nil, 0, errs.Wrap(errTx, errs.CodeDBTransactionError,
			"list orders tx failed", "user_id", userID)
	}

	return paged, total, nil
}

func (s *ServiceImpl) ScrollOrders(ctx context.Context, userID string, cur vo.ScrollCursor) ([]*models.Order, vo.ScrollCursor, error) {
	if userID == "" {
		return nil, vo.ScrollCursor{}, errs.New(
			errs.CodeValidationError, "empty user id",
		)
	}
	roCtx := s.tx.WithReadOnly(ctx)
	orders, next, err := s.ordRepo.NextBatchAfter(roCtx, userID, cur)
	if err != nil {
		return nil, vo.ScrollCursor{}, errs.Wrap(err, errs.CodeDatabaseError,
			"next batch query failed", "user_id", userID)
	}

	return orders, next, nil
}

func (s *ServiceImpl) ListReturns(
	ctx context.Context,
	pg vo.Pagination,
) ([]*models.ReturnRecord, error) {

	roCtx := s.tx.WithReadOnly(ctx)

	records, err := s.hrRepo.ListReturns(roCtx, pg)
	if err != nil {
		return nil, errs.Wrap(err,
			errs.CodeDatabaseError, "list returns failed")
	}
	return records, nil
}

func (s *ServiceImpl) OrderHistory(
	ctx context.Context,
	pg vo.Pagination,
) ([]*models.HistoryEvent, int, error) {

	roCtx := s.tx.WithReadOnly(ctx)

	events, err := s.hrRepo.History(roCtx, pg)
	if err != nil {
		return nil, 0, errs.Wrap(err,
			errs.CodeDatabaseError, "list history failed")
	}
	return events, len(events), nil
}

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

func (s *ServiceImpl) generateClientReport(
	ctx context.Context,
	sortBy string,
) ([]*models.ClientReport, error) {

	var allOrders []*models.Order

	errTx := s.tx.WithTx(
		ctx,
		txmanager.IsolationLevelReadCommitted,
		txmanager.AccessModeReadOnly,
		func(txCtx context.Context) error {
			var err error

			allOrders, err = s.ordRepo.ListAllOrders(txCtx)
			if err != nil {
				return errs.Wrap(err, errs.CodeDatabaseError,
					"list all orders failed")
			}

			return nil
		},
	)
	if errTx != nil {
		return nil, errs.Wrap(errTx, errs.CodeDBTransactionError,
			"generate client report tx failed")
	}

	clientsMap := make(map[string]*models.ClientReport)
	aggregateOrders(clientsMap, allOrders)

	reports := make([]*models.ClientReport, 0, len(clientsMap))
	for _, r := range clientsMap {
		reports = append(reports, r)
	}

	if err := sortReports(reports, sortBy); err != nil {
		return nil, errs.Wrap(err, errs.CodeValidationError,
			"invalid sort parameter")
	}

	return reports, nil
}

func aggregateOrders(
	clientsMap map[string]*models.ClientReport,
	orders []*models.Order,
) {
	for _, o := range orders {
		cr, exists := clientsMap[o.UserID]
		if !exists {
			cr = &models.ClientReport{UserID: o.UserID}
			clientsMap[o.UserID] = cr
		}
		cr.TotalOrders++
		if o.Status == models.StatusReturned {
			cr.ReturnedOrders++
		} else {
			cr.TotalPurchaseSum += o.Price
		}
	}
}

func (s *ServiceImpl) GenerateClientReportByte(ctx context.Context, sortBy string) ([]byte, error) {
	reports, err := s.generateClientReport(ctx, sortBy)
	if err != nil {
		return nil, err
	}

	f := excelize.NewFile()
	sheet := "ClientsReport"
	f.SetSheetName(f.GetSheetName(0), sheet)

	headers := []string{"UserID", "Total Orders", "Returned Orders", "Total Purchase Sum (₽)"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	for i, r := range reports {
		row := i + 2
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), r.UserID)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), r.TotalOrders)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), r.ReturnedOrders)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), float64(r.TotalPurchaseSum)/100)
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
