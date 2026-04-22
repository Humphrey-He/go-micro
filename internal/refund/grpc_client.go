package refund

import (
	"context"

	"go-micro/pkg/grpcx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GRPCClient struct {
	client RefundServiceClient
}

func NewGRPCClient(target string) (*GRPCClient, *grpc.ClientConn, error) {
	conn, err := grpc.Dial(target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.ForceCodec(grpcx.JSONCodec{})),
	)
	if err != nil {
		return nil, nil, err
	}
	return &GRPCClient{client: NewRefundServiceClient(conn)}, conn, nil
}

func (c *GRPCClient) Initiate(ctx context.Context, req InitiateRequest) (*Refund, error) {
	return c.client.Initiate(ctx, &req)
}

func (c *GRPCClient) Status(ctx context.Context, req StatusRequest) (*Refund, error) {
	return c.client.Status(ctx, &req)
}

func (c *GRPCClient) Rollback(ctx context.Context, req RollbackRequest) (*Refund, error) {
	return c.client.Rollback(ctx, &req)
}

func (c *GRPCClient) List(ctx context.Context, req ListRequest) (*ListResponse, error) {
	return c.client.List(ctx, &req)
}
