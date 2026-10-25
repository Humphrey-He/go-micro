package pricewatch

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jmoiron/sqlx"

	"go-micro/pkg/middleware"
)

func newPriceWatchService(t *testing.T) (*Service, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	sqlxDB := sqlx.NewDb(db, "mysql")
	t.Cleanup(func() { sqlxDB.Close() })
	return NewService(sqlxDB), mock
}

func watchRows(rows ...PriceWatch) *sqlmock.Rows {
	r := sqlmock.NewRows([]string{"id", "user_id", "sku_id", "target_price", "current_price", "notify_enabled", "last_notify_at", "notify_count", "status", "created_at"})
	for _, w := range rows {
		r.AddRow(w.ID, w.UserID, w.SkuID, w.TargetPrice, w.CurrentPrice, w.NotifyEnabled, w.LastNotifyAt, w.NotifyCount, w.Status, w.CreatedAt)
	}
	return r
}

func TestSetPriceWatch_UpdateExisting(t *testing.T) {
	svc, mock := newPriceWatchService(t)
	ctx := t.Context()

	mock.ExpectQuery(`SELECT \* FROM user_price_watches`).WithArgs(int64(42), "SKU-1").
		WillReturnRows(watchRows(PriceWatch{ID: 7, UserID: 42, SkuID: "SKU-1"}))
	mock.ExpectExec(`UPDATE user_price_watches`).WithArgs(99.0, 0.0, int64(42), "SKU-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	target := 99.0
	resp, err := svc.SetPriceWatch(ctx, 42, &SetPriceWatchRequest{SkuID: "SKU-1", TargetPrice: &target})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.WatchID != 7 {
		t.Fatalf("watch_id = %d, want 7 (existing)", resp.WatchID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSetPriceWatch_CreateNew(t *testing.T) {
	svc, mock := newPriceWatchService(t)
	ctx := t.Context()

	mock.ExpectQuery(`SELECT \* FROM user_price_watches`).WithArgs(int64(42), "SKU-2").
		WillReturnError(sql.ErrNoRows)
	mock.ExpectExec(`INSERT INTO user_price_watches`).WithArgs(int64(42), "SKU-2", 50.0, 0.0).
		WillReturnResult(sqlmock.NewResult(11, 1))

	target := 50.0
	resp, err := svc.SetPriceWatch(ctx, 42, &SetPriceWatchRequest{SkuID: "SKU-2", TargetPrice: &target})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.WatchID != 11 {
		t.Fatalf("watch_id = %d, want 11 (new)", resp.WatchID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestCancelPriceWatch(t *testing.T) {
	svc, mock := newPriceWatchService(t)

	mock.ExpectExec(`UPDATE user_price_watches SET status = 'stopped'`).WithArgs(int64(42), "SKU-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := svc.CancelPriceWatch(t.Context(), 42, "SKU-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdatePriceWatch_DynamicFields(t *testing.T) {
	svc, mock := newPriceWatchService(t)
	ctx := t.Context()

	// 只更新 target_price
	tp := 88.0
	mock.ExpectExec(`UPDATE user_price_watches SET updated_at = NOW\(\), target_price = \?`).
		WithArgs(tp, int64(42), "SKU-1").WillReturnResult(sqlmock.NewResult(0, 1))
	if err := svc.UpdatePriceWatch(ctx, 42, "SKU-1", &UpdatePriceWatchRequest{TargetPrice: &tp}); err != nil {
		t.Fatalf("target only: %v", err)
	}

	// 只更新 notify_enabled
	ne := false
	mock.ExpectExec(`, notify_enabled = \?`).WithArgs(ne, int64(42), "SKU-1").WillReturnResult(sqlmock.NewResult(0, 1))
	if err := svc.UpdatePriceWatch(ctx, 42, "SKU-1", &UpdatePriceWatchRequest{NotifyEnabled: &ne}); err != nil {
		t.Fatalf("notify only: %v", err)
	}

	// 空请求只刷新 updated_at
	mock.ExpectExec(`UPDATE user_price_watches SET updated_at = NOW\(\) WHERE`).
		WithArgs(int64(42), "SKU-1").WillReturnResult(sqlmock.NewResult(0, 1))
	if err := svc.UpdatePriceWatch(ctx, 42, "SKU-1", &UpdatePriceWatchRequest{}); err != nil {
		t.Fatalf("empty: %v", err)
	}
}

func TestGetPriceWatchList(t *testing.T) {
	svc, mock := newPriceWatchService(t)

	mock.ExpectQuery(`SELECT COUNT\(\*\)`).WithArgs(int64(42)).WillReturnRows(
		sqlmock.NewRows([]string{"total"}).AddRow(1))
	joined := sqlmock.NewRows([]string{"id", "user_id", "sku_id", "target_price", "current_price", "notify_enabled", "last_notify_at", "notify_count", "status", "created_at", "product_name", "image"}).
		AddRow(7, 42, "SKU-1", 99.0, 109.0, true, time.Time{}, 0, "active", time.Now(), "手机", "img.png")
	mock.ExpectQuery(`SELECT w\.\*`).WithArgs(int64(42), 20, 0).WillReturnRows(joined)

	resp, err := svc.GetPriceWatchList(t.Context(), 42, 0, 0, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Total != 1 || len(resp.Items) != 1 {
		t.Fatalf("total=%d items=%d, want 1/1", resp.Total, len(resp.Items))
	}
	if resp.Items[0].ProductName != "手机" || resp.Items[0].TargetPrice == nil || *resp.Items[0].TargetPrice != 99.0 {
		t.Fatalf("item = %+v", resp.Items[0])
	}
	if resp.Page != 1 || resp.PageSize != 20 {
		t.Fatalf("pagination defaults not applied: %d/%d", resp.Page, resp.PageSize)
	}
}

func TestGetPriceHistory_WithSnapshots(t *testing.T) {
	svc, mock := newPriceWatchService(t)

	day := func(i int) time.Time { return time.Date(2026, 10, i, 0, 0, 0, 0, time.UTC) }
	rows := sqlmock.NewRows([]string{"price", "created_at"}).
		AddRow(120.0, day(1)).
		AddRow(80.0, day(2)).
		AddRow(100.0, day(3))
	mock.ExpectQuery(`FROM product_price_snapshots`).WithArgs("SKU-1", 30).WillReturnRows(rows)

	resp, err := svc.GetPriceHistory(t.Context(), "SKU-1", "30d")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.LowestPrice != 80.0 || resp.HighestPrice != 120.0 || resp.CurrentPrice != 100.0 {
		t.Fatalf("stats wrong: %+v", resp)
	}
	if resp.AveragePrice != 100.0 {
		t.Fatalf("avg = %v, want 100", resp.AveragePrice)
	}
	if len(resp.PricePoints) != 3 {
		t.Fatalf("points = %d, want 3", len(resp.PricePoints))
	}
}

func TestGetPriceHistory_NoSnapshotsFallsBackToMock(t *testing.T) {
	svc, mock := newPriceWatchService(t)

	mock.ExpectQuery(`FROM product_price_snapshots`).WithArgs("SKU-9", 365).WillReturnRows(
		sqlmock.NewRows([]string{"price", "created_at"}))

	resp, err := svc.GetPriceHistory(t.Context(), "SKU-9", "1y")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.PricePoints) != 30 {
		t.Fatalf("mock points = %d, want 30", len(resp.PricePoints))
	}
}

func TestGetNotifications(t *testing.T) {
	svc, mock := newPriceWatchService(t)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM price_watch_notifications WHERE user_id = \?`).
		WithArgs(int64(42)).WillReturnRows(sqlmock.NewRows([]string{"c"}).AddRow(2))
	mock.ExpectQuery(`AND is_read = 0`).WithArgs(int64(42)).WillReturnRows(sqlmock.NewRows([]string{"c"}).AddRow(1))
	joined := sqlmock.NewRows([]string{"id", "user_id", "watch_id", "sku_id", "old_price", "new_price", "discount_amount", "discount_rate", "is_read", "read_at", "created_at", "product_name", "image"}).
		AddRow(3, 42, 7, "SKU-1", 120.0, 90.0, 30.0, 0.75, false, nil, time.Now(), "手机", "img.png")
	mock.ExpectQuery(`FROM price_watch_notifications n`).WithArgs(int64(42), 20, 0).WillReturnRows(joined)

	resp, err := svc.GetNotifications(t.Context(), 42, 0, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Total != 2 || resp.UnreadCount != 1 || len(resp.Items) != 1 {
		t.Fatalf("resp total=%d unread=%d items=%d", resp.Total, resp.UnreadCount, len(resp.Items))
	}
	if resp.Items[0].DiscountRate != 0.75 {
		t.Fatalf("item = %+v", resp.Items[0])
	}
}

// ---- handler：验证 JWT string user_id 转换链路 ----

func signTestToken(t *testing.T, userID string) string {
	t.Helper()
	claims := jwt.MapClaims{"user_id": userID, "username": "u", "role": "user",
		"exp": time.Now().Add(time.Hour).Unix()}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tok.SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatal(err)
	}
	return signed
}

func newPriceWatchRouter(svc *Service) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	rg := r.Group("/api/v1")
	rg.Use(middleware.JWTAuth())
	NewHandler(svc).Register(rg)
	return r
}

func TestHandler_AuthAndHappyPath(t *testing.T) {
	// JWTAuth 在构造时固化 secret，必须在创建路由前设置
	t.Setenv("JWT_SECRET", "test-secret")
	svc, mock := newPriceWatchService(t)
	r := newPriceWatchRouter(svc)

	// 未认证 → 401
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/price-watch/list", nil)
	r.ServeHTTP(w, req)
	if !strings.Contains(w.Body.String(), `"code":40101`) {
		t.Fatalf("unauthenticated body = %s", w.Body.String())
	}

	// 带 JWT（user_id 为 string "42"，与生产中间件一致）→ 应正确转换为 int64
	token := signTestToken(t, "42")
	mock.ExpectQuery(`SELECT COUNT\(\*\)`).WithArgs(int64(42)).WillReturnRows(
		sqlmock.NewRows([]string{"total"}).AddRow(0))
	mock.ExpectQuery(`SELECT w\.\*`).WithArgs(int64(42), 20, 0).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "sku_id", "target_price", "current_price", "notify_enabled", "last_notify_at", "notify_count", "status", "created_at", "product_name", "image"}))

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/price-watch/list", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	if !strings.Contains(w.Body.String(), `"code":0`) {
		t.Fatalf("authenticated body = %s", w.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
