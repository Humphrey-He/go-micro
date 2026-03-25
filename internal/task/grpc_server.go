package task

import (
	"context"

	"go-micro/proto/taskpb"
)

type GRPCServer struct {
	taskpb.UnimplementedTaskServiceServer
	svc *Service
}

func NewGRPCServer(svc *Service) *GRPCServer {
	return &GRPCServer{svc: svc}
}

func (s *GRPCServer) GetTaskByOrder(ctx context.Context, req *taskpb.GetTaskByOrderRequest) (*taskpb.Task, error) {
	t, err := s.svc.GetByOrder(req.OrderId)
	if err != nil {
		return nil, err
	}
	return &taskpb.Task{
		TaskId:      t.TaskID,
		BizNo:       t.BizNo,
		OrderId:     t.OrderID,
		Type:        t.Type,
		Status:      t.Status,
		RetryCount:  int32(t.RetryCount),
		NextRetryAt: t.NextRetryAt.Unix(),
	}, nil
}
