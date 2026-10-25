package inventory

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alicebob/miniredis/v2"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

func newPreStock(t *testing.T) (*PreStock, sqlmock.Sqlmock, *miniredis.Miniredis) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	sqlxDB := sqlx.NewDb(db, "mysql")

	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() {
		sqlxDB.Close()
		rdb.Close()
	})
	return NewPreStock(sqlxDB, rdb), mock, mr
}

func hLen(t *testing.T, rdb *redis.Client, key string) int64 {
	t.Helper()
	n, err := rdb.HLen(context.Background(), key).Result()
	if err != nil {
		t.Fatal(err)
	}
	return n
}

// R1-验收单测版：1000 并发扣 100 库存——最终额度=0、无超卖、无少卖
func TestPreStockReserve_ConcurrentNoOversell(t *testing.T) {
	p, _, mr := newPreStock(t)
	ctx := context.Background()
	const total, concurrency = 100, 1000

	if err := p.InitAvail(ctx, "SKU-1", total); err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	granted := 0
	rejected := 0
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			res, err := p.Reserve(ctx, "SKU-1", 1, fmt.Sprintf("req-%d", n))
			if err != nil {
				t.Errorf("reserve %d: %v", n, err)
				return
			}
			mu.Lock()
			defer mu.Unlock()
			switch res {
			case PreStockGranted, PreStockDuplicate:
				granted++
			case PreStockInsufficient:
				rejected++
			default:
				t.Errorf("unexpected result %d", res)
			}
		}(i)
	}
	wg.Wait()

	if granted != total {
		t.Fatalf("granted = %d, want exactly %d (oversell/undersell)", granted, total)
	}
	if rejected != concurrency-total {
		t.Fatalf("rejected = %d, want %d", rejected, concurrency-total)
	}
	avail, _ := mr.Get(availKey("SKU-1"))
	if avail != "0" {
		t.Fatalf("redis avail = %s, want 0", avail)
	}
	if n := hLen(t, p.rdb, logKey("SKU-1")); n != total {
		t.Fatalf("redis log entries = %d, want %d", n, total)
	}
}

func TestPreStockReserve_IdempotentReplay(t *testing.T) {
	p, _, mr := newPreStock(t)
	ctx := context.Background()

	if err := p.InitAvail(ctx, "SKU-1", 10); err != nil {
		t.Fatal(err)
	}
	if res, _ := p.Reserve(ctx, "SKU-1", 3, "req-A"); res != PreStockGranted {
		t.Fatalf("first = %d, want granted", res)
	}
	// 同一 request_id 重放：返回 DUPLICATE，额度只扣一次
	if res, _ := p.Reserve(ctx, "SKU-1", 3, "req-A"); res != PreStockDuplicate {
		t.Fatalf("replay = %d, want duplicate", res)
	}
	if avail, _ := mr.Get(availKey("SKU-1")); avail != "7" {
		t.Fatalf("avail = %s, want 7 (replay must not deduct again)", avail)
	}
	if n := hLen(t, p.rdb, logKey("SKU-1")); n != 1 {
		t.Fatalf("log entries = %d, want 1", n)
	}
}

func TestPreStockReserve_Insufficient(t *testing.T) {
	p, _, _ := newPreStock(t)
	ctx := context.Background()

	if err := p.InitAvail(ctx, "SKU-1", 2); err != nil {
		t.Fatal(err)
	}
	if res, _ := p.Reserve(ctx, "SKU-1", 5, "req-B"); res != PreStockInsufficient {
		t.Fatalf("res = %d, want insufficient", res)
	}
	if n := hLen(t, p.rdb, logKey("SKU-1")); n != 0 {
		t.Fatalf("failed reserve must not write log, got %d entries", n)
	}
}

func TestPreStockReserve_LazyInitFromDB(t *testing.T) {
	p, mock, mr := newPreStock(t)
	ctx := context.Background()

	// avail key 不存在 → 从 DB 读 available=50 惰性初始化后重试
	rows := sqlmock.NewRows([]string{"available"}).AddRow(50)
	mock.ExpectQuery(`SELECT available FROM inventory WHERE sku_id = \?`).WithArgs("SKU-2").WillReturnRows(rows)

	res, err := p.Reserve(ctx, "SKU-2", 20, "req-C")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res != PreStockGranted {
		t.Fatalf("res = %d, want granted", res)
	}
	if avail, _ := mr.Get(availKey("SKU-2")); avail != "30" {
		t.Fatalf("avail = %s, want 30", avail)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPreStockReserve_LazyInitMissingSKU(t *testing.T) {
	p, mock, _ := newPreStock(t)

	mock.ExpectQuery(`SELECT available FROM inventory WHERE sku_id = \?`).WithArgs("SKU-404").
		WillReturnError(sql.ErrNoRows)

	if _, err := p.Reserve(context.Background(), "SKU-404", 1, "req-D"); !errors.Is(err, ErrSkuNotFound) {
		t.Fatalf("err = %v, want ErrSkuNotFound", err)
	}
}

func TestPreStockRestore(t *testing.T) {
	p, _, mr := newPreStock(t)
	ctx := context.Background()

	if err := p.InitAvail(ctx, "SKU-1", 10); err != nil {
		t.Fatal(err)
	}
	if res, _ := p.Reserve(ctx, "SKU-1", 4, "req-E"); res != PreStockGranted {
		t.Fatalf("reserve = %d", res)
	}
	// 未持有的 request_id 归还 → 幂等返回 false
	if restored, _ := p.Restore(ctx, "SKU-1", "req-NONE"); restored {
		t.Fatal("restoring unknown request should return false")
	}
	if restored, _ := p.Restore(ctx, "SKU-1", "req-E"); !restored {
		t.Fatal("restore should succeed")
	}
	if avail, _ := mr.Get(availKey("SKU-1")); avail != "10" {
		t.Fatalf("avail = %s, want 10 after restore", avail)
	}
	if n := hLen(t, p.rdb, logKey("SKU-1")); n != 0 {
		t.Fatalf("log entries = %d, want 0 after restore", n)
	}
	// 重复归还幂等
	if restored, _ := p.Restore(ctx, "SKU-1", "req-E"); restored {
		t.Fatal("double restore must be idempotent no-op")
	}
	if avail, _ := mr.Get(availKey("SKU-1")); avail != "10" {
		t.Fatalf("avail = %s, want 10 (double restore must not inflate)", avail)
	}
}

// ---- R1-2 stock-sync worker ----

func TestStockSync_PersistsAndClears(t *testing.T) {
	p, mock, _ := newPreStock(t)
	ctx := context.Background()
	syncer := NewStockSyncer(p.db, p.rdb)

	if err := p.InitAvail(ctx, "SKU-1", 100); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		if _, err := p.Reserve(ctx, "SKU-1", 2, fmt.Sprintf("req-%d", i)); err != nil {
			t.Fatal(err)
		}
	}

	// 三条流水各自：journal 落库 + available 扣减
	for i := 0; i < 3; i++ {
		mock.ExpectBegin()
		mock.ExpectExec(`INSERT IGNORE INTO stock_journal`).
			WithArgs("SKU-1", fmt.Sprintf("req-%d", i), 2).
			WillReturnResult(sqlmock.NewResult(int64(i+1), 1))
		mock.ExpectExec(`UPDATE inventory SET available = available - \?`).
			WithArgs(2, "SKU-1").
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()
	}

	n, err := syncer.SyncOnce(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 3 {
		t.Fatalf("synced = %d, want 3", n)
	}
	if h := hLen(t, p.rdb, logKey("SKU-1")); h != 0 {
		t.Fatalf("redis log should be cleared after sync, got %d", h)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestStockSync_DuplicateDeliveryAbsorbed(t *testing.T) {
	p, mock, mr := newPreStock(t)
	ctx := context.Background()
	syncer := NewStockSyncer(p.db, p.rdb)

	mr.Set(availKey("SKU-1"), "98")
	mr.HSet(logKey("SKU-1"), "req-X", "2|"+fmt.Sprint(time.Now().Unix()))

	// 第一次：正常落库扣减
	mock.ExpectBegin()
	mock.ExpectExec(`INSERT IGNORE INTO stock_journal`).WithArgs("SKU-1", "req-X", 2).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`UPDATE inventory SET available = available - \?`).WithArgs(2, "SKU-1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if n, err := syncer.SyncOnce(ctx); err != nil || n != 1 {
		t.Fatalf("first sync: n=%d err=%v", n, err)
	}

	// 模拟"发送成功未标记"：手工把流水塞回 Redis（重投）
	mr.HSet(logKey("SKU-1"), "req-X", "2|"+fmt.Sprint(time.Now().Unix()))
	mock.ExpectBegin()
	// 撞唯一键：affected=0 → 不得再次扣减 inventory
	mock.ExpectExec(`INSERT IGNORE INTO stock_journal`).WithArgs("SKU-1", "req-X", 2).
		WillReturnResult(sqlmock.NewResult(1, 0))
	mock.ExpectRollback()

	if _, err := syncer.SyncOnce(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

// ---- R1-3 对账 ----

func TestReconcile_DetectsUnsyncedEntry(t *testing.T) {
	p, mock, mr := newPreStock(t)
	ctx := context.Background()
	rec := NewReconciler(p.db, p.rdb)

	mr.Set(availKey("SKU-1"), "99")
	// 老于宽限期的未落库流水 → 应捕获差异
	old := time.Now().Add(-2 * reconcileGrace).Unix()
	mr.HSet(logKey("SKU-1"), "req-OLD", "1|"+fmt.Sprint(old))
	// 宽限期内的流水 → 不报也不查库
	mr.HSet(logKey("SKU-1"), "req-NEW", "1|"+fmt.Sprint(time.Now().Unix()))

	mock.ExpectQuery(`SELECT COUNT\(1\) FROM stock_journal`).WithArgs("SKU-1", "req-OLD").
		WillReturnRows(sqlmock.NewRows([]string{"c"}).AddRow(0))
	mock.ExpectExec(`INSERT IGNORE INTO reconcile_diff`).
		WithArgs("SKU-1", "req-OLD", "REDIS_NOT_SYNCED", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	report, err := rec.RunOnce(ctx, time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.Checked != 2 || report.Diffs != 1 {
		t.Fatalf("report = %+v, want checked=2 diffs=1", report)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestReconcile_CleanWhenAllSynced(t *testing.T) {
	p, _, _ := newPreStock(t)
	ctx := context.Background()
	rec := NewReconciler(p.db, p.rdb)

	// 无任何未落库流水 → 零差异，不触 DB
	report, err := rec.RunOnce(ctx, time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.Diffs != 0 {
		t.Fatalf("diffs = %d, want 0", report.Diffs)
	}
}

func TestReconcile_ResolveDiff(t *testing.T) {
	p, mock, _ := newPreStock(t)

	mock.ExpectExec(`UPDATE reconcile_diff SET status = 'RESOLVED'`).
		WithArgs(int64(7)).WillReturnResult(sqlmock.NewResult(0, 1))

	rec := NewReconciler(p.db, p.rdb)
	if err := rec.ResolveDiff(context.Background(), 7); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
