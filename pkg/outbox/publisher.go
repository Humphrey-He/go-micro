package outbox

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	StatusPending = "PENDING"
	StatusSent    = "SENT"
	StatusFailed  = "FAILED"
	StatusDLQ     = "DLQ"
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
	ID         int64           `json:"id" db:"id"`
	EventType  string          `json:"event_type" db:"event_type"`
	Payload    json.RawMessage `json:"payload" db:"payload"`
	Status     string          `json:"status" db:"status"`
	RetryCount int             `json:"retry_count" db:"retry_count"`
	LastError  string          `json:"last_error" db:"last_error"`
	CreatedAt  time.Time       `json:"created_at" db:"created_at"`
	SentAt     sql.NullTime    `json:"sent_at" db:"sent_at"`
	Owner      string          `json:"owner" db:"owner"`
	LeaseUntil sql.NullTime    `json:"lease_until" db:"lease_until"`
}

// StatusClaimed 已被某实例认领、租约期内其他实例不得处理
const StatusClaimed = "CLAIMED"

type Publisher interface {
	Publish(ctx context.Context, topic string, payload json.RawMessage) error
}

type Config struct {
	BatchSize      int
	Interval       time.Duration
	MaxRetries     int
	BaseBackoff    time.Duration
	MaxBackoff     time.Duration
	StuckThreshold time.Duration
	DLQTopic       string
	// LeaseTTL 认领租约时长：持有者超过该时间未完成，消息可被其他实例重新认领
	LeaseTTL time.Duration
	// Owner 持有者标识（默认 hostname+pid）；多实例部署时用于 owner 校验
	Owner string
}

func DefaultConfig() *Config {
	return &Config{
		BatchSize:      100,
		Interval:       2 * time.Second,
		MaxRetries:     10,
		BaseBackoff:    1 * time.Second,
		MaxBackoff:     60 * time.Second,
		StuckThreshold: 1 * time.Hour,
		LeaseTTL:       30 * time.Second,
	}
}

type OutboxPublisher struct {
	db       DBInterface
	pub      Publisher
	cfg      *Config
	owner    string
	stopCh   chan struct{}
	stopOnce sync.Once
}

func (p *OutboxPublisher) Owner() string { return p.owner }

// DBInterface 由 *sqlx.DB 直接满足（认领/标记均为独立小语句，不再依赖长事务）。
type DBInterface interface {
	Get(dest interface{}, query string, args ...interface{}) error
	Select(dest interface{}, query string, args ...interface{}) error
	Exec(query string, args ...interface{}) (sql.Result, error)
}

func NewOutboxPublisher(db DBInterface, pub Publisher, cfg *Config) *OutboxPublisher {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	owner := cfg.Owner
	if owner == "" {
		host, _ := os.Hostname()
		owner = fmt.Sprintf("%s-%d", host, os.Getpid())
	}
	return &OutboxPublisher{
		db:     db,
		pub:    pub,
		cfg:    cfg,
		owner:  owner,
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

	// 候选 = PENDING 待投递 + 租约过期的 CLAIMED（持有者疑似崩溃，允许重投；
	// 重投可能造成下游重复消费，由消费端幂等表吸收——at-least-once + 幂等消费）
	var rows []Message
	if err := p.db.Select(&rows, `
		SELECT id, event_type, payload, retry_count, last_error, created_at, owner, lease_until
		FROM order_outbox
		WHERE (status = ? OR (status = ? AND lease_until < NOW())) AND retry_count < ?
		ORDER BY id ASC
		LIMIT ?`,
		StatusPending, StatusClaimed, p.cfg.MaxRetries, p.cfg.BatchSize); err != nil {
		return
	}

	for _, r := range rows {
		backoff := p.calculateBackoff(r.RetryCount)
		if time.Since(r.CreatedAt) < backoff {
			continue
		}

		// CAS 认领：仅当仍是 PENDING（或租约已过期）且未被他人抢先时生效
		res, err := p.db.Exec(`
			UPDATE order_outbox
			SET status = ?, owner = ?, lease_until = DATE_ADD(NOW(), INTERVAL ? SECOND), updated_at = NOW()
			WHERE id = ? AND (status = ? OR (status = ? AND lease_until < NOW()))`,
			StatusClaimed, p.owner, int(p.cfg.LeaseTTL.Seconds()), r.ID, StatusPending, StatusClaimed)
		if err != nil {
			continue
		}
		if affected, _ := res.RowsAffected(); affected == 0 {
			continue // 被其他实例抢先认领
		}

		if err := p.pub.Publish(ctx, r.EventType, r.Payload); err != nil {
			// 失败回退 PENDING（owner 校验：租约期内才允许本实例回退）
			_, _ = p.db.Exec(`
				UPDATE order_outbox
				SET status = ?, retry_count = retry_count + 1, last_error = ?, owner = '', lease_until = NULL, updated_at = NOW()
				WHERE id = ? AND owner = ? AND status = ?`,
				StatusPending, err.Error(), r.ID, p.owner, StatusClaimed)
			outboxPublishFailed.WithLabelValues(classifyError(err)).Inc()
			continue
		}

		// 成功标记 SENT：owner 校验防止租约过期被抢走后覆盖他人状态
		if _, err := p.db.Exec(`
			UPDATE order_outbox
			SET status = ?, sent_at = NOW(), updated_at = NOW()
			WHERE id = ? AND owner = ? AND status = ?`,
			StatusSent, r.ID, p.owner, StatusClaimed); err == nil {
			outboxPublishSuccess.Inc()
		}
	}
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
