package main

import (
	"os"

	reconcileapp "go-micro/internal/app/reconcile"
	"go-micro/pkg/logx"
	"go.uber.org/zap"
)

// @title Go-Micro Reconcile API
// @version 1.0
// @description Go-Micro 订单/库存对账服务 API
// @host localhost:8089
// @BasePath /
func main() {
	os.Exit(run())
}

func run() int {
	logger := logx.L()
	defer logx.Sync()
	logger.Info("service starting", zap.String("service", "reconcile-api"))
	if err := reconcileapp.Run(); err != nil {
		logger.Error("service exited with error", zap.String("service", "reconcile-api"), zap.Error(err))
		return 1
	}
	logger.Info("service stopped", zap.String("service", "reconcile-api"))
	return 0
}
