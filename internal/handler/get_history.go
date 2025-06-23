package handler

import (
	"context"
	"pvz-cli/internal/handler/mappers"
	"pvz-cli/pkg/errs"
	pvzpb "pvz-cli/pkg/pvz"

	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
)

func (s *OrderServiceServer) GetHistory(
	ctx context.Context,
	req *pvzpb.GetHistoryRequest,
) (*pvzpb.OrderHistoryList, error) {
	if req != nil {
		if err := req.Validate(); err != nil {
			return nil, grpcstatus.Error(codes.InvalidArgument, err.Error())
		}
	}

	pagination := mappers.ProtoToDomainPagination(req.Pagination)

	historyEvents, _, err := s.svc.OrderHistory(ctx, pagination)
	if err != nil {
		s.log.Errorw("GetHistory service error",
			"pagination", req.GetPagination(),
			"error", err,
		)
		if grpcErr := errs.GrpcError(err); grpcErr != nil {
			return nil, grpcErr
		}
		cause := errs.ErrorCause(err)
		return nil, grpcstatus.Error(codes.Internal, cause)
	}

	pbHistory := make([]*pvzpb.OrderHistory, 0, len(historyEvents))
	for _, h := range historyEvents {
		pbH, mapErr := mappers.DomainOrderHistoryToProtoOrderHistory(h)
		if mapErr != nil {
			return nil, grpcstatus.Error(codes.Internal, mapErr.Error())
		}
		pbHistory = append(pbHistory, pbH)
	}

	return &pvzpb.OrderHistoryList{
		History: pbHistory,
	}, nil
}
