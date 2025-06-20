package service

import (
	"fmt"
	"pvz-cli/internal/domain/codes"
	"pvz-cli/internal/domain/models"
	"pvz-cli/internal/domain/vo"
	"pvz-cli/pkg/errs"
	"sort"
	"time"
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

func validateAccept(orderID, userID string, exp time.Time, weight float64) error {
	if orderID == "" {
		return codes.ErrValidationFailed
	}
	if userID == "" {
		return codes.ErrValidationFailed
	}
	if exp.Before(time.Now()) {
		return codes.ErrValidationFailed
	}
	if weight <= 0 {
		return codes.ErrValidationFailed
	}
	return nil
}

func validateReturn(o *models.Order) error {
	if o.Status == models.StatusIssued {
		return errs.New(errs.CodeValidationError, "cannot return an issued order", "order_id", o.ID)
	}
	if time.Now().Before(o.ExpiresAt) {
		return errs.New(errs.CodeValidationError, "storage period not expired yet", "order_id", o.ID)
	}
	return nil
}

func validateIssue(o *models.Order, userID string, now time.Time) error {
	if o.UserID != userID {
		return errs.New(errs.CodeValidationError, "order belongs to another user", "order_id", o.ID)
	}
	if o.Status != models.StatusAccepted {
		return errs.New(errs.CodeValidationError, "order not in accepted status", "order_id", o.ID)
	}
	if now.After(o.ExpiresAt) {
		return errs.New(errs.CodeValidationError, "storage period expired", "order_id", o.ID)
	}
	return nil
}

func validateClientReturn(o *models.Order, userID string, now time.Time) error {
	if o.UserID != userID {
		return errs.New(errs.CodeValidationError,
			"order belongs to another user", "order_id", o.ID)
	}
	if o.Status != models.StatusIssued || o.IssuedAt == nil {
		return errs.New(errs.CodeValidationError,
			"order not in issued status", "order_id", o.ID)
	}
	if now.Sub(*o.IssuedAt) > 48*time.Hour {
		return errs.New(errs.CodeValidationError,
			"return window expired (>48h)", "order_id", o.ID)
	}
	return nil
}

func sortOrders(list []*models.Order) {
	sort.Slice(list, func(i, j int) bool {
		if !list[i].CreatedAt.Equal(list[j].CreatedAt) {
			return list[i].CreatedAt.Before(list[j].CreatedAt)
		}
		return list[i].ID < list[j].ID
	})
}

func paginate[T any](list []T, lastN int, pg vo.Pagination) ([]T, int) {
	total := len(list)

	if lastN > 0 && lastN < total {
		return list[total-lastN:], total
	}

	if pg.Page > 0 && pg.Limit > 0 {
		start := (pg.Page - 1) * pg.Limit
		if start >= total {
			return []T{}, total
		}
		end := start + pg.Limit
		if end > total {
			end = total
		}
		return list[start:end], total
	}

	return list, total
}
