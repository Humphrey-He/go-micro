package user

import (
	"context"

	"go-micro/proto/userpb"
)

type GRPCServer struct {
	userpb.UnimplementedUserServiceServer
	svc *Service
}

func NewGRPCServer(svc *Service) *GRPCServer {
	return &GRPCServer{svc: svc}
}

func (s *GRPCServer) GetUser(ctx context.Context, req *userpb.GetUserRequest) (*userpb.User, error) {
	u, err := s.svc.Get(req.UserId)
	if err != nil {
		return nil, err
	}
	return &userpb.User{UserId: u.UserID, Username: u.Username, Role: u.Role, Status: int32(u.Status)}, nil
}

func (s *GRPCServer) GetUserByName(ctx context.Context, req *userpb.GetUserByNameRequest) (*userpb.User, error) {
	u, err := s.svc.GetByUsername(req.Username)
	if err != nil {
		return nil, err
	}
	return &userpb.User{UserId: u.UserID, Username: u.Username, Role: u.Role, Status: int32(u.Status)}, nil
}
