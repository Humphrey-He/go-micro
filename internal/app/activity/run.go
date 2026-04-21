package activityapp

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go-micro/internal/activity"
	"go-micro/pkg/cache"
	"go-micro/pkg/config"
	"go-micro/pkg/db"
	"go-micro/pkg/grpcx"
	"go-micro/pkg/logx"
	"go-micro/pkg/middleware"
	"go.uber.org/zap"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
)

func Run() error {
	logger := logx.L()
	defer logx.Sync()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := config.ValidateService("activity-service"); err != nil {
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

	svc := activity.NewService(dbx, rdb)
	h := activity.NewHandler(svc)
	h.Register(r)

	grpcAddr := config.GetEnv("ACTIVITY_GRPC_ADDR", ":9087")
	grpcLis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		logger.Error("activity-grpc listen failed", zap.Error(err))
		return err
	}
	grpcServer := grpc.NewServer(grpc.ForceServerCodec(grpcx.JSONCodec{}))
	activity.RegisterActivityServiceServer(grpcServer, activity.NewGRPCServer(svc))
	go func() {
		if err := grpcServer.Serve(grpcLis); err != nil {
			logger.Error("activity-grpc serve failed", zap.Error(err))
		}
	}()

	addr := config.GetEnv("ACTIVITY_ADDR", ":8087")
	srv := &http.Server{Addr: addr, Handler: r}
	logger.Info("activity-service starting", zap.String("http_addr", addr), zap.String("grpc_addr", grpcAddr))
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("activity-service start failed", zap.Error(err))
		}
	}()

	<-ctx.Done()
	grpcServer.GracefulStop()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	logger.Info("activity-service shutting down")
	return srv.Shutdown(shutdownCtx)
}
