package inventory

import (
	"context"
	"sort"
	"strings"
	"time"

	"ecomm-micro/pkg/logx"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// StockSyncer 把 Redis 预扣流水异步落到 DB 真账：
// 事务内先 INSERT IGNORE stock_journal（唯一键 sku+request 幂等），
// 仅当新插入成功才扣减 inventory.available；落库成功后清理 Redis 流水。
// 崩溃恢复语义：Redis 流水未删 → 重投 → journal 撞键 → 不再重复扣减。
type StockSyncer struct {
	db  *sqlx.DB
	rdb *redis.Client
}

func NewStockSyncer(dbx *sqlx.DB, rdb *redis.Client) *StockSyncer {
	return &StockSyncer{db: dbx, rdb: rdb}
}

// SyncOnce 执行一轮同步，返回本轮落库条数。
func (s *StockSyncer) SyncOnce(ctx context.Context) (int, error) {
	// 演示规模用 KEYS；生产应换 SCAN 渐进遍历
	keys, err := s.rdb.Keys(ctx, "prestock:log:*").Result()
	if err != nil {
		return 0, err
	}
	synced := 0
	for _, key := range keys {
		skuID := strings.TrimPrefix(key, "prestock:log:")
		entries, err := s.rdb.HGetAll(ctx, key).Result()
		if err != nil {
			return synced, err
		}
		// map 迭代顺序随机：按 request_id 排序保证同步顺序确定
		ids := make([]string, 0, len(entries))
		for id := range entries {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		for _, requestID := range ids {
			val := entries[requestID]
			qty, _, err := parseLogEntry(val)
			if err != nil {
				logx.L().Warn("prestock: skip malformed log entry",
					zap.String("sku", skuID), zap.String("request_id", requestID), zap.Error(err))
				continue
			}
			ok, err := s.persistOne(ctx, skuID, requestID, qty)
			if err != nil {
				return synced, err
			}
			if ok {
				synced++
			}
			// 无论新落库还是撞键幂等，流水都已完成使命
			s.rdb.HDel(ctx, key, requestID)
		}
	}
	return synced, nil
}

// persistOne 在一个事务内完成 journal 落库 + 真账扣减。
// 返回 true 表示新落库；false 表示 journal 已存在（重复投递被幂等吸收）。
func (s *StockSyncer) persistOne(ctx context.Context, skuID, requestID string, qty int) (bool, error) {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx,
		`INSERT IGNORE INTO stock_journal (sku_id, request_id, quantity) VALUES (?, ?, ?)`,
		skuID, requestID, qty)
	if err != nil {
		return false, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	if affected == 0 {
		// 唯一键撞键：该流水曾落库（inventory 同事务已扣过），幂等跳过
		return false, tx.Rollback()
	}
	if _, err := tx.ExecContext(ctx,
		`UPDATE inventory SET available = available - ?, updated_at = NOW() WHERE sku_id = ?`,
		qty, skuID); err != nil {
		return false, err
	}
	if err := tx.Commit(); err != nil {
		return false, err
	}
	return true, nil
}

// StartStockSyncWorker 周期性执行 SyncOnce，直到 stop 关闭。
func StartStockSyncer(dbx *sqlx.DB, rdb *redis.Client, interval time.Duration, stop <-chan struct{}) {
	s := NewStockSyncer(dbx, rdb)
	if interval <= 0 {
		interval = time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			n, err := s.SyncOnce(ctx)
			cancel()
			if err != nil {
				logx.L().Warn("prestock sync round failed", zap.Error(err))
				continue
			}
			if n > 0 {
				logx.L().Info("prestock synced to db", zap.Int("entries", n))
			}
		}
	}
}
