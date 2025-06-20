package handler

import (
	"context"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
	"pvz-cli/internal/handler/mappers"
	"pvz-cli/pkg/errs"
	pvzpb "pvz-cli/pkg/pvz"
	"strconv"
)

func (s *OrderServiceServer) ReturnOrder(
	ctx context.Context,
	req *pvzpb.OrderIdRequest,
) (*pvzpb.OrderResponse, error) {
	if req == nil {
		return nil, grpcstatus.Error(codes.InvalidArgument, "OrderIdRequest is nil")
	}
	if err := req.Validate(); err != nil {
		return nil, grpcstatus.Error(codes.InvalidArgument, err.Error())
	}

	orderIDStr := strconv.FormatUint(req.OrderId, 10)
	if err := s.svc.ReturnOrder(ctx, orderIDStr); err != nil {
		s.log.Errorw("ReturnOrder service error",
			"order_id", orderIDStr,
			"error", err,
		)
		if grpcErr := errs.GrpcError(err); grpcErr != nil {
			return nil, grpcErr
		}
		return nil, grpcstatus.Error(codes.Internal, err.Error())
	}

	return mappers.DomainToProtoOrderResponse("ORDER_STATUS_DELETED", orderIDStr)
}
