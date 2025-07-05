package handler

import (
	"context"
	"pvz-cli/internal/handler/mappers"
	"pvz-cli/pkg/errs"
	pvzpb "pvz-cli/pkg/pvz"
	"pvz-cli/pkg/wpool"

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

	pg := mappers.ProtoToDomainPagination(req.GetPagination())

	resCh := make(chan wpool.Response, 1)

	s.wp.Submit(wpool.Job{
		Ctx:    ctx,
		Result: resCh,
		Do: func(c context.Context) (any, error) {
			historyEvents, _, err := s.svc.OrderHistory(c, pg)
			if err != nil {
				return nil, err
			}

			pbHistory := make([]*pvzpb.OrderHistory, 0, len(historyEvents))
			for _, h := range historyEvents {
				pbh, mapErr := mappers.DomainOrderHistoryToProtoOrderHistory(h)
				if mapErr != nil {
					return nil, mapErr
				}
				pbHistory = append(pbHistory, pbh)
			}
			return pbHistory, nil
		},
	})

	select {
	case <-ctx.Done():
		return nil, grpcstatus.Error(codes.Canceled, ctx.Err().Error())

	case r := <-resCh:
		if r.Err != nil {
			s.log.Errorw("GetHistory service error",
				"pagination", req.GetPagination(),
				"error", r.Err,
			)
			if grpcErr := errs.GrpcError(r.Err); grpcErr != nil {
				return nil, grpcErr
			}
			return nil, grpcstatus.Error(codes.Internal, errs.ErrorCause(r.Err))
		}

		return &pvzpb.OrderHistoryList{
			History: r.Val.([]*pvzpb.OrderHistory),
		}, nil
	}
}
