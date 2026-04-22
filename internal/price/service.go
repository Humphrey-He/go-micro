package price

import (
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"
)

var (
	ErrInvalid  = errors.New("invalid request")
	ErrNotFound = errors.New("price history not found")
)

type Service struct {
	db *sqlx.DB
}

func NewService(dbx *sqlx.DB) *Service {
	return &Service{db: dbx}
}

func (s *Service) Calculate(req CalculateRequest) (*CalculateResponse, error) {
	if req.SkuID == "" || req.BasePrice.LessThanOrEqual(decimal.Zero) {
		return nil, ErrInvalid
	}

	price := req.BasePrice.Sub(req.CouponAmount)
	if price.LessThan(decimal.Zero) {
		price = decimal.Zero
	}

	discount := decimal.NewFromFloat(1.0)
	reason := "none"
	if req.UserLevel >= 3 {
		discount = decimal.NewFromFloat(0.9)
		reason = "vip-level-3"
	} else if req.UserLevel >= 2 {
		discount = decimal.NewFromFloat(0.95)
		reason = "vip-level-2"
	}

	final := price.Mul(discount)
	resp := &CalculateResponse{
		SkuID:          req.SkuID,
		BasePrice:      req.BasePrice,
		CouponAmount:   req.CouponAmount,
		LevelDiscount:  discount,
		FinalPrice:     final,
		DiscountReason: reason,
	}
	if err := s.recordHistory(req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *Service) History(skuID string, limit int) ([]History, error) {
	if skuID == "" {
		return nil, ErrInvalid
	}
	if limit <= 0 {
		limit = 20
	}
	rows := []History{}
	if err := s.db.Select(&rows, `SELECT * FROM price_history WHERE sku_id = ? ORDER BY created_at DESC LIMIT ?`, skuID, limit); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return rows, nil
}

func (s *Service) recordHistory(req CalculateRequest, resp *CalculateResponse) error {
	if s.db == nil {
		return nil
	}
	reason := resp.DiscountReason
	if req.CouponAmount.GreaterThan(decimal.Zero) {
		if reason == "none" {
			reason = "coupon"
		} else {
			reason = "coupon+" + reason
		}
	}
	_, err := s.db.Exec(
		`INSERT INTO price_history(sku_id,old_price,new_price,reason,created_at) VALUES(?,?,?,?,?)`,
		req.SkuID,
		req.BasePrice.String(),
		resp.FinalPrice.String(),
		reason,
		time.Now(),
	)
	return err
}