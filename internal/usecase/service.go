package usecase

import (
	"encoding/json"
	"os"
	"pvz-cli/internal/domain/codes"
	"pvz-cli/internal/domain/models"
	"pvz-cli/internal/domain/vo"
	"pvz-cli/pkg/errs"
	"time"
)

type Service interface {
	AcceptOrder(orderID, userID string, expires time.Time) error
	ReturnOrder(orderID string) error

	IssueOrders(userID string, ids []string) (map[string]error, error)
	ReturnOrdersByClient(userID string, ids []string) (map[string]error, error)

	ListOrders(userID string, onlyInPVZ bool, lastN int, pg vo.Pagination) ([]*models.Order, int, error)

	ScrollOrders(userID string, cursor vo.ScrollCursor) (orders []*models.Order, next vo.ScrollCursor, err error)

	ListReturns(pg vo.Pagination) ([]*models.ReturnRecord, error)
	OrderHistory() ([]*models.HistoryEvent, error)

	ImportOrders(filePath string) (imported int, err error)
}

type ServiceImpl struct {
	repo Repository
}

func NewService(repo Repository) *ServiceImpl {
	return &ServiceImpl{
		repo: repo,
	}
}

func (s *ServiceImpl) AcceptOrder(orderID, userID string, exp time.Time) error {
	if orderID == "" || userID == "" {
		return codes.ErrValidationFailed
	}
	if exp.Before(time.Now()) {
		return codes.ErrValidationFailed
	}
	if _, err := s.repo.Get(orderID); err == nil {
		return codes.ErrOrderAlreadyExists
	}

	now := time.Now()
	o := &models.Order{
		ID:        orderID,
		UserID:    userID,
		Status:    models.StatusAccepted,
		ExpiresAt: exp,
		CreatedAt: now,
	}
	return s.repo.Create(o)
}

func (s *ServiceImpl) ReturnOrder(orderID string) error {
	o, err := s.repo.Get(orderID)
	if err != nil {
		return err
	}

	if o.Status == models.StatusIssued {
		return codes.ErrValidationFailed
	}
	if time.Now().Before(o.ExpiresAt) {
		return codes.ErrStorageExpired
	}

	return s.repo.Delete(orderID)
}

func (s *ServiceImpl) IssueOrders(userID string, ids []string) (map[string]error, error) {
	result := make(map[string]error, len(ids))
	now := time.Now()

	for _, id := range ids {
		o, err := s.repo.Get(id)
		if err != nil {
			result[id] = codes.ErrOrderNotFound
			continue
		}
		if o.UserID != userID {
			result[id] = codes.ErrValidationFailed
			continue
		}
		if o.Status != models.StatusAccepted {
			result[id] = codes.ErrValidationFailed
			continue
		}
		if now.After(o.ExpiresAt) {
			result[id] = codes.ErrStorageExpired
			continue
		}
		o.Status = models.StatusIssued
		o.IssuedAt = &now
		if err := s.repo.Update(o); err != nil {
			result[id] = err
			continue
		}
		result[id] = nil
	}
	return result, nil
}

func (s *ServiceImpl) ReturnOrdersByClient(userID string, ids []string) (map[string]error, error) {
	result := make(map[string]error, len(ids))
	now := time.Now()

	for _, id := range ids {
		o, err := s.repo.Get(id)
		if err != nil {
			result[id] = codes.ErrOrderNotFound
			continue
		}
		if o.UserID != userID {
			result[id] = codes.ErrValidationFailed
			continue
		}
		if o.Status != models.StatusIssued || o.IssuedAt == nil {
			result[id] = codes.ErrValidationFailed
			continue
		}
		if now.Sub(*o.IssuedAt) > 48*time.Hour {
			result[id] = codes.ErrStorageExpired
			continue
		}

		updated := *o
		updated.Status = models.StatusReturned
		updated.ReturnedAt = &now
		if err := s.repo.Update(&updated); err != nil {
			result[id] = err
			continue
		}
		result[id] = nil
	}
	return result, nil
}

func (s *ServiceImpl) ListOrders(
	userID string,
	onlyInPVZ bool,
	lastN int,
	pg vo.Pagination,
) ([]*models.Order, int, error) {
	if userID == "" {
		return nil, 0, codes.ErrValidationFailed
	}
	return s.repo.ListByUser(userID, onlyInPVZ, lastN, pg)
}

func (s *ServiceImpl) ScrollOrders(userID string, cur vo.ScrollCursor) (
	out []*models.Order, next vo.ScrollCursor, err error) {

	if userID == "" {
		return nil, vo.ScrollCursor{}, codes.ErrValidationFailed
	}
	return s.repo.NextBatchAfter(userID, cur)
}

func (s *ServiceImpl) ListReturns(pg vo.Pagination) ([]*models.ReturnRecord, error) {
	return s.repo.ListReturns(pg)
}

func (s *ServiceImpl) OrderHistory() ([]*models.HistoryEvent, error) {
	return s.repo.History()
}

func (s *ServiceImpl) ImportOrders(path string) (int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, errs.Wrap(err, errs.CodeFileReadError, "couldn't open the file")
	}
	defer f.Close()

	// читаю во временную структуру, чтобы не было багов с ExpiresAt
	var raw []struct {
		ID        string `json:"order_id"`
		UserID    string `json:"user_id"`
		ExpiresAt string `json:"expires_at"`
	}
	if err := json.NewDecoder(f).Decode(&raw); err != nil {
		return 0, errs.Wrap(err, errs.CodeParsingError, "couldn't decode")
	}

	now := time.Now()
	seen := make(map[string]struct{}, len(raw))
	batch := make([]*models.Order, 0, len(raw))

	for _, r := range raw {
		if r.ID == "" || r.UserID == "" {
			return 0, codes.ErrValidationFailed
		}
		if _, dup := seen[r.ID]; dup {
			return 0, codes.ErrOrderAlreadyExists
		}
		seen[r.ID] = struct{}{}

		exp, err := time.Parse("2006-01-02", r.ExpiresAt)
		if err != nil || exp.Before(now) {
			return 0, codes.ErrValidationFailed
		}

		batch = append(batch, &models.Order{
			ID:        r.ID,
			UserID:    r.UserID,
			Status:    models.StatusAccepted,
			ExpiresAt: exp,
			CreatedAt: now,
		})
	}

	if err := s.repo.ImportMany(batch); err != nil {
		return 0, err
	}
	return len(batch), nil
}
