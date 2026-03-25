package payment

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
)

type fakeOrderCanceler struct {
	calls int
}

func (f *fakeOrderCanceler) Cancel(ctx context.Context, orderID string) error {
	f.calls++
	return nil
}

type fakeInventoryReleaser struct {
	calls int
}

func (f *fakeInventoryReleaser) ReleaseByOrder(ctx context.Context, orderID string) error {
	f.calls++
	return nil
}

func newPaymentService(t *testing.T) (*Service, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	sqlxDB := sqlx.NewDb(db, "mysql")
	return NewService(sqlxDB, &fakeOrderCanceler{}, &fakeInventoryReleaser{}), mock
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
