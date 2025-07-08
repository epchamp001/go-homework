package handler

import (
	"pvz-cli/internal/usecase/service"
	"pvz-cli/pkg/logger"
	pvzpb "pvz-cli/pkg/pvz"
	"pvz-cli/pkg/wpool"

	"google.golang.org/grpc"
)

type OrderServiceServer struct {
	pvzpb.UnimplementedOrdersServiceServer
	svc service.Service
	log logger.Logger
	wp  *wpool.Pool
}

func NewOrderServiceServer(svc service.Service, log logger.Logger, wp *wpool.Pool) *OrderServiceServer {
	return &OrderServiceServer{
		svc: svc,
		log: log,
		wp:  wp,
	}
}

func RegisterOrderService(grpcServer *grpc.Server, svc service.Service, log logger.Logger, wp *wpool.Pool) {
	pvzpb.RegisterOrdersServiceServer(grpcServer, NewOrderServiceServer(svc, log, wp))
}
