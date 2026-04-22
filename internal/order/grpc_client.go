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

func (c *GRPCClient) Cancel(ctx context.Context, orderID string) error {
	_, err := c.client.CancelOrder(ctx, &orderpb.CancelOrderRequest{OrderId: orderID})
	return err
}

func (c *GRPCClient) List(ctx context.Context, req ListOrdersRequest) (*orderpb.ListOrdersResponse, error) {
	return c.client.ListOrders(ctx, &orderpb.ListOrdersRequest{
		Page:      req.Page,
		PageSize:  req.PageSize,
		OrderNo:   req.OrderNo,
		UserId:    req.UserID,
		Status:    req.Status,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		SortBy:    req.SortBy,
		SortOrder: req.SortOrder,
	})
}
