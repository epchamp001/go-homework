package handler

import (
	"context"
	errCodes "pvz-cli/internal/domain/codes"
	"pvz-cli/internal/handler/mappers"
	"pvz-cli/pkg/errs"
	pvzpb "pvz-cli/pkg/pvz"
	"pvz-cli/pkg/wpool"
	"strconv"

	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
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

	orderID := strconv.FormatUint(req.OrderId, 10)

	resCh := make(chan wpool.Response, 1)

	s.wp.Submit(wpool.Job{
		Ctx:    ctx,
		Result: resCh,
		Do: func(c context.Context) (any, error) {

			if err := s.svc.ReturnOrder(c, orderID); err != nil {
				return nil, err
			}

			return mappers.DomainToProtoOrderResponse(
				errCodes.CodeOrderReturned,
				orderID,
			)
		},
	})

	select {
	case <-ctx.Done():
		return nil, grpcstatus.Error(codes.Canceled, ctx.Err().Error())

	case r := <-resCh:
		if r.Err != nil {
			s.log.Errorw("ReturnOrder service error",
				"order_id", orderID,
				"error", r.Err,
			)
			if grpcErr := errs.GrpcError(r.Err); grpcErr != nil {
				return nil, grpcErr
			}
			return nil, grpcstatus.Error(codes.Internal, errs.ErrorCause(r.Err))
		}

		return r.Val.(*pvzpb.OrderResponse), nil
	}
}
