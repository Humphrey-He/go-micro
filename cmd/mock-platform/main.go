package main

import (
	"net/http"
	"os"
	"strconv"

	ecomm "ecomm-micro/internal/mockplatform"
	"github.com/gin-gonic/gin"
)

// mock-platform：电商平台的可配置 mock。
//
//	PORT        监听地址（如 :9091）
//	PLATFORM    平台名（展示与 task id 前缀）
//	QPS_LIMIT   限流阈值（默认 50）
//	LOSS_RATE   回执丢失率 0~1（默认 0）
func main() {
	port := envOr("PORT", ":9091")
	name := envOr("PLATFORM", "mock")
	qps := envFloat("QPS_LIMIT", 50)
	loss := envFloat("LOSS_RATE", 0)

	gin.SetMode(gin.ReleaseMode)
	p := ecomm.New(name, qps, loss)
	_ = http.ListenAndServe(port, p.Router())
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func envFloat(k string, def float64) float64 {
	if v := os.Getenv(k); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return def
}
