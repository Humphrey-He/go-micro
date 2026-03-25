package userapp

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go-micro/internal/user"
	"go-micro/pkg/cache"
	"go-micro/pkg/config"
	"go-micro/pkg/db"
	"go-micro/pkg/logx"
	"go-micro/pkg/middleware"
	"go-micro/proto/userpb"
	"go.uber.org/zap"
	"google.golang.org/grpc"

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

	svc := user.NewService(dbx, rdb)
	h := user.NewHandler(svc)
	h.Register(r)

	grpcAddr := config.GetEnv("USER_GRPC_ADDR", ":9083")
	grpcLis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		logger.Error("user-grpc listen failed", zap.Error(err))
		return err
	}
	grpcServer := grpc.NewServer()
	userpb.RegisterUserServiceServer(grpcServer, user.NewGRPCServer(svc))
	go func() {
		if err := grpcServer.Serve(grpcLis); err != nil {
			logger.Error("user-grpc serve failed", zap.Error(err))
		}
	}()

	addr := config.GetEnv("USER_ADDR", ":8083")
	srv := &http.Server{Addr: addr, Handler: r}
	logger.Info("user-service starting", zap.String("http_addr", addr), zap.String("grpc_addr", grpcAddr))
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("user-service start failed", zap.Error(err))
		}
	}()

	<-ctx.Done()

	grpcServer.GracefulStop()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	logger.Info("user-service shutting down")
	return srv.Shutdown(shutdownCtx)
}
