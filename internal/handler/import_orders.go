package handler

import (
	"context"
	"pvz-cli/internal/domain/models"
	"pvz-cli/internal/handler/mappers"
	"pvz-cli/pkg/errs"
	pvzpb "pvz-cli/pkg/pvz"
	"pvz-cli/pkg/wpool"

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

	resCh := make(chan wpool.Response, 1)

	s.wp.Submit(wpool.Job{
		Ctx:    ctx,
		Result: resCh,
		Do: func(c context.Context) (any, error) {

			importList := make([]*models.Order, 0, len(req.Orders))
			for _, o := range req.Orders {
				d, mapErr := mappers.ProtoToDomainOrderForImport(o)
				if mapErr != nil {
					return nil, mapErr
				}
				importList = append(importList, d)
			}

			count, err := s.svc.ImportOrders(c, importList)
			if err != nil {
				return nil, err
			}
			return count, nil
		},
	})

	select {
	case <-ctx.Done():
		return nil, grpcstatus.Error(codes.Canceled, ctx.Err().Error())

	case r := <-resCh:
		if r.Err != nil {
			s.log.Errorw("ImportOrders service error",
				"num_requests", len(req.Orders),
				"error", r.Err,
			)
			if grpcErr := errs.GrpcError(r.Err); grpcErr != nil {
				return nil, grpcErr
			}
			return nil, grpcstatus.Error(codes.Internal, errs.ErrorCause(r.Err))
		}

		return &pvzpb.ImportResult{
			Imported: int32(r.Val.(int)),
		}, nil
	}
}
