package task

import (
	"context"

	"go-micro/proto/taskpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GRPCClient struct {
	client taskpb.TaskServiceClient
}

func NewGRPCClient(target string) (*GRPCClient, *grpc.ClientConn, error) {
	conn, err := grpc.Dial(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}
	return &GRPCClient{client: taskpb.NewTaskServiceClient(conn)}, conn, nil
}

func (c *GRPCClient) GetByOrder(ctx context.Context, orderID string) (*taskpb.Task, error) {
	return c.client.GetTaskByOrder(ctx, &taskpb.GetTaskByOrderRequest{OrderId: orderID})
}
