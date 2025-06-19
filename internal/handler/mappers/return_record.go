package mappers

import (
	"google.golang.org/protobuf/types/known/timestamppb"
	"pvz-cli/internal/domain/models"
	"pvz-cli/pkg/errs"
	pvzpb "pvz-cli/pkg/pvz"
	"strconv"
)

func DomainReturnRecordToProtoReturnRecord(r *models.ReturnRecord) (*pvzpb.ReturnRecord, error) {
	if r == nil {
		return nil, errs.New(errs.CodeMissingParameter, "models.ReturnRecord is nil")
	}

	orderID, err := strconv.ParseUint(r.OrderID, 10, 64)
	if err != nil {
		return nil, errs.Wrap(err, errs.CodeParsingError, "invalid ReturnRecord.OrderID", "order_id", r.OrderID)
	}
	userID, err := strconv.ParseUint(r.UserID, 10, 64)
	if err != nil {
		return nil, errs.Wrap(err, errs.CodeParsingError, "invalid ReturnRecord.UserID", "user_id", r.UserID)
	}

	return &pvzpb.ReturnRecord{
		OrderId:    orderID,
		UserId:     userID,
		ReturnedAt: timestamppb.New(r.ReturnedAt),
	}, nil
}
