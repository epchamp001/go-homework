package handler

import (
	"context"
	"pvz-cli/internal/domain/models"
	"pvz-cli/internal/handler/mappers"
	"pvz-cli/pkg/errs"
	pvzpb "pvz-cli/pkg/pvz"

	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
)

func (s *OrderServiceServer) ImportOrders(
	ctx context.Context,
	req *pvzpb.ImportOrdersRequest,
) (*pvzpb.ImportResult, error) {
	if req == nil {
		return nil, grpcstatus.Error(codes.InvalidArgument, "ImportOrdersRequest is nil")
	}
	if err := req.Validate(); err != nil {
		return nil, grpcstatus.Error(codes.InvalidArgument, err.Error())
	}

	importList := make([]*models.Order, 0, len(req.Orders))
	for _, o := range req.Orders {
		domainOrder, mapErr := mappers.ProtoToDomainOrderForImport(o)
		if mapErr != nil {
			if grpcErr := errs.GrpcError(mapErr); grpcErr != nil {
				return nil, grpcErr
			}
			return nil, grpcstatus.Error(codes.Internal, mapErr.Error())
		}
		importList = append(importList, domainOrder)
	}

	count, err := s.svc.ImportOrders(ctx, importList)
	if err != nil {
		s.log.Errorw("ImportOrders service error",
			"num_requests", len(importList),
			"error", err,
		)
		if grpcErr := errs.GrpcError(err); grpcErr != nil {
			return nil, grpcErr
		}
		return nil, grpcstatus.Error(codes.Internal, err.Error())
	}

	return &pvzpb.ImportResult{
		Imported: int32(count),
	}, nil
}
