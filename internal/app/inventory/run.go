package inventoryapp

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
	"go-micro/pkg/cache"
	"go-micro/pkg/config"
	"go-micro/pkg/db"
	"go-micro/pkg/logx"
	"go-micro/pkg/middleware"
	"go-micro/proto/inventorypb"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func Run() error {
	logger := logx.L()
	defer logx.Sync()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := config.ValidateService("inventory-service"); err != nil {
		logger.Error("config validation failed", zap.Error(err))
		return err
	}

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
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/readyz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	svc := inventory.NewService(dbx, rdb)
	h := inventory.NewHandler(svc)
	h.Register(r)

	grpcAddr := config.GetEnv("INVENTORY_GRPC_ADDR", ":9082")
	grpcLis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		logger.Error("inventory-grpc listen failed", zap.Error(err))
		return err
	}
	grpcServer := grpc.NewServer()
	inventorypb.RegisterInventoryServiceServer(grpcServer, inventory.NewGRPCServer(svc))
	go func() {
		if err := grpcServer.Serve(grpcLis); err != nil {
			logger.Error("inventory-grpc serve failed", zap.Error(err))
		}
	}()

	addr := config.GetEnv("INVENTORY_ADDR", ":8082")
	srv := &http.Server{Addr: addr, Handler: r}
	logger.Info("inventory-service starting", zap.String("http_addr", addr), zap.String("grpc_addr", grpcAddr))
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("inventory-service start failed", zap.Error(err))
		}
	}()

	<-ctx.Done()

	grpcServer.GracefulStop()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	logger.Info("inventory-service shutting down")
	return srv.Shutdown(shutdownCtx)
}
