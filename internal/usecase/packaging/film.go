package packaging

import (
	"pvz-cli/internal/domain/codes"
	"pvz-cli/internal/domain/models"
)

type filmStrategy struct{}

// NewFilmStrategy возвращает стратегию упаковки "film".
func NewFilmStrategy() PackagingStrategy {
	return &filmStrategy{}
}

func (f *filmStrategy) Validate(weight float64) error {
	if weight <= 0 {
		return codes.ErrValidationFailed
	}
	return nil
}

func (f *filmStrategy) Surcharge() models.PriceKopecks {
	return models.SurchargeFilm
}
