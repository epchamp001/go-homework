package service

import (
	"context"
	"fmt"
	"pvz-cli/internal/domain/models"
	"pvz-cli/pkg/errs"
	"pvz-cli/pkg/txmanager"

	"github.com/xuri/excelize/v2"
)

func (s *ServiceImpl) generateClientReport(
	ctx context.Context,
	sortBy string,
) ([]*models.ClientReport, error) {

	var allOrders []*models.Order

	errTx := s.tx.WithTx(
		ctx,
		txmanager.IsolationLevelReadCommitted,
		txmanager.AccessModeReadOnly,
		func(txCtx context.Context) error {
			var err error

			allOrders, err = s.ordRepo.ListAllOrders(txCtx)
			if err != nil {
				return errs.Wrap(err, errs.CodeDatabaseError,
					"list all orders failed")
			}

			return nil
		},
	)
	if errTx != nil {
		return nil, errs.Wrap(errTx, errs.CodeDBTransactionError,
			"generate client report tx failed")
	}

	clientsMap := make(map[string]*models.ClientReport)
	aggregateOrders(clientsMap, allOrders)

	reports := make([]*models.ClientReport, 0, len(clientsMap))
	for _, r := range clientsMap {
		reports = append(reports, r)
	}

	if err := sortReports(reports, sortBy); err != nil {
		return nil, errs.Wrap(err, errs.CodeValidationError,
			"invalid sort parameter")
	}

	return reports, nil
}

func aggregateOrders(
	clientsMap map[string]*models.ClientReport,
	orders []*models.Order,
) {
	for _, o := range orders {
		cr, exists := clientsMap[o.UserID]
		if !exists {
			cr = &models.ClientReport{UserID: o.UserID}
			clientsMap[o.UserID] = cr
		}
		cr.TotalOrders++
		if o.Status == models.StatusReturned {
			cr.ReturnedOrders++
		} else {
			cr.TotalPurchaseSum += o.Price
		}
	}
}

func (s *ServiceImpl) GenerateClientReportByte(ctx context.Context, sortBy string) ([]byte, error) {
	reports, err := s.generateClientReport(ctx, sortBy)
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
