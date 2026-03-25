package price

import (
	"context"

	"go-micro/pkg/grpcx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GRPCClient struct {
	client PriceServiceClient
}

func NewGRPCClient(target string) (*GRPCClient, *grpc.ClientConn, error) {
	conn, err := grpc.Dial(target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.ForceCodec(grpcx.JSONCodec{})),
	)
	if err != nil {
		return nil, nil, err
	}
	return &GRPCClient{client: NewPriceServiceClient(conn)}, conn, nil
}

func (c *GRPCClient) Calculate(ctx context.Context, req CalculateRequest) (*CalculateResponse, error) {
	return c.client.Calculate(ctx, &req)
}

func (c *GRPCClient) History(ctx context.Context, skuID string, limit int) ([]History, error) {
	resp, err := c.client.History(ctx, &HistoryRequest{SkuID: skuID, Limit: int32(limit)})
	if err != nil {
		return nil, err
	}
	return resp.Items, nil
}
