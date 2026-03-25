package main

import (
	"os"

	taskapp "go-micro/internal/app/task"
	"go-micro/pkg/logx"
	"go.uber.org/zap"
)

func main() {
	os.Exit(run("task-service", taskapp.Run))
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
