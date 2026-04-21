package outbox

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	StatusPending   = "PENDING"
	StatusSent      = "SENT"
	StatusFailed    = "FAILED"
	StatusDLQ       = "DLQ"
)

var (
	outboxGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "gomicro",
		Subsystem: "outbox",
		Name:      "pending_count",
		Help:      "Current number of pending outbox messages",
	})

	outboxPublishSuccess = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "gomicro",
		Subsystem: "outbox",
		Name:      "publish_success_total",
		Help:      "Total outbox publish successes",
	})

	outboxPublishFailed = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "gomicro",
		Subsystem: "outbox",
		Name:      "publish_failed_total",
		Help:      "Total outbox publish failures",
	}, []string{"reason"})

	outboxDLQCount = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "gomicro",
		Subsystem: "outbox",
		Name:      "dlq_total",
		Help:      "Total messages sent to DLQ",
	})
)

func init() {
	prometheus.MustRegister(outboxGauge, outboxPublishSuccess, outboxPublishFailed, outboxDLQCount)
}

type Message struct {
	ID          int64           `json:"id" db:"id"`
	EventType   string          `json:"event_type" db:"event_type"`
	Payload     json.RawMessage `json:"payload" db:"payload"`
	Status      string          `json:"status" db:"status"`
	RetryCount int             `json:"retry_count" db:"retry_count"`
	LastError   string          `json:"last_error" db:"last_error"`
	CreatedAt   time.Time      `json:"created_at" db:"created_at"`
	SentAt      sql.NullTime    `json:"sent_at" db:"sent_at"`
}

type Publisher interface {
	Publish(ctx context.Context, topic string, payload json.RawMessage) error
}

type Config struct {
	BatchSize        int
	Interval         time.Duration
	MaxRetries       int
	BaseBackoff      time.Duration
	MaxBackoff       time.Duration
	StuckThreshold   time.Duration
	DLQTopic         string
}

func DefaultConfig() *Config {
	return &Config{
		BatchSize:      100,
		Interval:       2 * time.Second,
		MaxRetries:     10,
		BaseBackoff:    1 * time.Second,
		MaxBackoff:     60 * time.Second,
		StuckThreshold: 1 * time.Hour,
	}
}

type OutboxPublisher struct {
	db       DBInterface
	pub      Publisher
	cfg      *Config
	stopCh   chan struct{}
	stopOnce sync.Once
}

type DBInterface interface {
	Get(dest interface{}, query string, args ...interface{}) error
	Select(dest interface{}, query string, args ...interface{}) error
	Exec(query string, args ...interface{}) (sql.Result, error)
	Beginx() (TXInterface, error)
}

type TXInterface interface {
	Select(dest interface{}, query string, args ...interface{}) error
	Exec(query string, args ...interface{}) (sql.Result, error)
	Commit() error
	Rollback() error
}

func NewOutboxPublisher(db DBInterface, pub Publisher, cfg *Config) *OutboxPublisher {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &OutboxPublisher{
		db:    db,
		pub:   pub,
		cfg:   cfg,
		stopCh: make(chan struct{}),
	}
}

func (p *OutboxPublisher) Start(ctx context.Context) {
	go p.run(ctx)
}

func (p *OutboxPublisher) Stop() {
	p.stopOnce.Do(func() {
		close(p.stopCh)
	})
}

func (p *OutboxPublisher) run(ctx context.Context) {
	ticker := time.NewTicker(p.cfg.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-p.stopCh:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.processPending(ctx)
			p.processStuck(ctx)
		}
	}
}

func (p *OutboxPublisher) processPending(ctx context.Context) {
	var pending int64
	if err := p.db.Get(&pending, "SELECT COUNT(1) FROM order_outbox WHERE status = ?", StatusPending); err == nil {
		outboxGauge.Set(float64(pending))
	}

	tx, err := p.db.Beginx()
	if err != nil {
		return
	}

	var rows []Message
	if err := tx.Select(&rows, `
		SELECT id, event_type, payload, retry_count, last_error, created_at
		FROM order_outbox
		WHERE status = ? AND retry_count < ?
		ORDER BY id ASC
		LIMIT ?`,
		StatusPending, p.cfg.MaxRetries, p.cfg.BatchSize); err != nil {
		_ = tx.Rollback()
		return
	}

	for _, r := range rows {
		backoff := p.calculateBackoff(r.RetryCount)

		if time.Since(r.CreatedAt) < backoff {
			continue
		}

		if err := p.pub.Publish(ctx, r.EventType, r.Payload); err != nil {
			_, _ = tx.Exec(`
				UPDATE order_outbox
				SET retry_count = retry_count + 1, last_error = ?, updated_at = NOW()
				WHERE id = ?`,
				err.Error(), r.ID)
			outboxPublishFailed.WithLabelValues(classifyError(err)).Inc()
			continue
		}

		if _, err := tx.Exec(`
			UPDATE order_outbox
			SET status = ?, sent_at = NOW(), updated_at = NOW()
			WHERE id = ?`,
			StatusSent, r.ID); err != nil {
			_ = tx.Rollback()
			return
		}
		outboxPublishSuccess.Inc()
	}

	_ = tx.Commit()
}

func (p *OutboxPublisher) processStuck(ctx context.Context) {
	var stuck []Message
	if err := p.db.Select(&stuck, `
		SELECT id, event_type, payload, retry_count, last_error, created_at
		FROM order_outbox
		WHERE status = ? AND retry_count >= ? AND created_at < ?`,
		StatusPending, p.cfg.MaxRetries, time.Now().Add(-p.cfg.StuckThreshold)); err != nil {
		return
	}

	for _, r := range stuck {
		if p.cfg.DLQTopic != "" {
			if err := p.pub.Publish(ctx, p.cfg.DLQTopic, r.Payload); err != nil {
				_, _ = p.db.Exec(`
					UPDATE order_outbox
					SET last_error = ?, updated_at = NOW()
					WHERE id = ?`,
					fmt.Sprintf("DLQ publish failed: %s", err.Error()), r.ID)
				continue
			}
		}

		_, _ = p.db.Exec(`
			UPDATE order_outbox
			SET status = ?, last_error = ?, updated_at = NOW()
			WHERE id = ?`,
			StatusDLQ, fmt.Sprintf("Moved to DLQ after %d retries", r.RetryCount), r.ID)
		outboxDLQCount.Inc()
	}
}

func (p *OutboxPublisher) calculateBackoff(retryCount int) time.Duration {
	backoff := float64(p.cfg.BaseBackoff) * math.Pow(2, float64(retryCount))
	if backoff > float64(p.cfg.MaxBackoff) {
		backoff = float64(p.cfg.MaxBackoff)
	}
	return time.Duration(backoff)
}

func classifyError(err error) string {
	if err == nil {
		return "unknown"
	}
	errStr := err.Error()
	switch {
	case contains(errStr, "timeout"):
		return "timeout"
	case contains(errStr, "connection"):
		return "connection"
	case contains(errStr, "refused"):
		return "connection"
	case contains(errStr, "unavailable"):
		return "unavailable"
	default:
		return "other"
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func RecordOutboxPending(count float64) {
	outboxGauge.Set(count)
}
