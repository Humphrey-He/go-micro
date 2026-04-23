package recommendationapp

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go-micro/internal/recommendation"
	"go-micro/pkg/config"
	"go-micro/pkg/db"
	"go-micro/pkg/logx"
	"go-micro/pkg/middleware"
	"go-micro/pkg/metrics"
	"go-micro/pkg/mq"
	"go-micro/pkg/tracing"
	"go.uber.org/zap"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func Run() error {
	logger := logx.L()
	defer logx.Sync()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger(logger))
	r.Use(metrics.HTTPMiddleware())
	r.Use(tracing.Middleware("recommendation-api"))

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/readyz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Database
	dbx, err := db.NewMySQL()
	if err != nil {
		logger.Error("mysql connect failed", zap.Error(err))
		return err
	}

	// Redis
	rdb := redis.NewClient(&redis.Options{
		Addr: config.GetEnv("REDIS_ADDR", "localhost:6379"),
		DB:   config.GetInt("REDIS_DB", 1),
	})

	// RabbitMQ (optional - service can run without it)
	rabbit, err := mq.NewRabbit(
		config.GetEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		"recommendation_exchange",
		"user_behavior_topic",
		"user.behavior.#",
		"recommendation_dlx",
		"user_behavior_dlq",
	)
	if err != nil {
		logger.Warn("rabbitmq connect failed, running without MQ", zap.Error(err))
		rabbit = nil
	}

	svc := recommendation.NewService(dbx, rdb)
	h := recommendation.NewHandler(svc)

	// Start MQ consumer if available
	if rabbit != nil {
		consumer := recommendation.NewConsumer(svc, rabbit)
		go consumer.Start(ctx)
	}

	h.Register(r)

	addr := config.GetEnv("RECOMMENDATION_ADDR", ":8085")
	srv := &http.Server{Addr: addr, Handler: r}
	logger.Info("recommendation-api starting", zap.String("addr", addr))

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("recommendation-api start failed", zap.Error(err))
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	logger.Info("recommendation-api shutting down")
	return srv.Shutdown(shutdownCtx)
}