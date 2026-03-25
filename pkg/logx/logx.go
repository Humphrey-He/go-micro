package logx

import (
	"sync"

	"go.uber.org/zap"
)

var (
	logger *zap.Logger
	once   sync.Once
)

func L() *zap.Logger {
	once.Do(func() {
		logger, _ = zap.NewProduction()
	})
	return logger
}

func Sync() {
	if logger != nil {
		_ = logger.Sync()
	}
}
