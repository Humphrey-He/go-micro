package payment

import (
	"context"

	"go-micro/proto/paymentpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GRPCClient struct {
	client paymentpb.PaymentServiceClient
}

func NewGRPCClient(target string) (*GRPCClient, *grpc.ClientConn, error) {
	conn, err := grpc.Dial(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}
	return &GRPCClient{client: paymentpb.NewPaymentServiceClient(conn)}, conn, nil
}

func (c *GRPCClient) Create(ctx context.Context, req CreatePaymentRequest) (*paymentpb.CreatePaymentResponse, error) {
	return c.client.CreatePayment(ctx, &paymentpb.CreatePaymentRequest{
		OrderId:   req.OrderID,
		Amount:    req.Amount,
		RequestId: req.RequestID,
	})
}

func (c *GRPCClient) Get(ctx context.Context, paymentID string) (*paymentpb.Payment, error) {
	return c.client.GetPayment(ctx, &paymentpb.GetPaymentRequest{
		PaymentId: paymentID,
	})
}

func (c *GRPCClient) MarkSuccess(ctx context.Context, paymentID string) error {
	_, err := c.client.MarkSuccess(ctx, &paymentpb.ChangePaymentStatusRequest{
		PaymentId: paymentID,
	})
	return err
}

func (c *GRPCClient) MarkFailed(ctx context.Context, paymentID string) error {
	_, err := c.client.MarkFailed(ctx, &paymentpb.ChangePaymentStatusRequest{
		PaymentId: paymentID,
	})
	return err
}

func (c *GRPCClient) MarkTimeout(ctx context.Context, paymentID string) error {
	_, err := c.client.MarkTimeout(ctx, &paymentpb.ChangePaymentStatusRequest{
		PaymentId: paymentID,
	})
	return err
}

func (c *GRPCClient) Refund(ctx context.Context, paymentID string) error {
	_, err := c.client.Refund(ctx, &paymentpb.ChangePaymentStatusRequest{
		PaymentId: paymentID,
	})
	return err
}
