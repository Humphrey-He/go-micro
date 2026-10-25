package main

import (
	"os"

	distributionapp "ecomm-micro/internal/app/distribution"
	"ecomm-micro/pkg/logx"
	"go.uber.org/zap"
)

// @title Go-Micro Distribution API
// @version 1.0
// @description Go-Micro 多平台分发调度服务 API
// @host localhost:8090
// @BasePath /
func main() {
	logger := logx.L()
	defer logx.Sync()
	logger.Info("service starting", zap.String("service", "distribution-api"))
	if err := distributionapp.Run(); err != nil {
		logger.Error("service exited with error", zap.String("service", "distribution-api"), zap.Error(err))
		os.Exit(1)
	}
	logger.Info("service stopped", zap.String("service", "distribution-api"))
}
