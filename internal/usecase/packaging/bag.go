package packaging

import (
	"pvz-cli/internal/domain/codes"
	"pvz-cli/internal/domain/models"
)

type bagStrategy struct{}

// NewBagStrategy возвращает стратегию упаковки "bag".
func NewBagStrategy() PackagingStrategy {
	return &bagStrategy{}
}

func (b *bagStrategy) Validate(w float64) error {
	if w <= 0 {
		return codes.ErrValidationFailed
	}
	if w >= 10 {
		return codes.ErrWeightTooHeavy
	}
	return nil
}

func (b *bagStrategy) Surcharge() models.PriceKopecks {
	return models.SurchargeBag
}
