package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	Namespace = "gomicro"
	Subsystem = "service"
)

var (
	// HTTP metrics
	HTTPRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: "http",
			Name:      "requests_total",
			Help:      "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	HTTPRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: Namespace,
			Subsystem: "http",
			Name:      "request_duration_seconds",
			Help:      "HTTP request latency in seconds",
			Buckets:   []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"method", "path"},
	)

	HTTPUsecondsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: "http",
			Name:      "request_duration_seconds_total",
			Help:      "Total HTTP request duration in seconds",
		},
		[]string{"method", "path"},
	)

	// gRPC metrics
	GRPCRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: "grpc",
			Name:      "requests_total",
			Help:      "Total number of gRPC requests",
		},
		[]string{"method", "status"},
	)

	GRPCRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: Namespace,
			Subsystem: "grpc",
			Name:      "request_duration_seconds",
			Help:      "gRPC request latency in seconds",
			Buckets:   []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5},
		},
		[]string{"method"},
	)

	// Business metrics - Order
	OrderCreatedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: "order",
			Name:      "created_total",
			Help:      "Total number of orders created",
		},
		[]string{"status"},
	)

	OrderCreateFailedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: "order",
			Name:      "create_failed_total",
			Help:      "Total number of order creation failures",
		},
		[]string{"reason"},
	)

	OrderCanceledTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: "order",
			Name:      "canceled_total",
			Help:      "Total number of orders canceled",
		},
		[]string{"reason"},
	)

	// Business metrics - Payment
	PaymentCreatedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: "payment",
			Name:      "created_total",
			Help:      "Total number of payments created",
		},
		[]string{"status"},
	)

	PaymentSuccessTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: "payment",
			Name:      "success_total",
			Help:      "Total number of successful payments",
		},
	)

	PaymentFailedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: "payment",
			Name:      "failed_total",
			Help:      "Total number of failed payments",
		},
		[]string{"reason"},
	)

	PaymentTimeoutTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: "payment",
			Name:      "timeout_total",
			Help:      "Total number of payment timeouts",
		},
	)

	// Business metrics - Inventory
	InventoryReservedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: "inventory",
			Name:      "reserved_total",
			Help:      "Total number of inventory reservations",
		},
		[]string{"status"},
	)

	InventoryReserveFailedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: "inventory",
			Name:      "reserve_failed_total",
			Help:      "Total number of inventory reservation failures",
		},
		[]string{"reason"},
	)

	InventoryInsufficientTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: "inventory",
			Name:      "insufficient_total",
			Help:      "Total number of insufficient inventory events",
		},
	)

	// Business metrics - Outbox
	OutboxPublishedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: "outbox",
			Name:      "published_total",
			Help:      "Total number of outbox messages published",
		},
	)

	OutboxPublishFailedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: "outbox",
			Name:      "publish_failed_total",
			Help:      "Total number of outbox publish failures",
		},
		[]string{"reason"},
	)

	OutboxPendingGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: Namespace,
			Subsystem: "outbox",
			Name:      "pending_count",
			Help:      "Current number of pending outbox messages",
		},
	)

	// Business metrics - Refund
	RefundInitiatedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: "refund",
			Name:      "initiated_total",
			Help:      "Total number of refunds initiated",
		},
		[]string{"status"},
	)

	RefundFailedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: "refund",
			Name:      "failed_total",
			Help:      "Total number of refund failures",
		},
		[]string{"reason"},
	)

	// MQ metrics
	MQMessagesPublished = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: "mq",
			Name:      "messages_published_total",
			Help:      "Total number of messages published to MQ",
		},
		[]string{"exchange", "routing_key"},
	)

	MQMessagesConsumed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: "mq",
			Name:      "messages_consumed_total",
			Help:      "Total number of messages consumed from MQ",
		},
		[]string{"queue"},
	)

	MQConsumerErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: "mq",
			Name:      "consumer_errors_total",
			Help:      "Total number of MQ consumer errors",
		},
		[]string{"queue", "error_type"},
	)

	MQQueueDepth = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: Namespace,
			Subsystem: "mq",
			Name:      "queue_depth",
			Help:      "Current depth of MQ queues",
		},
		[]string{"queue"},
	)

	MQDeadLetterTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: "mq",
			Name:      "dead_letter_total",
			Help:      "Total number of messages sent to dead letter queue",
		},
		[]string{"queue"},
	)

	// Task metrics
	TaskCreatedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: "task",
			Name:      "created_total",
			Help:      "Total number of tasks created",
		},
	)

	TaskCompletedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: "task",
			Name:      "completed_total",
			Help:      "Total number of tasks completed",
		},
	)

	TaskFailedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: "task",
			Name:      "failed_total",
			Help:      "Total number of task failures",
		},
		[]string{"reason"},
	)

	TaskRetryTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: "task",
			Name:      "retry_total",
			Help:      "Total number of task retries",
		},
	)

	// Database metrics
	DBQueryDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: Namespace,
			Subsystem: "db",
			Name:      "query_duration_seconds",
			Help:      "Database query duration in seconds",
			Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
		},
		[]string{"operation", "table"},
	)

	DBConnectionPoolGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: Namespace,
			Subsystem: "db",
			Name:      "connection_pool_size",
			Help:      "Current database connection pool size",
		},
		[]string{"state"},
	)

	// Error metrics
	ErrorTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: "error",
			Name:      "total",
			Help:      "Total number of errors",
		},
		[]string{"type", "service"},
	)
)

func init() {
	prometheus.MustRegister(
		HTTPRequestsTotal,
		HTTPRequestDuration,
		HTTPUsecondsTotal,
		GRPCRequestsTotal,
		GRPCRequestDuration,
		OrderCreatedTotal,
		OrderCreateFailedTotal,
		OrderCanceledTotal,
		PaymentCreatedTotal,
		PaymentSuccessTotal,
		PaymentFailedTotal,
		PaymentTimeoutTotal,
		InventoryReservedTotal,
		InventoryReserveFailedTotal,
		InventoryInsufficientTotal,
		OutboxPublishedTotal,
		OutboxPublishFailedTotal,
		OutboxPendingGauge,
		RefundInitiatedTotal,
		RefundFailedTotal,
		MQMessagesPublished,
		MQMessagesConsumed,
		MQConsumerErrors,
		MQQueueDepth,
		MQDeadLetterTotal,
		TaskCreatedTotal,
		TaskCompletedTotal,
		TaskFailedTotal,
		TaskRetryTotal,
		DBQueryDuration,
		DBConnectionPoolGauge,
		ErrorTotal,
	)
}

func RecordHTTPRequest(method, path, status string, duration time.Duration) {
	HTTPRequestsTotal.WithLabelValues(method, path, status).Inc()
	HTTPRequestDuration.WithLabelValues(method, path).Observe(duration.Seconds())
	HTTPUsecondsTotal.WithLabelValues(method, path).Add(duration.Seconds())
}

func RecordGRPCRequest(method, status string, duration time.Duration) {
	GRPCRequestsTotal.WithLabelValues(method, status).Inc()
	GRPCRequestDuration.WithLabelValues(method).Observe(duration.Seconds())
}

func RecordOrderCreated(status string) {
	OrderCreatedTotal.WithLabelValues(status).Inc()
}

func RecordOrderCreateFailed(reason string) {
	OrderCreateFailedTotal.WithLabelValues(reason).Inc()
}

func RecordPaymentSuccess() {
	PaymentSuccessTotal.Inc()
}

func RecordPaymentFailed(reason string) {
	PaymentFailedTotal.WithLabelValues(reason).Inc()
}

func RecordInventoryInsufficient() {
	InventoryInsufficientTotal.Inc()
}

func RecordOutboxPending(count float64) {
	OutboxPendingGauge.Set(count)
}
