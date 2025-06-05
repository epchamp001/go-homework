package mappers

import (
	"google.golang.org/protobuf/types/known/timestamppb"
	"pvz-cli/internal/domain/models"
	"pvz-cli/pkg/errs"
	pvzpb "pvz-cli/pkg/pvz"
	"strconv"
)

func DomainOrderToProtoOrder(o *models.Order) (*pvzpb.Order, error) {
	if o == nil {
		return nil, errs.New(errs.CodeMissingParameter, "models.Order is nil")
	}

	orderID, err := strconv.ParseUint(o.ID, 10, 64)
	if err != nil {
		return nil, errs.Wrap(err, errs.CodeParsingError, "invalid Order.ID", "order_id", o.ID)
	}
	userID, err := strconv.ParseUint(o.UserID, 10, 64)
	if err != nil {
		return nil, errs.Wrap(err, errs.CodeParsingError, "invalid Order.UserID", "user_id", o.UserID)
	}

	pbStatus := DomainToProtoOrderStatus(o.Status)

	pbPkg := DomainToProtoPackage(o.Package)

	totalRubles := float32(o.Price) / 100.0

	return &pvzpb.Order{
		OrderId:    orderID,
		UserId:     userID,
		Status:     pbStatus,
		ExpiresAt:  timestamppb.New(o.ExpiresAt),
		Weight:     float32(o.Weight),
		TotalPrice: totalRubles,
		Package:    &pbPkg,
	}, nil
}
