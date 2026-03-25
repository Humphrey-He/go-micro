package gateway

import (
	"context"
	"time"

	"go-micro/internal/order"
	"go-micro/pkg/httpx"
)

type Service struct {
	grpc *order.GRPCClient
}

func NewService(grpcClient *order.GRPCClient) *Service {
	return &Service{
		grpc: grpcClient,
	}
}

func (s *Service) CreateOrder(req CreateOrderRequest, requestID string) (httpx.Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	items := make([]order.Item, 0, len(req.Items))
	for _, it := range req.Items {
		items = append(items, order.Item{SkuID: it.SkuID, Quantity: it.Quantity, Price: it.Price})
	}
	resp, err := s.grpc.Create(ctx, order.CreateOrderRequest{
		RequestID: req.RequestID,
		UserID:    req.UserID,
		Items:     items,
		Remark:    req.Remark,
	})
	if err != nil {
		return httpx.Response{}, err
	}
	return httpx.Response{Code: 0, Message: "OK", Data: resp}, nil
}

func (s *Service) GetOrder(orderID, requestID string) (httpx.Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	resp, err := s.grpc.Get(ctx, orderID)
	if err != nil {
		return httpx.Response{}, err
	}
	return httpx.Response{Code: 0, Message: "OK", Data: resp}, nil
}
