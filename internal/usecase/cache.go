package usecase

import (
	"pvz-cli/internal/domain/models"
)

type OrderCache interface {
	Get(id string) (*models.Order, bool)
	Set(id string, ord *models.Order)
	Delete(id string)
}

func OrderKey(id string) string {
	return "orders:" + id
}
