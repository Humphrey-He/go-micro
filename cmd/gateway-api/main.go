package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	_ "go-micro/docs/swagger"
	"go-micro/internal/gateway"
	"go-micro/internal/inventory"
	"go-micro/internal/order"
	"go-micro/internal/task"
	"go-micro/internal/user"
	"go-micro/pkg/config"
	"go-micro/pkg/logx"
	"go-micro/pkg/middleware"
	"go.uber.org/zap"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Go-Micro Gateway API
// @version 1.0
// @description ???? API ??
// @host localhost:8080
// @BasePath /
func main() {
	logger := logx.L()
	defer logx.Sync()

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger(logger))
	r.Use(middleware.RateLimit())

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	orderTarget := config.GetEnv("ORDER_GRPC_TARGET", "localhost:9081")
	orderClient, orderConn, err := order.NewGRPCClient(orderTarget)
	if err != nil {
		logger.Fatal("order grpc dial failed", zap.Error(err))
	}
	defer orderConn.Close()

	userTarget := config.GetEnv("USER_GRPC_TARGET", "localhost:9083")
	userClient, userConn, err := user.NewGRPCClient(userTarget)
	if err != nil {
		logger.Fatal("user grpc dial failed", zap.Error(err))
	}
	defer userConn.Close()

	invTarget := config.GetEnv("INVENTORY_GRPC_TARGET", "localhost:9082")
	invClient, invConn, err := inventory.NewGRPCClient(invTarget)
	if err != nil {
		logger.Fatal("inventory grpc dial failed", zap.Error(err))
	}
	defer invConn.Close()

	taskTarget := config.GetEnv("TASK_GRPC_TARGET", "localhost:9084")
	taskClient, taskConn, err := task.NewGRPCClient(taskTarget)
	if err != nil {
		logger.Fatal("task grpc dial failed", zap.Error(err))
	}
	defer taskConn.Close()

	svc := gateway.NewService(orderClient, userClient, invClient, taskClient)
	h := gateway.NewHandler(svc)
	h.Register(r)

	addr := config.GetEnv("GATEWAY_ADDR", ":8080")
	srv := &http.Server{Addr: addr, Handler: r}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("gateway-api start failed", zap.Error(err))
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}
