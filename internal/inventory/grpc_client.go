package inventory

import (
	"context"

	"go-micro/proto/inventorypb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GRPCClient struct {
	client inventorypb.InventoryServiceClient
}

func NewGRPCClient(target string) (*GRPCClient, *grpc.ClientConn, error) {
	conn, err := grpc.Dial(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}
	return &GRPCClient{client: inventorypb.NewInventoryServiceClient(conn)}, conn, nil
}

func (c *GRPCClient) Reserve(ctx context.Context, orderID string, items []Item) (string, error) {
	req := &inventorypb.ReserveRequest{OrderId: orderID}
	for _, it := range items {
		req.Items = append(req.Items, &inventorypb.Item{SkuId: it.SkuID, Quantity: int32(it.Quantity)})
	}
	resp, err := c.client.Reserve(ctx, req)
	if err != nil {
		return "", err
	}
	return resp.ReservedId, nil
}
