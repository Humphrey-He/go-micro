package middleware

import (
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// MetricsCollector holds HTTP metrics
type MetricsCollector struct {
	mu          sync.RWMutex
	Requests    int64
	Errors      int64
	Latencies   []time.Duration
	StatusCodes map[int]int64
	ByPath      map[string]*pathMetrics
}

type pathMetrics struct {
	Count    int64
	Errors   int64
	Latencies []time.Duration
}

var (
	globalMetrics = &MetricsCollector{
		StatusCodes: make(map[int]int64),
		ByPath:      make(map[string]*pathMetrics),
	}
)

// HTTPMetrics returns a middleware that collects HTTP metrics
// Metrics collected: request count, error count, latency, status codes
func HTTPMetrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		path := c.FullPath()
		if path == "" {
			path = "unknown"
		}

		globalMetrics.RecordRequest(path, latency, status, status >= 400)
	}
}

// RecordRequest records a request in the metrics
func (m *MetricsCollector) RecordRequest(path string, latency time.Duration, status int, isError bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Requests++
	if isError {
		m.Errors++
	}
	m.StatusCodes[status]++

	if pm, ok := m.ByPath[path]; ok {
		pm.Count++
		if isError {
			pm.Errors++
		}
		pm.Latencies = append(pm.Latencies, latency)
	} else {
		m.ByPath[path] = &pathMetrics{
			Count:    1,
			Errors:   0,
			Latencies: []time.Duration{latency},
		}
		if isError {
			m.ByPath[path].Errors++
		}
	}
}

// GetMetrics returns current metrics
func (m *MetricsCollector) GetMetrics() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]interface{}{
		"requests_total": m.Requests,
		"errors_total":   m.Errors,
		"error_rate":    m.errorRate(),
		"status_codes":  m.StatusCodes,
		"by_path":       m.pathMetricsSummary(),
	}
}

func (m *MetricsCollector) errorRate() float64 {
	if m.Requests == 0 {
		return 0
	}
	return float64(m.Errors) / float64(m.Requests) * 100
}

func (m *MetricsCollector) pathMetricsSummary() map[string]map[string]interface{} {
	result := make(map[string]map[string]interface{})
	for path, pm := range m.ByPath {
		result[path] = map[string]interface{}{
			"count":      pm.Count,
			"errors":     pm.Errors,
			"error_rate": m.errorRateFor(pm),
			"avg_latency": m.avgLatency(pm.Latencies),
		}
	}
	return result
}

func (m *MetricsCollector) errorRateFor(pm *pathMetrics) float64 {
	if pm.Count == 0 {
		return 0
	}
	return float64(pm.Errors) / float64(pm.Count) * 100
}

func (m *MetricsCollector) avgLatency(latencies []time.Duration) time.Duration {
	if len(latencies) == 0 {
		return 0
	}
	var sum time.Duration
	for _, l := range latencies {
		sum += l
	}
	return sum / time.Duration(len(latencies))
}

// PrometheusMetrics returns a handler that exposes metrics in Prometheus format
func PrometheusMetrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		metrics := globalMetrics.GetMetrics()

		// Build Prometheus text format
		lines := []string{
			"# HELP go_micro_http_requests_total Total HTTP requests",
			"# TYPE go_micro_http_requests_total counter",
			"go_micro_http_requests_total " + strconv.FormatInt(metrics["requests_total"].(int64), 10),
			"# HELP go_micro_http_errors_total Total HTTP errors",
			"# TYPE go_micro_http_errors_total counter",
			"go_micro_http_errors_total " + strconv.FormatInt(metrics["errors_total"].(int64), 10),
		}

		// Add status code metrics
		statusCodes := metrics["status_codes"].(map[int]int64)
		for status, count := range statusCodes {
			lines = append(lines, "go_micro_http_requests_by_status{status=\""+strconv.Itoa(status)+"\"} "+strconv.FormatInt(count, 10))
		}

		// Add path metrics
		byPath := metrics["by_path"].(map[string]map[string]interface{})
		for path, pm := range byPath {
			lines = append(lines, "go_micro_http_requests_by_path{path=\""+path+"\"} "+strconv.FormatInt(pm["count"].(int64), 10))
		}

		c.Data(200, "text/plain; charset=utf-8", []byte(joinLines(lines)))
	}
}

func joinLines(lines []string) string {
	result := ""
	for _, line := range lines {
		result += line + "\n"
	}
	return result
}
