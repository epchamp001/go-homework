package handler

import (
	"google.golang.org/grpc"
	"pvz-cli/internal/usecase"
	"pvz-cli/pkg/logger"
	pvzpb "pvz-cli/pkg/pvz"
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
