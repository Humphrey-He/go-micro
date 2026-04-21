package paymentapp

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
	"go-micro/internal/payment"
	"go-micro/pkg/config"
	"go-micro/pkg/db"
	"go-micro/pkg/logx"
	"go-micro/pkg/middleware"
	"go-micro/proto/paymentpb"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func Run() error {
	logger := logx.L()
	defer logx.Sync()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := config.ValidateService("payment-service"); err != nil {
		logger.Error("config validation failed", zap.Error(err))
		return err
	}

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

	svc := payment.NewService(
		dbx,
		&orderCancelAdapter{c: orderClient},
		&inventoryReleaseAdapter{c: invClient},
	)

	h := payment.NewHandler(svc)
	h.Register(r)

	grpcAddr := config.GetEnv("PAYMENT_GRPC_ADDR", ":9085")
	grpcLis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		logger.Error("payment-grpc listen failed", zap.Error(err))
		return err
	}

	grpcServer := grpc.NewServer()
	paymentpb.RegisterPaymentServiceServer(grpcServer, payment.NewGRPCServer(svc))

	go func() {
		if err := grpcServer.Serve(grpcLis); err != nil {
			logger.Error("payment-grpc serve failed", zap.Error(err))
		}
	}()

	stopTimeout := make(chan struct{})
	go payment.StartTimeoutWorker(svc, stopTimeout)

	addr := config.GetEnv("PAYMENT_ADDR", ":8085")
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	logger.Info("payment-service starting",
		zap.String("http_addr", addr),
		zap.String("grpc_addr", grpcAddr),
	)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("payment-service start failed", zap.Error(err))
		}
	}()

	<-ctx.Done()

	close(stopTimeout)
	grpcServer.GracefulStop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	logger.Info("payment-service shutting down")
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
