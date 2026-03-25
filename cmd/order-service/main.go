package main

import (
	"context"
	"net"

	"github.com/gin-gonic/gin"
	"go-micro/internal/inventory"
	"go-micro/internal/order"
	"go-micro/pkg/cache"
	"go-micro/pkg/config"
	"go-micro/pkg/db"
	"go-micro/pkg/logx"
	"go-micro/pkg/middleware"
	"go-micro/pkg/mq"
	"go-micro/proto/orderpb"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	logger := logx.L()
	defer logx.Sync()

	dbx, err := db.NewMySQL()
	if err != nil {
		logger.Fatal("mysql connect failed", zap.Error(err))
	}
	rdb := cache.NewRedis()

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger(logger))

	invTarget := config.GetEnv("INVENTORY_GRPC_TARGET", "localhost:9082")
	invClient, invConn, err := inventory.NewGRPCClient(invTarget)
	if err != nil {
		logger.Fatal("inventory grpc dial failed", zap.Error(err))
	}
	defer invConn.Close()

	mqURL := config.GetEnv("MQ_URL", "amqp://guest:guest@localhost:5672/")
	exchange := config.GetEnv("MQ_EXCHANGE", "order.events")
	queue := config.GetEnv("MQ_QUEUE", "order.created")
	routeKey := config.GetEnv("MQ_ROUTING_KEY", "order.created")
	dlx := config.GetEnv("MQ_DLX", "order.events.dlx")
	dlq := config.GetEnv("MQ_DLQ", "order.created.dlq")
	publisher, err := mq.NewRabbit(mqURL, exchange, queue, routeKey, dlx, dlq)
	if err != nil {
		logger.Fatal("mq connect failed", zap.Error(err))
	}
	defer publisher.Close()

	svc := order.NewService(dbx, rdb, &inventoryAdapter{c: invClient}, publisher)
	h := order.NewHandler(svc)
	h.Register(r)

	stop := make(chan struct{})
	go svc.StartOutboxPublisher(stop)

	go func() {
		grpcAddr := config.GetEnv("ORDER_GRPC_ADDR", ":9081")
		lis, err := net.Listen("tcp", grpcAddr)
		if err != nil {
			logger.Fatal("order-grpc listen failed", zap.Error(err))
		}
		grpcServer := grpc.NewServer()
		orderpb.RegisterOrderServiceServer(grpcServer, order.NewGRPCServer(svc))
		if err := grpcServer.Serve(lis); err != nil {
			logger.Fatal("order-grpc serve failed", zap.Error(err))
		}
	}()

	addr := config.GetEnv("ORDER_ADDR", ":8081")
	if err := r.Run(addr); err != nil {
		logger.Fatal("order-service start failed", zap.Error(err))
	}
}

type inventoryAdapter struct {
	c *inventory.GRPCClient
}

func (a *inventoryAdapter) Reserve(ctx context.Context, orderID string, items []order.Item) (string, error) {
	invItems := make([]inventory.Item, 0, len(items))
	for _, it := range items {
		invItems = append(invItems, inventory.Item{SkuID: it.SkuID, Quantity: it.Quantity})
	}
	return a.c.Reserve(ctx, orderID, invItems)
}
