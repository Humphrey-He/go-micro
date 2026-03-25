package main

import (
	"net"

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
)

func main() {
	logger := logx.L()
	defer logx.Sync()

	dbx, err := db.NewMySQL()
	if err != nil {
		logger.Fatal("mysql connect failed", zap.Error(err))
	}
	rdb := cache.NewRedis()

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger(logger))

	svc := inventory.NewService(dbx, rdb)
	h := inventory.NewHandler(svc)
	h.Register(r)

	go func() {
		grpcAddr := config.GetEnv("INVENTORY_GRPC_ADDR", ":9082")
		lis, err := net.Listen("tcp", grpcAddr)
		if err != nil {
			logger.Fatal("inventory-grpc listen failed", zap.Error(err))
		}
		grpcServer := grpc.NewServer()
		inventorypb.RegisterInventoryServiceServer(grpcServer, inventory.NewGRPCServer(svc))
		if err := grpcServer.Serve(lis); err != nil {
			logger.Fatal("inventory-grpc serve failed", zap.Error(err))
		}
	}()

	addr := config.GetEnv("INVENTORY_ADDR", ":8082")
	if err := r.Run(addr); err != nil {
		logger.Fatal("inventory-service start failed", zap.Error(err))
	}
}
