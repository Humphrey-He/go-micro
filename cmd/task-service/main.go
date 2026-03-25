package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go-micro/internal/inventory"
	"go-micro/internal/order"
	"go-micro/internal/task"
	"go-micro/pkg/config"
	"go-micro/pkg/db"
	"go-micro/pkg/logx"
	"go-micro/pkg/middleware"
	"go-micro/pkg/mq"
	"go-micro/proto/taskpb"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/prometheus/client_golang/prometheus/promhttp"
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
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	svc := task.NewService(dbx)
	h := task.NewHandler(svc)
	h.Register(r)

	grpcAddr := config.GetEnv("TASK_GRPC_ADDR", ":9084")
	grpcLis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		logger.Fatal("task-grpc listen failed", zap.Error(err))
	}
	grpcServer := grpc.NewServer()
	taskpb.RegisterTaskServiceServer(grpcServer, task.NewGRPCServer(svc))
	go func() {
		if err := grpcServer.Serve(grpcLis); err != nil {
			logger.Fatal("task-grpc serve failed", zap.Error(err))
		}
	}()

	mqURL := config.GetEnv("MQ_URL", "amqp://guest:guest@localhost:5672/")
	exchange := config.GetEnv("MQ_EXCHANGE", "order.events")
	queue := config.GetEnv("MQ_QUEUE", "order_reserved")
	routeKey := config.GetEnv("MQ_ROUTING_KEY", "order_reserved")
	dlx := config.GetEnv("MQ_DLX", "order.events.dlx")
	dlq := config.GetEnv("MQ_DLQ", "order_reserved.dlq")
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

	orderTarget := config.GetEnv("ORDER_GRPC_TARGET", "localhost:9081")
	orderClient, orderConn, err := order.NewGRPCClient(orderTarget)
	if err != nil {
		logger.Fatal("order grpc dial failed", zap.Error(err))
	}
	defer orderConn.Close()

	go task.StartRabbitConsumer(consumer, svc, &inventoryReleaseAdapter{c: invClient}, &orderUpdateAdapter{c: orderClient})

	addr := config.GetEnv("TASK_ADDR", ":8084")
	srv := &http.Server{Addr: addr, Handler: r}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("task-service start failed", zap.Error(err))
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	grpcServer.GracefulStop()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}

type inventoryReleaseAdapter struct {
	c *inventory.GRPCClient
}

func (a *inventoryReleaseAdapter) Release(ctx context.Context, reservedID string) error {
	return a.c.Release(ctx, reservedID)
}

type orderUpdateAdapter struct {
	c *order.GRPCClient
}

func (a *orderUpdateAdapter) UpdateStatus(ctx context.Context, orderID, from, to string) error {
	return a.c.UpdateStatus(ctx, orderID, from, to)
}
