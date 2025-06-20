package mappers

import (
	"pvz-cli/internal/domain/models"
	"pvz-cli/pkg/errs"
	pvzpb "pvz-cli/pkg/pvz"
	"strconv"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func DomainOrderHistoryToProtoOrderHistory(h *models.HistoryEvent) (*pvzpb.OrderHistory, error) {
	if h == nil {
		return nil, errs.New(errs.CodeMissingParameter, "models.OrderHistory is nil")
	}

	orderID, err := strconv.ParseUint(h.OrderID, 10, 64)
	if err != nil {
		return nil, errs.Wrap(err, errs.CodeParsingError, "invalid OrderHistory.OrderID", "order_id", h.OrderID)
	}

	pbStatus := DomainToProtoOrderStatus(h.Status)

	return &pvzpb.OrderHistory{
		OrderId:   orderID,
		Status:    pbStatus,
		CreatedAt: timestamppb.New(h.Time),
	}, nil
}
