package price

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
)

func newPriceService(t *testing.T) (*Service, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	sqlxDB := sqlx.NewDb(db, "mysql")
	return NewService(sqlxDB), mock
}

func TestCalculate_BasicPrice(t *testing.T) {
	svc := &Service{}

	req := CalculateRequest{
		SkuID:        "SKU-1",
		BasePrice:    10000, // 100.00 元
		CouponAmount: 0,
		UserLevel:    1,
	}

	resp, err := svc.Calculate(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.SkuID != req.SkuID {
		t.Fatalf("expected sku_id %s, got %s", req.SkuID, resp.SkuID)
	}
	if resp.BasePrice != req.BasePrice {
		t.Fatalf("expected base_price %f, got %f", req.BasePrice, resp.BasePrice)
	}
	if resp.CouponAmount != req.CouponAmount {
		t.Fatalf("expected coupon_amount %f, got %f", req.CouponAmount, resp.CouponAmount)
	}
	if resp.FinalPrice != 10000 {
		t.Fatalf("expected final_price 10000, got %f", resp.FinalPrice)
	}
	if resp.LevelDiscount != 1.0 {
		t.Fatalf("expected level_discount 1.0, got %f", resp.LevelDiscount)
	}
	if resp.DiscountReason != "none" {
		t.Fatalf("expected discount_reason 'none', got %s", resp.DiscountReason)
	}
}

func TestCalculate_WithCoupon(t *testing.T) {
	svc := &Service{}

	req := CalculateRequest{
		SkuID:        "SKU-1",
		BasePrice:    10000,
		CouponAmount: 2000, // 20 元 coupon
		UserLevel:    1,
	}

	resp, err := svc.Calculate(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.FinalPrice != 8000 {
		t.Fatalf("expected final_price 8000, got %f", resp.FinalPrice)
	}
}

func TestCalculate_CouponExceedsPrice(t *testing.T) {
	svc := &Service{}

	req := CalculateRequest{
		SkuID:        "SKU-1",
		BasePrice:    1000,
		CouponAmount: 2000, // coupon > price
		UserLevel:    1,
	}

	resp, err := svc.Calculate(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.FinalPrice != 0 {
		t.Fatalf("expected final_price 0 (coupon exceeds), got %f", resp.FinalPrice)
	}
}

func TestCalculate_VIPLevel2(t *testing.T) {
	svc := &Service{}

	req := CalculateRequest{
		SkuID:        "SKU-1",
		BasePrice:    10000,
		CouponAmount: 0,
		UserLevel:    2,
	}

	resp, err := svc.Calculate(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.LevelDiscount != 0.95 {
		t.Fatalf("expected level_discount 0.95 for vip-2, got %f", resp.LevelDiscount)
	}
	if resp.FinalPrice != 9500 {
		t.Fatalf("expected final_price 9500, got %f", resp.FinalPrice)
	}
	if resp.DiscountReason != "vip-level-2" {
		t.Fatalf("expected discount_reason 'vip-level-2', got %s", resp.DiscountReason)
	}
}

func TestCalculate_VIPLevel3(t *testing.T) {
	svc := &Service{}

	req := CalculateRequest{
		SkuID:        "SKU-1",
		BasePrice:    10000,
		CouponAmount: 0,
		UserLevel:    3,
	}

	resp, err := svc.Calculate(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.LevelDiscount != 0.9 {
		t.Fatalf("expected level_discount 0.9 for vip-3, got %f", resp.LevelDiscount)
	}
	if resp.FinalPrice != 9000 {
		t.Fatalf("expected final_price 9000, got %f", resp.FinalPrice)
	}
	if resp.DiscountReason != "vip-level-3" {
		t.Fatalf("expected discount_reason 'vip-level-3', got %s", resp.DiscountReason)
	}
}

func TestCalculate_VIPWithCoupon(t *testing.T) {
	svc := &Service{}

	req := CalculateRequest{
		SkuID:        "SKU-1",
		BasePrice:    10000,
		CouponAmount: 1000,
		UserLevel:    3,
	}

	resp, err := svc.Calculate(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// (10000 - 1000) * 0.9 = 8100
	if resp.FinalPrice != 8100 {
		t.Fatalf("expected final_price 8100, got %f", resp.FinalPrice)
	}
}

func TestCalculate_RecordsHistory(t *testing.T) {
	svc, mock := newPriceService(t)

	mock.ExpectExec(`INSERT INTO price_history`).
		WithArgs("SKU-1", 10000.0, 8100.0, "coupon+vip-level-3", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	resp, err := svc.Calculate(CalculateRequest{
		SkuID:        "SKU-1",
		BasePrice:    10000,
		CouponAmount: 1000,
		UserLevel:    3,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.FinalPrice != 8100 {
		t.Fatalf("expected final_price 8100, got %f", resp.FinalPrice)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestCalculate_RecordHistoryError(t *testing.T) {
	svc, mock := newPriceService(t)

	mock.ExpectExec(`INSERT INTO price_history`).
		WithArgs("SKU-1", 10000.0, 10000.0, "none", sqlmock.AnyArg()).
		WillReturnError(errors.New("insert failed"))

	_, err := svc.Calculate(CalculateRequest{
		SkuID:        "SKU-1",
		BasePrice:    10000,
		CouponAmount: 0,
		UserLevel:    1,
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestCalculate_InvalidRequest(t *testing.T) {
	svc := &Service{}

	cases := []struct {
		name string
		req  CalculateRequest
	}{
		{"empty sku_id", CalculateRequest{SkuID: "", BasePrice: 1000}},
		{"zero base_price", CalculateRequest{SkuID: "SKU-1", BasePrice: 0}},
		{"negative base_price", CalculateRequest{SkuID: "SKU-1", BasePrice: -100}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := svc.Calculate(c.req)
			if err == nil {
				t.Fatalf("expected error for %s", c.name)
			}
		})
	}
}

func TestHistory_Success(t *testing.T) {
	svc, mock := newPriceService(t)

	rows := sqlmock.NewRows([]string{"id", "sku_id", "old_price", "new_price", "reason", "created_at"}).
		AddRow(1, "SKU-1", 10000, 10000, "original", time.Now()).
		AddRow(2, "SKU-1", 10000, 9500, "discount", time.Now())
	mock.ExpectQuery(`SELECT \* FROM price_history WHERE sku_id = \? ORDER BY created_at DESC LIMIT \?`).
		WithArgs("SKU-1", 20).
		WillReturnRows(rows)

	history, err := svc.History("SKU-1", 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(history) != 2 {
		t.Fatalf("expected 2 history records, got %d", len(history))
	}
	if history[0].OldPrice != 10000 {
		t.Fatalf("expected first old_price 10000, got %f", history[0].OldPrice)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestHistory_DefaultLimit(t *testing.T) {
	svc, mock := newPriceService(t)

	mock.ExpectQuery(`SELECT \* FROM price_history WHERE sku_id = \? ORDER BY created_at DESC LIMIT \?`).
		WithArgs("SKU-1", 20).
		WillReturnRows(sqlmock.NewRows([]string{"id", "sku_id", "old_price", "new_price", "reason", "created_at"}))

	_, err := svc.History("SKU-1", 0) // Should use default 20
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestHistory_NotFound(t *testing.T) {
	svc, mock := newPriceService(t)

	mock.ExpectQuery(`SELECT \* FROM price_history WHERE sku_id = \? ORDER BY created_at DESC LIMIT \?`).
		WithArgs("SKU-404", 20).
		WillReturnError(sql.ErrNoRows)

	_, err := svc.History("SKU-404", 20)
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestHistory_EmptySkuID(t *testing.T) {
	svc, _ := newPriceService(t)

	_, err := svc.History("", 20)
	if err != ErrInvalid {
		t.Fatalf("expected ErrInvalid, got %v", err)
	}
}
