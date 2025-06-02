package packaging

import "pvz-cli/internal/domain/models"

type compositeStrategy struct {
	parts []PackagingStrategy
}

// NewCompositeStrategy создает составную стратегию упаковки из нескольких частей.
func NewCompositeStrategy(parts ...PackagingStrategy) PackagingStrategy {
	return &compositeStrategy{parts: parts}
}

func (c *compositeStrategy) Validate(weight float64) error {
	for _, s := range c.parts {
		if err := s.Validate(weight); err != nil {
			return err
		}
	}
	return nil
}

func (c *compositeStrategy) Surcharge() models.PriceKopecks {
	var total models.PriceKopecks
	for _, s := range c.parts {
		total += s.Surcharge()
	}
	return total
}
