package price

import (
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
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
	if req.SkuID == "" || req.BasePrice <= 0 {
		return nil, ErrInvalid
	}
	price := req.BasePrice - req.CouponAmount
	if price < 0 {
		price = 0
	}
	discount := 1.0
	reason := "none"
	if req.UserLevel >= 3 {
		discount = 0.9
		reason = "vip-level-3"
	} else if req.UserLevel >= 2 {
		discount = 0.95
		reason = "vip-level-2"
	}
	final := price * discount
	return &CalculateResponse{
		SkuID:          req.SkuID,
		BasePrice:      req.BasePrice,
		CouponAmount:   req.CouponAmount,
		LevelDiscount:  discount,
		FinalPrice:     final,
		DiscountReason: reason,
	}, nil
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
