package mappers

import (
	"pvz-cli/pkg/errs"
	pvzpb "pvz-cli/pkg/pvz"
	"strconv"
)

func DomainToProtoOrderResponse(status, orderID string) (*pvzpb.OrderResponse, error) {
	if orderID == "" {
		return nil, errs.New(errs.CodeMissingParameter, "orderID пустой")
	}
	idUint, err := strconv.ParseUint(orderID, 10, 64)
	if err != nil {
		return nil, errs.Wrap(err, errs.CodeParsingError, "invalid orderID", "order_id", orderID)
	}

	var pbStatus pvzpb.OrderStatus = pvzpb.OrderStatus_ORDER_STATUS_UNSPECIFIED
	if status != "" {
		if sVal, ok := pvzpb.OrderStatus_value[status]; ok {
			pbStatus = pvzpb.OrderStatus(sVal)
		} else {
			return nil, errs.New(errs.CodeInvalidParameter, "unknown order status", "status", status)
		}
	}

	return &pvzpb.OrderResponse{
		Status:  pbStatus,
		OrderId: idUint,
	}, nil
}
