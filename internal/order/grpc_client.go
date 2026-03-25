package order

import (
	"context"

	"go-micro/proto/orderpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GRPCClient struct {
	client orderpb.OrderServiceClient
}

func NewGRPCClient(target string) (*GRPCClient, *grpc.ClientConn, error) {
	conn, err := grpc.Dial(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}
	return &GRPCClient{client: orderpb.NewOrderServiceClient(conn)}, conn, nil
}

func (c *GRPCClient) Create(ctx context.Context, req CreateOrderRequest) (*orderpb.CreateOrderResponse, error) {
	pb := &orderpb.CreateOrderRequest{RequestId: req.RequestID, UserId: req.UserID, Remark: req.Remark}
	for _, it := range req.Items {
		pb.Items = append(pb.Items, &orderpb.Item{SkuId: it.SkuID, Quantity: int32(it.Quantity), Price: it.Price})
	}
	return c.client.CreateOrder(ctx, pb)
}

func (c *GRPCClient) Get(ctx context.Context, orderID string) (*orderpb.Order, error) {
	return c.client.GetOrder(ctx, &orderpb.GetOrderRequest{OrderId: orderID})
}

func (c *GRPCClient) GetByBizNo(ctx context.Context, bizNo string) (*orderpb.Order, error) {
	return c.client.GetOrderByBizNo(ctx, &orderpb.GetOrderByBizNoRequest{BizNo: bizNo})
}

func (c *GRPCClient) UpdateStatus(ctx context.Context, orderID, from, to string) error {
	_, err := c.client.UpdateOrderStatus(ctx, &orderpb.UpdateOrderStatusRequest{OrderId: orderID, FromStatus: from, ToStatus: to})
	return err
}
