package usecase

import (
	"context"
	"pvz-cli/internal/domain/models"
	"pvz-cli/internal/domain/vo"
)

type OrdersRepository interface {
	// Create сохраняет новый заказ
	Create(ctx context.Context, o *models.Order) error

	// Update изменяет существующий заказ
	Update(ctx context.Context, o *models.Order) error

	// Get возвращает заказ по его ID
	Get(ctx context.Context, id string) (*models.Order, error)

	// Delete удаляет заказ по его ID
	Delete(ctx context.Context, id string) error

	// ListByUser возвращает заказы пользователя
	ListByUser(ctx context.Context, userID string, onlyInPVZ bool, lastN int, pg *vo.Pagination) ([]*models.Order, error)

	// NextBatchAfter делает бесконечную пагинацию
	NextBatchAfter(ctx context.Context, userID string, cur vo.ScrollCursor) ([]*models.Order, vo.ScrollCursor, error)

	// ImportMany вставляет пачку заказов
	ImportMany(ctx context.Context, list []*models.Order) error

	ListAllOrders(ctx context.Context) ([]*models.Order, error)
}

type HistoryAndReturnsRepository interface {
	// ListReturns возвращает постраничный список записей о возвратах
	ListReturns(ctx context.Context, pg vo.Pagination) ([]*models.ReturnRecord, error)

	// History возвращает постраничный список истории статусов
	History(ctx context.Context, pg vo.Pagination) ([]*models.HistoryEvent, error)

	// AddHistory вставляет запись события истории заказа
	AddHistory(ctx context.Context, e *models.HistoryEvent) error

	// AddReturn вставляет запись о возврате заказа
	AddReturn(ctx context.Context, rec *models.ReturnRecord) error

	ListReturnsByUser(ctx context.Context, userID string) ([]*models.ReturnRecord, error)
}
