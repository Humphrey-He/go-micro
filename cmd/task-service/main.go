package main

import (
	"context"

	"github.com/gin-gonic/gin"
	"go-micro/internal/inventory"
	"go-micro/internal/task"
	"go-micro/pkg/config"
	"go-micro/pkg/db"
	"go-micro/pkg/logx"
	"go-micro/pkg/middleware"
	"go-micro/pkg/mq"
	"go.uber.org/zap"
)

func main() {
	logger := logx.L()
	defer logx.Sync()

	dbx, err := db.NewMySQL()
	if err != nil {
		logger.Fatal("mysql connect failed", zap.Error(err))
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger(logger))

	svc := task.NewService(dbx)
	h := task.NewHandler(svc)
	h.Register(r)

	mqURL := config.GetEnv("MQ_URL", "amqp://guest:guest@localhost:5672/")
	exchange := config.GetEnv("MQ_EXCHANGE", "order.events")
	queue := config.GetEnv("MQ_QUEUE", "order.created")
	routeKey := config.GetEnv("MQ_ROUTING_KEY", "order.created")
	dlx := config.GetEnv("MQ_DLX", "order.events.dlx")
	dlq := config.GetEnv("MQ_DLQ", "order.created.dlq")
	consumer, err := mq.NewRabbit(mqURL, exchange, queue, routeKey, dlx, dlq)
	if err != nil {
		logger.Fatal("mq connect failed", zap.Error(err))
	}
	defer consumer.Close()

	invTarget := config.GetEnv("INVENTORY_GRPC_TARGET", "localhost:9082")
	invClient, invConn, err := inventory.NewGRPCClient(invTarget)
	if err != nil {
		logger.Fatal("inventory grpc dial failed", zap.Error(err))
	}
	defer invConn.Close()

	go task.StartRabbitConsumer(consumer, svc, &inventoryReleaseAdapter{c: invClient})

	addr := config.GetEnv("TASK_ADDR", ":8084")
	if err := r.Run(addr); err != nil {
		logger.Fatal("task-service start failed", zap.Error(err))
	}
}

type inventoryReleaseAdapter struct {
	c *inventory.GRPCClient
}

func (a *inventoryReleaseAdapter) Release(ctx context.Context, reservedID string) error {
	return a.c.Release(ctx, reservedID)
}
