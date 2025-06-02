// Package usecase содержит реализацию бизнес-логики приложения.
package usecase

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"pvz-cli/internal/domain/codes"
	"pvz-cli/internal/domain/models"
	"pvz-cli/internal/domain/vo"
	"pvz-cli/internal/usecase/packaging"
	"pvz-cli/pkg/errs"

	"github.com/xuri/excelize/v2"
)

// Service определяет бизнес-логику работы Пункта Выдачи Заказов.
type Service interface {
	// AcceptOrder регистрирует новый заказ и рассчитывает итоговую стоимость с учётом упаковки.
	AcceptOrder(orderID, userID string, expires time.Time, weight float64, price models.PriceKopecks, pkgType models.PackageType) (models.PriceKopecks, error)

	// ReturnOrder выполняет возврат заказа по его ID (если срок хранения истёк и не был выдан).
	ReturnOrder(orderID string) error

	// IssueOrders выполняет массовую выдачу заказов клиенту.
	IssueOrders(userID string, ids []string) (map[string]error, error)

	// ReturnOrdersByClient обрабатывает массовый возврат заказов клиентом в течение 48 часов после выдачи.
	ReturnOrdersByClient(userID string, ids []string) (map[string]error, error)

	// ListOrders возвращает заказы клиента с возможностью фильтрации, пагинации и лимита.
	ListOrders(userID string, onlyInPVZ bool, lastN int, pg vo.Pagination) ([]*models.Order, int, error)

	// ScrollOrders возвращает порцию заказов по курсору (постраничная прокрутка).
	ScrollOrders(userID string, cursor vo.ScrollCursor) (orders []*models.Order, next vo.ScrollCursor, err error)

	// ListReturns возвращает список возвратов с пагинацией.
	ListReturns(pg vo.Pagination) ([]*models.ReturnRecord, error)

	// OrderHistory возвращает полную историю по заказам.
	OrderHistory() ([]*models.HistoryEvent, error)

	// ImportOrders импортирует заказы из JSON-файла, возвращает количество успешно добавленных.
	ImportOrders(filePath string) (imported int, err error)

	// GenerateClientReportByte генерирует отчет по заказам клиентов. Возвращает слайс []byte для дальнейшего преобразования в формат .xlsx
	GenerateClientReportByte(sortBy string) ([]byte, error)
}

type ServiceImpl struct {
	repo Repository
}

func NewService(repo Repository) *ServiceImpl {
	return &ServiceImpl{
		repo: repo,
	}
}

func (s *ServiceImpl) AcceptOrder(orderID, userID string, exp time.Time, weight float64, price models.PriceKopecks, pkgType models.PackageType) (models.PriceKopecks, error) {
	if orderID == "" || userID == "" {
		return 0, codes.ErrValidationFailed
	}
	if exp.Before(time.Now()) {
		return 0, codes.ErrValidationFailed
	}
	if _, err := s.repo.Get(orderID); err == nil {
		return 0, codes.ErrOrderAlreadyExists
	}

	strat, err := packaging.GetStrategy(pkgType)
	if err != nil {
		return 0, err
	}

	if err := strat.Validate(weight); err != nil {
		return 0, err
	}

	total := price + strat.Surcharge()

	now := time.Now()
	o := &models.Order{
		ID:         orderID,
		UserID:     userID,
		Status:     models.StatusAccepted,
		ExpiresAt:  exp,
		CreatedAt:  now,
		Weight:     weight,
		Price:      int64(price),
		TotalPrice: int64(total),
		Package:    string(pkgType),
	}

	if err := s.repo.Create(o); err != nil {
		return 0, err
	}

	return total, nil
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

func (s *ServiceImpl) generateClientReport(sortBy string) ([]*models.ClientReport, error) {
	activeOrders, err := s.repo.ListAllOrders()
	if err != nil {
		return nil, err
	}

	returnsList, err := s.repo.ListReturns(vo.Pagination{})
	if err != nil {
		return nil, err
	}

	clientsMap := make(map[string]*models.ClientReport)

	s.aggregateActiveOrders(clientsMap, activeOrders)
	s.aggregateReturnRecords(clientsMap, returnsList)

	reports := make([]*models.ClientReport, 0, len(clientsMap))
	for _, v := range clientsMap {
		reports = append(reports, v)
	}

	if err := sortReports(reports, sortBy); err != nil {
		return nil, err
	}

	return reports, nil
}

// aggregateActiveOrders добавляет к clientsMap данные по не возвращённым заказам.
func (s *ServiceImpl) aggregateActiveOrders(clientsMap map[string]*models.ClientReport, activeOrders []*models.Order) {
	for _, o := range activeOrders {
		cr, exists := clientsMap[o.UserID]
		if !exists {
			cr = &models.ClientReport{UserID: o.UserID}
			clientsMap[o.UserID] = cr
		}
		cr.TotalOrders++
		cr.TotalPurchaseSum += models.PriceKopecks(o.Price)
	}
}

// aggregateReturnRecords добавляет к clientsMap данные по всем возвратам.
func (s *ServiceImpl) aggregateReturnRecords(clientsMap map[string]*models.ClientReport, returnsList []*models.ReturnRecord) {
	for _, rec := range returnsList {
		cr, exists := clientsMap[rec.UserID]
		if !exists {
			cr = &models.ClientReport{UserID: rec.UserID}
			clientsMap[rec.UserID] = cr
		}
		cr.TotalOrders++
		cr.ReturnedOrders++
		// цену не добавляю, т.к. покупка не состоялась
	}
}

func (s *ServiceImpl) GenerateClientReportByte(sortBy string) ([]byte, error) {
	reports, err := s.generateClientReport(sortBy)
	if err != nil {
		return nil, err
	}

	f := excelize.NewFile()
	sheet := "ClientsReport"
	f.SetSheetName(f.GetSheetName(0), sheet)

	headers := []string{"UserID", "Total Orders", "Returned Orders", "Total Purchase Sum (₽)"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	for i, r := range reports {
		row := i + 2
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), r.UserID)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), r.TotalOrders)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), r.ReturnedOrders)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), float64(r.TotalPurchaseSum)/100)
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
