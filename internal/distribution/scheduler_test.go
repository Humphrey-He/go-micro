package distribution

import (
	"context"
	"fmt"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	ecomm "ecomm-micro/internal/mockplatform"
)

func startMockPlatform(t *testing.T, name string, qps float64, loss float64) (*ecomm.Platform, *httptest.Server) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	p := ecomm.New(name, qps, loss)
	srv := httptest.NewServer(p.Router())
	t.Cleanup(srv.Close)
	return p, srv
}

func taskRows(tasks ...DispatchTask) *sqlmock.Rows {
	r := sqlmock.NewRows([]string{"id", "client_msg_id", "batch_id", "sku_id", "platform", "site", "shop_id",
		"status", "redispatch_count", "sent_at"})
	for _, t := range tasks {
		r.AddRow(t.ID, t.ClientMsgID, t.BatchID, t.SkuID, t.Platform, t.Site, t.ShopID,
			t.Status, t.RedispatchCount, t.SentAt)
	}
	return r
}

// 验收 1：店铺令牌桶低于平台配额 → 100 任务全部送出、平台零 429。
// 店铺桶 burst=200 一次放行整批（burst 覆盖批次是生产合理配置）。
func TestScheduler_Zero429UnderBurst(t *testing.T) {
	mockPlat, srv := startMockPlatform(t, "tmall", 1000, 0)
	store := newMemStore()
	const n = 100
	for i := 0; i < n; i++ {
		store.add(DispatchTask{ID: int64(i + 1), ClientMsgID: fmt.Sprintf("m-%d", i), BatchID: "b1",
			SkuID: "SKU-1", Platform: "tmall", Site: "CN", ShopID: "shop-1", Status: TaskQueued})
	}
	sched := NewSchedulerWithStore(store, map[string]string{"tmall": srv.URL}, 200) // 店铺 200 QPS << 平台 1000

	sent, queued, rejected, err := sched.Tick(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if sent != n || queued != 0 || rejected != 0 {
		t.Fatalf("tick = sent %d queued %d rejected %d, want %d/0/0", sent, queued, rejected, n)
	}
	stats := platformStats(t, mockPlat)
	if stats.Received != n || stats.Rejected429 != 0 {
		t.Fatalf("platform stats = %+v, want received=%d rejected=0", stats, n)
	}
	if store.countByStatus(TaskPendingAck) != n {
		t.Fatalf("pending_ack = %d, want %d", store.countByStatus(TaskPendingAck), n)
	}
}

// 验收 2：店铺限速高于平台配额（配置错误）→ 平台 429 → 任务回队列不丢失，
// 等令牌回填后继续送出（削峰吸收，绝不 hard-fail）。
func TestScheduler_Absorbs429(t *testing.T) {
	mockPlat, srv := startMockPlatform(t, "jd", 5, 0) // 平台只给 5 QPS
	store := newMemStore()
	const n = 12
	for i := 0; i < n; i++ {
		store.add(DispatchTask{ID: int64(i + 1), ClientMsgID: fmt.Sprintf("m-%d", i), BatchID: "b1",
			SkuID: "SKU-1", Platform: "jd", Site: "CN", ShopID: "shop-1", Status: TaskQueued})
	}
	sched := NewSchedulerWithStore(store, map[string]string{"jd": srv.URL}, 200) // 店铺桶远超平台配额

	ctx := context.Background()
	saw429 := false
	for round := 0; store.countByStatus(TaskQueued) > 0 && round < 20; round++ {
		_, _, rejected, err := sched.Tick(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if rejected > 0 {
			saw429 = true
		}
		time.Sleep(300 * time.Millisecond) // 等平台 5 QPS 慢慢消化 + 店铺桶回填
	}
	if !saw429 {
		t.Fatal("expected at least one 429 with misconfigured limiter")
	}
	stats := platformStats(t, mockPlat)
	if stats.Received != n {
		t.Fatalf("platform received = %d, want %d (all tasks eventually delivered)", stats.Received, n)
	}
	if store.countByStatus(TaskQueued) != 0 {
		t.Fatalf("queue not drained: %d left", store.countByStatus(TaskQueued))
	}
}

// ---- 内存 Store ----

type memStore struct {
	mu    sync.Mutex
	tasks map[int64]DispatchTask
}

func newMemStore() *memStore { return &memStore{tasks: map[int64]DispatchTask{}} }

func (m *memStore) add(t DispatchTask) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tasks[t.ID] = t
}

func (m *memStore) countByStatus(status string) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	n := 0
	for _, t := range m.tasks {
		if t.Status == status {
			n++
		}
	}
	return n
}

func (m *memStore) ListQueued(_ context.Context, limit int) ([]DispatchTask, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]DispatchTask, 0, limit)
	for _, t := range m.tasks {
		if t.Status == TaskQueued {
			out = append(out, t)
		}
	}
	// 按 id 稳定排序，模拟 SQL ORDER BY
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j].ID < out[j-1].ID; j-- {
			out[j], out[j-1] = out[j-1], out[j]
		}
	}
	if len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func (m *memStore) MarkSent(_ context.Context, id int64, platformTaskID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	t := m.tasks[id]
	t.Status = TaskPendingAck
	t.PlatformTaskID = platformTaskID
	m.tasks[id] = t
	return nil
}

func (m *memStore) MarkError(_ context.Context, id int64, msg string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	t := m.tasks[id]
	t.LastError = msg
	m.tasks[id] = t
	return nil
}

// 验收 3：回执丢失 → diff 抓取（404）→ 台账 + 同 client_msg_id 补发 →
// 重投唤醒回执 → 下一轮 diff 确认 CONFIRMED。完整闭环。
func TestReconcile_CatchesMissingAndRedispatches(t *testing.T) {
	mockPlat, srv := startMockPlatform(t, "tiktok", 1000, 0)
	db, sqlm, _ := sqlmock.New()
	defer db.Close()
	dbx := sqlx.NewDb(db, "mysql")

	// 先真实投递一条
	sched := NewScheduler(dbx, map[string]string{"tiktok": srv.URL}, 100)
	sqlm.ExpectQuery(`FROM dispatch_tasks WHERE status = \?`).WillReturnRows(
		taskRows(DispatchTask{ID: 1, ClientMsgID: "m-lost", SkuID: "SKU-9", Platform: "tiktok", Site: "GLOBAL", ShopID: "shop-1", Status: TaskQueued}))
	sqlm.ExpectExec(`UPDATE dispatch_tasks SET status = \?`).
		WithArgs(TaskPendingAck, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	if sent, _, _, err := sched.Tick(context.Background()); err != nil || sent != 1 {
		t.Fatalf("send: sent=%d err=%v", sent, err)
	}

	// 平台侧把这条标记为回执丢失
	mockPlat.SetLost("m-lost", true)

	recon := NewReconciler(dbx, map[string]string{"tiktok": srv.URL}, 0)
	pendingRow := func() *sqlmock.Rows {
		return taskRows(DispatchTask{ID: 1, ClientMsgID: "m-lost", SkuID: "SKU-9",
			Platform: "tiktok", Site: "GLOBAL", ShopID: "shop-1", Status: TaskPendingAck})
	}

	// 轮 1：回执 404 → 台账 RECEIPT_MISSING + 重新入队
	sqlm.ExpectQuery(`FROM dispatch_tasks`).WillReturnRows(pendingRow())
	sqlm.ExpectExec(`INSERT IGNORE INTO reconcile_diff`).
		WithArgs(LedgerBizType, "SKU-9", "m-lost", LedgerDiffType, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	sqlm.ExpectExec(`UPDATE dispatch_tasks SET status = \?, redispatch_count`).
		WithArgs(TaskQueued, int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	_, missing, _, err := recon.Tick(context.Background())
	if err != nil || missing != 1 {
		t.Fatalf("round1: missing=%d err=%v", missing, err)
	}

	// 调度器把补发任务再投出去（同 client_msg_id）
	sqlm.ExpectQuery(`FROM dispatch_tasks WHERE status = \?`).WillReturnRows(
		taskRows(DispatchTask{ID: 1, ClientMsgID: "m-lost", SkuID: "SKU-9", Platform: "tiktok", Site: "GLOBAL", ShopID: "shop-1", Status: TaskQueued, RedispatchCount: 1}))
	sqlm.ExpectExec(`UPDATE dispatch_tasks SET status = \?`).
		WithArgs(TaskPendingAck, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	if sent, _, _, err := sched.Tick(context.Background()); err != nil || sent != 1 {
		t.Fatalf("redispatch: sent=%d err=%v", sent, err)
	}

	// 轮 2：重投唤醒回执 → 200 → CONFIRMED
	sqlm.ExpectQuery(`FROM dispatch_tasks`).WillReturnRows(pendingRow())
	sqlm.ExpectExec(`UPDATE dispatch_tasks SET status = \?, updated_at`).
		WithArgs(TaskConfirmed, int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	_, missing2, confirmed, err := recon.Tick(context.Background())
	if err != nil || missing2 != 0 || confirmed != 1 {
		t.Fatalf("round2: missing=%d confirmed=%d err=%v", missing2, confirmed, err)
	}

	stats := platformStats(t, mockPlat)
	if stats.Duplicates != 1 {
		t.Fatalf("platform duplicates = %d, want 1 (idempotent redispatch)", stats.Duplicates)
	}
	if err := sqlm.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func platformStats(t *testing.T, p *ecomm.Platform) ecomm.Stats {
	t.Helper()
	return p.StatsSnapshot()
}
