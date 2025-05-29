package packaging

import (
	"pvz-cli/internal/domain/codes"
	"pvz-cli/internal/domain/models"
)

// Registry содержит стратегии упаковки, сопоставляя тип упаковки с соответствующей реализацией PackagingStrategy.
var Registry = map[models.PackageType]PackagingStrategy{}

func init() {
	Registry[models.PackageNone] = NewCompositeStrategy()
	Registry[models.PackageBag] = NewBagStrategy()
	Registry[models.PackageBox] = NewBoxStrategy()
	Registry[models.PackageFilm] = NewFilmStrategy()
	Registry[models.PackageBagFilm] = NewCompositeStrategy(NewBagStrategy(), NewFilmStrategy())
	Registry[models.PackageBoxFilm] = NewCompositeStrategy(NewBoxStrategy(), NewFilmStrategy())
}

// GetStrategy возвращает упаковочную стратегию по типу.
func GetStrategy(pkg models.PackageType) (PackagingStrategy, error) {
	strat, ok := Registry[pkg]
	if !ok {
		return nil, codes.ErrInvalidPackage
	}
	return strat, nil
}
