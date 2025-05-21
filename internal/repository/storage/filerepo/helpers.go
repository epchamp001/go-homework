package filerepo

import (
	"encoding/json"
	"os"
	"pvz-cli/internal/domain/models"
	"pvz-cli/internal/domain/vo"
	"sort"
)

func atomicWrite(path string, data any) error {
	tmp := path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	if err := json.NewEncoder(f).Encode(data); err != nil {
		f.Close()
		_ = os.Remove(tmp)
		return err
	}
	f.Close()
	return os.Rename(tmp, path)
}

func filterOrders(src map[string]*models.Order, userID string, onlyInPVZ bool) []*models.Order {
	out := make([]*models.Order, 0, len(src))
	for _, o := range src {
		if o.UserID != userID {
			continue
		}
		if onlyInPVZ && o.Status != models.StatusAccepted {
			continue
		}
		out = append(out, o)
	}
	return out
}

func sortOrdersByCreatedAt(list []*models.Order) {
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
