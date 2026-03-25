package user

import (
	"context"

	"go-micro/proto/userpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GRPCClient struct {
	client userpb.UserServiceClient
}

func NewGRPCClient(target string) (*GRPCClient, *grpc.ClientConn, error) {
	conn, err := grpc.Dial(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}
	return &GRPCClient{client: userpb.NewUserServiceClient(conn)}, conn, nil
}

func (c *GRPCClient) GetUser(ctx context.Context, userID string) (*userpb.User, error) {
	return c.client.GetUser(ctx, &userpb.GetUserRequest{UserId: userID})
}

func (c *GRPCClient) GetUserByName(ctx context.Context, username string) (*userpb.User, error) {
	return c.client.GetUserByName(ctx, &userpb.GetUserByNameRequest{Username: username})
}

func (c *GRPCClient) Authenticate(ctx context.Context, username, password string) (*userpb.User, error) {
	return c.client.Authenticate(ctx, &userpb.AuthRequest{Username: username, Password: password})
}
