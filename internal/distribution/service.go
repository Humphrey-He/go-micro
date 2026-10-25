package distribution

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var ErrNoMapping = errors.New("no active channel mapping for sku")

// Service 分发入口：按渠道映射把一次分发请求展开为 N 个平台任务。
type Service struct {
	db *sqlx.DB
}

func NewService(dbx *sqlx.DB) *Service {
	return &Service{db: dbx}
}

// Dispatch 把 sku 分发到所有已映射平台。
// client_msg_id = batchID:platform:site：补发复用同一 ID，平台侧幂等吸收。
func (s *Service) Dispatch(ctx context.Context, skuID string) (*DispatchBatch, error) {
	var mappings []ChannelMapping
	if err := s.db.SelectContext(ctx, &mappings,
		`SELECT id, sku_id, platform, site, shop_id, rate_limit_qps, status
		 FROM channel_mappings WHERE sku_id = ? AND status = 'ACTIVE'`, skuID); err != nil {
		return nil, err
	}
	if len(mappings) == 0 {
		return nil, ErrNoMapping
	}

	batchID := uuid.NewString()[:8]
	res := &DispatchBatch{BatchID: batchID, SkuID: skuID, Tasks: make([]DispatchTask, 0, len(mappings))}
	for _, m := range mappings {
		msgID := fmt.Sprintf("%s:%s:%s", batchID, m.Platform, m.Site)
		r, err := s.db.ExecContext(ctx, `
			INSERT INTO dispatch_tasks (client_msg_id, batch_id, sku_id, platform, site, shop_id, status)
			VALUES (?, ?, ?, ?, ?, ?, ?)`,
			msgID, batchID, skuID, m.Platform, m.Site, m.ShopID, TaskQueued)
		if err != nil {
			return res, err
		}
		id, _ := r.LastInsertId()
		res.Tasks = append(res.Tasks, DispatchTask{
			ID: id, ClientMsgID: msgID, BatchID: batchID, SkuID: skuID,
			Platform: m.Platform, Site: m.Site, ShopID: m.ShopID, Status: TaskQueued,
		})
	}
	return res, nil
}

type DispatchBatch struct {
	BatchID string         `json:"batch_id"`
	SkuID   string         `json:"sku_id"`
	Tasks   []DispatchTask `json:"tasks"`
}

// CountByStatus 演示/验收用：各状态任务数。
func (s *Service) CountByStatus(ctx context.Context) (map[string]int, error) {
	rows := []struct {
		Status string `db:"status"`
		N      int    `db:"n"`
	}{}
	if err := s.db.SelectContext(ctx, &rows,
		`SELECT status, COUNT(*) n FROM dispatch_tasks GROUP BY status`); err != nil {
		if err == sql.ErrNoRows {
			return map[string]int{}, nil
		}
		return nil, err
	}
	out := make(map[string]int, len(rows))
	for _, r := range rows {
		out[r.Status] = r.N
	}
	return out, nil
}
