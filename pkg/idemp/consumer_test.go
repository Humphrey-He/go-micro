package idemp

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
)

func newRecorder(t *testing.T) (*ConsumeRecorder, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	sqlxDB := sqlx.NewDb(db, "mysql")
	t.Cleanup(func() { sqlxDB.Close() })
	return NewConsumeRecorder(sqlxDB), mock
}

func TestTryRecord_FirstConsumption(t *testing.T) {
	rec, mock := newRecorder(t)

	// 首次消费：INSERT IGNORE 影响一行
	mock.ExpectExec(`INSERT IGNORE INTO consumed_messages`).
		WithArgs("task-order-created", "msg-1").
		WillReturnResult(sqlmock.NewResult(1, 1))

	first, err := rec.TryRecord(context.Background(), "task-order-created", "msg-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !first {
		t.Fatal("first consumption should return true")
	}
}

func TestTryRecord_DuplicateAbsorbed(t *testing.T) {
	rec, mock := newRecorder(t)

	// 重复投递：唯一键撞键，affected=0 → 应跳过业务
	mock.ExpectExec(`INSERT IGNORE INTO consumed_messages`).
		WithArgs("task-order-created", "msg-1").
		WillReturnResult(sqlmock.NewResult(1, 0))

	first, err := rec.TryRecord(context.Background(), "task-order-created", "msg-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if first {
		t.Fatal("redelivered message should return false (already consumed)")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestTryRecord_DifferentGroupsIndependent(t *testing.T) {
	rec, mock := newRecorder(t)

	// 同一消息被不同消费组消费：各自独立判定
	mock.ExpectExec(`INSERT IGNORE INTO consumed_messages`).
		WithArgs("group-a", "msg-9").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`INSERT IGNORE INTO consumed_messages`).
		WithArgs("group-b", "msg-9").WillReturnResult(sqlmock.NewResult(2, 1))

	for _, g := range []string{"group-a", "group-b"} {
		first, err := rec.TryRecord(context.Background(), g, "msg-9")
		if err != nil || !first {
			t.Fatalf("group %s: first=%v err=%v, want true", g, first, err)
		}
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
