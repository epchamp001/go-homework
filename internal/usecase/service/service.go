// Package usecase содержит реализацию бизнес-логики приложения.
package service

import (
	"context"
	"pvz-cli/internal/usecase"
	"pvz-cli/pkg/txmanager"
	"time"

	"pvz-cli/internal/domain/models"
	"pvz-cli/internal/domain/vo"
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
	ordRepo usecase.OrdersRepository
	hrRepo  usecase.HistoryAndReturnsRepository
}

func NewService(tx txmanager.TxManager, ordRepo usecase.OrdersRepository, hrRepo usecase.HistoryAndReturnsRepository) *ServiceImpl {
	return &ServiceImpl{
		tx:      tx,
		ordRepo: ordRepo,
		hrRepo:  hrRepo,
	}
}
