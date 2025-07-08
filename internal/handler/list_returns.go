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

func (s *OrderServiceServer) ListReturns(
	ctx context.Context,
	req *pvzpb.ListReturnsRequest,
) (*pvzpb.ReturnsList, error) {
	if req != nil {
		if err := req.Validate(); err != nil {
			return nil, grpcstatus.Error(codes.InvalidArgument, err.Error())
		}
	}

	pagination := mappers.ProtoToDomainPagination(req.Pagination)

	resCh := make(chan wpool.Response, 1)

	s.wp.Submit(wpool.Job{
		Ctx:    ctx,
		Result: resCh,
		Do: func(c context.Context) (any, error) {

			returnRecords, err := s.svc.ListReturns(c, pagination)
			if err != nil {
				return nil, err
			}

			pbReturns := make([]*pvzpb.ReturnRecord, 0, len(returnRecords))
			for _, r := range returnRecords {
				pbr, mapErr := mappers.DomainReturnRecordToProtoReturnRecord(r)
				if mapErr != nil {
					return nil, mapErr
				}
				pbReturns = append(pbReturns, pbr)
			}
			return pbReturns, nil
		},
	})

	select {
	case <-ctx.Done():
		return nil, grpcstatus.Error(codes.Canceled, ctx.Err().Error())

	case r := <-resCh:
		if r.Err != nil {
			s.log.Errorw("ListReturns service error",
				"pagination", req.GetPagination(),
				"error", r.Err,
			)
			if grpcErr := errs.GrpcError(r.Err); grpcErr != nil {
				return nil, grpcErr
			}
			return nil, grpcstatus.Error(codes.Internal, errs.ErrorCause(r.Err))
		}

		return &pvzpb.ReturnsList{
			Returns: r.Val.([]*pvzpb.ReturnRecord),
		}, nil
	}
}
