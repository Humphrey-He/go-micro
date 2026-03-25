package orderapp

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
	"go-micro/pkg/cache"
	"go-micro/pkg/config"
	"go-micro/pkg/db"
	"go-micro/pkg/logx"
	"go-micro/pkg/middleware"
	"go-micro/pkg/mq"
	"go-micro/proto/orderpb"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func Run() error {
	logger := logx.L()
	defer logx.Sync()

	dbx, err := db.NewMySQL()
	if err != nil {
		logger.Error("mysql connect failed", zap.Error(err))
		return err
	}
	rdb := cache.NewRedis()

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger(logger))
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	invTarget := config.GetEnv("INVENTORY_GRPC_TARGET", "localhost:9082")
	invClient, invConn, err := inventory.NewGRPCClient(invTarget)
	if err != nil {
		logger.Error("inventory grpc dial failed", zap.Error(err))
		return err
	}
	defer invConn.Close()

	mqURL := config.GetEnv("MQ_URL", "amqp://guest:guest@localhost:5672/")
	exchange := config.GetEnv("MQ_EXCHANGE", "order.events")
	queue := config.GetEnv("MQ_QUEUE", "order_reserved")
	routeKey := config.GetEnv("MQ_ROUTING_KEY", "order_reserved")
	dlx := config.GetEnv("MQ_DLX", "order.events.dlx")
	dlq := config.GetEnv("MQ_DLQ", "order_reserved.dlq")
	producer, err := mq.NewRabbit(mqURL, exchange, queue, routeKey, dlx, dlq)
	if err != nil {
		logger.Error("mq connect failed", zap.Error(err))
		return err
	}
	defer producer.Close()

	svc := order.NewService(dbx, rdb, &inventoryAdapter{c: invClient}, producer)
	h := order.NewHandler(svc)
	h.Register(r)

	grpcAddr := config.GetEnv("ORDER_GRPC_ADDR", ":9081")
	grpcLis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		logger.Error("order-grpc listen failed", zap.Error(err))
		return err
	}
	grpcServer := grpc.NewServer()
	orderpb.RegisterOrderServiceServer(grpcServer, order.NewGRPCServer(svc))
	go func() {
		if err := grpcServer.Serve(grpcLis); err != nil {
			logger.Error("order-grpc serve failed", zap.Error(err))
		}
	}()

	stopOutbox := make(chan struct{})
	go svc.StartOutboxPublisher(stopOutbox)

	addr := config.GetEnv("ORDER_ADDR", ":8081")
	srv := &http.Server{Addr: addr, Handler: r}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("order-service start failed", zap.Error(err))
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	close(stopOutbox)
	grpcServer.GracefulStop()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return srv.Shutdown(ctx)
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
