// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v6.30.1
// source: pvz/order_service.proto

package pvzpb

import (
	context "context"

	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	OrdersService_AcceptOrder_FullMethodName   = "/orders.OrdersService/AcceptOrder"
	OrdersService_ReturnOrder_FullMethodName   = "/orders.OrdersService/ReturnOrder"
	OrdersService_ProcessOrders_FullMethodName = "/orders.OrdersService/ProcessOrders"
	OrdersService_ListOrders_FullMethodName    = "/orders.OrdersService/ListOrders"
	OrdersService_ListReturns_FullMethodName   = "/orders.OrdersService/ListReturns"
	OrdersService_GetHistory_FullMethodName    = "/orders.OrdersService/GetHistory"
	OrdersService_ImportOrders_FullMethodName  = "/orders.OrdersService/ImportOrders"
)

// OrdersServiceClient is the client API for OrdersService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type OrdersServiceClient interface {
	// Принять заказ от курьера
	AcceptOrder(ctx context.Context, in *AcceptOrderRequest, opts ...grpc.CallOption) (*OrderResponse, error)
	// Вернуть заказ курьеру
	ReturnOrder(ctx context.Context, in *OrderIdRequest, opts ...grpc.CallOption) (*OrderResponse, error)
	// Выдать заказы или принять возврат клиента
	ProcessOrders(ctx context.Context, in *ProcessOrdersRequest, opts ...grpc.CallOption) (*ProcessResult, error)
	// Получить список заказов клиента
	ListOrders(ctx context.Context, in *ListOrdersRequest, opts ...grpc.CallOption) (*OrdersList, error)
	// Получить список возвратов клиентов (постранично, от новых к старым)
	ListReturns(ctx context.Context, in *ListReturnsRequest, opts ...grpc.CallOption) (*ReturnsList, error)
	// Получить историю изменения заказов
	GetHistory(ctx context.Context, in *GetHistoryRequest, opts ...grpc.CallOption) (*OrderHistoryList, error)
	// Импортировать заказы
	ImportOrders(ctx context.Context, in *ImportOrdersRequest, opts ...grpc.CallOption) (*ImportResult, error)
}

type ordersServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewOrdersServiceClient(cc grpc.ClientConnInterface) OrdersServiceClient {
	return &ordersServiceClient{cc}
}

func (c *ordersServiceClient) AcceptOrder(ctx context.Context, in *AcceptOrderRequest, opts ...grpc.CallOption) (*OrderResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(OrderResponse)
	err := c.cc.Invoke(ctx, OrdersService_AcceptOrder_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *ordersServiceClient) ReturnOrder(ctx context.Context, in *OrderIdRequest, opts ...grpc.CallOption) (*OrderResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(OrderResponse)
	err := c.cc.Invoke(ctx, OrdersService_ReturnOrder_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *ordersServiceClient) ProcessOrders(ctx context.Context, in *ProcessOrdersRequest, opts ...grpc.CallOption) (*ProcessResult, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ProcessResult)
	err := c.cc.Invoke(ctx, OrdersService_ProcessOrders_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *ordersServiceClient) ListOrders(ctx context.Context, in *ListOrdersRequest, opts ...grpc.CallOption) (*OrdersList, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(OrdersList)
	err := c.cc.Invoke(ctx, OrdersService_ListOrders_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *ordersServiceClient) ListReturns(ctx context.Context, in *ListReturnsRequest, opts ...grpc.CallOption) (*ReturnsList, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ReturnsList)
	err := c.cc.Invoke(ctx, OrdersService_ListReturns_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *ordersServiceClient) GetHistory(ctx context.Context, in *GetHistoryRequest, opts ...grpc.CallOption) (*OrderHistoryList, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(OrderHistoryList)
	err := c.cc.Invoke(ctx, OrdersService_GetHistory_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *ordersServiceClient) ImportOrders(ctx context.Context, in *ImportOrdersRequest, opts ...grpc.CallOption) (*ImportResult, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ImportResult)
	err := c.cc.Invoke(ctx, OrdersService_ImportOrders_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// OrdersServiceServer is the server API for OrdersService service.
// All implementations must embed UnimplementedOrdersServiceServer
// for forward compatibility.
type OrdersServiceServer interface {
	// Принять заказ от курьера
	AcceptOrder(context.Context, *AcceptOrderRequest) (*OrderResponse, error)
	// Вернуть заказ курьеру
	ReturnOrder(context.Context, *OrderIdRequest) (*OrderResponse, error)
	// Выдать заказы или принять возврат клиента
	ProcessOrders(context.Context, *ProcessOrdersRequest) (*ProcessResult, error)
	// Получить список заказов клиента
	ListOrders(context.Context, *ListOrdersRequest) (*OrdersList, error)
	// Получить список возвратов клиентов (постранично, от новых к старым)
	ListReturns(context.Context, *ListReturnsRequest) (*ReturnsList, error)
	// Получить историю изменения заказов
	GetHistory(context.Context, *GetHistoryRequest) (*OrderHistoryList, error)
	// Импортировать заказы
	ImportOrders(context.Context, *ImportOrdersRequest) (*ImportResult, error)
	mustEmbedUnimplementedOrdersServiceServer()
}

// UnimplementedOrdersServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedOrdersServiceServer struct{}

func (UnimplementedOrdersServiceServer) AcceptOrder(context.Context, *AcceptOrderRequest) (*OrderResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AcceptOrder not implemented")
}
func (UnimplementedOrdersServiceServer) ReturnOrder(context.Context, *OrderIdRequest) (*OrderResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ReturnOrder not implemented")
}
func (UnimplementedOrdersServiceServer) ProcessOrders(context.Context, *ProcessOrdersRequest) (*ProcessResult, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ProcessOrders not implemented")
}
func (UnimplementedOrdersServiceServer) ListOrders(context.Context, *ListOrdersRequest) (*OrdersList, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListOrders not implemented")
}
func (UnimplementedOrdersServiceServer) ListReturns(context.Context, *ListReturnsRequest) (*ReturnsList, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListReturns not implemented")
}
func (UnimplementedOrdersServiceServer) GetHistory(context.Context, *GetHistoryRequest) (*OrderHistoryList, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetHistory not implemented")
}
func (UnimplementedOrdersServiceServer) ImportOrders(context.Context, *ImportOrdersRequest) (*ImportResult, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ImportOrders not implemented")
}
func (UnimplementedOrdersServiceServer) mustEmbedUnimplementedOrdersServiceServer() {}
func (UnimplementedOrdersServiceServer) testEmbeddedByValue()                       {}

// UnsafeOrdersServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to OrdersServiceServer will
// result in compilation errors.
type UnsafeOrdersServiceServer interface {
	mustEmbedUnimplementedOrdersServiceServer()
}

func RegisterOrdersServiceServer(s grpc.ServiceRegistrar, srv OrdersServiceServer) {
	// If the following call pancis, it indicates UnimplementedOrdersServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&OrdersService_ServiceDesc, srv)
}

func _OrdersService_AcceptOrder_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AcceptOrderRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OrdersServiceServer).AcceptOrder(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: OrdersService_AcceptOrder_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OrdersServiceServer).AcceptOrder(ctx, req.(*AcceptOrderRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _OrdersService_ReturnOrder_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(OrderIdRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OrdersServiceServer).ReturnOrder(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: OrdersService_ReturnOrder_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OrdersServiceServer).ReturnOrder(ctx, req.(*OrderIdRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _OrdersService_ProcessOrders_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ProcessOrdersRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OrdersServiceServer).ProcessOrders(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: OrdersService_ProcessOrders_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OrdersServiceServer).ProcessOrders(ctx, req.(*ProcessOrdersRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _OrdersService_ListOrders_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListOrdersRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OrdersServiceServer).ListOrders(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: OrdersService_ListOrders_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OrdersServiceServer).ListOrders(ctx, req.(*ListOrdersRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _OrdersService_ListReturns_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListReturnsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OrdersServiceServer).ListReturns(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: OrdersService_ListReturns_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OrdersServiceServer).ListReturns(ctx, req.(*ListReturnsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _OrdersService_GetHistory_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetHistoryRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OrdersServiceServer).GetHistory(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: OrdersService_GetHistory_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OrdersServiceServer).GetHistory(ctx, req.(*GetHistoryRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _OrdersService_ImportOrders_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ImportOrdersRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OrdersServiceServer).ImportOrders(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: OrdersService_ImportOrders_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OrdersServiceServer).ImportOrders(ctx, req.(*ImportOrdersRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// OrdersService_ServiceDesc is the grpc.ServiceDesc for OrdersService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var OrdersService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "orders.OrdersService",
	HandlerType: (*OrdersServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "AcceptOrder",
			Handler:    _OrdersService_AcceptOrder_Handler,
		},
		{
			MethodName: "ReturnOrder",
			Handler:    _OrdersService_ReturnOrder_Handler,
		},
		{
			MethodName: "ProcessOrders",
			Handler:    _OrdersService_ProcessOrders_Handler,
		},
		{
			MethodName: "ListOrders",
			Handler:    _OrdersService_ListOrders_Handler,
		},
		{
			MethodName: "ListReturns",
			Handler:    _OrdersService_ListReturns_Handler,
		},
		{
			MethodName: "GetHistory",
			Handler:    _OrdersService_GetHistory_Handler,
		},
		{
			MethodName: "ImportOrders",
			Handler:    _OrdersService_ImportOrders_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "pvz/order_service.proto",
}
