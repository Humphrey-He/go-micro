package order

import (
	"context"

	"go-micro/proto/orderpb"
)

type GRPCServer struct {
	orderpb.UnimplementedOrderServiceServer
	svc *Service
}

func NewGRPCServer(svc *Service) *GRPCServer {
	return &GRPCServer{svc: svc}
}

func (s *GRPCServer) CreateOrder(ctx context.Context, req *orderpb.CreateOrderRequest) (*orderpb.CreateOrderResponse, error) {
	items := make([]Item, 0, len(req.Items))
	for _, it := range req.Items {
		items = append(items, Item{SkuID: it.SkuId, Quantity: int(it.Quantity), Price: it.Price})
	}
	resp, err := s.svc.Create(CreateOrderRequest{RequestID: req.RequestId, UserID: req.UserId, Items: items, Remark: req.Remark})
	if err != nil {
		return nil, err
	}
	return &orderpb.CreateOrderResponse{OrderId: resp.OrderID, BizNo: resp.BizNo, Status: resp.Status}, nil
}

func (s *GRPCServer) GetOrder(ctx context.Context, req *orderpb.GetOrderRequest) (*orderpb.Order, error) {
	ord, err := s.svc.Get(req.OrderId)
	if err != nil {
		return nil, err
	}
	resp := &orderpb.Order{OrderId: ord.OrderID, BizNo: ord.BizNo, UserId: ord.UserID, Status: ord.Status, TotalAmount: ord.TotalAmount}
	for _, it := range ord.Items {
		resp.Items = append(resp.Items, &orderpb.Item{SkuId: it.SkuID, Quantity: int32(it.Quantity), Price: it.Price})
	}
	return resp, nil
}
