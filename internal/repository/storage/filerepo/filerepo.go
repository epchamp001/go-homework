// Package filerepo реализует файловое хранилище для заказов, возвратов и событий истории.
package filerepo

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"pvz-cli/internal/domain/codes"
	"pvz-cli/internal/domain/models"
	"pvz-cli/internal/domain/vo"
	"pvz-cli/pkg/errs"
	"time"
)

// FileRepo представляет файловое хранилище заказов, возвратов и истории.
// Данные сохраняются в формате JSON на диск и кэшируются в памяти.
type FileRepo struct {
	pathOrders  string
	pathReturns string
	pathHist    string

	orders  map[string]*models.Order
	returns []*models.ReturnRecord
	history []*models.HistoryEvent
}

// NewFileRepo инициализирует файловое хранилище и загружает существующие данные из JSON-файлов.
func NewFileRepo(dataDir string) (*FileRepo, error) {
	r := &FileRepo{
		pathOrders:  filepath.Join(dataDir, "orders.json"),
		pathReturns: filepath.Join(dataDir, "returns.json"),
		pathHist:    filepath.Join(dataDir, "history.json"),
		orders:      make(map[string]*models.Order),
	}

	load := func(path string, dst interface{}) error {
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		if err != nil {
			return err
		}

		return json.NewDecoder(f).Decode(dst)
	}
	_ = load(r.pathOrders, &r.orders)
	_ = load(r.pathReturns, &r.returns)
	_ = load(r.pathHist, &r.history)

	return r, nil
}

func (r *FileRepo) dumpOrders() error {
	return atomicWrite(r.pathOrders, r.orders)
}

func (r *FileRepo) dumpReturns() error {
	return atomicWrite(r.pathReturns, r.returns)
}

func (r *FileRepo) dumpHistory() error {
	return atomicWrite(r.pathHist, r.history)
}

func (r *FileRepo) Create(o *models.Order) error {
	if _, ok := r.orders[o.ID]; ok {
		return codes.ErrOrderAlreadyExists
	}
	r.orders[o.ID] = o
	r.history = append(r.history, &models.HistoryEvent{
		OrderID: o.ID, Status: o.Status, Time: time.Now(),
	})

	if err := r.dumpOrders(); err != nil {
		return err
	}
	return r.dumpHistory()
}

func (r *FileRepo) Update(o *models.Order) error {
	orig, ok := r.orders[o.ID]
	if !ok {
		return codes.ErrOrderNotFound
	}

	if o.Status == models.StatusReturned && orig.Status != models.StatusReturned {
		r.returns = append(r.returns, &models.ReturnRecord{
			OrderID:    o.ID,
			UserID:     o.UserID,
			ReturnedAt: time.Now(),
		})
		if err := r.dumpReturns(); err != nil {
			return err
		}
	}

	r.orders[o.ID] = o
	r.history = append(r.history, &models.HistoryEvent{
		OrderID: o.ID, Status: o.Status, Time: time.Now(),
	})

	if err := r.dumpOrders(); err != nil {
		return err
	}
	return r.dumpHistory()
}

func (r *FileRepo) Get(id string) (*models.Order, error) {
	if o, ok := r.orders[id]; ok {
		return o, nil
	}
	return nil, codes.ErrOrderNotFound
}

func (r *FileRepo) Delete(id string) error {
	o, ok := r.orders[id]
	if !ok {
		return codes.ErrOrderNotFound
	}

	r.returns = append(r.returns, &models.ReturnRecord{
		OrderID:    o.ID,
		UserID:     o.UserID,
		ReturnedAt: time.Now(),
	})
	if err := r.dumpReturns(); err != nil {
		return err
	}

	delete(r.orders, id)
	return r.dumpOrders()
}

func (r *FileRepo) ListByUser(userID string, onlyInPVZ bool, lastN int,
	pg vo.Pagination) ([]*models.Order, int, error) {

	orders := filterOrders(r.orders, userID, onlyInPVZ)
	sortOrdersByCreatedAt(orders)
	orders, total := paginate[*models.Order](orders, lastN, pg)

	return orders, total, nil
}

func (r *FileRepo) NextBatchAfter(userID string, cur vo.ScrollCursor) (
	out []*models.Order, next vo.ScrollCursor, _ error) {

	list := filterOrders(r.orders, userID, false)
	sortOrdersByCreatedAt(list)

	pos := 0
	if cur.LastID != "" {
		for i, o := range list {
			if o.ID == cur.LastID {
				pos = i + 1
				break
			}
		}
	}

	lim := cur.Limit
	if lim <= 0 {
		lim = 20
	}
	end := pos + lim
	if end > len(list) {
		end = len(list)
	}
	out = list[pos:end]

	if end < len(list) {
		next = vo.ScrollCursor{LastID: list[end-1].ID, Limit: lim}
	}
	return
}

func (r *FileRepo) ListReturns(pg vo.Pagination) ([]*models.ReturnRecord, error) {
	out, _ := paginate(r.returns, 0, pg)
	return out, nil
}

func (r *FileRepo) History(pg vo.Pagination) ([]*models.HistoryEvent, int, error) {
	all := make([]*models.HistoryEvent, len(r.history))
	copy(all, r.history)

	paged, total := paginate[*models.HistoryEvent](all, 0, pg)
	return paged, total, nil
}

func (r *FileRepo) ImportMany(list []*models.Order) error {
	for _, o := range list {
		if _, dup := r.orders[o.ID]; dup {
			return errs.New(errs.CodeRecordAlreadyExists,
				"order already exists",
				"orderID", o.ID)
		}
		r.orders[o.ID] = o
		r.history = append(r.history, &models.HistoryEvent{
			OrderID: o.ID, Status: o.Status, Time: time.Now(),
		})
	}
	if err := r.dumpOrders(); err != nil {
		return err
	}
	return r.dumpHistory()
}

func (r *FileRepo) ListAllOrders() ([]*models.Order, error) {
	orders := make([]*models.Order, 0, len(r.orders))
	for _, o := range r.orders {
		orders = append(orders, o)
	}
	return orders, nil
}
