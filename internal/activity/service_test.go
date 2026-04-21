package activity

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alicebob/miniredis/v2"
	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

func newActivityService(t *testing.T) (*Service, sqlmock.Sqlmock, *miniredis.Miniredis) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	sqlxDB := sqlx.NewDb(db, "mysql")

	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	return NewService(sqlxDB, rdb), mock, mr
}

func TestIssueCoupon_Success(t *testing.T) {
	svc, mock, mr := newActivityService(t)
	defer mr.Close()

	req := CouponRequest{
		CouponID: "COUPON-1",
		UserID:   "U-1",
		Amount:   100,
	}

	mock.ExpectExec(`INSERT INTO activity_coupons`).
		WithArgs(req.CouponID, req.UserID, req.Amount, CouponStatusIssued, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	coupon, err := svc.IssueCoupon(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if coupon.CouponID != req.CouponID {
		t.Fatalf("expected coupon_id %s, got %s", req.CouponID, coupon.CouponID)
	}
	if coupon.UserID != req.UserID {
		t.Fatalf("expected user_id %s, got %s", req.UserID, coupon.UserID)
	}
	if coupon.Amount != req.Amount {
		t.Fatalf("expected amount %f, got %f", req.Amount, coupon.Amount)
	}
	if coupon.Status != CouponStatusIssued {
		t.Fatalf("expected status %s, got %s", CouponStatusIssued, coupon.Status)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestIssueCoupon_InvalidRequest(t *testing.T) {
	svc, _, mr := newActivityService(t)
	defer mr.Close()

	cases := []struct {
		name string
		req  CouponRequest
	}{
		{"empty coupon_id", CouponRequest{CouponID: "", UserID: "U-1", Amount: 100}},
		{"empty user_id", CouponRequest{CouponID: "COUPON-1", UserID: "", Amount: 100}},
		{"zero amount", CouponRequest{CouponID: "COUPON-1", UserID: "U-1", Amount: 0}},
		{"negative amount", CouponRequest{CouponID: "COUPON-1", UserID: "U-1", Amount: -10}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := svc.IssueCoupon(c.req)
			if err == nil {
				t.Fatalf("expected error for %s", c.name)
			}
		})
	}
}

func TestIssueCoupon_IdempotentHit(t *testing.T) {
	svc, mock, mr := newActivityService(t)
	defer mr.Close()

	req := CouponRequest{
		CouponID: "COUPON-1",
		UserID:   "U-1",
		Amount:   100,
	}

	// First insert fails with duplicate
	mock.ExpectExec(`INSERT INTO activity_coupons`).
		WithArgs(req.CouponID, req.UserID, req.Amount, CouponStatusIssued, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(&mysql.MySQLError{Number: 1062})

	// Then Get succeeds
	now := time.Now()
	rows := sqlmock.NewRows([]string{"coupon_id", "user_id", "amount", "status", "issued_at", "created_at", "updated_at"}).
		AddRow(req.CouponID, req.UserID, req.Amount, CouponStatusIssued, now, now, now)
	mock.ExpectQuery(`SELECT \* FROM activity_coupons WHERE coupon_id`).
		WithArgs(req.CouponID).
		WillReturnRows(rows)

	_, err := svc.IssueCoupon(req)
	if err != ErrIdempotentHit {
		t.Fatalf("expected ErrIdempotentHit, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGetCoupon_Success(t *testing.T) {
	svc, mock, mr := newActivityService(t)
	defer mr.Close()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"coupon_id", "user_id", "amount", "status", "issued_at", "created_at", "updated_at"}).
		AddRow("COUPON-1", "U-1", 100, CouponStatusIssued, now, now, now)
	mock.ExpectQuery(`SELECT \* FROM activity_coupons WHERE coupon_id`).
		WithArgs("COUPON-1").
		WillReturnRows(rows)

	coupon, err := svc.GetCoupon("COUPON-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if coupon.CouponID != "COUPON-1" {
		t.Fatalf("expected coupon_id COUPON-1, got %s", coupon.CouponID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGetCoupon_NotFound(t *testing.T) {
	svc, mock, mr := newActivityService(t)
	defer mr.Close()

	mock.ExpectQuery(`SELECT \* FROM activity_coupons WHERE coupon_id`).
		WithArgs("COUPON-404").
		WillReturnError(sql.ErrNoRows)

	_, err := svc.GetCoupon("COUPON-404")
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGetCoupon_EmptyCouponID(t *testing.T) {
	svc, _, mr := newActivityService(t)
	defer mr.Close()

	_, err := svc.GetCoupon("")
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestSeckill_Success(t *testing.T) {
	svc, mock, mr := newActivityService(t)
	defer mr.Close()

	req := SeckillRequest{
		SkuID:    "SKU-SECKILL-1",
		UserID:   "U-1",
		Quantity: 1,
	}

	now := time.Now()
	startTime := now.Add(-time.Hour)
	endTime := now.Add(time.Hour)

	// Begin transaction
	mock.ExpectBegin()

	// SELECT seckill FOR UPDATE
	seckillRows := sqlmock.NewRows([]string{"sku_id", "stock", "reserved", "status", "start_time", "end_time", "created_at", "updated_at"}).
		AddRow(req.SkuID, 100, 0, SeckillStatusActive, startTime, endTime, now, now)
	mock.ExpectQuery(`SELECT \* FROM activity_seckill WHERE sku_id = \? FOR UPDATE`).
		WithArgs(req.SkuID).
		WillReturnRows(seckillRows)

	// Check no existing order
	mock.ExpectQuery(`SELECT \* FROM seckill_orders WHERE sku_id = \? AND user_id = \?`).
		WithArgs(req.SkuID, req.UserID).
		WillReturnError(sql.ErrNoRows)

	// Update reserved
	mock.ExpectExec(`UPDATE activity_seckill SET reserved = reserved \+ \?, updated_at = \? WHERE sku_id = \?`).
		WithArgs(1, sqlmock.AnyArg(), req.SkuID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Insert seckill order
	mock.ExpectExec(`INSERT INTO seckill_orders`).
		WithArgs(req.SkuID, req.UserID, 1, OrderStatusReserved, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	order, err := svc.Seckill(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if order.SkuID != req.SkuID {
		t.Fatalf("expected sku_id %s, got %s", req.SkuID, order.SkuID)
	}
	if order.Status != OrderStatusReserved {
		t.Fatalf("expected status %s, got %s", OrderStatusReserved, order.Status)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSeckill_InvalidRequest(t *testing.T) {
	svc, _, mr := newActivityService(t)
	defer mr.Close()

	cases := []struct {
		name string
		req  SeckillRequest
	}{
		{"empty sku_id", SeckillRequest{SkuID: "", UserID: "U-1"}},
		{"empty user_id", SeckillRequest{SkuID: "SKU-1", UserID: ""}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := svc.Seckill(c.req)
			if err == nil {
				t.Fatalf("expected error for %s", c.name)
			}
		})
	}
}

func TestSeckill_SecondPurchaseIdempotent(t *testing.T) {
	svc, mock, mr := newActivityService(t)
	defer mr.Close()

	req := SeckillRequest{
		SkuID:    "SKU-SECKILL-1",
		UserID:   "U-1",
		Quantity: 1,
	}

	now := time.Now()
	startTime := now.Add(-time.Hour)
	endTime := now.Add(time.Hour)

	mock.ExpectBegin()

	// SELECT seckill FOR UPDATE
	seckillRows := sqlmock.NewRows([]string{"sku_id", "stock", "reserved", "status", "start_time", "end_time", "created_at", "updated_at"}).
		AddRow(req.SkuID, 100, 0, SeckillStatusActive, startTime, endTime, now, now)
	mock.ExpectQuery(`SELECT \* FROM activity_seckill WHERE sku_id = \? FOR UPDATE`).
		WithArgs(req.SkuID).
		WillReturnRows(seckillRows)

	// User already has order - returns existing
	existingRows := sqlmock.NewRows([]string{"sku_id", "user_id", "quantity", "status", "created_at", "updated_at"}).
		AddRow(req.SkuID, req.UserID, 1, OrderStatusReserved, now, now)
	mock.ExpectQuery(`SELECT \* FROM seckill_orders WHERE sku_id = \? AND user_id = \?`).
		WithArgs(req.SkuID, req.UserID).
		WillReturnRows(existingRows)

	mock.ExpectCommit()

	_, err := svc.Seckill(req)
	if err != ErrIdempotentHit {
		t.Fatalf("expected ErrIdempotentHit, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSeckill_OutOfStock(t *testing.T) {
	svc, mock, mr := newActivityService(t)
	defer mr.Close()

	req := SeckillRequest{
		SkuID:    "SKU-SECKILL-1",
		UserID:   "U-1",
		Quantity: 1,
	}

	now := time.Now()
	startTime := now.Add(-time.Hour)
	endTime := now.Add(time.Hour)

	mock.ExpectBegin()

	seckillRows := sqlmock.NewRows([]string{"sku_id", "stock", "reserved", "status", "start_time", "end_time", "created_at", "updated_at"}).
		AddRow(req.SkuID, 100, 100, SeckillStatusActive, startTime, endTime, now, now)
	mock.ExpectQuery(`SELECT \* FROM activity_seckill WHERE sku_id = \? FOR UPDATE`).
		WithArgs(req.SkuID).
		WillReturnRows(seckillRows)

	mock.ExpectRollback()

	_, err := svc.Seckill(req)
	if err != ErrOutOfStock {
		t.Fatalf("expected ErrOutOfStock, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSeckill_SeckillNotFound(t *testing.T) {
	svc, mock, mr := newActivityService(t)
	defer mr.Close()

	req := SeckillRequest{
		SkuID:    "SKU-404",
		UserID:   "U-1",
		Quantity: 1,
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT \* FROM activity_seckill WHERE sku_id = \? FOR UPDATE`).
		WithArgs(req.SkuID).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()

	_, err := svc.Seckill(req)
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSeckill_SeckillInactive(t *testing.T) {
	svc, mock, mr := newActivityService(t)
	defer mr.Close()

	req := SeckillRequest{
		SkuID:    "SKU-SECKILL-1",
		UserID:   "U-1",
		Quantity: 1,
	}

	now := time.Now()
	startTime := now.Add(-2 * time.Hour)
	endTime := now.Add(-time.Hour) // ended

	mock.ExpectBegin()

	seckillRows := sqlmock.NewRows([]string{"sku_id", "stock", "reserved", "status", "start_time", "end_time", "created_at", "updated_at"}).
		AddRow(req.SkuID, 100, 0, SeckillStatusActive, startTime, endTime, now, now)
	mock.ExpectQuery(`SELECT \* FROM activity_seckill WHERE sku_id = \? FOR UPDATE`).
		WithArgs(req.SkuID).
		WillReturnRows(seckillRows)

	mock.ExpectRollback()

	_, err := svc.Seckill(req)
	if err != ErrInvalid {
		t.Fatalf("expected ErrInvalid, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSeckill_NotStarted(t *testing.T) {
	svc, mock, mr := newActivityService(t)
	defer mr.Close()

	req := SeckillRequest{
		SkuID:    "SKU-SECKILL-1",
		UserID:   "U-1",
		Quantity: 1,
	}

	now := time.Now()
	startTime := now.Add(time.Hour)   // future
	endTime := now.Add(2 * time.Hour) // future

	mock.ExpectBegin()

	seckillRows := sqlmock.NewRows([]string{"sku_id", "stock", "reserved", "status", "start_time", "end_time", "created_at", "updated_at"}).
		AddRow(req.SkuID, 100, 0, SeckillStatusActive, startTime, endTime, now, now)
	mock.ExpectQuery(`SELECT \* FROM activity_seckill WHERE sku_id = \? FOR UPDATE`).
		WithArgs(req.SkuID).
		WillReturnRows(seckillRows)

	mock.ExpectRollback()

	_, err := svc.Seckill(req)
	if err != ErrInvalid {
		t.Fatalf("expected ErrInvalid, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGetSeckill_Success(t *testing.T) {
	svc, mock, mr := newActivityService(t)
	defer mr.Close()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"sku_id", "stock", "reserved", "status", "start_time", "end_time", "created_at", "updated_at"}).
		AddRow("SKU-1", 100, 50, SeckillStatusActive, now, now.Add(time.Hour), now, now)
	mock.ExpectQuery(`SELECT \* FROM activity_seckill WHERE sku_id = \?`).
		WithArgs("SKU-1").
		WillReturnRows(rows)

	seckill, err := svc.GetSeckill("SKU-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if seckill.SkuID != "SKU-1" {
		t.Fatalf("expected sku_id SKU-1, got %s", seckill.SkuID)
	}
	if seckill.Stock != 100 {
		t.Fatalf("expected stock 100, got %d", seckill.Stock)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGetSeckill_NotFound(t *testing.T) {
	svc, mock, mr := newActivityService(t)
	defer mr.Close()

	mock.ExpectQuery(`SELECT \* FROM activity_seckill WHERE sku_id = \?`).
		WithArgs("SKU-404").
		WillReturnError(sql.ErrNoRows)

	_, err := svc.GetSeckill("SKU-404")
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGetSeckill_EmptySkuID(t *testing.T) {
	svc, _, mr := newActivityService(t)
	defer mr.Close()

	_, err := svc.GetSeckill("")
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
