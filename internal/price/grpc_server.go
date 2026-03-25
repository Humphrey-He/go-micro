package price

import (
	"context"

	"google.golang.org/grpc"
)

const (
	PriceService_Calculate_FullMethodName = "/price.PriceService/Calculate"
	PriceService_History_FullMethodName   = "/price.PriceService/History"
)

type HistoryRequest struct {
	SkuID string `json:"sku_id"`
	Limit int32  `json:"limit"`
}

type HistoryResponse struct {
	Items []History `json:"items"`
}

type PriceServiceClient interface {
	Calculate(ctx context.Context, in *CalculateRequest, opts ...grpc.CallOption) (*CalculateResponse, error)
	History(ctx context.Context, in *HistoryRequest, opts ...grpc.CallOption) (*HistoryResponse, error)
}

type priceServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewPriceServiceClient(cc grpc.ClientConnInterface) PriceServiceClient {
	return &priceServiceClient{cc}
}

func (c *priceServiceClient) Calculate(ctx context.Context, in *CalculateRequest, opts ...grpc.CallOption) (*CalculateResponse, error) {
	out := new(CalculateResponse)
	err := c.cc.Invoke(ctx, PriceService_Calculate_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *priceServiceClient) History(ctx context.Context, in *HistoryRequest, opts ...grpc.CallOption) (*HistoryResponse, error) {
	out := new(HistoryResponse)
	err := c.cc.Invoke(ctx, PriceService_History_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

type PriceServiceServer interface {
	Calculate(context.Context, *CalculateRequest) (*CalculateResponse, error)
	History(context.Context, *HistoryRequest) (*HistoryResponse, error)
}

func RegisterPriceServiceServer(s grpc.ServiceRegistrar, srv PriceServiceServer) {
	s.RegisterService(&PriceService_ServiceDesc, srv)
}

func _PriceService_Calculate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CalculateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PriceServiceServer).Calculate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: PriceService_Calculate_FullMethodName}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PriceServiceServer).Calculate(ctx, req.(*CalculateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _PriceService_History_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HistoryRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PriceServiceServer).History(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: PriceService_History_FullMethodName}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PriceServiceServer).History(ctx, req.(*HistoryRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var PriceService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "price.PriceService",
	HandlerType: (*PriceServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{MethodName: "Calculate", Handler: _PriceService_Calculate_Handler},
		{MethodName: "History", Handler: _PriceService_History_Handler},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/price.proto",
}

type GRPCServer struct {
	svc *Service
}

func NewGRPCServer(svc *Service) *GRPCServer {
	return &GRPCServer{svc: svc}
}

func (s *GRPCServer) Calculate(ctx context.Context, in *CalculateRequest) (*CalculateResponse, error) {
	return s.svc.Calculate(*in)
}

func (s *GRPCServer) History(ctx context.Context, in *HistoryRequest) (*HistoryResponse, error) {
	items, err := s.svc.History(in.SkuID, int(in.Limit))
	if err != nil {
		return nil, err
	}
	return &HistoryResponse{Items: items}, nil
}
