// Package packaging предоставляет стратегии упаковки заказов.
package packaging

import "pvz-cli/internal/domain/models"

// PackagingStrategy описывает логику валидации и наценки.
type PackagingStrategy interface {
	// Validate проверяет ограничения.
	Validate(weight float64) error
	// Surcharge возвращает добавочную стоимость.
	Surcharge() models.PriceKopecks
}
