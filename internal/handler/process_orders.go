package handler

import (
	"context"
	"pvz-cli/internal/handler/mappers"
	"pvz-cli/pkg/errs"
	pvzpb "pvz-cli/pkg/pvz"
	"strconv"

	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
)

func (s *OrderServiceServer) ProcessOrders(
	ctx context.Context,
	req *pvzpb.ProcessOrdersRequest,
) (*pvzpb.ProcessResult, error) {
	if req == nil {
		return nil, grpcstatus.Error(codes.InvalidArgument, "ProcessOrdersRequest is nil")
	}
	if err := req.Validate(); err != nil {
		return nil, grpcstatus.Error(codes.InvalidArgument, err.Error())
	}

	userIDStr := strconv.FormatUint(req.UserId, 10)
	ids := make([]string, 0, len(req.OrderIds))
	for _, id := range req.OrderIds {
		ids = append(ids, strconv.FormatUint(id, 10))
	}

	var (
		result map[string]error
		err    error
	)

	switch req.Action {
	case pvzpb.ActionType_ACTION_TYPE_ISSUE:
		result, err = s.svc.IssueOrders(ctx, userIDStr, ids)
	case pvzpb.ActionType_ACTION_TYPE_RETURN:
		result, err = s.svc.ReturnOrdersByClient(ctx, userIDStr, ids)
	default:
		return nil, grpcstatus.Error(codes.InvalidArgument, "unknown action")
	}

	if err != nil {
		s.log.Errorw("ProcessOrders service error",
			"action", req.Action,
			"user_id", userIDStr,
			"error", err,
		)
		cause := errs.ErrorCause(err)
		return nil, grpcstatus.Error(codes.Internal, cause)
	}

	protoResult, mapErr := mappers.DomainProcessResultToProtoProcessResult(result)
	if mapErr != nil {
		return nil, grpcstatus.Error(codes.Internal, mapErr.Error())
	}
	return protoResult, nil
}
