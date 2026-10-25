package inventory

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

// PreStock 是库存的 Redis 预扣加速层：Lua 内完成「读、判、扣、记」四步原子操作。
// Redis 只承担加速与防超卖，DB（inventory + stock_journal）是真账；
// 两者由 stock-sync worker 异步对齐，reconciler 对账兜底。
//
// 与现有 DB 事务三段式 Reserve/Confirm/Release 并存：
// 订单主链路仍走 TCC 预占；高并发场景（秒杀/抢购）走本层，确认时同步真账。
type PreStock struct {
	db  *sqlx.DB
	rdb *redis.Client
}

func NewPreStock(dbx *sqlx.DB, rdb *redis.Client) *PreStock {
	return &PreStock{db: dbx, rdb: rdb}
}

// Lua 结果码
const (
	luaOK           = 1
	luaInsufficient = 0
	luaDuplicate    = 2
	luaUninit       = -1
)

// prestockLua KEYS[1]=avail KEYS[2]=log ARGV[1]=qty ARGV[2]=request_id ARGV[3]=now(unix)
// 读额度 → 判幂等/判足量 → 扣额度 → 记流水，四步在同一脚本内原子执行。
const prestockLua = `
if redis.call('HEXISTS', KEYS[2], ARGV[2]) == 1 then
  return 2
end
local avail = redis.call('GET', KEYS[1])
if not avail then
  return -1
end
local q = tonumber(ARGV[1])
if tonumber(avail) < q then
  return 0
end
redis.call('DECRBY', KEYS[1], q)
redis.call('HSET', KEYS[2], ARGV[2], table.concat({ARGV[1], ARGV[3]}, '|'))
return 1
`

// prestockRestoreLua 归还流水：仅当流水存在时加回额度并删除流水（防双重归还）。
const prestockRestoreLua = `
local v = redis.call('HGET', KEYS[2], ARGV[1])
if not v then
  return 0
end
local q = tonumber(string.match(v, '^(%-?%d+)'))
redis.call('HDEL', KEYS[2], ARGV[1])
redis.call('INCRBY', KEYS[1], q)
return 1
`

func availKey(skuID string) string { return "prestock:avail:" + skuID }
func logKey(skuID string) string   { return "prestock:log:" + skuID }

// InitAvail 用 DB 真账初始化 Redis 预扣额度（SETNX，只在首次生效）。
func (p *PreStock) InitAvail(ctx context.Context, skuID string, avail int) error {
	return p.rdb.SetNX(ctx, availKey(skuID), avail, 0).Err()
}

// PreStockResult 预扣结果
type PreStockResult int

const (
	PreStockGranted PreStockResult = iota
	PreStockInsufficient
	PreStockDuplicate // 同一 request_id 重放，幂等成功
)

// Reserve 预扣一份库存。额度 key 不存在时从 DB 惰性加载并 SETNX 后重试一次。
func (p *PreStock) Reserve(ctx context.Context, skuID string, qty int, requestID string) (PreStockResult, error) {
	if qty <= 0 || requestID == "" {
		return PreStockInsufficient, errors.New("invalid prestock request")
	}
	run := func() (int64, error) {
		res, err := p.rdb.Eval(ctx, prestockLua,
			[]string{availKey(skuID), logKey(skuID)},
			qty, requestID, time.Now().Unix()).Int64()
		if err != nil {
			return 0, err
		}
		return res, nil
	}

	code, err := run()
	if errors.Is(err, redis.Nil) {
		err = nil
	}
	if err != nil {
		return PreStockInsufficient, err
	}
	if code == luaUninit {
		if err := p.lazyInit(ctx, skuID); err != nil {
			return PreStockInsufficient, err
		}
		code, err = run()
		if errors.Is(err, redis.Nil) {
			err = nil
		}
		if err != nil {
			return PreStockInsufficient, err
		}
	}

	switch code {
	case luaOK:
		return PreStockGranted, nil
	case luaDuplicate:
		return PreStockDuplicate, nil
	case luaInsufficient:
		return PreStockInsufficient, nil
	default:
		return PreStockInsufficient, fmt.Errorf("unexpected lua result %d", code)
	}
}

// Restore 归还一次预扣（订单取消）。流水不存在时幂等返回 false。
func (p *PreStock) Restore(ctx context.Context, skuID, requestID string) (bool, error) {
	n, err := p.rdb.Eval(ctx, prestockRestoreLua,
		[]string{availKey(skuID), logKey(skuID)},
		requestID).Int64()
	if errors.Is(err, redis.Nil) {
		err = nil
	}
	if err != nil {
		return false, err
	}
	return n == 1, nil
}

// GetAvail 返回当前 Redis 额度（未初始化返回 -1）。
func (p *PreStock) GetAvail(ctx context.Context, skuID string) (int, error) {
	v, err := p.rdb.Get(ctx, availKey(skuID)).Int()
	if errors.Is(err, redis.Nil) {
		return -1, nil
	}
	return v, err
}

func (p *PreStock) lazyInit(ctx context.Context, skuID string) error {
	var avail int
	err := p.db.GetContext(ctx, &avail, `SELECT available FROM inventory WHERE sku_id = ?`, skuID)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrSkuNotFound
	}
	if err != nil {
		return err
	}
	return p.rdb.SetNX(ctx, availKey(skuID), avail, 0).Err()
}

// parseLogEntry 解析 "qty|ts" 流水值
func parseLogEntry(v string) (qty int, ts int64, err error) {
	var i int
	for i < len(v) && v[i] != '|' {
		i++
	}
	if i >= len(v) {
		return 0, 0, fmt.Errorf("bad log entry %q", v)
	}
	qty, err = strconv.Atoi(v[:i])
	if err != nil {
		return 0, 0, err
	}
	ts, err = strconv.ParseInt(v[i+1:], 10, 64)
	if err != nil {
		return 0, 0, err
	}
	return qty, ts, nil
}
