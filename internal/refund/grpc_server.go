package refund

import (
	"context"

	"google.golang.org/grpc"
)

const (
	RefundService_Initiate_FullMethodName = "/refund.RefundService/Initiate"
	RefundService_Status_FullMethodName   = "/refund.RefundService/Status"
	RefundService_Rollback_FullMethodName = "/refund.RefundService/Rollback"
)

type RefundServiceClient interface {
	Initiate(ctx context.Context, in *InitiateRequest, opts ...grpc.CallOption) (*Refund, error)
	Status(ctx context.Context, in *StatusRequest, opts ...grpc.CallOption) (*Refund, error)
	Rollback(ctx context.Context, in *RollbackRequest, opts ...grpc.CallOption) (*Refund, error)
}

type refundServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewRefundServiceClient(cc grpc.ClientConnInterface) RefundServiceClient {
	return &refundServiceClient{cc}
}

func (c *refundServiceClient) Initiate(ctx context.Context, in *InitiateRequest, opts ...grpc.CallOption) (*Refund, error) {
	out := new(Refund)
	err := c.cc.Invoke(ctx, RefundService_Initiate_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *refundServiceClient) Status(ctx context.Context, in *StatusRequest, opts ...grpc.CallOption) (*Refund, error) {
	out := new(Refund)
	err := c.cc.Invoke(ctx, RefundService_Status_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *refundServiceClient) Rollback(ctx context.Context, in *RollbackRequest, opts ...grpc.CallOption) (*Refund, error) {
	out := new(Refund)
	err := c.cc.Invoke(ctx, RefundService_Rollback_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

type RefundServiceServer interface {
	Initiate(context.Context, *InitiateRequest) (*Refund, error)
	Status(context.Context, *StatusRequest) (*Refund, error)
	Rollback(context.Context, *RollbackRequest) (*Refund, error)
}

func RegisterRefundServiceServer(s grpc.ServiceRegistrar, srv RefundServiceServer) {
	s.RegisterService(&RefundService_ServiceDesc, srv)
}

func _RefundService_Initiate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(InitiateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RefundServiceServer).Initiate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: RefundService_Initiate_FullMethodName}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RefundServiceServer).Initiate(ctx, req.(*InitiateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RefundService_Status_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StatusRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RefundServiceServer).Status(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: RefundService_Status_FullMethodName}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RefundServiceServer).Status(ctx, req.(*StatusRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RefundService_Rollback_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RollbackRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RefundServiceServer).Rollback(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: RefundService_Rollback_FullMethodName}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RefundServiceServer).Rollback(ctx, req.(*RollbackRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var RefundService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "refund.RefundService",
	HandlerType: (*RefundServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{MethodName: "Initiate", Handler: _RefundService_Initiate_Handler},
		{MethodName: "Status", Handler: _RefundService_Status_Handler},
		{MethodName: "Rollback", Handler: _RefundService_Rollback_Handler},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/refund.proto",
}

type GRPCServer struct {
	svc *Service
}

func NewGRPCServer(svc *Service) *GRPCServer {
	return &GRPCServer{svc: svc}
}

func (s *GRPCServer) Initiate(ctx context.Context, in *InitiateRequest) (*Refund, error) {
	return s.svc.Initiate(*in)
}

func (s *GRPCServer) Status(ctx context.Context, in *StatusRequest) (*Refund, error) {
	return s.svc.Get(in.RefundID)
}

func (s *GRPCServer) Rollback(ctx context.Context, in *RollbackRequest) (*Refund, error) {
	return s.svc.Rollback(*in)
}
