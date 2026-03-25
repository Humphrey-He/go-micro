package main

import (
	"github.com/gin-gonic/gin"
	"go-micro/internal/gateway"
	"go-micro/internal/order"
	"go-micro/pkg/config"
	"go-micro/pkg/logx"
	"go-micro/pkg/middleware"
	"go.uber.org/zap"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "go-micro/docs/swagger"
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

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	orderTarget := config.GetEnv("ORDER_GRPC_TARGET", "localhost:9081")
	orderClient, orderConn, err := order.NewGRPCClient(orderTarget)
	if err != nil {
		logger.Fatal("order grpc dial failed", zap.Error(err))
	}
	defer orderConn.Close()

	svc := gateway.NewService(orderClient)
	h := gateway.NewHandler(svc)
	h.Register(r)

	addr := config.GetEnv("GATEWAY_ADDR", ":8080")
	if err := r.Run(addr); err != nil {
		logger.Fatal("gateway-api start failed", zap.Error(err))
	}
}
