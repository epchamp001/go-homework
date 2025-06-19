package handler

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
	"pvz-cli/internal/domain/models"
	"pvz-cli/internal/handler/mappers"
	"pvz-cli/internal/usecase"
	"pvz-cli/pkg/errs"
	"pvz-cli/pkg/logger"
	pvzpb "pvz-cli/pkg/pvz"
	"strconv"
)

type OrderServiceServer struct {
	pvzpb.UnimplementedOrdersServiceServer
	svc usecase.Service
	log logger.Logger
}

func NewOrderServiceServer(svc usecase.Service, log logger.Logger) *OrderServiceServer {
	return &OrderServiceServer{
		svc: svc,
		log: log,
	}
}

func RegisterOrderService(grpcServer *grpc.Server, svc usecase.Service, log logger.Logger) {
	pvzpb.RegisterOrdersServiceServer(grpcServer, NewOrderServiceServer(svc, log))
}

func (s *OrderServiceServer) AcceptOrder(
	ctx context.Context,
	req *pvzpb.AcceptOrderRequest,
) (*pvzpb.OrderResponse, error) {
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

	_, err = s.svc.AcceptOrder(
		ctx,
		domainOrder.ID,
		domainOrder.UserID,
		domainOrder.ExpiresAt,
		domainOrder.Weight,
		domainOrder.Price,
		domainOrder.Package,
	)
	if err != nil {
		// логируем только ошибку svc
		s.log.Errorw("AcceptOrder service error",
			"order_id", domainOrder.ID,
			"error", err,
		)
		if grpcErr := errs.GrpcError(err); grpcErr != nil {
			return nil, grpcErr
		}
		return nil, grpcstatus.Error(codes.Internal, err.Error())
	}

	return &pvzpb.OrderResponse{
		Status:  pvzpb.OrderStatus_ORDER_STATUS_EXPECTS,
		OrderId: req.OrderId,
	}, nil
}

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

	orderIDStr := strconv.FormatUint(req.OrderId, 10)
	if err := s.svc.ReturnOrder(ctx, orderIDStr); err != nil {
		s.log.Errorw("ReturnOrder service error",
			"order_id", orderIDStr,
			"error", err,
		)
		if grpcErr := errs.GrpcError(err); grpcErr != nil {
			return nil, grpcErr
		}
		return nil, grpcstatus.Error(codes.Internal, err.Error())
	}

	return mappers.DomainToProtoOrderResponse("ORDER_STATUS_DELETED", orderIDStr)
}

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

	userIDStr := strconv.FormatUint(req.UserId, 10)
	ids := make([]string, 0, len(req.OrderIds))
	for _, id := range req.OrderIds {
		ids = append(ids, strconv.FormatUint(id, 10))
	}

	var (
		result map[string]error
		err    error
	)

	switch req.Action {
	case pvzpb.ActionType_ACTION_TYPE_ISSUE:
		result, err = s.svc.IssueOrders(ctx, userIDStr, ids)
	case pvzpb.ActionType_ACTION_TYPE_RETURN:
		result, err = s.svc.ReturnOrdersByClient(ctx, userIDStr, ids)
	default:
		return nil, grpcstatus.Error(codes.InvalidArgument, "unknown action")
	}

	if err != nil {
		s.log.Errorw("ProcessOrders service error",
			"action", req.Action,
			"user_id", userIDStr,
			"error", err,
		)
		return nil, grpcstatus.Error(codes.Internal, err.Error())
	}

	protoResult, mapErr := mappers.DomainProcessResultToProtoProcessResult(result)
	if mapErr != nil {
		return nil, grpcstatus.Error(codes.Internal, mapErr.Error())
	}
	return protoResult, nil
}

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
		return nil, grpcstatus.Error(codes.Internal, err.Error())
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

func (s *OrderServiceServer) GetHistory(
	ctx context.Context,
	req *pvzpb.GetHistoryRequest,
) (*pvzpb.OrderHistoryList, error) {
	if req != nil {
		if err := req.Validate(); err != nil {
			return nil, grpcstatus.Error(codes.InvalidArgument, err.Error())
		}
	}

	pagination := mappers.ProtoToDomainPagination(req.Pagination)

	historyEvents, _, err := s.svc.OrderHistory(ctx, pagination)
	if err != nil {
		s.log.Errorw("GetHistory service error",
			"pagination", req.GetPagination(),
			"error", err,
		)
		if grpcErr := errs.GrpcError(err); grpcErr != nil {
			return nil, grpcErr
		}
		return nil, grpcstatus.Error(codes.Internal, err.Error())
	}

	pbHistory := make([]*pvzpb.OrderHistory, 0, len(historyEvents))
	for _, h := range historyEvents {
		pbH, mapErr := mappers.DomainOrderHistoryToProtoOrderHistory(h)
		if mapErr != nil {
			return nil, grpcstatus.Error(codes.Internal, mapErr.Error())
		}
		pbHistory = append(pbHistory, pbH)
	}

	return &pvzpb.OrderHistoryList{
		History: pbHistory,
	}, nil
}

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
