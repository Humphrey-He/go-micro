package refund

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var (
	ErrNotFound      = errors.New("refund not found")
	ErrInvalidState  = errors.New("invalid refund state")
	ErrIdempotentHit = errors.New("idempotent hit")
)

const (
	TypeManual  = "manual"
	TypeFailure = "payment_failed"
	TypeCancel  = "order_cancel"
)

type OrderCanceler interface {
	Cancel(ctx context.Context, orderID string) error
}

type InventoryReleaser interface {
	ReleaseByOrder(ctx context.Context, orderID string) error
}

type Publisher interface {
	Publish(ctx context.Context, body []byte) error
}

type Service struct {
	db  *sqlx.DB
	mq  Publisher
	ord OrderCanceler
	inv InventoryReleaser
}

func NewService(dbx *sqlx.DB, mq Publisher, ord OrderCanceler, inv InventoryReleaser) *Service {
	return &Service{db: dbx, mq: mq, ord: ord, inv: inv}
}

func (s *Service) Initiate(req InitiateRequest) (*Refund, error) {
	if req.RefundID == "" || req.OrderID == "" {
		return nil, errors.New("invalid request")
	}
	if req.RefundType == "" {
		req.RefundType = TypeManual
	}
	_, err := s.db.Exec(`INSERT INTO refunds(refund_id,order_id,refund_type,status,reason,created_at,updated_at)
		VALUES(?,?,?,?,?,NOW(),NOW())`, req.RefundID, req.OrderID, req.RefundType, StatusPending, req.Reason)
	if err != nil {
		if isDuplicate(err) {
			ref, err := s.Get(req.RefundID)
			if err == nil {
				return ref, ErrIdempotentHit
			}
		}
		return nil, err
	}
	ref := &Refund{RefundID: req.RefundID, OrderID: req.OrderID, RefundType: req.RefundType, Status: StatusPending, Reason: req.Reason, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	_ = s.publish(ref)
	return ref, nil
}

func (s *Service) Rollback(req RollbackRequest) (*Refund, error) {
	initReq := InitiateRequest{RefundID: req.RefundID, OrderID: req.OrderID, RefundType: TypeCancel, Reason: req.Reason}
	return s.Initiate(initReq)
}

func (s *Service) Get(refundID string) (*Refund, error) {
	if refundID == "" {
		return nil, ErrNotFound
	}
	ref := Refund{}
	if err := s.db.Get(&ref, `SELECT * FROM refunds WHERE refund_id = ?`, refundID); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &ref, nil
}

func (s *Service) Process(refundID string) error {
	ref, err := s.Get(refundID)
	if err != nil {
		return err
	}
	if ref.Status == StatusSuccess {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if s.ord != nil {
		if err := s.ord.Cancel(ctx, ref.OrderID); err != nil {
			return s.markFailed(ref, err)
		}
	}
	if s.inv != nil {
		if err := s.inv.ReleaseByOrder(ctx, ref.OrderID); err != nil {
			return s.markFailed(ref, err)
		}
	}
	return s.markSuccess(ref.RefundID)
}

func (s *Service) markFailed(ref *Refund, err error) error {
	retry := ref.RetryCount + 1
	next := time.Now().Add(retryDelay(retry))
	_, execErr := s.db.Exec(`UPDATE refunds SET status=?, retry_count=?, next_retry_time=?, last_error=?, updated_at=NOW() WHERE refund_id=?`,
		StatusFailed, retry, next, truncateErr(err), ref.RefundID)
	if execErr != nil {
		return execErr
	}
	return err
}

func (s *Service) markSuccess(refundID string) error {
	res, err := s.db.Exec(`UPDATE refunds SET status=?, updated_at=NOW() WHERE refund_id=? AND status <> ?`, StatusSuccess, refundID, StatusSuccess)
	if err != nil {
		return err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return ErrInvalidState
	}
	return nil
}

func (s *Service) ListRetryDue(limit int) ([]Refund, error) {
	if limit <= 0 {
		limit = 20
	}
	rows := []Refund{}
	if err := s.db.Select(&rows, `SELECT * FROM refunds WHERE status = ? AND next_retry_time <= NOW() ORDER BY next_retry_time ASC LIMIT ?`, StatusFailed, limit); err != nil {
		return nil, err
	}
	return rows, nil
}

func (s *Service) publish(ref *Refund) error {
	if s.mq == nil || ref == nil {
		return nil
	}
	body, _ := json.Marshal(map[string]string{"refund_id": ref.RefundID, "order_id": ref.OrderID, "refund_type": ref.RefundType})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return s.mq.Publish(ctx, body)
}

func retryDelay(retry int) time.Duration {
	switch retry {
	case 1:
		return time.Minute
	case 2:
		return 5 * time.Minute
	default:
		return 15 * time.Minute
	}
}

func truncateErr(err error) string {
	if err == nil {
		return ""
	}
	msg := err.Error()
	if len(msg) > 240 {
		return msg[:240]
	}
	return msg
}

func isDuplicate(err error) bool {
	var me *mysql.MySQLError
	if errors.As(err, &me) {
		return me.Number == 1062
	}
	return false
}
