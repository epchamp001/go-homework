package handler

import (
	"context"
	pvzpb "pvz-cli/pkg/pvz"
	"pvz-cli/pkg/wpool"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ServiceServer struct {
	pvzpb.UnimplementedAdminServiceServer
	wp *wpool.Pool
}

func NewAdminServiceServer(wp *wpool.Pool) *ServiceServer {
	return &ServiceServer{wp: wp}
}

func RegisterAdminService(grpcServer *grpc.Server, wp *wpool.Pool) {
	pvzpb.RegisterAdminServiceServer(grpcServer, NewAdminServiceServer(wp))
}

func (s *ServiceServer) ResizePool(ctx context.Context, req *pvzpb.ResizeRequest) (*emptypb.Empty, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	s.wp.Resize(int(req.GetSize()))
	return &emptypb.Empty{}, nil
}
