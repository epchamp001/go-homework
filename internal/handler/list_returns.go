package handler

import (
	"context"
	"pvz-cli/internal/handler/mappers"
	"pvz-cli/pkg/errs"
	pvzpb "pvz-cli/pkg/pvz"

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

	returnRecords, err := s.svc.ListReturns(ctx, pagination)
	if err != nil {
		s.log.Errorw("ListReturns service error",
			"pagination", req.GetPagination(),
			"error", err,
		)
		if grpcErr := errs.GrpcError(err); grpcErr != nil {
			return nil, grpcErr
		}
		cause := errs.ErrorCause(err)
		return nil, grpcstatus.Error(codes.Internal, cause)
	}

	pbReturns := make([]*pvzpb.ReturnRecord, 0, len(returnRecords))
	for _, r := range returnRecords {
		pbR, mapErr := mappers.DomainReturnRecordToProtoReturnRecord(r)
		if mapErr != nil {
			return nil, grpcstatus.Error(codes.Internal, mapErr.Error())
		}
		pbReturns = append(pbReturns, pbR)
	}

	return &pvzpb.ReturnsList{
		Returns: pbReturns,
	}, nil
}
