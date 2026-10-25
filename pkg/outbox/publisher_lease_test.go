package outbox

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
)

type fakePublisher struct {
	calls     []string // 每次调用的 event topic
	errOn     int      // 第 N 次（1-based）调用返回错误；0 表示全部成功
	callCount int
}

func (f *fakePublisher) Publish(_ context.Context, topic string, _ json.RawMessage) error {
	f.callCount++
	if f.callCount == f.errOn {
		return errors.New("connection refused")
	}
	f.calls = append(f.calls, topic)
	return nil
}

func newPublisher(t *testing.T, owner string, pub Publisher) (*OutboxPublisher, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	sqlxDB := sqlx.NewDb(db, "mysql")
	t.Cleanup(func() { sqlxDB.Close() })
	cfg := DefaultConfig()
	cfg.Owner = owner
	return NewOutboxPublisher(sqlxDB, pub, cfg), mock
}

func msgRows(rows ...Message) *sqlmock.Rows {
	r := sqlmock.NewRows([]string{"id", "event_type", "payload", "retry_count", "last_error", "created_at", "owner", "lease_until"})
	for _, m := range rows {
		r.AddRow(m.ID, m.EventType, m.Payload, m.RetryCount, m.LastError, m.CreatedAt, m.Owner, m.LeaseUntil)
	}
	return r
}

func TestProcessPending_ClaimAndSendWithOwnerCheck(t *testing.T) {
	pub := &fakePublisher{}
	p, mock := newPublisher(t, "instance-A", pub)

	// 指标查询 + 候选查询
	mock.ExpectQuery(`SELECT COUNT\(1\) FROM order_outbox`).WillReturnRows(
		sqlmock.NewRows([]string{"c"}).AddRow(1))
	mock.ExpectQuery(`ORDER BY id ASC`).WillReturnRows(
		msgRows(Message{ID: 1, EventType: "order_reserved", Payload: json.RawMessage(`{}`)}))
	// CAS 认领成功
	mock.ExpectExec(`SET status = \?`).
		WithArgs(StatusClaimed, "instance-A", 30, int64(1), StatusPending, StatusClaimed).
		WillReturnResult(sqlmock.NewResult(0, 1))
	// 发送成功 → SENT 标记必须带 owner 校验
	mock.ExpectExec(`SET status = \?, sent_at`).
		WithArgs(StatusSent, int64(1), "instance-A", StatusClaimed).
		WillReturnResult(sqlmock.NewResult(0, 1))

	p.processPending(context.Background())

	if len(pub.calls) != 1 {
		t.Fatalf("publish calls = %d, want 1", len(pub.calls))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestProcessPending_LostCASRace(t *testing.T) {
	pub := &fakePublisher{}
	p, mock := newPublisher(t, "instance-B", pub)

	mock.ExpectQuery(`SELECT COUNT\(1\) FROM order_outbox`).WillReturnRows(
		sqlmock.NewRows([]string{"c"}).AddRow(1))
	mock.ExpectQuery(`ORDER BY id ASC`).WillReturnRows(
		msgRows(Message{ID: 1, EventType: "order_reserved", Payload: json.RawMessage(`{}`)}))
	// 另一实例抢先认领（affected=0）→ 本实例不得发送
	mock.ExpectExec(`SET status = \?`).
		WithArgs(StatusClaimed, "instance-B", 30, int64(1), StatusPending, StatusClaimed).
		WillReturnResult(sqlmock.NewResult(0, 0))

	p.processPending(context.Background())

	if len(pub.calls) != 0 {
		t.Fatalf("lost claim race must not publish, calls = %d", len(pub.calls))
	}
}

func TestProcessPending_PublishFailRevertsToPending(t *testing.T) {
	pub := &fakePublisher{errOn: 1}
	p, mock := newPublisher(t, "instance-A", pub)

	mock.ExpectQuery(`SELECT COUNT\(1\) FROM order_outbox`).WillReturnRows(
		sqlmock.NewRows([]string{"c"}).AddRow(1))
	mock.ExpectQuery(`ORDER BY id ASC`).WillReturnRows(
		msgRows(Message{ID: 2, EventType: "order_reserved", Payload: json.RawMessage(`{}`)}))
	mock.ExpectExec(`SET status = \?`).
		WithArgs(StatusClaimed, "instance-A", 30, int64(2), StatusPending, StatusClaimed).
		WillReturnResult(sqlmock.NewResult(0, 1))
	// 失败回退 PENDING：retry+1 且带 owner 校验
	mock.ExpectExec(`SET status = \?, retry_count`).
		WithArgs(StatusPending, "connection refused", int64(2), "instance-A", StatusClaimed).
		WillReturnResult(sqlmock.NewResult(0, 1))

	p.processPending(context.Background())

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

// R3-2 核心演示：发送成功后进程崩溃（未标记 SENT）→ 租约过期 → 重新认领重投 →
// 下游用消费幂等表吸收重复消息——「at-least-once 发送 + 幂等消费 = 业务恰好一次」。
func TestProcessPending_CrashAfterPublish_RedeliveryAbsorbed(t *testing.T) {
	pub := &fakePublisher{}
	p, mock := newPublisher(t, "instance-A", pub)

	// ===== 轮 1：认领成功、发送成功，随后进程崩溃（SENT 标记永远没有执行）=====
	mock.ExpectQuery(`SELECT COUNT\(1\) FROM order_outbox`).WillReturnRows(
		sqlmock.NewRows([]string{"c"}).AddRow(1))
	mock.ExpectQuery(`ORDER BY id ASC`).WillReturnRows(
		msgRows(Message{ID: 9, EventType: "order_reserved", Payload: json.RawMessage(`{"order_id":"O-9"}`)}))
	mock.ExpectExec(`SET status = \?`).
		WithArgs(StatusClaimed, "instance-A", 30, int64(9), StatusPending, StatusClaimed).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`SET status = \?, sent_at`).
		WithArgs(StatusSent, int64(9), "instance-A", StatusClaimed).
		WillReturnError(errors.New("process killed before marking SENT"))

	p.processPending(context.Background())
	firstRoundPublishes := len(pub.calls)
	if firstRoundPublishes != 1 {
		t.Fatalf("round 1 publishes = %d, want 1", firstRoundPublishes)
	}

	// ===== 轮 2（重启后）：行仍为 CLAIMED 且租约过期 → 候选命中 → 重投 =====
	mock.ExpectQuery(`SELECT COUNT\(1\) FROM order_outbox`).WillReturnRows(
		sqlmock.NewRows([]string{"c"}).AddRow(0))
	mock.ExpectQuery(`ORDER BY id ASC`).WillReturnRows(
		msgRows(Message{ID: 9, EventType: "order_reserved", Payload: json.RawMessage(`{"order_id":"O-9"}`),
			Status: StatusClaimed, Owner: "instance-A-dead", LeaseUntil: sqlmockNullTimePast()}))
	mock.ExpectExec(`SET status = \?`).
		WithArgs(StatusClaimed, "instance-A", 30, int64(9), StatusPending, StatusClaimed).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`SET status = \?, sent_at`).
		WithArgs(StatusSent, int64(9), "instance-A", StatusClaimed).
		WillReturnResult(sqlmock.NewResult(0, 1))

	p.processPending(context.Background())

	if len(pub.calls) != 2 {
		t.Fatalf("redelivery publishes = %d, want 2 (crash caused duplicate send)", len(pub.calls))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
	// 下游第二次收到同一条消息：由 consumed_messages(consumer_group, message_id)
	// 唯一键吸收（pkg/idemp.ConsumeRecorder，测试见 pkg/idemp/consumer_test.go）。
}

func sqlmockNullTimePast() sql.NullTime {
	return sql.NullTime{Time: time.Now().Add(-time.Minute), Valid: true}
}
