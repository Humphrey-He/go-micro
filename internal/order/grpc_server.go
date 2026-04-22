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

func (s *GRPCServer) GetOrderByBizNo(ctx context.Context, req *orderpb.GetOrderByBizNoRequest) (*orderpb.Order, error) {
	ord, err := s.svc.GetByBizNo(req.BizNo)
	if err != nil {
		return nil, err
	}
	resp := &orderpb.Order{OrderId: ord.OrderID, BizNo: ord.BizNo, UserId: ord.UserID, Status: ord.Status, TotalAmount: ord.TotalAmount}
	for _, it := range ord.Items {
		resp.Items = append(resp.Items, &orderpb.Item{SkuId: it.SkuID, Quantity: int32(it.Quantity), Price: it.Price})
	}
	return resp, nil
}

func (s *GRPCServer) UpdateOrderStatus(ctx context.Context, req *orderpb.UpdateOrderStatusRequest) (*orderpb.SimpleResponse, error) {
	err := s.svc.UpdateStatus(req.OrderId, req.FromStatus, req.ToStatus)
	if err != nil {
		return nil, err
	}
	return &orderpb.SimpleResponse{Success: true}, nil
}

func (s *GRPCServer) CancelOrder(ctx context.Context, req *orderpb.CancelOrderRequest) (*orderpb.SimpleResponse, error) {
	if err := s.svc.Cancel(req.OrderId); err != nil {
		return nil, err
	}
	return &orderpb.SimpleResponse{Success: true}, nil
}

func (s *GRPCServer) ListOrders(ctx context.Context, req *orderpb.ListOrdersRequest) (*orderpb.ListOrdersResponse, error) {
	resp, err := s.svc.ListOrders(ListOrdersRequest{
		Page:      req.Page,
		PageSize:  req.PageSize,
		OrderNo:   req.OrderNo,
		UserID:    req.UserId,
		Status:    req.Status,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		SortBy:    req.SortBy,
		SortOrder: req.SortOrder,
	})
	if err != nil {
		return nil, err
	}
	result := &orderpb.ListOrdersResponse{
		Orders:   make([]*orderpb.OrderListItem, 0, len(resp.Orders)),
		Total:    resp.Total,
		Page:     resp.Page,
		PageSize: resp.PageSize,
	}
	for _, o := range resp.Orders {
		result.Orders = append(result.Orders, &orderpb.OrderListItem{
			OrderId:       o.OrderID,
			BizNo:         o.BizNo,
			UserId:        o.UserID,
			Status:        o.Status,
			TotalAmount:   o.TotalAmount,
			ItemCount:     int32(o.ItemCount),
			PaymentStatus: o.PaymentStatus,
			CreatedAt:     o.CreatedAt,
			UpdatedAt:     o.UpdatedAt,
		})
	}
	return result, nil
}
