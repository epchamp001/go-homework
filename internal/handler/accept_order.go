package handler

import (
	"context"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
	"pvz-cli/internal/handler/mappers"
	"pvz-cli/pkg/errs"
	pvzpb "pvz-cli/pkg/pvz"
)

func (s *OrderServiceServer) AcceptOrder(
	ctx context.Context,
	req *pvzpb.AcceptOrderRequest,
) (*pvzpb.OrderResponse, error) {
	if req == nil {
		return nil, grpcstatus.Error(codes.InvalidArgument, "AcceptOrderRequest is nil")
	}
	if err := req.Validate(); err != nil {
		return nil, grpcstatus.Error(codes.InvalidArgument, err.Error())
	}

	domainOrder, err := mappers.ProtoToDomainOrderForImport(req)
	if err != nil {
		if grpcErr := errs.GrpcError(err); grpcErr != nil {
			return nil, grpcErr
		}
		return nil, grpcstatus.Error(codes.Internal, err.Error())
	}

	_, err = s.svc.AcceptOrder(
		ctx,
		domainOrder.ID,
		domainOrder.UserID,
		domainOrder.ExpiresAt,
		domainOrder.Weight,
		domainOrder.Price,
		domainOrder.Package,
	)
	if err != nil {
		// логируем только ошибку svc
		s.log.Errorw("AcceptOrder service error",
			"order_id", domainOrder.ID,
			"error", err,
		)
		if grpcErr := errs.GrpcError(err); grpcErr != nil {
			return nil, grpcErr
		}
		return nil, grpcstatus.Error(codes.Internal, err.Error())
	}

	return &pvzpb.OrderResponse{
		Status:  pvzpb.OrderStatus_ORDER_STATUS_EXPECTS,
		OrderId: req.OrderId,
	}, nil
}
