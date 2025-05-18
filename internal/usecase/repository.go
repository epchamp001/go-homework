package usecase

import (
	"pvz-cli/internal/domain/models"
	"pvz-cli/internal/domain/vo"
)

type Repository interface {
	Create(o *models.Order) error
	Update(o *models.Order) error
	Get(id string) (*models.Order, error)
	Delete(id string) error

	ListByUser(userID string, filterOnlyInPVZ bool, lastN int, pg vo.Pagination) (orders []*models.Order, total int, err error)

	NextBatchAfter(userID string, cursor vo.ScrollCursor) (orders []*models.Order, next vo.ScrollCursor, err error)

	ListReturns(pg vo.Pagination) ([]*models.ReturnRecord, error)
	History() ([]*models.HistoryEvent, error)
	ImportMany(orders []*models.Order) error
}
