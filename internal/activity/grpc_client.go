package activity

import (
	"context"

	"go-micro/pkg/grpcx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GRPCClient struct {
	client ActivityServiceClient
}

func NewGRPCClient(target string) (*GRPCClient, *grpc.ClientConn, error) {
	conn, err := grpc.Dial(target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.ForceCodec(grpcx.JSONCodec{})),
	)
	if err != nil {
		return nil, nil, err
	}
	return &GRPCClient{client: NewActivityServiceClient(conn)}, conn, nil
}

func (c *GRPCClient) IssueCoupon(ctx context.Context, req CouponRequest) (*Coupon, error) {
	return c.client.IssueCoupon(ctx, &req)
}

func (c *GRPCClient) Seckill(ctx context.Context, req SeckillRequest) (*SeckillOrder, error) {
	return c.client.Seckill(ctx, &req)
}

func (c *GRPCClient) GetCoupon(ctx context.Context, couponID string) (*Coupon, error) {
	return c.client.GetCoupon(ctx, &StatusQuery{CouponID: couponID})
}

func (c *GRPCClient) GetSeckill(ctx context.Context, skuID string) (*Seckill, error) {
	return c.client.GetSeckill(ctx, &StatusQuery{SkuID: skuID})
}
