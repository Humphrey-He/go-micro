package inventory

import (
	"context"

	"go-micro/proto/inventorypb"
)

type GRPCServer struct {
	inventorypb.UnimplementedInventoryServiceServer
	svc *Service
}

func NewGRPCServer(svc *Service) *GRPCServer {
	return &GRPCServer{svc: svc}
}

func (s *GRPCServer) Reserve(ctx context.Context, req *inventorypb.ReserveRequest) (*inventorypb.ReserveResponse, error) {
	items := make([]Item, 0, len(req.Items))
	for _, it := range req.Items {
		items = append(items, Item{SkuID: it.SkuId, Quantity: int(it.Quantity)})
	}
	resp, err := s.svc.Reserve(ReserveRequest{OrderID: req.OrderId, Items: items})
	if err != nil {
		return nil, err
	}
	return &inventorypb.ReserveResponse{ReservedId: resp.ReservedID}, nil
}

func (s *GRPCServer) Release(ctx context.Context, req *inventorypb.ReleaseRequest) (*inventorypb.SimpleResponse, error) {
	err := s.svc.Release(req.ReservedId)
	if err != nil {
		return nil, err
	}
	return &inventorypb.SimpleResponse{Success: true}, nil
}

func (s *GRPCServer) Confirm(ctx context.Context, req *inventorypb.ConfirmRequest) (*inventorypb.SimpleResponse, error) {
	err := s.svc.Confirm(req.ReservedId)
	if err != nil {
		return nil, err
	}
	return &inventorypb.SimpleResponse{Success: true}, nil
}

func (s *GRPCServer) GetReservation(ctx context.Context, req *inventorypb.GetReservationRequest) (*inventorypb.Reservation, error) {
	resv, err := s.svc.GetReservation(req.OrderId)
	if err != nil {
		return nil, err
	}
	return &inventorypb.Reservation{ReservedId: resv.ReservedID, OrderId: resv.OrderID, Status: resv.Status}, nil
}
