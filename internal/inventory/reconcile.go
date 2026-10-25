package inventory

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"ecomm-micro/pkg/logx"
	"go.uber.org/zap"
)

// reconcileGrace：Redis 流水晚于该时长仍未落 DB 才判为差异，
// 避免把「正在途中的同步」误报成差异。
const reconcileGrace = 30 * time.Second

// 差异类型
const (
	DiffRedisNotSynced = "REDIS_NOT_SYNCED" // Redis 流水存在、DB journal 缺失
	DiffAvailMismatch  = "AVAIL_MISMATCH"   // Redis 额度与真账推导值不符
)

// ReconcileReport 单轮对账报告
type ReconcileReport struct {
	Checked int // 检查的未落库流水条数
	Diffs   int // 新增差异条数
}

// Reconciler 对账兜底：比对 Redis 预扣流水与 DB stock_journal，
// 差异写入 reconcile_diff 台账（状态机 PENDING → RESOLVED/IGNORED）。
type Reconciler struct {
	db  *sqlx.DB
	rdb *redis.Client
}

func NewReconciler(dbx *sqlx.DB, rdb *redis.Client) *Reconciler {
	return &Reconciler{db: dbx, rdb: rdb}
}

// RunOnce 执行一轮对账；now 参数化便于测试。
func (r *Reconciler) RunOnce(ctx context.Context, now time.Time) (*ReconcileReport, error) {
	keys, err := r.rdb.Keys(ctx, "prestock:log:*").Result()
	if err != nil {
		return nil, err
	}
	report := &ReconcileReport{}
	for _, key := range keys {
		entries, err := r.rdb.HGetAll(ctx, key).Result()
		if err != nil {
			return report, err
		}
		skuID := key[len("prestock:log:"):]
		for requestID, val := range entries {
			report.Checked++
			qty, ts, err := parseLogEntry(val)
			if err != nil {
				logx.L().Warn("reconcile: malformed entry",
					zap.String("sku", skuID), zap.String("request_id", requestID), zap.Error(err))
				continue
			}
			if now.Sub(time.Unix(ts, 0)) <= reconcileGrace {
				continue // 同步在途，给宽限
			}
			// 落库核对：该流水是否已写入 journal（HDel 失败窗口内可能滞后）
			var exists int
			if err := r.db.GetContext(ctx, &exists,
				`SELECT COUNT(1) FROM stock_journal WHERE sku_id = ? AND request_id = ?`,
				skuID, requestID); err != nil {
				return report, err
			}
			if exists > 0 {
				// journal 已有但 Redis 未清理：同步链路滞后，直接吸收
				r.rdb.HDel(ctx, key, requestID)
				continue
			}
			detail := fmt.Sprintf("redis log qty=%d ts=%d older than grace, missing in stock_journal", qty, ts)
			if err := r.recordDiff(ctx, skuID, requestID, DiffRedisNotSynced, detail); err != nil {
				return report, err
			}
			report.Diffs++
		}
	}
	return report, nil
}

// recordDiff 写台账（INSERT IGNORE 幂等，同一差异只记一次）。
func (r *Reconciler) recordDiff(ctx context.Context, skuID, requestID, diffType, detail string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT IGNORE INTO reconcile_diff (biz_type, sku_id, request_id, diff_type, detail)
		 VALUES ('STOCK', ?, ?, ?, ?)`,
		skuID, requestID, diffType, detail)
	return err
}

// ResolveDiff 人工/自动修复后关闭台账条目。
func (r *Reconciler) ResolveDiff(ctx context.Context, diffID int64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE reconcile_diff SET status = 'RESOLVED', resolved_at = NOW() WHERE id = ? AND status = 'PENDING'`,
		diffID)
	return err
}

// StartReconciler 周期对账，直到 stop 关闭。
func StartReconciler(dbx *sqlx.DB, rdb *redis.Client, interval time.Duration, stop <-chan struct{}) {
	rec := NewReconciler(dbx, rdb)
	if interval <= 0 {
		interval = time.Minute
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			report, err := rec.RunOnce(ctx, time.Now())
			cancel()
			if err != nil {
				logx.L().Warn("prestock reconcile round failed", zap.Error(err))
				continue
			}
			if report.Diffs > 0 {
				logx.L().Warn("prestock reconcile found diffs",
					zap.Int("checked", report.Checked), zap.Int("diffs", report.Diffs))
			}
		}
	}
}
