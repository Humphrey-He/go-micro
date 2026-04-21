package payment

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type fakeOrderCanceler struct {
	calls int
	err   error
}

func (f *fakeOrderCanceler) Cancel(ctx context.Context, orderID string) error {
	f.calls++
	return f.err
}

type fakeInventoryReleaser struct {
	calls int
	err   error
}

func (f *fakeInventoryReleaser) ReleaseByOrder(ctx context.Context, orderID string) error {
	f.calls++
	return f.err
}

func newPaymentService(t *testing.T) (*Service, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	sqlxDB := sqlx.NewDb(db, "mysql")
	return NewService(sqlxDB, &fakeOrderCanceler{}, &fakeInventoryReleaser{}), mock
}

func TestCreate_Success(t *testing.T) {
	svc, mock := newPaymentService(t)

	mock.ExpectExec(`INSERT INTO payments`).
		WithArgs(sqlmock.AnyArg(), "O-1", int64(10000), statusCreated, "REQ-1").
		WillReturnResult(sqlmock.NewResult(1, 1))

	p, err := svc.Create(CreatePaymentRequest{OrderID: "O-1", Amount: 10000, RequestID: "REQ-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.OrderID != "O-1" {
		t.Fatalf("expected order_id O-1, got %s", p.OrderID)
	}
	if p.Amount != 10000 {
		t.Fatalf("expected amount 10000, got %d", p.Amount)
	}
	if p.Status != statusCreated {
		t.Fatalf("expected status CREATED, got %s", p.Status)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestCreate_IdempotentHit(t *testing.T) {
	svc, mock := newPaymentService(t)

	mock.ExpectExec(`INSERT INTO payments`).
		WillReturnError(&mysql.MySQLError{Number: 1062})

	now := time.Now()
	mock.ExpectQuery(`SELECT \* FROM payments WHERE request_id = \?`).
		WithArgs("REQ-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "payment_id", "order_id", "amount", "status", "request_id", "created_at", "updated_at"}).
			AddRow(1, "P-1", "O-1", int64(10000), statusCreated, "REQ-1", now, now))

	_, err := svc.Create(CreatePaymentRequest{OrderID: "O-1", Amount: 10000, RequestID: "REQ-1"})
	if err != ErrIdempotentHit {
		t.Fatalf("expected ErrIdempotentHit, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestCreate_InvalidRequest(t *testing.T) {
	svc, _ := newPaymentService(t)

	cases := []struct {
		name string
		req  CreatePaymentRequest
	}{
		{"empty order_id", CreatePaymentRequest{OrderID: "", Amount: 100, RequestID: "REQ-1"}},
		{"empty request_id", CreatePaymentRequest{OrderID: "O-1", Amount: 100, RequestID: ""}},
		{"zero amount", CreatePaymentRequest{OrderID: "O-1", Amount: 0, RequestID: "REQ-1"}},
		{"negative amount", CreatePaymentRequest{OrderID: "O-1", Amount: -100, RequestID: "REQ-1"}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := svc.Create(c.req)
			if err == nil {
				t.Fatalf("expected error for %s", c.name)
			}
		})
	}
}

func TestGet_Success(t *testing.T) {
	svc, mock := newPaymentService(t)

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "payment_id", "order_id", "amount", "status", "request_id", "created_at", "updated_at"}).
		AddRow(1, "P-1", "O-1", int64(10000), statusCreated, "REQ-1", now, now)
	mock.ExpectQuery(`SELECT \* FROM payments WHERE payment_id = \?`).
		WithArgs("P-1").
		WillReturnRows(rows)

	p, err := svc.Get("P-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.PaymentID != "P-1" {
		t.Fatalf("expected payment_id P-1, got %s", p.PaymentID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGet_NotFound(t *testing.T) {
	svc, mock := newPaymentService(t)

	mock.ExpectQuery(`SELECT \* FROM payments WHERE payment_id = \?`).
		WithArgs("P-404").
		WillReturnError(sql.ErrNoRows)

	_, err := svc.Get("P-404")
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGet_EmptyPaymentID(t *testing.T) {
	svc, _ := newPaymentService(t)

	_, err := svc.Get("")
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestMarkSuccess_Success(t *testing.T) {
	svc, mock := newPaymentService(t)

	mock.ExpectQuery(`SELECT status FROM payments WHERE payment_id = \?`).
		WithArgs("P-1").
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow(statusCreated))
	mock.ExpectExec(`UPDATE payments SET status=\?, updated_at=NOW\(\) WHERE payment_id=\? AND status=\?`).
		WithArgs(statusSuccess, "P-1", statusCreated).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := svc.MarkSuccess("P-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMarkSuccess_InvalidState(t *testing.T) {
	svc, mock := newPaymentService(t)

	mock.ExpectQuery(`SELECT status FROM payments WHERE payment_id = \?`).
		WithArgs("P-1").
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow(statusFailed))

	err := svc.MarkSuccess("P-1")
	if err != ErrInvalidState {
		t.Fatalf("expected ErrInvalidState, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMarkFailed_Compensate(t *testing.T) {
	svc, mock := newPaymentService(t)

	mock.ExpectQuery(`SELECT status FROM payments WHERE payment_id = \?`).
		WithArgs("P-1").
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow(statusCreated))
	mock.ExpectExec(`UPDATE payments SET status=\?, updated_at=NOW\(\) WHERE payment_id=\? AND status=\?`).
		WithArgs(statusFailed, "P-1", statusCreated).
		WillReturnResult(sqlmock.NewResult(1, 1))

	now := time.Now()
	mock.ExpectQuery(`SELECT \* FROM payments WHERE payment_id = \?`).
		WithArgs("P-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "payment_id", "order_id", "amount", "status", "request_id", "created_at", "updated_at"}).
			AddRow(1, "P-1", "O-1", int64(100), statusFailed, "REQ-1", now, now))

	if err := svc.MarkFailed("P-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMarkTimeout_Compensate(t *testing.T) {
	svc, mock := newPaymentService(t)

	mock.ExpectQuery(`SELECT status FROM payments WHERE payment_id = \?`).
		WithArgs("P-2").
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow(statusCreated))
	mock.ExpectExec(`UPDATE payments SET status=\?, updated_at=NOW\(\) WHERE payment_id=\? AND status=\?`).
		WithArgs(statusTimeout, "P-2", statusCreated).
		WillReturnResult(sqlmock.NewResult(1, 1))

	now := time.Now()
	mock.ExpectQuery(`SELECT \* FROM payments WHERE payment_id = \?`).
		WithArgs("P-2").
		WillReturnRows(sqlmock.NewRows([]string{"id", "payment_id", "order_id", "amount", "status", "request_id", "created_at", "updated_at"}).
			AddRow(1, "P-2", "O-2", int64(200), statusTimeout, "REQ-2", now, now))

	if err := svc.MarkTimeout("P-2"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRefund_Success(t *testing.T) {
	svc, mock := newPaymentService(t)

	mock.ExpectQuery(`SELECT status FROM payments WHERE payment_id = \?`).
		WithArgs("P-1").
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow(statusSuccess))
	mock.ExpectExec(`UPDATE payments SET status=\?, updated_at=NOW\(\) WHERE payment_id=\? AND status=\?`).
		WithArgs(statusRefund, "P-1", statusSuccess).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := svc.Refund("P-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRefund_InvalidState(t *testing.T) {
	svc, mock := newPaymentService(t)

	mock.ExpectQuery(`SELECT status FROM payments WHERE payment_id = \?`).
		WithArgs("P-1").
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow(statusCreated))

	err := svc.Refund("P-1")
	if err != ErrInvalidState {
		t.Fatalf("expected ErrInvalidState, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUpdateStatus_NotFound(t *testing.T) {
	svc, mock := newPaymentService(t)

	mock.ExpectQuery(`SELECT status FROM payments WHERE payment_id = \?`).
		WithArgs("P-404").
		WillReturnError(sql.ErrNoRows)

	err := svc.MarkSuccess("P-404")
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUpdateStatus_AlreadyInTargetState(t *testing.T) {
	svc, mock := newPaymentService(t)

	mock.ExpectQuery(`SELECT status FROM payments WHERE payment_id = \?`).
		WithArgs("P-1").
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow(statusSuccess))

	err := svc.MarkSuccess("P-1")
	if err != nil {
		t.Fatalf("expected no error when already in target state, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUpdateStatus_NoRowsAffected(t *testing.T) {
	svc, mock := newPaymentService(t)

	mock.ExpectQuery(`SELECT status FROM payments WHERE payment_id = \?`).
		WithArgs("P-1").
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow(statusCreated))
	mock.ExpectExec(`UPDATE payments SET status=\?, updated_at=NOW\(\) WHERE payment_id=\? AND status=\?`).
		WithArgs(statusSuccess, "P-1", statusCreated).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := svc.MarkSuccess("P-1")
	if err != ErrInvalidState {
		t.Fatalf("expected ErrInvalidState when no rows affected, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestListTimeoutPayments_Success(t *testing.T) {
	svc, mock := newPaymentService(t)

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "payment_id", "order_id", "amount", "status", "request_id", "created_at", "updated_at"}).
		AddRow(1, "P-1", "O-1", int64(100), statusCreated, "REQ-1", now.Add(-20*time.Minute), now).
		AddRow(2, "P-2", "O-2", int64(200), statusCreated, "REQ-2", now.Add(-30*time.Minute), now)
	mock.ExpectQuery(`SELECT \* FROM payments WHERE status = \? AND created_at <= \? ORDER BY created_at ASC LIMIT \?`).
		WithArgs(statusCreated, sqlmock.AnyArg(), 20).
		WillReturnRows(rows)

	result, err := svc.ListTimeoutPayments(now, 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 payments, got %d", len(result))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestListTimeoutPayments_DefaultLimit(t *testing.T) {
	svc, mock := newPaymentService(t)

	now := time.Now()
	mock.ExpectQuery(`SELECT \* FROM payments WHERE status = \? AND created_at <= \? ORDER BY created_at ASC LIMIT \?`).
		WithArgs(statusCreated, sqlmock.AnyArg(), 20).
		WillReturnRows(sqlmock.NewRows([]string{"id", "payment_id", "order_id", "amount", "status", "request_id", "created_at", "updated_at"}))

	_, err := svc.ListTimeoutPayments(now, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
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

	err3 := errors.New("other error")
	if isDuplicate(err3) {
		t.Fatal("expected false for non-MySQL error")
	}
}

func TestCompensate_NoOrderCanceler(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	sqlxDB := sqlx.NewDb(db, "mysql")
	svc := NewService(sqlxDB, nil, nil)

	mock.ExpectQuery(`SELECT status FROM payments WHERE payment_id = \?`).
		WithArgs("P-1").
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow(statusCreated))
	mock.ExpectExec(`UPDATE payments SET status=\?, updated_at=NOW\(\) WHERE payment_id=\? AND status=\?`).
		WithArgs(statusFailed, "P-1", statusCreated).
		WillReturnResult(sqlmock.NewResult(1, 1))

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "payment_id", "order_id", "amount", "status", "request_id", "created_at", "updated_at"}).
		AddRow(1, "P-1", "O-1", int64(100), statusFailed, "REQ-1", now, now)
	mock.ExpectQuery(`SELECT \* FROM payments WHERE payment_id = \?`).
		WithArgs("P-1").
		WillReturnRows(rows)

	err = svc.MarkFailed("P-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGetByRequestID_Success(t *testing.T) {
	svc, mock := newPaymentService(t)

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "payment_id", "order_id", "amount", "status", "request_id", "created_at", "updated_at"}).
		AddRow(1, "P-1", "O-1", int64(10000), statusCreated, "REQ-1", now, now)
	mock.ExpectQuery(`SELECT \* FROM payments WHERE request_id = \?`).
		WithArgs("REQ-1").
		WillReturnRows(rows)

	p, err := svc.getByRequestID("REQ-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.PaymentID != "P-1" {
		t.Fatalf("expected payment_id P-1, got %s", p.PaymentID)
	}
}
