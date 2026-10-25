package inventoryapp

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ecomm-micro/internal/inventory"
	"ecomm-micro/pkg/cache"
	"ecomm-micro/pkg/config"
	"ecomm-micro/pkg/db"
	"ecomm-micro/pkg/logx"
	"ecomm-micro/pkg/middleware"
	"ecomm-micro/proto/inventorypb"
	"github.com/gin-gonic/gin"
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
	pre := inventory.NewPreStock(dbx, rdb)
	syncer := inventory.NewStockSyncer(dbx, rdb)
	reconciler := inventory.NewReconciler(dbx, rdb)
	h := inventory.NewHandler(svc, pre, syncer, reconciler)
	h.Register(r)

	// R1: 预扣流水异步落库 + 周期对账兜底
	go inventory.StartStockSyncer(dbx, rdb, time.Second, ctx.Done())
	go inventory.StartReconciler(dbx, rdb, time.Minute, ctx.Done())

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
