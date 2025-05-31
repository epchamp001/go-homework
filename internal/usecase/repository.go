package usecase

import (
	"pvz-cli/internal/domain/models"
	"pvz-cli/internal/domain/vo"
)

// Repository определяет контракт для хранилища заказов, возвратов и истории операций.
type Repository interface {
	// Create сохраняет новый заказ в хранилище. Возвращает ошибку, если заказ с таким ID уже существует.
	Create(o *models.Order) error

	// Update обновляет существующий заказ и сохраняет событие в историю. Если заказ возвращён — сохраняет его в список возвратов.
	Update(o *models.Order) error

	// Get возвращает заказ по ID. Если заказ не найден, возвращает ошибку.
	Get(id string) (*models.Order, error)

	// Delete удаляет заказ и сохраняет информацию о возврате. Возвращает ошибку, если заказ не найден.
	Delete(id string) error

	// ListByUser возвращает список заказов пользователя с поддержкой пагинации и фильтрацией по наличию в ПВЗ.
	ListByUser(userID string, filterOnlyInPVZ bool, lastN int, pg vo.Pagination) (orders []*models.Order, total int, err error)

	// NextBatchAfter возвращает следующую порцию заказов пользователя после указанного идентификатора (реализация scroll-пагинации).
	NextBatchAfter(userID string, cursor vo.ScrollCursor) (orders []*models.Order, next vo.ScrollCursor, err error)

	// ListReturns возвращает список возвратов с поддержкой пагинации.
	ListReturns(pg vo.Pagination) ([]*models.ReturnRecord, error)

	// History возвращает историю изменений заказов.
	History() ([]*models.HistoryEvent, error)

	// ImportMany импортирует список заказов из внешнего файла в хранилище. Возвращает ошибку при попытке вставить дубликат.
	ImportMany(orders []*models.Order) error

	// ListAllOrders возвращает список всех заказов.
	ListAllOrders() ([]*models.Order, error)
}
