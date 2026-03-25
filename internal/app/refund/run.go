package refundapp

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go-micro/internal/inventory"
	"go-micro/internal/order"
	"go-micro/internal/refund"
	"go-micro/pkg/config"
	"go-micro/pkg/db"
	"go-micro/pkg/logx"
	"go-micro/pkg/middleware"
	"go-micro/pkg/mq"
	"go.uber.org/zap"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func Run() error {
	logger := logx.L()
	defer logx.Sync()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	dbx, err := db.NewMySQL()
	if err != nil {
		logger.Error("mysql connect failed", zap.Error(err))
		return err
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger(logger))
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/readyz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	mqURL := config.GetEnv("MQ_URL", "amqp://guest:guest@localhost:5672/")
	exchange := config.GetEnv("REFUND_MQ_EXCHANGE", "refund.events")
	queue := config.GetEnv("REFUND_MQ_QUEUE", "refund.process")
	routeKey := config.GetEnv("REFUND_MQ_ROUTING_KEY", "refund.initiated")
	dlx := config.GetEnv("REFUND_MQ_DLX", "refund.events.dlx")
	dlq := config.GetEnv("REFUND_MQ_DLQ", "refund.process.dlq")
	consumer, err := mq.NewRabbit(mqURL, exchange, queue, routeKey, dlx, dlq)
	if err != nil {
		logger.Error("mq connect failed", zap.Error(err))
		return err
	}
	defer consumer.Close()

	orderTarget := config.GetEnv("ORDER_GRPC_TARGET", "localhost:9081")
	orderClient, orderConn, err := order.NewGRPCClient(orderTarget)
	if err != nil {
		logger.Error("order grpc dial failed", zap.Error(err))
		return err
	}
	defer orderConn.Close()

	invTarget := config.GetEnv("INVENTORY_GRPC_TARGET", "localhost:9082")
	invClient, invConn, err := inventory.NewGRPCClient(invTarget)
	if err != nil {
		logger.Error("inventory grpc dial failed", zap.Error(err))
		return err
	}
	defer invConn.Close()

	svc := refund.NewService(dbx, consumer, &orderCancelAdapter{c: orderClient}, &inventoryReleaseAdapter{c: invClient})
	h := refund.NewHandler(svc)
	h.Register(r)

	workerStop := make(chan struct{})
	go refund.StartConsumer(consumer, svc, workerStop)
	go refund.StartRetryWorker(svc, consumer, workerStop)

	addr := config.GetEnv("REFUND_ADDR", ":8086")
	srv := &http.Server{Addr: addr, Handler: r}
	logger.Info("refund-service starting", zap.String("http_addr", addr))
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("refund-service start failed", zap.Error(err))
		}
	}()

	<-ctx.Done()
	close(workerStop)
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	logger.Info("refund-service shutting down")
	return srv.Shutdown(shutdownCtx)
}

type orderCancelAdapter struct {
	c *order.GRPCClient
}

func (a *orderCancelAdapter) Cancel(ctx context.Context, orderID string) error {
	return a.c.Cancel(ctx, orderID)
}

type inventoryReleaseAdapter struct {
	c *inventory.GRPCClient
}

func (a *inventoryReleaseAdapter) ReleaseByOrder(ctx context.Context, orderID string) error {
	return a.c.ReleaseByOrder(ctx, orderID)
}
