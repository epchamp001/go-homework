package handler

import (
	"context"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
	"pvz-cli/internal/handler/mappers"
	pvzpb "pvz-cli/pkg/pvz"
	"strconv"
)

func (s *OrderServiceServer) ListOrders(
	ctx context.Context,
	req *pvzpb.ListOrdersRequest,
) (*pvzpb.OrdersList, error) {
	if req == nil {
		return nil, grpcstatus.Error(codes.InvalidArgument, "ListOrdersRequest is nil")
	}
	if err := req.Validate(); err != nil {
		return nil, grpcstatus.Error(codes.InvalidArgument, err.Error())
	}

	lastN := 0
	if req.LastN != nil {
		lastN = int(*req.LastN)
	}
	pagination := mappers.ProtoToDomainPagination(req.Pagination)
	inPVZ := req.InPvz

	domainOrders, total, err := s.svc.ListOrders(
		ctx,
		strconv.FormatUint(req.UserId, 10),
		inPVZ,
		lastN,
		pagination,
	)
	if err != nil {
		s.log.Errorw("ListOrders service error",
			"user_id", req.UserId,
			"in_pvz", inPVZ,
			"lastN", lastN,
			"error", err,
		)
		return nil, grpcstatus.Error(codes.Internal, err.Error())
	}

	pbOrders := make([]*pvzpb.Order, 0, len(domainOrders))
	for _, o := range domainOrders {
		pbOrder, mapErr := mappers.DomainOrderToProtoOrder(o)
		if mapErr != nil {
			return nil, grpcstatus.Error(codes.Internal, mapErr.Error())
		}
		pbOrders = append(pbOrders, pbOrder)
	}

	return &pvzpb.OrdersList{
		Orders: pbOrders,
		Total:  int32(total),
	}, nil
}
