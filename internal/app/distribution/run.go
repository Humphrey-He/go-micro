package distributionapp

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"ecomm-micro/internal/distribution"
	"ecomm-micro/pkg/config"
	"ecomm-micro/pkg/db"
	"ecomm-micro/pkg/logx"
	"ecomm-micro/pkg/middleware"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Run distribution-service：渠道映射分发 + 按店铺令牌桶限频 + 轮次回执比对兜底。
func Run() error {
	logger := logx.L()
	defer logx.Sync()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := config.ValidateService("distribution-service"); err != nil {
		logger.Error("config validation failed", zap.Error(err))
		return err
	}

	dbx, err := db.NewMySQL()
	if err != nil {
		logger.Error("mysql connect failed", zap.Error(err))
		return err
	}

	endpoints := parseEndpoints(config.GetEnv("PLATFORM_ENDPOINTS", ""))
	shopQPS := parseFloat(config.GetEnv("SHOP_QPS", ""), 15)
	reconInterval := time.Duration(config.GetInt("RECONCILE_INTERVAL_SECONDS", 900)) * time.Second
	receiptGrace := time.Duration(config.GetInt("RECEIPT_GRACE_SECONDS", 30)) * time.Second

	svc := distribution.NewService(dbx)
	sched := distribution.NewScheduler(dbx, endpoints, shopQPS)
	recon := distribution.NewReconciler(dbx, endpoints, receiptGrace)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger(logger))
	r.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	r.GET("/readyz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ready"}) })
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	distribution.NewHandler(svc, sched, recon).Register(r)

	go distribution.StartScheduler(sched, 500*time.Millisecond, ctx.Done())
	go distribution.StartReconciler(recon, reconInterval, ctx.Done())

	addr := config.GetEnv("DISTRIBUTION_ADDR", ":8090")
	srv := &http.Server{Addr: addr, Handler: r}
	logger.Info("distribution-service starting",
		zap.String("http_addr", addr),
		zap.Any("platforms", keysOf(endpoints)),
		zap.Float64("shop_qps", shopQPS),
		zap.Duration("reconcile_interval", reconInterval))
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("distribution-service start failed", zap.Error(err))
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	logger.Info("distribution-service shutting down")
	return srv.Shutdown(shutdownCtx)
}

// parseEndpoints "tmall=http://a:9091,jd=http://b:9092"
func parseEndpoints(s string) map[string]string {
	out := make(map[string]string)
	for _, pair := range strings.Split(s, ",") {
		if pair == "" {
			continue
		}
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) == 2 && kv[0] != "" && kv[1] != "" {
			out[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
	}
	return out
}

func parseFloat(s string, def float64) float64 {
	if s == "" {
		return def
	}
	if f, err := strconv.ParseFloat(s, 64); err == nil && f > 0 {
		return f
	}
	return def
}

func keysOf(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
