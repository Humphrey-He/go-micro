package main

import (
	"github.com/gin-gonic/gin"
	"go-micro/internal/user"
	"go-micro/pkg/cache"
	"go-micro/pkg/config"
	"go-micro/pkg/db"
	"go-micro/pkg/logx"
	"go-micro/pkg/middleware"
	"go.uber.org/zap"
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

	svc := user.NewService(dbx, rdb)
	h := user.NewHandler(svc)
	h.Register(r)

	addr := config.GetEnv("USER_ADDR", ":8083")
	if err := r.Run(addr); err != nil {
		logger.Fatal("user-service start failed", zap.Error(err))
	}
}
