package distribution

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	"golang.org/x/time/rate"

	"ecomm-micro/pkg/logx"
	"go.uber.org/zap"
)

// Scheduler 按店铺令牌桶削峰的投递调度器：
// 每个店铺一个 x/time/rate.Limiter，拿不到令牌的任务留在 QUEUED（排队，不失败重试），
// 平台 429 也回 QUEUED——「限频是硬约束，超限封号不是报错」。
// Store 调度器对持久层的窄依赖，*sqlx.DB 直接满足；测试用内存实现。
type Store interface {
	ListQueued(ctx context.Context, limit int) ([]DispatchTask, error)
	MarkSent(ctx context.Context, id int64, platformTaskID string) error
	MarkError(ctx context.Context, id int64, msg string) error
}

type Scheduler struct {
	store     Store
	endpoints map[string]string // platform -> base url（如 tmall=http://mock-platform-a:9091）
	client    *http.Client
	shopQPS   float64

	mu       sync.Mutex
	limiters map[string]*rate.Limiter
}

func NewScheduler(dbx *sqlx.DB, endpoints map[string]string, shopQPS float64) *Scheduler {
	return NewSchedulerWithStore(NewService(dbx), endpoints, shopQPS)
}

func NewSchedulerWithStore(store Store, endpoints map[string]string, shopQPS float64) *Scheduler {
	return &Scheduler{
		store:     store,
		endpoints: endpoints,
		client:    &http.Client{Timeout: 3 * time.Second},
		shopQPS:   shopQPS,
		limiters:  make(map[string]*rate.Limiter),
	}
}

// ---- *sqlx.DB 的 Store 实现 ----

func (s *Service) ListQueued(ctx context.Context, limit int) ([]DispatchTask, error) {
	var tasks []DispatchTask
	err := s.db.SelectContext(ctx, &tasks, `
		SELECT id, client_msg_id, batch_id, sku_id, platform, site, shop_id, status, redispatch_count
		FROM dispatch_tasks WHERE status = ? ORDER BY id ASC LIMIT ?`, TaskQueued, limit)
	return tasks, err
}

func (s *Service) MarkSent(ctx context.Context, id int64, platformTaskID string) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE dispatch_tasks SET status = ?, platform_task_id = ?, sent_at = NOW(), last_error = '', updated_at = NOW()
		WHERE id = ?`, TaskPendingAck, platformTaskID, id)
	return err
}

func (s *Service) MarkError(ctx context.Context, id int64, msg string) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE dispatch_tasks SET last_error = ?, updated_at = NOW() WHERE id = ?`, msg, id)
	return err
}

func (s *Scheduler) limiterFor(shopID string) *rate.Limiter {
	s.mu.Lock()
	defer s.mu.Unlock()
	if l, ok := s.limiters[shopID]; ok {
		return l
	}
	// burst = qps：允许一个瞬时批次，之后严格按速率放行
	l := rate.NewLimiter(rate.Limit(s.shopQPS), int(s.shopQPS))
	s.limiters[shopID] = l
	return l
}

// Tick 执行一轮投递。返回本轮发出/排队/被 429 的任务数（验收断言用）。
func (s *Scheduler) Tick(ctx context.Context) (sent, queued, rejected int, err error) {
	tasks, err := s.store.ListQueued(ctx, 500)
	if err != nil {
		return
	}

	byShop := make(map[string][]DispatchTask, len(tasks))
	for _, t := range tasks {
		byShop[t.ShopID] = append(byShop[t.ShopID], t)
	}

	for shop, group := range byShop {
		lim := s.limiterFor(shop)
		for _, t := range group {
			if !lim.Allow() {
				queued += len(group) - indexOfSent(group, t)
				break // 本店铺剩余任务继续排队
			}
			st, perr := s.sendOne(ctx, t)
			switch st {
			case http.StatusOK:
				sent++
			case http.StatusTooManyRequests:
				rejected++
				// 429：任务留在 QUEUED，下一轮再取
			default:
				queued++ // 网络错误等：仍回队列，last_error 已记录
			}
			if perr != nil {
				logx.L().Warn("dispatch send failed",
					zap.String("client_msg_id", t.ClientMsgID), zap.Error(perr))
			}
		}
	}
	return
}

func indexOfSent(group []DispatchTask, cur DispatchTask) int {
	for i, t := range group {
		if t.ID == cur.ID {
			return i
		}
	}
	return 0
}

// sendOne 投递单条：2xx → PENDING_ACK；429 → 留队列；其他 → 记错误回队列。
func (s *Scheduler) sendOne(ctx context.Context, t DispatchTask) (int, error) {
	base, ok := s.endpoints[t.Platform]
	if !ok {
		_ = s.store.MarkError(ctx, t.ID, "no endpoint for platform "+t.Platform)
		return 0, fmt.Errorf("no endpoint for platform %s", t.Platform)
	}
	body, _ := json.Marshal(map[string]string{
		"client_msg_id": t.ClientMsgID,
		"sku_id":        t.SkuID,
		"shop_id":       t.ShopID,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		base+"/open/api/listing", bytes.NewReader(body))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.client.Do(req)
	if err != nil {
		_ = s.store.MarkError(ctx, t.ID, err.Error())
		return 0, err
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	switch {
	case resp.StatusCode >= 200 && resp.StatusCode < 300:
		var out struct {
			Data struct {
				PlatformTaskID string `json:"platform_task_id"`
			} `json:"data"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&out)
		_ = s.store.MarkSent(ctx, t.ID, out.Data.PlatformTaskID)
		return resp.StatusCode, nil
	default:
		_ = s.store.MarkError(ctx, t.ID, fmt.Sprintf("platform http %d", resp.StatusCode))
		return resp.StatusCode, nil
	}
}

// StartScheduler 周期投递，直到 stop 关闭。
func StartScheduler(s *Scheduler, interval time.Duration, stop <-chan struct{}) {
	if interval <= 0 {
		interval = 500 * time.Millisecond
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			sent, _, rejected, err := s.Tick(ctx)
			cancel()
			if err != nil {
				logx.L().Warn("scheduler tick failed", zap.Error(err))
				continue
			}
			if sent > 0 || rejected > 0 {
				logx.L().Info("scheduler tick", zap.Int("sent", sent), zap.Int("rejected_429", rejected))
			}
		}
	}
}
