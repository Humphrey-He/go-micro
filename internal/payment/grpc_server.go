package payment

import (
	"context"

	"go-micro/proto/paymentpb"
)

type GRPCServer struct {
	paymentpb.UnimplementedPaymentServiceServer
	svc *Service
}

func NewGRPCServer(svc *Service) *GRPCServer {
	return &GRPCServer{svc: svc}
}

func (s *GRPCServer) CreatePayment(ctx context.Context, req *paymentpb.CreatePaymentRequest) (*paymentpb.CreatePaymentResponse, error) {
	p, err := s.svc.Create(CreatePaymentRequest{
		OrderID:   req.OrderId,
		Amount:    req.Amount,
		RequestID: req.RequestId,
	})
	if err == ErrIdempotentHit {
		return &paymentpb.CreatePaymentResponse{
			Payment:       toPaymentPB(p),
			IdempotentHit: true,
		}, nil
	}
	if err != nil {
		return nil, err
	}
	return &paymentpb.CreatePaymentResponse{
		Payment:       toPaymentPB(p),
		IdempotentHit: false,
	}, nil
}

func (s *GRPCServer) GetPayment(ctx context.Context, req *paymentpb.GetPaymentRequest) (*paymentpb.Payment, error) {
	p, err := s.svc.Get(req.PaymentId)
	if err != nil {
		return nil, err
	}
	return toPaymentPB(p), nil
}

func (s *GRPCServer) MarkSuccess(ctx context.Context, req *paymentpb.ChangePaymentStatusRequest) (*paymentpb.SimpleResponse, error) {
	if err := s.svc.MarkSuccess(req.PaymentId); err != nil {
		return nil, err
	}
	return &paymentpb.SimpleResponse{Success: true}, nil
}

func (s *GRPCServer) MarkFailed(ctx context.Context, req *paymentpb.ChangePaymentStatusRequest) (*paymentpb.SimpleResponse, error) {
	if err := s.svc.MarkFailed(req.PaymentId); err != nil {
		return nil, err
	}
	return &paymentpb.SimpleResponse{Success: true}, nil
}

func (s *GRPCServer) MarkTimeout(ctx context.Context, req *paymentpb.ChangePaymentStatusRequest) (*paymentpb.SimpleResponse, error) {
	if err := s.svc.MarkTimeout(req.PaymentId); err != nil {
		return nil, err
	}
	return &paymentpb.SimpleResponse{Success: true}, nil
}

func (s *GRPCServer) Refund(ctx context.Context, req *paymentpb.ChangePaymentStatusRequest) (*paymentpb.SimpleResponse, error) {
	if err := s.svc.Refund(req.PaymentId); err != nil {
		return nil, err
	}
	return &paymentpb.SimpleResponse{Success: true}, nil
}

func toPaymentPB(p *Payment) *paymentpb.Payment {
	if p == nil {
		return nil
	}
	return &paymentpb.Payment{
		PaymentId: p.PaymentID,
		OrderId:   p.OrderID,
		Amount:    p.Amount,
		Status:    p.Status,
		RequestId: p.RequestID,
		CreatedAt: p.CreatedAt.Unix(),
		UpdatedAt: p.UpdatedAt.Unix(),
	}
}
