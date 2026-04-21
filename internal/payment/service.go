package payment

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var (
	ErrNotFound      = errors.New("payment not found")
	ErrInvalidState  = errors.New("invalid payment state")
	ErrIdempotentHit = errors.New("idempotent hit")
)

const (
	statusCreated = "CREATED"
	statusSuccess = "SUCCESS"
	statusFailed  = "FAILED"
	statusTimeout = "TIMEOUT"
	statusRefund  = "REFUNDED"
)

type OrderCanceler interface {
	Cancel(ctx context.Context, orderID string) error
}

type InventoryReleaser interface {
	ReleaseByOrder(ctx context.Context, orderID string) error
}

type Service struct {
	db    *sqlx.DB
	order OrderCanceler
	inv   InventoryReleaser
}

func NewService(dbx *sqlx.DB, order OrderCanceler, inv InventoryReleaser) *Service {
	return &Service{db: dbx, order: order, inv: inv}
}

func (s *Service) Create(req CreatePaymentRequest) (*Payment, error) {
	if req.OrderID == "" || req.RequestID == "" || req.Amount <= 0 {
		return nil, errors.New("invalid request")
	}
	payID := uuid.NewString()
	_, err := s.db.Exec(`INSERT INTO payments(payment_id,order_id,amount,status,request_id,created_at,updated_at) VALUES(?,?,?,?,?,NOW(),NOW())`,
		payID, req.OrderID, req.Amount, statusCreated, req.RequestID)
	if err != nil {
		if isDuplicate(err) {
			p, err := s.getByRequestID(req.RequestID)
			if err == nil {
				return p, ErrIdempotentHit
			}
		}
		return nil, err
	}
	return &Payment{PaymentID: payID, OrderID: req.OrderID, Amount: req.Amount, Status: statusCreated, RequestID: req.RequestID, CreatedAt: time.Now(), UpdatedAt: time.Now()}, nil
}

func (s *Service) Get(paymentID string) (*Payment, error) {
	if paymentID == "" {
		return nil, ErrNotFound
	}
	p := Payment{}
	if err := s.db.Get(&p, `SELECT * FROM payments WHERE payment_id = ?`, paymentID); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (s *Service) GetByOrderID(orderID string) (*Payment, error) {
	if orderID == "" {
		return nil, ErrNotFound
	}
	p := Payment{}
	if err := s.db.Get(&p, `SELECT * FROM payments WHERE order_id = ? LIMIT 1`, orderID); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (s *Service) MarkSuccess(paymentID string) error {
	return s.updateStatus(paymentID, statusCreated, statusSuccess)
}

func (s *Service) MarkFailed(paymentID string) error {
	if err := s.updateStatus(paymentID, statusCreated, statusFailed); err != nil {
		return err
	}
	return s.compensate(paymentID, "failed")
}

func (s *Service) MarkTimeout(paymentID string) error {
	if err := s.updateStatus(paymentID, statusCreated, statusTimeout); err != nil {
		return err
	}
	return s.compensate(paymentID, "timeout")
}

func (s *Service) Refund(paymentID string) error {
	return s.updateStatus(paymentID, statusSuccess, statusRefund)
}

func (s *Service) compensate(paymentID, reason string) error {
	p, err := s.Get(paymentID)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if s.order != nil {
		_ = s.order.Cancel(ctx, p.OrderID)
	}
	if s.inv != nil {
		_ = s.inv.ReleaseByOrder(ctx, p.OrderID)
	}
	return nil
}

func (s *Service) updateStatus(paymentID, from, to string) error {
	if paymentID == "" {
		return ErrNotFound
	}
	var status string
	if err := s.db.QueryRow(`SELECT status FROM payments WHERE payment_id = ?`, paymentID).Scan(&status); err != nil {
		if err == sql.ErrNoRows {
			return ErrNotFound
		}
		return err
	}
	if status == to {
		return nil
	}
	if status != from {
		return ErrInvalidState
	}
	res, err := s.db.Exec(`UPDATE payments SET status=?, updated_at=NOW() WHERE payment_id=? AND status=?`, to, paymentID, from)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return ErrInvalidState
	}
	return nil
}

func (s *Service) ListTimeoutPayments(before time.Time, limit int) ([]Payment, error) {
	if limit <= 0 {
		limit = 20
	}
	var rows []Payment
	if err := s.db.Select(&rows, `SELECT * FROM payments WHERE status = ? AND created_at <= ? ORDER BY created_at ASC LIMIT ?`, statusCreated, before, limit); err != nil {
		return nil, err
	}
	return rows, nil
}

func (s *Service) getByRequestID(requestID string) (*Payment, error) {
	p := Payment{}
	if err := s.db.Get(&p, `SELECT * FROM payments WHERE request_id = ?`, requestID); err != nil {
		return nil, err
	}
	return &p, nil
}

func isDuplicate(err error) bool {
	var me *mysql.MySQLError
	if errors.As(err, &me) {
		return me.Number == 1062
	}
	return false
}
