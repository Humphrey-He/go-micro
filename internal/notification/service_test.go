package notification

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func newNotificationService(t *testing.T) (*Service, *Store, sqlmock.Sqlmock) {
	t.Helper()
	t.Setenv("SMTP_USERNAME", "")
	t.Setenv("SMTP_PASSWORD", "")

	// 异步 goroutine 可能在测试结束后仍在运行，钩子必须始终有值
	BroadcastNotification = func(string, *Notification) {}
	t.Cleanup(func() { BroadcastNotification = nil })

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })

	store := NewStore(db)
	return NewService(store, NewEmailService()), store, mock
}

func notifRows(ns ...*Notification) *sqlmock.Rows {
	r := sqlmock.NewRows([]string{"id", "user_id", "type", "title", "content", "is_read", "created_at"})
	for _, n := range ns {
		r.AddRow(n.ID, n.UserID, n.Type, n.Title, n.Content, n.IsRead, n.CreatedAt)
	}
	return r
}

func TestStore_Create(t *testing.T) {
	_, store, mock := newNotificationService(t)

	mock.ExpectExec(`INSERT INTO notifications`).
		WithArgs("u-1", TypeLowStock, "低库存", "SKU-1 不足", false, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(9, 1))

	n := &Notification{UserID: "u-1", Type: TypeLowStock, Title: "低库存", Content: "SKU-1 不足"}
	if err := store.Create(context.Background(), n); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n.ID != 9 {
		t.Fatalf("id = %d, want 9", n.ID)
	}
}

func TestStore_ListByUser(t *testing.T) {
	_, store, mock := newNotificationService(t)

	now := time.Now()
	mock.ExpectQuery(`FROM notifications WHERE user_id = \?`).
		WithArgs("u-1", 20, 0).
		WillReturnRows(notifRows(
			&Notification{ID: 2, UserID: "u-1", Type: TypeDailyReport, Title: "b", Content: "x", CreatedAt: now},
			&Notification{ID: 1, UserID: "u-1", Type: TypeRefundPending, Title: "a", Content: "y", CreatedAt: now},
		))

	ns, err := store.ListByUser(context.Background(), "u-1", 20, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ns) != 2 || ns[0].ID != 2 || ns[0].Type != TypeDailyReport {
		t.Fatalf("notifications = %+v", ns)
	}
}

func TestStore_CountAndMark(t *testing.T) {
	_, store, mock := newNotificationService(t)
	ctx := context.Background()

	mock.ExpectQuery(`AND is_read = 0`).WithArgs("u-1").
		WillReturnRows(sqlmock.NewRows([]string{"c"}).AddRow(3))
	if n, err := store.CountUnread(ctx, "u-1"); err != nil || n != 3 {
		t.Fatalf("count = %d err = %v", n, err)
	}

	mock.ExpectExec(`SET is_read = 1 WHERE id = \?`).WithArgs(int64(5)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	if err := store.MarkRead(ctx, 5); err != nil {
		t.Fatalf("mark read: %v", err)
	}

	mock.ExpectExec(`SET is_read = 1 WHERE user_id = \?`).WithArgs("u-1").
		WillReturnResult(sqlmock.NewResult(0, 3))
	if err := store.MarkAllRead(ctx, "u-1"); err != nil {
		t.Fatalf("mark all read: %v", err)
	}
}

func TestStore_Config(t *testing.T) {
	_, store, mock := newNotificationService(t)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"id", "user_id", "type", "email_enabled", "push_enabled", "threshold"}).
		AddRow(1, "u-1", TypeLowStock, true, false, 10)
	mock.ExpectQuery(`FROM notification_configs`).
		WithArgs("u-1", TypeLowStock).WillReturnRows(rows)

	cfg, err := store.GetConfig(ctx, "u-1", TypeLowStock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.EmailEnabled || cfg.PushEnabled || cfg.Threshold != 10 {
		t.Fatalf("cfg = %+v", cfg)
	}

	mock.ExpectQuery(`FROM notification_configs`).
		WithArgs("u-1", "nope").WillReturnError(sql.ErrNoRows)
	if _, err := store.GetConfig(ctx, "u-1", "nope"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("err = %v, want sql.ErrNoRows", err)
	}

	mock.ExpectExec(`INSERT INTO notification_configs`).
		WithArgs("u-1", TypeLowStock, true, false, 10, true, false, 10).
		WillReturnResult(sqlmock.NewResult(1, 1))
	if err := store.UpsertConfig(ctx, &NotificationConfig{UserID: "u-1", Type: TypeLowStock, EmailEnabled: true, Threshold: 10}); err != nil {
		t.Fatalf("upsert: %v", err)
	}
}

func TestService_CreateNotification_Broadcasts(t *testing.T) {
	svc, _, mock := newNotificationService(t)

	// 异步分支：查配置不存在 → 默认推送开启 → 广播
	mock.ExpectExec(`INSERT INTO notifications`).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(`FROM notification_configs`).WillReturnError(sql.ErrNoRows)

	broadcasted := make(chan *Notification, 1)
	BroadcastNotification = func(userID string, n *Notification) {
		if userID == "u-1" {
			broadcasted <- n
		}
	}
	t.Cleanup(func() { BroadcastNotification = nil })

	if err := svc.CreateNotification(context.Background(), "u-1", TypeLowStock, "低库存", "SKU-1 不足"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	select {
	case n := <-broadcasted:
		if n.Title != "低库存" || n.UserID != "u-1" {
			t.Fatalf("broadcast n = %+v", n)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("notification was not broadcast within 2s")
	}
}

func TestService_CreateNotification_RespectsDisabledPush(t *testing.T) {
	svc, _, mock := newNotificationService(t)

	mock.ExpectExec(`INSERT INTO notifications`).WillReturnResult(sqlmock.NewResult(2, 1))
	// 配置存在且两者都关 → 不广播不发邮件
	mock.ExpectQuery(`FROM notification_configs`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "type", "email_enabled", "push_enabled", "threshold"}).
			AddRow(1, "u-2", TypeDailyReport, false, false, 0))

	called := false
	BroadcastNotification = func(string, *Notification) { called = true }
	t.Cleanup(func() { BroadcastNotification = nil })

	if err := svc.CreateNotification(context.Background(), "u-2", TypeDailyReport, "每日报告", "内容"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 给 goroutine 一点时间，确认没有误广播
	time.Sleep(100 * time.Millisecond)
	if called {
		t.Fatal("should not broadcast when push disabled")
	}
}

func TestService_ListAndUnread(t *testing.T) {
	svc, _, mock := newNotificationService(t)
	ctx := context.Background()

	mock.ExpectQuery(`FROM notifications WHERE user_id = \?`).WithArgs("u-1", 20, 0).
		WillReturnRows(notifRows(&Notification{ID: 1, UserID: "u-1"}))
	mock.ExpectQuery(`AND is_read = 0`).WithArgs("u-1").
		WillReturnRows(sqlmock.NewRows([]string{"c"}).AddRow(1))

	ns, unread, err := svc.ListNotifications(ctx, "u-1", 1, 20)
	if err != nil || len(ns) != 1 || unread != 1 {
		t.Fatalf("ns=%d unread=%d err=%v", len(ns), unread, err)
	}
}

func TestBuildMessage(t *testing.T) {
	msg := buildMessage("from@a.c", "to@b.c", "标题", "<b>正文</b>")
	s := string(msg)
	for _, want := range []string{"From: from@a.c", "To: to@b.c", "Subject: 标题", "MIME-Version: 1.0", `<b>正文</b>`} {
		if !strings.Contains(s, want) {
			t.Fatalf("message missing %q:\n%s", want, s)
		}
	}
}

func TestGenerateReportContent(t *testing.T) {
	now := time.Date(2026, 10, 25, 10, 0, 0, 0, time.UTC)
	content := generateReportContent("每日", now)
	if !strings.Contains(content, "运营每日报告") || !strings.Contains(content, "2026年10月25日") {
		t.Fatalf("content = %s", content)
	}
}

func TestEmailService_SkipsWhenNotConfigured(t *testing.T) {
	_, _, _ = newNotificationService(t) // 确保 env 已清空
	e := NewEmailService()
	// 未配置 SMTP 凭据时静默跳过，不应报错
	if err := e.Send("to@b.c", "s", "b"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSendDailyReport(t *testing.T) {
	svc, _, mock := newNotificationService(t)

	mock.ExpectExec(`INSERT INTO notifications`).
		WithArgs("admin", TypeDailyReport, "每日运营报告", sqlmock.AnyArg(), false, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(`FROM notification_configs`).WillReturnRows(
		sqlmock.NewRows([]string{"id", "user_id", "type", "email_enabled", "push_enabled", "threshold"}).
			AddRow(1, "admin", TypeDailyReport, false, false, 0))

	if err := svc.SendDailyReport(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
