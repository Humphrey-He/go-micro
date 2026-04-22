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

func (c *GRPCClient) Release(ctx context.Context, reservedID string) error {
	_, err := c.client.Release(ctx, &inventorypb.ReleaseRequest{ReservedId: reservedID})
	return err
}

func (c *GRPCClient) ReleaseByOrder(ctx context.Context, orderID string) error {
	_, err := c.client.ReleaseByOrder(ctx, &inventorypb.ReleaseByOrderRequest{OrderId: orderID})
	return err
}

func (c *GRPCClient) Confirm(ctx context.Context, reservedID string) error {
	_, err := c.client.Confirm(ctx, &inventorypb.ConfirmRequest{ReservedId: reservedID})
	return err
}

func (c *GRPCClient) GetReservation(ctx context.Context, orderID string) (*inventorypb.Reservation, error) {
	return c.client.GetReservation(ctx, &inventorypb.GetReservationRequest{OrderId: orderID})
}

func (c *GRPCClient) ListInventory(ctx context.Context) ([]InventoryItem, error) {
	resp, err := c.client.ListInventory(ctx, &inventorypb.ListInventoryRequest{})
	if err != nil {
		return nil, err
	}
	items := make([]InventoryItem, 0, len(resp.Items))
	for _, it := range resp.Items {
		items = append(items, InventoryItem{
			SkuID:     it.SkuId,
			Available: int(it.Available),
			Reserved:  int(it.Reserved),
		})
	}
	return items, nil
}
