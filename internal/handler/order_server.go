package handler

import (
	"pvz-cli/internal/usecase/service"
	"pvz-cli/pkg/logger"
	pvzpb "pvz-cli/pkg/pvz"

	"google.golang.org/grpc"
)

type OrderServiceServer struct {
	pvzpb.UnimplementedOrdersServiceServer
	svc service.Service
	log logger.Logger
}

func NewOrderServiceServer(svc service.Service, log logger.Logger) *OrderServiceServer {
	return &OrderServiceServer{
		svc: svc,
		log: log,
	}
}

func RegisterOrderService(grpcServer *grpc.Server, svc service.Service, log logger.Logger) {
	pvzpb.RegisterOrdersServiceServer(grpcServer, NewOrderServiceServer(svc, log))
}
