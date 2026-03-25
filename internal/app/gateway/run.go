package gatewayapp

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	_ "go-micro/docs/swagger"
	"go-micro/internal/activity"
	"go-micro/internal/gateway"
	"go-micro/internal/inventory"
	"go-micro/internal/order"
	"go-micro/internal/payment"
	"go-micro/internal/price"
	"go-micro/internal/refund"
	"go-micro/internal/task"
	"go-micro/internal/user"
	"go-micro/pkg/config"
	"go-micro/pkg/db"
	"go-micro/pkg/logx"
	"go-micro/pkg/middleware"
	"go.uber.org/zap"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Run() error {
	logger := logx.L()
	defer logx.Sync()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger(logger))
	r.Use(middleware.RateLimit())
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/readyz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	orderTarget := config.GetEnv("ORDER_GRPC_TARGET", "localhost:9081")
	orderClient, orderConn, err := order.NewGRPCClient(orderTarget)
	if err != nil {
		logger.Error("order grpc dial failed", zap.Error(err))
		return err
	}
	defer orderConn.Close()

	userTarget := config.GetEnv("USER_GRPC_TARGET", "localhost:9083")
	userClient, userConn, err := user.NewGRPCClient(userTarget)
	if err != nil {
		logger.Error("user grpc dial failed", zap.Error(err))
		return err
	}
	defer userConn.Close()

	invTarget := config.GetEnv("INVENTORY_GRPC_TARGET", "localhost:9082")
	invClient, invConn, err := inventory.NewGRPCClient(invTarget)
	if err != nil {
		logger.Error("inventory grpc dial failed", zap.Error(err))
		return err
	}
	defer invConn.Close()

	taskTarget := config.GetEnv("TASK_GRPC_TARGET", "localhost:9084")
	taskClient, taskConn, err := task.NewGRPCClient(taskTarget)
	if err != nil {
		logger.Error("task grpc dial failed", zap.Error(err))
		return err
	}
	defer taskConn.Close()

	refundTarget := config.GetEnv("REFUND_GRPC_TARGET", "localhost:9086")
	refundClient, refundConn, err := refund.NewGRPCClient(refundTarget)
	if err != nil {
		logger.Error("refund grpc dial failed", zap.Error(err))
		return err
	}
	defer refundConn.Close()

	activityTarget := config.GetEnv("ACTIVITY_GRPC_TARGET", "localhost:9087")
	activityClient, activityConn, err := activity.NewGRPCClient(activityTarget)
	if err != nil {
		logger.Error("activity grpc dial failed", zap.Error(err))
		return err
	}
	defer activityConn.Close()

	priceTarget := config.GetEnv("PRICE_GRPC_TARGET", "localhost:9088")
	priceClient, priceConn, err := price.NewGRPCClient(priceTarget)
	if err != nil {
		logger.Error("price grpc dial failed", zap.Error(err))
		return err
	}
	defer priceConn.Close()

	dbx, err := db.NewMySQL()
	if err != nil {
		logger.Error("mysql connect failed", zap.Error(err))
		return err
	}
	paySvc := payment.NewService(dbx, &orderCancelAdapter{c: orderClient}, &inventoryReleaseAdapter{c: invClient})
	svc := gateway.NewService(orderClient, userClient, invClient, taskClient, refundClient, activityClient, priceClient)
	svc.SetPayment(paySvc)
	h := gateway.NewHandler(svc)
	h.Register(r)

	addr := config.GetEnv("GATEWAY_ADDR", ":8080")
	srv := &http.Server{Addr: addr, Handler: r}
	logger.Info("gateway-api starting", zap.String("http_addr", addr))

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("gateway-api start failed", zap.Error(err))
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	logger.Info("gateway-api shutting down")
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
