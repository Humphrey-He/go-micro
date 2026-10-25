package reconcileapp

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go-micro/internal/reconcile"
	"go-micro/pkg/config"
	"go-micro/pkg/db"
	"go-micro/pkg/logx"
	"go-micro/pkg/middleware"
	"go.uber.org/zap"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Run reconcile-service：订单/库存对账 job + 差异台账管理端点。
func Run() error {
	logger := logx.L()
	defer logx.Sync()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := config.ValidateService("reconcile-service"); err != nil {
		logger.Error("config validation failed", zap.Error(err))
		return err
	}

	dbx, err := db.NewMySQL()
	if err != nil {
		logger.Error("mysql connect failed", zap.Error(err))
		return err
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger(logger))
	r.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	r.GET("/readyz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ready"}) })
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	svc := reconcile.NewService(dbx)
	reconcile.NewHandler(svc).Register(r)

	// 周期对账（每 5 分钟），手动触发走 POST /reconcile/run
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				jobCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
				if _, err := svc.RunOnce(jobCtx); err != nil {
					logx.L().Warn("reconcile round failed", zap.Error(err))
				}
				cancel()
			}
		}
	}()

	addr := config.GetEnv("RECONCILE_ADDR", ":8089")
	srv := &http.Server{Addr: addr, Handler: r}
	logger.Info("reconcile-service starting", zap.String("http_addr", addr))
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("reconcile-service start failed", zap.Error(err))
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	logger.Info("reconcile-service shutting down")
	return srv.Shutdown(shutdownCtx)
}
