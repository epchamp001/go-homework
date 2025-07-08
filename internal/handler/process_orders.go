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

func (s *OrderServiceServer) ProcessOrders(
	ctx context.Context,
	req *pvzpb.ProcessOrdersRequest,
) (*pvzpb.ProcessResult, error) {
	if req == nil {
		return nil, grpcstatus.Error(codes.InvalidArgument, "ProcessOrdersRequest is nil")
	}
	if err := req.Validate(); err != nil {
		return nil, grpcstatus.Error(codes.InvalidArgument, err.Error())
	}

	userID := strconv.FormatUint(req.UserId, 10)

	ids := make([]string, 0, len(req.OrderIds))
	for _, id := range req.OrderIds {
		ids = append(ids, strconv.FormatUint(id, 10))
	}

	if req.Action != pvzpb.ActionType_ACTION_TYPE_ISSUE &&
		req.Action != pvzpb.ActionType_ACTION_TYPE_RETURN {
		return nil, grpcstatus.Error(codes.InvalidArgument, "unknown action")
	}

	resCh := make(chan wpool.Response, 1)

	s.wp.Submit(wpool.Job{
		Ctx:    ctx,
		Result: resCh,
		Do: func(c context.Context) (any, error) {

			var (
				res map[string]error
				err error
			)

			switch req.Action {
			case pvzpb.ActionType_ACTION_TYPE_ISSUE:
				res, err = s.svc.IssueOrders(c, userID, ids)
			case pvzpb.ActionType_ACTION_TYPE_RETURN:
				res, err = s.svc.ReturnOrdersByClient(c, userID, ids)
			}
			if err != nil {
				return nil, err
			}

			return mappers.DomainProcessResultToProtoProcessResult(res)
		},
	})

	select {
	case <-ctx.Done():
		return nil, grpcstatus.Error(codes.Canceled, ctx.Err().Error())

	case r := <-resCh:
		if r.Err != nil {
			s.log.Errorw("ProcessOrders service error",
				"action", req.Action,
				"user_id", userID,
				"error", r.Err,
			)
			if grpcErr := errs.GrpcError(r.Err); grpcErr != nil {
				return nil, grpcErr
			}
			return nil, grpcstatus.Error(codes.Internal, errs.ErrorCause(r.Err))
		}

		return r.Val.(*pvzpb.ProcessResult), nil
	}
}
