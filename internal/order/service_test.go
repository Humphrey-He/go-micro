package order

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alicebob/miniredis/v2"
	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

type fakeInventory struct {
	reservedID string
	err        error
}

func (f *fakeInventory) Reserve(ctx context.Context, orderID string, items []Item) (string, error) {
	return f.reservedID, f.err
}

type fakePublisher struct {
	last []byte
	err  error
}

func (f *fakePublisher) Publish(ctx context.Context, body []byte) error {
	f.last = body
	return f.err
}

func newTestService(t *testing.T) (*Service, sqlmock.Sqlmock, *miniredis.Miniredis) {
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

	inv := &fakeInventory{reservedID: "RESV-1"}
	pub := &fakePublisher{}
	return NewService(sqlxDB, rdb, inv, pub), mock, mr
}

func TestCreateOrder_Success(t *testing.T) {
	svc, mock, mr := newTestService(t)
	defer mr.Close()

	req := CreateOrderRequest{
		RequestID: "REQ-1",
		UserID:    "U-1",
		Items:     []Item{{SkuID: "SKU-1001", Quantity: 2, Price: 100}},
	}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO orders").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), req.UserID, statusCreated, int64(200), req.RequestID, "", int64(0)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO order_items").
		WithArgs(sqlmock.AnyArg(), "SKU-1001", 2, int64(100)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE orders SET status=").
		WithArgs(statusReserved, "RESV-1", sqlmock.AnyArg(), statusCreated, int64(0)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO order_outbox").
		WithArgs("order_reserved", sqlmock.AnyArg(), outboxPending, 0, "").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	resp, err := svc.Create(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != statusReserved {
		t.Fatalf("status mismatch: %s", resp.Status)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestCreateOrder_IdempotentHit(t *testing.T) {
	svc, mock, mr := newTestService(t)
	defer mr.Close()

	req := CreateOrderRequest{
		RequestID: "REQ-1",
		UserID:    "U-1",
		Items:     []Item{{SkuID: "SKU-1001", Quantity: 1, Price: 50}},
	}

	dupErr := &mysql.MySQLError{Number: 1062}
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO orders").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), req.UserID, statusCreated, int64(50), req.RequestID, "", int64(0)).
		WillReturnError(dupErr)
	mock.ExpectRollback()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "order_id", "biz_no", "user_id", "status", "total_amount", "idempotent_key", "reserved_id", "version", "created_at", "updated_at"}).
		AddRow(1, "O-1", "B-1", req.UserID, statusReserved, int64(50), req.RequestID, "", int64(1), now, now)
	mock.ExpectQuery(`SELECT \* FROM orders WHERE idempotent_key`).
		WithArgs(req.RequestID).
		WillReturnRows(rows)

	items := sqlmock.NewRows([]string{"sku_id", "quantity", "price"}).AddRow("SKU-1001", 1, int64(50))
	mock.ExpectQuery(`SELECT sku_id,quantity,price FROM order_items WHERE order_id`).
		WithArgs("O-1").
		WillReturnRows(items)

	_, err := svc.Create(req)
	if err != ErrIdempotentHit {
		t.Fatalf("expected idempotent hit, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGetOrder_NotFound(t *testing.T) {
	svc, mock, mr := newTestService(t)
	defer mr.Close()

	mock.ExpectQuery(`SELECT \* FROM orders WHERE order_id`).
		WithArgs("O-404").
		WillReturnError(sql.ErrNoRows)

	_, err := svc.Get("O-404")
	if err != ErrNotFound {
		t.Fatalf("expected not found, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGetByBizNo_Success(t *testing.T) {
	svc, mock, mr := newTestService(t)
	defer mr.Close()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "order_id", "biz_no", "user_id", "status", "total_amount", "idempotent_key", "reserved_id", "version", "created_at", "updated_at"}).
		AddRow(1, "O-1", "BIZ-1", "U-1", statusReserved, int64(1000), "REQ-1", "RESV-1", int64(1), now, now)
	mock.ExpectQuery(`SELECT \* FROM orders WHERE biz_no`).
		WithArgs("BIZ-1").
		WillReturnRows(rows)

	items := sqlmock.NewRows([]string{"sku_id", "quantity", "price"}).AddRow("SKU-1", 2, int64(500))
	mock.ExpectQuery(`SELECT sku_id,quantity,price FROM order_items WHERE order_id`).
		WithArgs("O-1").
		WillReturnRows(items)

	order, err := svc.GetByBizNo("BIZ-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if order.BizNo != "BIZ-1" {
		t.Fatalf("expected biz_no BIZ-1, got %s", order.BizNo)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGetByBizNo_NotFound(t *testing.T) {
	svc, mock, mr := newTestService(t)
	defer mr.Close()

	mock.ExpectQuery(`SELECT \* FROM orders WHERE biz_no`).
		WithArgs("BIZ-404").
		WillReturnError(sql.ErrNoRows)

	_, err := svc.GetByBizNo("BIZ-404")
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestCancel_Success(t *testing.T) {
	svc, mock, mr := newTestService(t)
	defer mr.Close()

	mock.ExpectQuery(`SELECT status,version FROM orders WHERE order_id = \?`).
		WithArgs("O-1").
		WillReturnRows(sqlmock.NewRows([]string{"status", "version"}).AddRow(statusReserved, int64(1)))

	mock.ExpectExec(`UPDATE orders SET status=\?, reserved_id=\?, version=version\+1, updated_at=NOW\(\) WHERE order_id=\? AND status=\? AND version=\?`).
		WithArgs(statusCanceled, "", "O-1", statusReserved, int64(1)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := svc.Cancel("O-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestCancel_AlreadyCanceled(t *testing.T) {
	svc, mock, mr := newTestService(t)
	defer mr.Close()

	mock.ExpectQuery(`SELECT status,version FROM orders WHERE order_id = \?`).
		WithArgs("O-1").
		WillReturnRows(sqlmock.NewRows([]string{"status", "version"}).AddRow(statusCanceled, int64(1)))

	err := svc.Cancel("O-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateStatus_Success(t *testing.T) {
	svc, mock, mr := newTestService(t)
	defer mr.Close()

	mock.ExpectQuery(`SELECT status,version FROM orders WHERE order_id = \?`).
		WithArgs("O-1").
		WillReturnRows(sqlmock.NewRows([]string{"status", "version"}).AddRow(statusProcessing, int64(1)))

	mock.ExpectExec(`UPDATE orders SET status=\?, reserved_id=\?, version=version\+1, updated_at=NOW\(\) WHERE order_id=\? AND status=\? AND version=\?`).
		WithArgs(statusSuccess, "", "O-1", statusProcessing, int64(1)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := svc.UpdateStatus("O-1", statusProcessing, statusSuccess)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUpdateStatus_InvalidTransition(t *testing.T) {
	svc, mock, mr := newTestService(t)
	defer mr.Close()

	mock.ExpectQuery(`SELECT status,version FROM orders WHERE order_id = \?`).
		WithArgs("O-1").
		WillReturnRows(sqlmock.NewRows([]string{"status", "version"}).AddRow(statusCanceled, int64(1)))

	err := svc.UpdateStatus("O-1", statusReserved, statusSuccess)
	if err == nil {
		t.Fatalf("expected error for invalid transition")
	}
}

func TestCanTransition(t *testing.T) {
	cases := []struct {
		from, to string
		valid    bool
	}{
		{statusCreated, statusReserved, true},
		{statusCreated, statusCanceled, true},
		{statusReserved, statusProcessing, true},
		{statusReserved, statusFailed, true},
		{statusReserved, statusCanceled, true},
		{statusReserved, statusSuccess, false},
		{statusProcessing, statusSuccess, true},
		{statusCanceled, statusSuccess, false},
		{statusSuccess, statusCanceled, false},
	}

	for _, c := range cases {
		t.Run(c.from+"_"+c.to, func(t *testing.T) {
			got := canTransition(c.from, c.to)
			if got != c.valid {
				t.Fatalf("canTransition(%s, %s): got %v, want %v", c.from, c.to, got, c.valid)
			}
		})
	}
}

func TestGetOrder_Success(t *testing.T) {
	svc, mock, mr := newTestService(t)
	defer mr.Close()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "order_id", "biz_no", "user_id", "status", "total_amount", "idempotent_key", "reserved_id", "version", "created_at", "updated_at"}).
		AddRow(1, "O-1", "BIZ-1", "U-1", statusReserved, int64(1000), "REQ-1", "RESV-1", int64(1), now, now)
	mock.ExpectQuery(`SELECT \* FROM orders WHERE order_id`).
		WithArgs("O-1").
		WillReturnRows(rows)

	items := sqlmock.NewRows([]string{"sku_id", "quantity", "price"}).AddRow("SKU-1", 2, int64(500))
	mock.ExpectQuery(`SELECT sku_id,quantity,price FROM order_items WHERE order_id`).
		WithArgs("O-1").
		WillReturnRows(items)

	order, err := svc.Get("O-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if order.OrderID != "O-1" {
		t.Fatalf("expected order_id O-1, got %s", order.OrderID)
	}
	if len(order.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(order.Items))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestIsDuplicate(t *testing.T) {
	err := &mysql.MySQLError{Number: 1062}
	if !isDuplicate(err) {
		t.Fatal("expected true for MySQL error 1062")
	}

	err2 := &mysql.MySQLError{Number: 1045}
	if isDuplicate(err2) {
		t.Fatal("expected false for MySQL error 1045")
	}
}

func TestReserveInventory(t *testing.T) {
	svc, _, mr := newTestService(t)
	defer mr.Close()

	// reserveInventory calls invClient.Reserve which is mocked via fakeInventory
	id, err := svc.reserveInventory("O-1", []Item{{SkuID: "SKU-1", Quantity: 2}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "RESV-1" {
		t.Fatalf("expected RESV-1, got %s", id)
	}
}
