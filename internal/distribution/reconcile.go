package distribution

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"ecomm-micro/pkg/logx"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// Reconciler 15 分钟轮次差异兜底（演示可配短周期）：
// PENDING_ACK 超过宽限期仍无回执 → 查平台 → 404 即漏发 →
// 记台账（与 R4 共用 reconcile_diff，biz_type=DISTRIBUTION）→ 同 client_msg_id 重新入队补发。
type Reconciler struct {
	db        *sqlx.DB
	endpoints map[string]string
	client    *http.Client
	grace     time.Duration
}

func NewReconciler(dbx *sqlx.DB, endpoints map[string]string, grace time.Duration) *Reconciler {
	return &Reconciler{
		db:        dbx,
		endpoints: endpoints,
		client:    &http.Client{Timeout: 3 * time.Second},
		grace:     grace,
	}
}

// Tick 执行一轮回执比对。返回 {checked, missing, confirmed}。
func (r *Reconciler) Tick(ctx context.Context) (checked, missing, confirmed int, err error) {
	var tasks []DispatchTask
	cutoff := time.Now().Add(-r.grace)
	if err = r.db.SelectContext(ctx, &tasks, `
		SELECT id, client_msg_id, sku_id, platform, site, shop_id, status, sent_at
		FROM dispatch_tasks
		WHERE status = ? AND sent_at < ? LIMIT 200`, TaskPendingAck, cutoff); err != nil {
		return
	}

	for _, t := range tasks {
		checked++
		st, err2 := r.queryReceipt(ctx, t)
		if err2 != nil {
			logx.L().Warn("receipt query failed",
				zap.String("client_msg_id", t.ClientMsgID), zap.Error(err2))
			continue
		}
		if st == http.StatusOK {
			confirmed++
			_, _ = r.db.ExecContext(ctx, `
				UPDATE dispatch_tasks SET status = ?, updated_at = NOW() WHERE id = ?`,
				TaskConfirmed, t.ID)
			continue
		}
		// 404：回执缺失 = 漏发嫌疑 → 台账 + 同 client_msg_id 补发（平台侧幂等吸收）
		missing++
		detail := fmt.Sprintf("receipt missing for %s on %s after grace; redispatched", t.ClientMsgID, t.Platform)
		_, _ = r.db.ExecContext(ctx, `
			INSERT IGNORE INTO reconcile_diff (biz_type, sku_id, request_id, diff_type, detail)
			VALUES (?, ?, ?, ?, ?)`,
			LedgerBizType, t.SkuID, t.ClientMsgID, LedgerDiffType, detail)
		_, _ = r.db.ExecContext(ctx, `
			UPDATE dispatch_tasks
			SET status = ?, redispatch_count = redispatch_count + 1, sent_at = NULL, updated_at = NOW()
			WHERE id = ?`, TaskQueued, t.ID)
	}
	return
}

func (r *Reconciler) queryReceipt(ctx context.Context, t DispatchTask) (int, error) {
	base, ok := r.endpoints[t.Platform]
	if !ok {
		return 0, fmt.Errorf("no endpoint for platform %s", t.Platform)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		base+"/open/api/receipts/"+t.ClientMsgID, nil)
	if err != nil {
		return 0, err
	}
	resp, err := r.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	return resp.StatusCode, nil
}

// StartReconciler 周期回执比对，直到 stop 关闭。
func StartReconciler(r *Reconciler, interval time.Duration, stop <-chan struct{}) {
	if interval <= 0 {
		interval = 15 * time.Minute
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			_, missing, _, err := r.Tick(ctx)
			cancel()
			if err != nil {
				logx.L().Warn("receipt reconcile tick failed", zap.Error(err))
				continue
			}
			if missing > 0 {
				logx.L().Warn("receipt reconcile redispatched", zap.Int("missing", missing))
			}
		}
	}
}

// ListTasks 演示/验收用：按状态查任务。
func (s *Service) ListTasks(ctx context.Context, status string, limit int) ([]DispatchTask, error) {
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	query := `SELECT id, client_msg_id, batch_id, sku_id, platform, site, shop_id, status,
		platform_task_id, redispatch_count, last_error, sent_at, created_at, updated_at
		FROM dispatch_tasks`
	var args []interface{}
	if status != "" {
		query += " WHERE status = ?"
		args = append(args, status)
	}
	query += " ORDER BY id DESC LIMIT ?"
	args = append(args, limit)
	var tasks []DispatchTask
	if err := s.db.SelectContext(ctx, &tasks, query, args...); err != nil {
		return nil, err
	}
	return tasks, nil
}

var _ = json.Marshal // keep encoding/json referenced for future payload fields
