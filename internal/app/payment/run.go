package paymentapp

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
	"go-micro/internal/payment"
	"go-micro/pkg/config"
	"go-micro/pkg/db"
	"go-micro/pkg/logx"
	"go-micro/pkg/middleware"
	"go.uber.org/zap"
)

func Run() error {
	logger := logx.L()
	defer logx.Sync()

	dbx, err := db.NewMySQL()
	if err != nil {
		logger.Error("mysql connect failed", zap.Error(err))
		return err
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger(logger))

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

	svc := payment.NewService(dbx, &orderCancelAdapter{c: orderClient}, &inventoryReleaseAdapter{c: invClient})
	h := payment.NewHandler(svc)
	h.Register(r)

	stopTimeout := make(chan struct{})
	go payment.StartTimeoutWorker(svc, stopTimeout)

	addr := config.GetEnv("PAYMENT_ADDR", ":8085")
	srv := &http.Server{Addr: addr, Handler: r}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("payment-service start failed", zap.Error(err))
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	close(stopTimeout)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return srv.Shutdown(ctx)
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
