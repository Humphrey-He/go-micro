package reconcile

import (
	"context"
	"errors"
	"fmt"
	"time"

	"ecomm-micro/pkg/logx"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// 差异类型（订单 vs 库存侧）
const (
	// DiffOrderNoReservation 订单声称占用库存（RESERVED/PROCESSING/SUCCESS），
	// 但 inventory_reserved 无该订单的预占记录
	DiffOrderNoReservation = "ORDER_NO_RESERVATION"
	// DiffReservationOrphan 库存侧仍持有预占（status=RESERVED），
	// 但订单已终结（CANCELED/FAILED）或不存在
	DiffReservationOrphan = "RESERVATION_ORPHAN"
)

// 台账状态机：PENDING → RESOLVED | IGNORED（终态不可逆）
const (
	DiffPending  = "PENDING"
	DiffResolved = "RESOLVED"
	DiffIgnored  = "IGNORED"
)

var ErrInvalidTransition = errors.New("diff is not in PENDING state")

// Report 单轮对账报告
type Report struct {
	OrdersChecked int `json:"orders_checked"`
	OrphansFound  int `json:"orphans_found"`
	NewDiffs      int `json:"new_diffs"`
}

// Service 订单侧 vs 库存侧流水对账（R4）。
// 与 R1-3 的 Redis-vs-journal 对账互补：那边看「加速层 vs 真账」，
// 这边看「业务单据 vs 库存凭据」，差异统一落 reconcile_diff 台账。
type Service struct {
	db *sqlx.DB
}

func NewService(dbx *sqlx.DB) *Service {
	return &Service{db: dbx}
}

// RunOnce 执行一轮订单/库存双向 diff。
func (s *Service) RunOnce(ctx context.Context) (*Report, error) {
	report := &Report{}

	// 方向 A：订单占用库存但库存侧无预占凭据
	var missing []string
	if err := s.db.SelectContext(ctx, &missing, `
		SELECT o.order_id FROM orders o
		WHERE o.status IN ('RESERVED','PROCESSING','SUCCESS')
		  AND NOT EXISTS (SELECT 1 FROM inventory_reserved r WHERE r.order_id = o.order_id)
		LIMIT 100`); err != nil {
		return report, err
	}
	report.OrdersChecked = len(missing)
	for _, orderID := range missing {
		detail := fmt.Sprintf("order %s holds stock-claiming status but no inventory_reserved row exists", orderID)
		if err := s.recordDiff(ctx, orderID, DiffOrderNoReservation, detail); err != nil {
			return report, err
		}
		report.NewDiffs++
	}

	// 方向 B：库存侧仍持有预占但订单已终结/不存在（幽灵占用）
	var orphans []string
	if err := s.db.SelectContext(ctx, &orphans, `
		SELECT r.order_id FROM inventory_reserved r
		WHERE r.status = 'RESERVED'
		  AND (NOT EXISTS (SELECT 1 FROM orders o WHERE o.order_id = r.order_id)
		       OR EXISTS (SELECT 1 FROM orders o WHERE o.order_id = r.order_id AND o.status IN ('CANCELED','FAILED')))
		LIMIT 100`); err != nil {
		return report, err
	}
	report.OrphansFound = len(orphans)
	for _, orderID := range orphans {
		detail := fmt.Sprintf("inventory reservation held for order %s but order is terminal or missing", orderID)
		if err := s.recordDiff(ctx, orderID, DiffReservationOrphan, detail); err != nil {
			return report, err
		}
		report.NewDiffs++
	}

	if report.NewDiffs > 0 {
		logx.L().Warn("reconcile found order/inventory diffs",
			zap.Int("no_reservation", report.OrdersChecked),
			zap.Int("orphans", report.OrphansFound),
			zap.Int("new_diffs", report.NewDiffs))
	}
	return report, nil
}

// Diff 台账条目
type Diff struct {
	ID         int64      `db:"id" json:"id"`
	BizType    string     `db:"biz_type" json:"biz_type"`
	SkuID      string     `db:"sku_id" json:"sku_id"`
	RequestID  string     `db:"request_id" json:"request_id"`
	DiffType   string     `db:"diff_type" json:"diff_type"`
	Detail     string     `db:"detail" json:"detail"`
	Status     string     `db:"status" json:"status"`
	CreatedAt  time.Time  `db:"created_at" json:"created_at"`
	ResolvedAt *time.Time `db:"resolved_at" json:"resolved_at,omitempty"`
}

// ListDiffs 台账查询（status 为空返回全部）
func (s *Service) ListDiffs(ctx context.Context, status string, limit int) ([]Diff, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	var diffs []Diff
	query := `SELECT id, biz_type, sku_id, request_id, diff_type, detail, status, created_at, resolved_at
		FROM reconcile_diff WHERE biz_type = 'ORDER_INV'`
	args := []interface{}{}
	if status != "" {
		query += " AND status = ?"
		args = append(args, status)
	}
	query += " ORDER BY id DESC LIMIT ?"
	args = append(args, limit)
	if err := s.db.SelectContext(ctx, &diffs, query, args...); err != nil {
		return nil, err
	}
	return diffs, nil
}

// Resolve 关闭差异（PENDING → RESOLVED，非法流转返回 ErrInvalidTransition）
func (s *Service) Resolve(ctx context.Context, id int64) error {
	return s.transition(ctx, id, DiffResolved)
}

// Ignore 忽略差异（PENDING → IGNORED）
func (s *Service) Ignore(ctx context.Context, id int64) error {
	return s.transition(ctx, id, DiffIgnored)
}

func (s *Service) transition(ctx context.Context, id int64, to string) error {
	res, err := s.db.ExecContext(ctx,
		`UPDATE reconcile_diff SET status = ?, resolved_at = NOW() WHERE id = ? AND status = ?`,
		to, id, DiffPending)
	if err != nil {
		return err
	}
	if affected, _ := res.RowsAffected(); affected == 0 {
		return ErrInvalidTransition
	}
	return nil
}

func (s *Service) recordDiff(ctx context.Context, orderID, diffType, detail string) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT IGNORE INTO reconcile_diff (biz_type, sku_id, request_id, diff_type, detail)
		 VALUES ('ORDER_INV', '', ?, ?, ?)`,
		orderID, diffType, detail)
	return err
}
