package activity

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"go-micro/pkg/cache"
)

var (
	ErrNotFound      = errors.New("activity not found")
	ErrInvalid       = errors.New("invalid request")
	ErrBusy          = errors.New("seckill busy")
	ErrOutOfStock    = errors.New("out of stock")
	ErrIdempotentHit = errors.New("idempotent hit")
)

type Service struct {
	db  *sqlx.DB
	rdb *redis.Client
}

func NewService(dbx *sqlx.DB, rdb *redis.Client) *Service {
	return &Service{db: dbx, rdb: rdb}
}

func (s *Service) IssueCoupon(req CouponRequest) (*Coupon, error) {
	if req.CouponID == "" || req.UserID == "" || req.Amount <= 0 {
		return nil, ErrInvalid
	}
	now := time.Now()
	_, err := s.db.Exec(`INSERT INTO activity_coupons(coupon_id,user_id,amount,status,issued_at,created_at,updated_at)
		VALUES(?,?,?,?,?,?,?)`, req.CouponID, req.UserID, req.Amount, CouponStatusIssued, now, now, now)
	if err != nil {
		if isDuplicate(err) {
			cp, err := s.GetCoupon(req.CouponID)
			if err == nil {
				return cp, ErrIdempotentHit
			}
		}
		return nil, err
	}
	return &Coupon{CouponID: req.CouponID, UserID: req.UserID, Amount: req.Amount, Status: CouponStatusIssued, IssuedAt: now, CreatedAt: now, UpdatedAt: now}, nil
}

func (s *Service) GetCoupon(couponID string) (*Coupon, error) {
	if couponID == "" {
		return nil, ErrNotFound
	}
	cp := Coupon{}
	if err := s.db.Get(&cp, `SELECT * FROM activity_coupons WHERE coupon_id = ?`, couponID); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &cp, nil
}

func (s *Service) Seckill(req SeckillRequest) (*SeckillOrder, error) {
	if req.SkuID == "" || req.UserID == "" {
		return nil, ErrInvalid
	}
	if req.Quantity <= 0 {
		req.Quantity = 1
	}
	lockKey := "lock:seckill:" + req.SkuID
	locked, err := cache.TryLock(context.Background(), s.rdb, lockKey, 3*time.Second)
	if err != nil || !locked {
		return nil, ErrBusy
	}
	defer func() { _ = cache.Unlock(context.Background(), s.rdb, lockKey) }()

	tx, err := s.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	var sk Seckill
	if err := tx.Get(&sk, `SELECT * FROM activity_seckill WHERE sku_id = ? FOR UPDATE`, req.SkuID); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	now := time.Now()
	if sk.Status != SeckillStatusActive || now.Before(sk.StartTime) || now.After(sk.EndTime) {
		return nil, ErrInvalid
	}

	var existing SeckillOrder
	if err := tx.Get(&existing, `SELECT * FROM seckill_orders WHERE sku_id = ? AND user_id = ?`, req.SkuID, req.UserID); err == nil {
		_ = tx.Commit()
		return &existing, ErrIdempotentHit
	}

	if sk.Stock-sk.Reserved < req.Quantity {
		return nil, ErrOutOfStock
	}

	_, err = tx.Exec(`UPDATE activity_seckill SET reserved = reserved + ?, updated_at = ? WHERE sku_id = ?`, req.Quantity, now, req.SkuID)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(`INSERT INTO seckill_orders(sku_id,user_id,quantity,status,created_at,updated_at) VALUES(?,?,?,?,?,?)`,
		req.SkuID, req.UserID, req.Quantity, OrderStatusReserved, now, now)
	if err != nil {
		if isDuplicate(err) {
			if err := tx.Get(&existing, `SELECT * FROM seckill_orders WHERE sku_id = ? AND user_id = ?`, req.SkuID, req.UserID); err == nil {
				_ = tx.Commit()
				return &existing, ErrIdempotentHit
			}
		}
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &SeckillOrder{SkuID: req.SkuID, UserID: req.UserID, Quantity: req.Quantity, Status: OrderStatusReserved, CreatedAt: now, UpdatedAt: now}, nil
}

func (s *Service) GetSeckill(skuID string) (*Seckill, error) {
	if skuID == "" {
		return nil, ErrNotFound
	}
	sk := Seckill{}
	if err := s.db.Get(&sk, `SELECT * FROM activity_seckill WHERE sku_id = ?`, skuID); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &sk, nil
}

func isDuplicate(err error) bool {
	var me *mysql.MySQLError
	if errors.As(err, &me) {
		return me.Number == 1062
	}
	return false
}
