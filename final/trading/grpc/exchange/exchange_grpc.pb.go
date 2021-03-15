// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package exchange

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// ExchangeClient is the client API for Exchange service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ExchangeClient interface {
	// поток ценовых данных от биржи к брокеру
	// мы каждую секнуду будем получать отсюда событие с ценами, которые броке аггрегирует у себя в минуты и показывает клиентам
	// устанавливается 1 раз брокером
	Statistic(ctx context.Context, in *BrokerID, opts ...grpc.CallOption) (Exchange_StatisticClient, error)
	// отправка на биржу заявки от брокера
	Create(ctx context.Context, in *Deal, opts ...grpc.CallOption) (*DealID, error)
	// отмена заявки
	Cancel(ctx context.Context, in *DealID, opts ...grpc.CallOption) (*CancelResult, error)
	// исполнение заявок от биржи к брокеру
	// устанавливается 1 раз брокером и при исполнении какой-то заявки
	Results(ctx context.Context, in *BrokerID, opts ...grpc.CallOption) (Exchange_ResultsClient, error)
}

type exchangeClient struct {
	cc grpc.ClientConnInterface
}

func NewExchangeClient(cc grpc.ClientConnInterface) ExchangeClient {
	return &exchangeClient{cc}
}

func (c *exchangeClient) Statistic(ctx context.Context, in *BrokerID, opts ...grpc.CallOption) (Exchange_StatisticClient, error) {
	stream, err := c.cc.NewStream(ctx, &Exchange_ServiceDesc.Streams[0], "/exchange.Exchange/Statistic", opts...)
	if err != nil {
		return nil, err
	}
	x := &exchangeStatisticClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Exchange_StatisticClient interface {
	Recv() (*OHLCV, error)
	grpc.ClientStream
}

type exchangeStatisticClient struct {
	grpc.ClientStream
}

func (x *exchangeStatisticClient) Recv() (*OHLCV, error) {
	m := new(OHLCV)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *exchangeClient) Create(ctx context.Context, in *Deal, opts ...grpc.CallOption) (*DealID, error) {
	out := new(DealID)
	err := c.cc.Invoke(ctx, "/exchange.Exchange/Create", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *exchangeClient) Cancel(ctx context.Context, in *DealID, opts ...grpc.CallOption) (*CancelResult, error) {
	out := new(CancelResult)
	err := c.cc.Invoke(ctx, "/exchange.Exchange/Cancel", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *exchangeClient) Results(ctx context.Context, in *BrokerID, opts ...grpc.CallOption) (Exchange_ResultsClient, error) {
	stream, err := c.cc.NewStream(ctx, &Exchange_ServiceDesc.Streams[1], "/exchange.Exchange/Results", opts...)
	if err != nil {
		return nil, err
	}
	x := &exchangeResultsClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Exchange_ResultsClient interface {
	Recv() (*Deal, error)
	grpc.ClientStream
}

type exchangeResultsClient struct {
	grpc.ClientStream
}

func (x *exchangeResultsClient) Recv() (*Deal, error) {
	m := new(Deal)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// ExchangeServer is the server API for Exchange service.
// All implementations must embed UnimplementedExchangeServer
// for forward compatibility
type ExchangeServer interface {
	// поток ценовых данных от биржи к брокеру
	// мы каждую секнуду будем получать отсюда событие с ценами, которые броке аггрегирует у себя в минуты и показывает клиентам
	// устанавливается 1 раз брокером
	Statistic(*BrokerID, Exchange_StatisticServer) error
	// отправка на биржу заявки от брокера
	Create(context.Context, *Deal) (*DealID, error)
	// отмена заявки
	Cancel(context.Context, *DealID) (*CancelResult, error)
	// исполнение заявок от биржи к брокеру
	// устанавливается 1 раз брокером и при исполнении какой-то заявки
	Results(*BrokerID, Exchange_ResultsServer) error
	mustEmbedUnimplementedExchangeServer()
}

// UnimplementedExchangeServer must be embedded to have forward compatible implementations.
type UnimplementedExchangeServer struct {
}

func (UnimplementedExchangeServer) Statistic(*BrokerID, Exchange_StatisticServer) error {
	return status.Errorf(codes.Unimplemented, "method Statistic not implemented")
}
func (UnimplementedExchangeServer) Create(context.Context, *Deal) (*DealID, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Create not implemented")
}
func (UnimplementedExchangeServer) Cancel(context.Context, *DealID) (*CancelResult, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Cancel not implemented")
}
func (UnimplementedExchangeServer) Results(*BrokerID, Exchange_ResultsServer) error {
	return status.Errorf(codes.Unimplemented, "method Results not implemented")
}
func (UnimplementedExchangeServer) mustEmbedUnimplementedExchangeServer() {}

// UnsafeExchangeServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ExchangeServer will
// result in compilation errors.
type UnsafeExchangeServer interface {
	mustEmbedUnimplementedExchangeServer()
}

func RegisterExchangeServer(s grpc.ServiceRegistrar, srv ExchangeServer) {
	s.RegisterService(&Exchange_ServiceDesc, srv)
}

func _Exchange_Statistic_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(BrokerID)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(ExchangeServer).Statistic(m, &exchangeStatisticServer{stream})
}

type Exchange_StatisticServer interface {
	Send(*OHLCV) error
	grpc.ServerStream
}

type exchangeStatisticServer struct {
	grpc.ServerStream
}

func (x *exchangeStatisticServer) Send(m *OHLCV) error {
	return x.ServerStream.SendMsg(m)
}

func _Exchange_Create_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Deal)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ExchangeServer).Create(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/exchange.Exchange/Create",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ExchangeServer).Create(ctx, req.(*Deal))
	}
	return interceptor(ctx, in, info, handler)
}

func _Exchange_Cancel_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DealID)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ExchangeServer).Cancel(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/exchange.Exchange/Cancel",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ExchangeServer).Cancel(ctx, req.(*DealID))
	}
	return interceptor(ctx, in, info, handler)
}

func _Exchange_Results_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(BrokerID)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(ExchangeServer).Results(m, &exchangeResultsServer{stream})
}

type Exchange_ResultsServer interface {
	Send(*Deal) error
	grpc.ServerStream
}

type exchangeResultsServer struct {
	grpc.ServerStream
}

func (x *exchangeResultsServer) Send(m *Deal) error {
	return x.ServerStream.SendMsg(m)
}

// Exchange_ServiceDesc is the grpc.ServiceDesc for Exchange service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Exchange_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "exchange.Exchange",
	HandlerType: (*ExchangeServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Create",
			Handler:    _Exchange_Create_Handler,
		},
		{
			MethodName: "Cancel",
			Handler:    _Exchange_Cancel_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Statistic",
			Handler:       _Exchange_Statistic_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "Results",
			Handler:       _Exchange_Results_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "exchange.proto",
}
