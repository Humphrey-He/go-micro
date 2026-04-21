package refund

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

func newRefundService(t *testing.T) (*Service, sqlmock.Sqlmock, *fakeMQ) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	sqlxDB := sqlx.NewDb(db, "mysql")
	return NewService(sqlxDB, &fakeMQ{}, nil, nil), mock, &fakeMQ{}
}

type fakeMQ struct {
	published []byte
	err       error
}

func (f *fakeMQ) Publish(ctx context.Context, body []byte) error {
	f.published = body
	return f.err
}

type fakeOrderCancelerRef struct {
	calls int
	err   error
}

func (f *fakeOrderCancelerRef) Cancel(ctx context.Context, orderID string) error {
	f.calls++
	return f.err
}

type fakeInventoryReleaserRef struct {
	calls int
	err   error
}

func (f *fakeInventoryReleaserRef) ReleaseByOrder(ctx context.Context, orderID string) error {
	f.calls++
	return f.err
}

func TestInitiate_Success(t *testing.T) {
	svc, mock, _ := newRefundService(t)

	req := InitiateRequest{
		RefundID:   "REF-1",
		OrderID:    "O-1",
		RefundType: TypeManual,
		Reason:     "user request",
	}

	mock.ExpectExec(`INSERT INTO refunds`).
		WithArgs(req.RefundID, req.OrderID, req.RefundType, StatusPending, req.Reason).
		WillReturnResult(sqlmock.NewResult(1, 1))

	refund, err := svc.Initiate(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if refund.RefundID != req.RefundID {
		t.Fatalf("expected refund_id %s, got %s", req.RefundID, refund.RefundID)
	}
	if refund.OrderID != req.OrderID {
		t.Fatalf("expected order_id %s, got %s", req.OrderID, refund.OrderID)
	}
	if refund.Status != StatusPending {
		t.Fatalf("expected status %s, got %s", StatusPending, refund.Status)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestInitiate_DefaultType(t *testing.T) {
	svc, mock, _ := newRefundService(t)

	req := InitiateRequest{
		RefundID: "REF-1",
		OrderID:  "O-1",
	}

	mock.ExpectExec(`INSERT INTO refunds`).
		WithArgs(req.RefundID, req.OrderID, TypeManual, StatusPending, "").
		WillReturnResult(sqlmock.NewResult(1, 1))

	refund, err := svc.Initiate(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if refund.RefundType != TypeManual {
		t.Fatalf("expected default type %s, got %s", TypeManual, refund.RefundType)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestInitiate_IdempotentHit(t *testing.T) {
	svc, mock, _ := newRefundService(t)

	req := InitiateRequest{
		RefundID: "REF-1",
		OrderID:  "O-1",
	}

	// First insert fails with duplicate
	mock.ExpectExec(`INSERT INTO refunds`).
		WithArgs(req.RefundID, req.OrderID, TypeManual, StatusPending, "").
		WillReturnError(&mysql.MySQLError{Number: 1062})

	// Then Get succeeds
	now := time.Now()
	rows := sqlmock.NewRows([]string{"refund_id", "order_id", "refund_type", "status", "reason", "retry_count", "last_error", "created_at", "updated_at"}).
		AddRow(req.RefundID, req.OrderID, TypeManual, StatusPending, "", 0, "", now, now)
	mock.ExpectQuery(`SELECT \* FROM refunds WHERE refund_id`).
		WithArgs(req.RefundID).
		WillReturnRows(rows)

	_, err := svc.Initiate(req)
	if err != ErrIdempotentHit {
		t.Fatalf("expected ErrIdempotentHit, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestInitiate_InvalidRequest(t *testing.T) {
	svc, _, _ := newRefundService(t)

	cases := []struct {
		name string
		req  InitiateRequest
	}{
		{"empty refund_id", InitiateRequest{RefundID: "", OrderID: "O-1"}},
		{"empty order_id", InitiateRequest{RefundID: "REF-1", OrderID: ""}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := svc.Initiate(c.req)
			if err == nil {
				t.Fatalf("expected error for %s", c.name)
			}
		})
	}
}

func TestRollback_Success(t *testing.T) {
	svc, mock, _ := newRefundService(t)

	req := RollbackRequest{
		RefundID: "REF-ROLLBACK-1",
		OrderID:  "O-1",
		Reason:   "order canceled",
	}

	mock.ExpectExec(`INSERT INTO refunds`).
		WithArgs(req.RefundID, req.OrderID, TypeCancel, StatusPending, req.Reason).
		WillReturnResult(sqlmock.NewResult(1, 1))

	refund, err := svc.Rollback(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if refund.RefundType != TypeCancel {
		t.Fatalf("expected refund_type %s, got %s", TypeCancel, refund.RefundType)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGet_Success(t *testing.T) {
	svc, mock, _ := newRefundService(t)

	now := time.Now()
	rows := sqlmock.NewRows([]string{"refund_id", "order_id", "refund_type", "status", "reason", "retry_count", "last_error", "created_at", "updated_at"}).
		AddRow("REF-1", "O-1", TypeManual, StatusPending, "user request", 0, "", now, now)
	mock.ExpectQuery(`SELECT \* FROM refunds WHERE refund_id`).
		WithArgs("REF-1").
		WillReturnRows(rows)

	refund, err := svc.Get("REF-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if refund.RefundID != "REF-1" {
		t.Fatalf("expected refund_id REF-1, got %s", refund.RefundID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGet_NotFound(t *testing.T) {
	svc, mock, _ := newRefundService(t)

	mock.ExpectQuery(`SELECT \* FROM refunds WHERE refund_id`).
		WithArgs("REF-404").
		WillReturnError(sql.ErrNoRows)

	_, err := svc.Get("REF-404")
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGet_EmptyRefundID(t *testing.T) {
	svc, _, _ := newRefundService(t)

	_, err := svc.Get("")
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestProcess_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	sqlxDB := sqlx.NewDb(db, "mysql")

	canceler := &fakeOrderCancelerRef{}
	releaser := &fakeInventoryReleaserRef{}
	svc := NewService(sqlxDB, &fakeMQ{}, canceler, releaser)

	// Get refund
	now := time.Now()
	rows := sqlmock.NewRows([]string{"refund_id", "order_id", "refund_type", "status", "reason", "retry_count", "last_error", "created_at", "updated_at"}).
		AddRow("REF-1", "O-1", TypeManual, StatusPending, "", 0, "", now, now)
	mock.ExpectQuery(`SELECT \* FROM refunds WHERE refund_id`).
		WithArgs("REF-1").
		WillReturnRows(rows)

	// mark_success
	mock.ExpectExec(`UPDATE refunds SET status=\?, updated_at=NOW\(\) WHERE refund_id=\? AND status <> \?`).
		WithArgs(StatusSuccess, "REF-1", StatusSuccess).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = svc.Process("REF-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if canceler.calls != 1 {
		t.Fatalf("expected canceler to be called 1 time, got %d", canceler.calls)
	}
	if releaser.calls != 1 {
		t.Fatalf("expected releaser to be called 1 time, got %d", releaser.calls)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestProcess_AlreadySuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	sqlxDB := sqlx.NewDb(db, "mysql")

	svc := NewService(sqlxDB, &fakeMQ{}, nil, nil)

	// Get refund - already SUCCESS
	now := time.Now()
	rows := sqlmock.NewRows([]string{"refund_id", "order_id", "refund_type", "status", "reason", "retry_count", "last_error", "created_at", "updated_at"}).
		AddRow("REF-1", "O-1", TypeManual, StatusSuccess, "", 0, "", now, now)
	mock.ExpectQuery(`SELECT \* FROM refunds WHERE refund_id`).
		WithArgs("REF-1").
		WillReturnRows(rows)

	err = svc.Process("REF-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestProcess_CancelFails(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	sqlxDB := sqlx.NewDb(db, "mysql")

	canceler := &fakeOrderCancelerRef{err: context.DeadlineExceeded}
	svc := NewService(sqlxDB, &fakeMQ{}, canceler, nil)

	// Get refund
	now := time.Now()
	rows := sqlmock.NewRows([]string{"refund_id", "order_id", "refund_type", "status", "reason", "retry_count", "last_error", "created_at", "updated_at"}).
		AddRow("REF-1", "O-1", TypeManual, StatusPending, "", 0, "", now, now)
	mock.ExpectQuery(`SELECT \* FROM refunds WHERE refund_id`).
		WithArgs("REF-1").
		WillReturnRows(rows)

	// markFailed with retry
	mock.ExpectExec(`UPDATE refunds SET status=\?, retry_count=\?, next_retry_time=\?, last_error=\?, updated_at=NOW\(\) WHERE refund_id=\?`).
		WithArgs(StatusFailed, 1, sqlmock.AnyArg(), sqlmock.AnyArg(), "REF-1").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = svc.Process("REF-1")
	if err == nil {
		t.Fatalf("expected error when cancel fails")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestProcess_CancelNotSet(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	sqlxDB := sqlx.NewDb(db, "mysql")

	// No canceler set
	releaser := &fakeInventoryReleaserRef{}
	svc := NewService(sqlxDB, &fakeMQ{}, nil, releaser)

	// Get refund
	now := time.Now()
	rows := sqlmock.NewRows([]string{"refund_id", "order_id", "refund_type", "status", "reason", "retry_count", "last_error", "created_at", "updated_at"}).
		AddRow("REF-1", "O-1", TypeManual, StatusPending, "", 0, "", now, now)
	mock.ExpectQuery(`SELECT \* FROM refunds WHERE refund_id`).
		WithArgs("REF-1").
		WillReturnRows(rows)

	// mark_success
	mock.ExpectExec(`UPDATE refunds SET status=\?, updated_at=NOW\(\) WHERE refund_id=\? AND status <> \?`).
		WithArgs(StatusSuccess, "REF-1", StatusSuccess).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = svc.Process("REF-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if releaser.calls != 1 {
		t.Fatalf("expected releaser to be called 1 time, got %d", releaser.calls)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestListRetryDue_Success(t *testing.T) {
	svc, mock, _ := newRefundService(t)

	now := time.Now()
	rows := sqlmock.NewRows([]string{"refund_id", "order_id", "refund_type", "status", "reason", "retry_count", "last_error", "created_at", "updated_at"}).
		AddRow("REF-1", "O-1", TypeManual, StatusFailed, "", 1, "timeout", now, now).
		AddRow("REF-2", "O-2", TypeCancel, StatusFailed, "canceled", 2, "", now, now)
	mock.ExpectQuery(`SELECT \* FROM refunds WHERE status = \? AND next_retry_time <= NOW\(\) ORDER BY next_retry_time ASC LIMIT \?`).
		WithArgs(StatusFailed, 20).
		WillReturnRows(rows)

	refunds, err := svc.ListRetryDue(20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(refunds) != 2 {
		t.Fatalf("expected 2 refunds, got %d", len(refunds))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestListRetryDue_DefaultLimit(t *testing.T) {
	svc, mock, _ := newRefundService(t)

	mock.ExpectQuery(`SELECT \* FROM refunds WHERE status = \? AND next_retry_time <= NOW\(\) ORDER BY next_retry_time ASC LIMIT \?`).
		WithArgs(StatusFailed, 20).
		WillReturnRows(sqlmock.NewRows([]string{"refund_id", "order_id", "refund_type", "status", "reason", "retry_count", "last_error", "created_at", "updated_at"}))

	_, err := svc.ListRetryDue(0) // Should use default 20
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRetryDelay(t *testing.T) {
	cases := []struct {
		retry   int
		waitMin int
	}{
		{1, 1},
		{2, 5},
		{3, 15},
		{4, 15},
	}

	for _, c := range cases {
		got := retryDelay(c.retry)
		expected := time.Duration(c.waitMin) * time.Minute
		if got != expected {
			t.Fatalf("retry %d: expected %v, got %v", c.retry, expected, got)
		}
	}
}

func TestTruncateErr(t *testing.T) {
	cases := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"short error", 240, "short error"},
		{string(make([]byte, 300)), 240, string(make([]byte, 240))},
		{"", 240, ""},
	}

	for _, c := range cases {
		got := truncateErr(errors.New(c.input))
		if len(got) != len(c.expected) {
			t.Fatalf("truncateErr(%q): expected len %d, got %d", c.input, len(c.expected), len(got))
		}
	}
}

func TestTruncateErr_Nil(t *testing.T) {
	got := truncateErr(nil)
	if got != "" {
		t.Fatalf("expected empty string for nil, got %s", got)
	}
}
