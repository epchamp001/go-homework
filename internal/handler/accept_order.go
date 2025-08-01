package handler

import (
	"context"
	"go.opentelemetry.io/otel"
	"pvz-cli/internal/handler/mappers"
	"pvz-cli/pkg/errs"
	pvzpb "pvz-cli/pkg/pvz"
	"pvz-cli/pkg/wpool"

	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
)

func (s *OrderServiceServer) AcceptOrder(
	ctx context.Context,
	req *pvzpb.AcceptOrderRequest,
) (*pvzpb.OrderResponse, error) {
	ctx, span := otel.Tracer("pvz-cli/handler/orders").Start(ctx, "Handler.AcceptOrder")
	defer span.End()
	
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

	resCh := make(chan wpool.Response, 1)

	s.wp.Submit(wpool.Job{
		Ctx:    ctx,
		Result: resCh,
		Do: func(c context.Context) (any, error) {
			_, err := s.svc.AcceptOrder(
				c,
				domainOrder.ID,
				domainOrder.UserID,
				domainOrder.ExpiresAt,
				domainOrder.Weight,
				domainOrder.Price,
				domainOrder.Package,
			)
			return nil, err
		},
	})

	select {
	case <-ctx.Done():
		return nil, grpcstatus.Error(codes.Canceled, ctx.Err().Error())

	case r := <-resCh:
		if r.Err != nil {
			s.log.Errorw("AcceptOrder service error",
				"order_id", domainOrder.ID,
				"error", r.Err,
			)
			if grpcErr := errs.GrpcError(r.Err); grpcErr != nil {
				return nil, grpcErr
			}
			cause := errs.ErrorCause(r.Err)
			return nil, grpcstatus.Error(codes.Internal, cause)
		}

		return &pvzpb.OrderResponse{
			Status:  pvzpb.OrderStatus_ORDER_STATUS_EXPECTS,
			OrderId: req.OrderId,
		}, nil
	}
}
