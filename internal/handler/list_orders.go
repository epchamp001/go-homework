package handler

import (
	"context"
	"pvz-cli/internal/handler/mappers"
	"pvz-cli/pkg/errs"
	pvzpb "pvz-cli/pkg/pvz"
	"pvz-cli/pkg/wpool"
	"strconv"

	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
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
	userIDStr := strconv.FormatUint(req.UserId, 10)

	resCh := make(chan wpool.Response, 1)

	s.wp.Submit(wpool.Job{
		Ctx:    ctx,
		Result: resCh,
		Do: func(c context.Context) (any, error) {

			domainOrders, total, err := s.svc.ListOrders(
				c, userIDStr, inPVZ, lastN, pagination,
			)
			if err != nil {
				return nil, err
			}

			pbOrders := make([]*pvzpb.Order, 0, len(domainOrders))
			for _, o := range domainOrders {
				pbo, mapErr := mappers.DomainOrderToProtoOrder(o)
				if mapErr != nil {
					return nil, mapErr
				}
				pbOrders = append(pbOrders, pbo)
			}

			return struct {
				orders []*pvzpb.Order
				total  int
			}{pbOrders, total}, nil
		},
	})

	select {
	case <-ctx.Done():
		return nil, grpcstatus.Error(codes.Canceled, ctx.Err().Error())

	case r := <-resCh:
		if r.Err != nil {
			s.log.Errorw("ListOrders service error",
				"user_id", req.UserId,
				"in_pvz", inPVZ,
				"lastN", lastN,
				"error", r.Err,
			)
			return nil, grpcstatus.Error(codes.Internal, errs.ErrorCause(r.Err))
		}

		data := r.Val.(struct {
			orders []*pvzpb.Order
			total  int
		})

		return &pvzpb.OrdersList{
			Orders: data.orders,
			Total:  int32(data.total),
		}, nil
	}
}
