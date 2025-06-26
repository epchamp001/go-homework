package packaging

import (
	"pvz-cli/internal/domain/codes"
	"pvz-cli/internal/domain/models"
	"sync"
)

type Provider interface {
	Strategy(models.PackageType) (PackagingStrategy, error)
}

type defaultProvider struct {
	mu   sync.RWMutex
	regs map[models.PackageType]PackagingStrategy
}

func NewDefaultProvider() Provider {
	return &defaultProvider{
		regs: map[models.PackageType]PackagingStrategy{
			models.PackageNone:    NewCompositeStrategy(),
			models.PackageBag:     NewBagStrategy(),
			models.PackageBox:     NewBoxStrategy(),
			models.PackageFilm:    NewFilmStrategy(),
			models.PackageBagFilm: NewCompositeStrategy(NewBagStrategy(), NewFilmStrategy()),
			models.PackageBoxFilm: NewCompositeStrategy(NewBoxStrategy(), NewFilmStrategy()),
		},
	}
}

func (p *defaultProvider) Strategy(t models.PackageType) (PackagingStrategy, error) {
	p.mu.RLock()
	strat, ok := p.regs[t]
	p.mu.RUnlock()
	if !ok {
		return nil, codes.ErrInvalidPackage
	}
	return strat, nil
}
