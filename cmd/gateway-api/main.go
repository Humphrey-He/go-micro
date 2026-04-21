package main

import (
	"os"

	gatewayapp "go-micro/internal/app/gateway"
	"go-micro/pkg/logx"
	"go.uber.org/zap"
)

// @title Go-Micro Gateway API
// @version 1.0
// @description Go-Micro 网关服务 API
// @host localhost:8080
// @BasePath /
func main() {
	os.Exit(run("gateway-api", gatewayapp.Run))
}

func run(name string, exec func() error) int {
	logger := logx.L()
	defer logx.Sync()
	logger.Info("service starting", zap.String("service", name))
	if err := exec(); err != nil {
		logger.Error("service exited with error", zap.String("service", name), zap.Error(err))
		return 1
	}
	logger.Info("service stopped", zap.String("service", name))
	return 0
}
