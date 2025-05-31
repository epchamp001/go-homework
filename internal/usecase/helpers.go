package usecase

import (
	"fmt"
	"pvz-cli/internal/domain/models"
	"sort"
)

func sortReports(reports []*models.ClientReport, sortBy string) error {
	switch sortBy {
	case "orders":
		sort.Slice(reports, func(i, j int) bool {
			if reports[i].TotalOrders != reports[j].TotalOrders {
				return reports[i].TotalOrders > reports[j].TotalOrders
			}
			// при равенстве количества заказов сравниваю по UserID
			return reports[i].UserID < reports[j].UserID
		})
	case "sum":
		sort.Slice(reports, func(i, j int) bool {
			if reports[i].TotalPurchaseSum != reports[j].TotalPurchaseSum {
				return reports[i].TotalPurchaseSum > reports[j].TotalPurchaseSum
			}
			// тоже самое по id
			return reports[i].UserID < reports[j].UserID
		})
	default:
		return fmt.Errorf("invalid sort option: %s", sortBy)
	}
	return nil
}
