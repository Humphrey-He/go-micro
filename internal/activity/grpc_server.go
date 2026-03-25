package activity

import (
	"context"

	"google.golang.org/grpc"
)

const (
	ActivityService_IssueCoupon_FullMethodName = "/activity.ActivityService/IssueCoupon"
	ActivityService_Seckill_FullMethodName     = "/activity.ActivityService/Seckill"
	ActivityService_GetCoupon_FullMethodName   = "/activity.ActivityService/GetCoupon"
	ActivityService_GetSeckill_FullMethodName  = "/activity.ActivityService/GetSeckill"
)

type ActivityServiceClient interface {
	IssueCoupon(ctx context.Context, in *CouponRequest, opts ...grpc.CallOption) (*Coupon, error)
	Seckill(ctx context.Context, in *SeckillRequest, opts ...grpc.CallOption) (*SeckillOrder, error)
	GetCoupon(ctx context.Context, in *StatusQuery, opts ...grpc.CallOption) (*Coupon, error)
	GetSeckill(ctx context.Context, in *StatusQuery, opts ...grpc.CallOption) (*Seckill, error)
}

type activityServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewActivityServiceClient(cc grpc.ClientConnInterface) ActivityServiceClient {
	return &activityServiceClient{cc}
}

func (c *activityServiceClient) IssueCoupon(ctx context.Context, in *CouponRequest, opts ...grpc.CallOption) (*Coupon, error) {
	out := new(Coupon)
	err := c.cc.Invoke(ctx, ActivityService_IssueCoupon_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *activityServiceClient) Seckill(ctx context.Context, in *SeckillRequest, opts ...grpc.CallOption) (*SeckillOrder, error) {
	out := new(SeckillOrder)
	err := c.cc.Invoke(ctx, ActivityService_Seckill_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *activityServiceClient) GetCoupon(ctx context.Context, in *StatusQuery, opts ...grpc.CallOption) (*Coupon, error) {
	out := new(Coupon)
	err := c.cc.Invoke(ctx, ActivityService_GetCoupon_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *activityServiceClient) GetSeckill(ctx context.Context, in *StatusQuery, opts ...grpc.CallOption) (*Seckill, error) {
	out := new(Seckill)
	err := c.cc.Invoke(ctx, ActivityService_GetSeckill_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

type ActivityServiceServer interface {
	IssueCoupon(context.Context, *CouponRequest) (*Coupon, error)
	Seckill(context.Context, *SeckillRequest) (*SeckillOrder, error)
	GetCoupon(context.Context, *StatusQuery) (*Coupon, error)
	GetSeckill(context.Context, *StatusQuery) (*Seckill, error)
}

func RegisterActivityServiceServer(s grpc.ServiceRegistrar, srv ActivityServiceServer) {
	s.RegisterService(&ActivityService_ServiceDesc, srv)
}

func _ActivityService_IssueCoupon_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CouponRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ActivityServiceServer).IssueCoupon(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: ActivityService_IssueCoupon_FullMethodName}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ActivityServiceServer).IssueCoupon(ctx, req.(*CouponRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ActivityService_Seckill_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SeckillRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ActivityServiceServer).Seckill(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: ActivityService_Seckill_FullMethodName}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ActivityServiceServer).Seckill(ctx, req.(*SeckillRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ActivityService_GetCoupon_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StatusQuery)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ActivityServiceServer).GetCoupon(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: ActivityService_GetCoupon_FullMethodName}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ActivityServiceServer).GetCoupon(ctx, req.(*StatusQuery))
	}
	return interceptor(ctx, in, info, handler)
}

func _ActivityService_GetSeckill_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StatusQuery)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ActivityServiceServer).GetSeckill(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: ActivityService_GetSeckill_FullMethodName}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ActivityServiceServer).GetSeckill(ctx, req.(*StatusQuery))
	}
	return interceptor(ctx, in, info, handler)
}

var ActivityService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "activity.ActivityService",
	HandlerType: (*ActivityServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{MethodName: "IssueCoupon", Handler: _ActivityService_IssueCoupon_Handler},
		{MethodName: "Seckill", Handler: _ActivityService_Seckill_Handler},
		{MethodName: "GetCoupon", Handler: _ActivityService_GetCoupon_Handler},
		{MethodName: "GetSeckill", Handler: _ActivityService_GetSeckill_Handler},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/activity.proto",
}

type GRPCServer struct {
	svc *Service
}

func NewGRPCServer(svc *Service) *GRPCServer {
	return &GRPCServer{svc: svc}
}

func (s *GRPCServer) IssueCoupon(ctx context.Context, in *CouponRequest) (*Coupon, error) {
	return s.svc.IssueCoupon(*in)
}

func (s *GRPCServer) Seckill(ctx context.Context, in *SeckillRequest) (*SeckillOrder, error) {
	return s.svc.Seckill(*in)
}

func (s *GRPCServer) GetCoupon(ctx context.Context, in *StatusQuery) (*Coupon, error) {
	return s.svc.GetCoupon(in.CouponID)
}

func (s *GRPCServer) GetSeckill(ctx context.Context, in *StatusQuery) (*Seckill, error) {
	return s.svc.GetSeckill(in.SkuID)
}
