package packaging

import (
	"pvz-cli/internal/domain/codes"
	"pvz-cli/internal/domain/models"
)

type boxStrategy struct{}

// NewBoxStrategy возвращает стратегию упаковки "box".
func NewBoxStrategy() PackagingStrategy {
	return &boxStrategy{}
}

func (b *boxStrategy) Validate(weight float64) error {
	if weight <= 0 {
		return codes.ErrValidationFailed
	}
	if weight >= 30 {
		return codes.ErrWeightTooHeavy
	}
	return nil
}

func (b *boxStrategy) Surcharge() models.PriceKopecks {
	return models.SurchargeBox
}
